package stats

import (
	"container/ring"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// Packetloss is stored as a 1 = 100% and 0 = 0% loss.
type ipRings struct {
	Mu              sync.Mutex
	Ip              net.IP
	Stats100        *ring.Ring
	Stats1k         *ring.Ring
	Packetloss100   float64
	Packetloss1000  float64
	TotalSent       int
	TotalLoss       int
	TotalReceived   int
	TotalDuplicates int
	shutdown        chan bool
}

type RingStats struct {
	Hostname string
	Ips      []ipRings
}

var RingHosts map[string]*RingStats

/*
Add a new Ring Host for monitoring; we don't lock it since we aren't messing with the ring
We do DNS resolution and for each IP address we find we are going to init a stats ring for
the last 100 and 1k packets.
*/
func RegisterRingHost(host string) error {

	stats := new(RingStats)
	RingHosts[host] = stats

	RingHosts[host].Hostname = host

	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		var newRing ipRings
		newRing.Ip = ip
		newRing.Stats1k = ring.New(1000)
		newRing.Stats100 = ring.New(100)

		// Here we don't care that we are copying a struct with a mutex because this is the initialization of the ring.
		RingHosts[host].Ips = append(RingHosts[host].Ips, newRing)

		fmt.Println("Registered Hostname: " + host + " With Ip Address: " + ip.String())
	}

	fmt.Println("---- Done adding ----")

	ringCollector(host, 1, 1)

	return nil
}

/*
Low Level ping thread, Takes seconds between runs and number of packets to send.
Can be shutdown by writing (technically any value to the shutdown channel) runs
forever until shutdown.

*/
func pingThread(pIp *ipRings, seconds int, packets int, host string) {
	ticker := time.NewTicker(time.Second * time.Duration(seconds))
	for range ticker.C {
		// Check for incoming shutdown and return if we get one.
		select {
		case <-pIp.shutdown:
			fmt.Println("thread shutdown for : " + host + " ---> " + pIp.Ip.String())
			return
		default:
		}
		startTime := time.Now()
		pinger, err := probing.NewPinger(pIp.Ip.String())
		if err != nil {
			fmt.Println(err)
			return
		}
		pinger.Count = packets
		pinger.Timeout = time.Second * time.Duration(1)
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			fmt.Println(err)
			return
		}

		stats := pinger.Statistics()

		ringParseStats(*stats, pIp, host, startTime)
	}
}

/*
Func to kick off the pingThreads for the first time. Can be called directly from a future API
For now only call with 1 packet and 1 second.
*/
func ringCollector(host string, seconds int, packets int) {
	// Loop this way so we aren't copying the RingStats struct and can reference it directly
	for index := 0; index < len(RingHosts[host].Ips); index++ {
		go pingThread(&RingHosts[host].Ips[index], seconds, packets, host)
	}
}

// Todo: check the value is there in the first place
func deleteHost(hostname string) error {
	var hostRing RingStats

	hostRing = *RingHosts[hostname]

	for x, _ := range hostRing.Ips {
		hostRing.Ips[x].shutdown <- true
	}

	return nil
}

/*
Function to parse the ping stats from each pingThread.

I am not sure I really need to be locking this technically this the only place where the each ipRings
struct (that name seems bad now). But I will be reading this from outside this package so I think it
won't hurt to lock the data struct when accessing it.
*/
func ringParseStats(s probing.Statistics, pIp *ipRings, hostname string, startTime time.Time) {
	// Generate arrays of ping packets for storage long term
	pingPackets, err := generatePingPackets(s, startTime)
	if err != nil {
		fmt.Println("unable to generate ping packets")
	}

	pIp.Mu.Lock()
	defer pIp.Mu.Unlock()

	// Update Totals Counters
	pIp.TotalSent = pIp.TotalSent + s.PacketsSent
	pIp.TotalReceived = pIp.TotalReceived + s.PacketsRecv
	pIp.TotalLoss = pIp.TotalSent - pIp.TotalReceived
	pIp.TotalDuplicates = pIp.TotalDuplicates + s.PacketsRecvDuplicates

	for _, ping := range pingPackets {
		err := ringAddStats(ping, pIp.Stats100)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" Host: " + hostname + " --->  100 ring")
		}
		err = ringAddStats(ping, pIp.Stats1k)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" Host: " + hostname + " ---> 1000 ring")
		}
	}

	pIp.Packetloss100 = genPacketloss(pIp.Stats100)
	pIp.Packetloss1000 = genPacketloss(pIp.Stats1k)

}

/*
Take a probing.Statistics and return an slice of pings.
If there are no pings in the out we still create blank packets
Also we don't a super accurate sent time so we are just going add 1000ms to the start
for every additional packet in s. Luckly since the plan is to run only 1 packet pings
the startTime should be very close to the true sent. But this might be an issue if we
many packets in a s stats.
*/
func generatePingPackets(s probing.Statistics, startTime time.Time) ([]ping, error) {
	var packets []ping
	/*
		For packets that are received for real; If we are getting stats for a single packet (the default).
		Then this will loop only once but if we have multiple packets we are going to assume they are
		sent once a second. Sadly we don't get the real time in the Statistics.
	*/
	for i := range s.Rtts {
		var p ping
		p.rtts = s.Rtts[i]
		p.sent = startTime.Add(time.Duration(i) * time.Second)
		p.replyReceived = true
		packets = append(packets, p)
	}

	if s.PacketLoss > 0 {
		droppedPackets := s.PacketsSent - s.PacketsRecv
		for x := 0; x < droppedPackets; x++ {
			var p ping
			// We are taking the number of packets we got and adding the dropped packets at the end of the window
			p.sent = startTime.Add(time.Duration(len(s.Rtts)+x) * time.Second)
			p.replyReceived = false
			packets = append(packets, p)
		}
	}

	return packets, nil
}

/*
Add a ping packet into a stats ring; insert first into empty slots.
Or into slots older than the transmit time + ring size (100/1000 seconds).
*/
func ringAddStats(packet ping, stats *ring.Ring) error {
	ringSize := stats.Len()
	// Set an expiration time of for 100 or 1000 seconds Before the packet was sent.
	expireTime := packet.sent.Add(time.Duration(-ringSize) * time.Second)
	var inserted bool

	// yes, we are going all the way around the ring + one element.
	for i := 0; i <= ringSize; i++ {
		// Checking that we have a ping packet in the ring
		switch v := stats.Value.(type) {
		case ping:
			if stats.Value.(ping).sent.Before(expireTime) {
				stats.Value = packet
				inserted = true
			}
		case int:
			fmt.Println(v)
		// blank value so we can insert.
		default:
			stats.Value = packet
			inserted = true
		}

		if inserted {
			break
		}

		stats = stats.Next()
	}

	if !inserted {
		fmt.Println("Expire time      : " + expireTime.Format("2006-01-02T15:04:05.999999999Z07:00"))
		fmt.Println("Value sent time  : " + stats.Value.(ping).sent.Format("2006-01-02T15:04:05.999999999Z07:00"))
		fmt.Println("Packet sent time : " + packet.sent.Format("2006-01-02T15:04:05.999999999Z07:00"))
		err := errors.New("unable to insert into ring")
		return err
	}

	return nil
}

/*
Recalculate current packet loss from the long term statistics
*/
func genPacketloss(ring *ring.Ring) float64 {
	var packetLoss float64
	var droppedPackets int
	ringSize := ring.Len()
	for i := 0; i < ringSize; i++ {
		switch v := ring.Value.(type) {
		case ping:
			if !ring.Value.(ping).replyReceived {
				droppedPackets++
			}
		case int:
			fmt.Println(v)
		default:
		}
		ring = ring.Next()
	}

	//fmt.Printf("Found dropped packets = %d\n", droppedPackets)
	packetLoss = float64(droppedPackets) / float64(ringSize)

	return packetLoss
}
