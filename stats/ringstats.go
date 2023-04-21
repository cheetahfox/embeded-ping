package stats

import (
	"container/ring"
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
	Stats100        ring.Ring
	Stats1k         ring.Ring
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
		newRing.Stats1k = *ring.New(1000)
		newRing.Stats100 = *ring.New(100)

		// Here we don't care that we are copying a struct with a mutex because this is the initialization of the data.
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
		pinger.Timeout = time.Second * time.Duration(5)
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			fmt.Println(err)
			return
		}

		stats := pinger.Statistics()

		ringParseStats(*stats, pIp, host, startTime)
	}
}

// Func to kick off the pingThreads for the first time. Can be called directly from a future API
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
	pIp.Mu.Lock()
	defer pIp.Mu.Unlock()

	// Update Totals Counters
	pIp.TotalSent = pIp.TotalSent + s.PacketsSent
	pIp.TotalReceived = pIp.TotalReceived + s.PacketsRecv
	pIp.TotalLoss = pIp.TotalSent - pIp.TotalReceived
	pIp.TotalDuplicates = pIp.TotalDuplicates + s.PacketsRecvDuplicates

}
