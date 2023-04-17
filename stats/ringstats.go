package stats

import (
	"container/ring"
	"errors"
	"fmt"
	"net"
	"sync"

	probing "github.com/prometheus-community/pro-bing"
)

// Packetloss is stored as a 1 = 100% and 0 = 0% loss.
type ipRings struct {
	mu             sync.Mutex
	ip             net.IP
	stats100       ring.Ring
	stats1k        ring.Ring
	packetloss100  float64
	packetloss1000 float64
}

type ringStats struct {
	hostname string
	ips      []ipRings
}

var RingHosts map[string]*ringStats

/*
Add a new Ring Host for monitoring; we don't lock it since we aren't messing with the ring
We do DNS resolution and for each IP address we find we are going to init a stats ring for
the last 100 and 1k packets.
*/
func RegisterRingHost(host string) error {
	stats := new(ringStats)
	RingHosts[host] = stats

	RingHosts[host].hostname = host

	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}

	for x, ip := range ips {
		RingHosts[host].ips[x].ip = ip
		RingHosts[host].ips[x].stats1k = *ring.New(1000)
		RingHosts[host].ips[x].stats100 = *ring.New(100)
	}

	ringCollector(*RingHosts[host])

	return nil
}

// Func to kick off the pingers
func ringCollector(rs ringStats) {

	for ip, _ := range rs.ips {
		go ringPinger(&rs.ips[ip], 1)
	}

}

// Func that pings the hosts
func ringPinger(ring *ipRings, count int) {

	pinger, err := probing.NewPinger(ring.ip.String())
	if err != nil {
		fmt.Println(err)
		return
	}
	pinger.Count = count
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		fmt.Println(err)
		return
	}

	stats := pinger.Statistics()
	return
}

// Create new ping primatives and append them to the longterm data.
func ringParseStats(s probing.Statistics, host string) error {
	if s.PacketsRecvDuplicates != 0 {
		return errors.New("dup packets: failing for now")
	}

	// create valid received ping packets
	for _, rtts := range s.Rtts {
		var p ping
		p.rtts = rtts
		p.replyReceived = true
		hostStats[host].PushFront(p)

	}

	// Insert needed blank sample. Since the pinger doesn't list dropped packets with a rtts of 0.
	if s.PacketsRecv < s.PacketsSent {
		for s.PacketsRecv <= s.PacketsSent {
			var p ping
			p.replyReceived = false
			hostStats[host].PushFront(p)
		}
	}

	// Trim to 1000 packets
	err := tripStats(1000, host)
	if err != nil {
		return err
	}

	err = updateStats(host)
	if err != nil {
		return err
	}

	return nil
}
