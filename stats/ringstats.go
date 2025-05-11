/*
Here we store all of the packet loss and latency data for each host. We store the packets using a
ring buffer so we can keep track of the last 100 and 1000 packets. We also store the total number of packets dropped.

This is mostly an attempt for me to use a complicated memory struct vs a more standard slice.
*/
package stats

import (
	"container/ring"
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/cheetahfox/embeded-ping/config"

	probing "github.com/prometheus-community/pro-bing"
)

// Packetloss is stored as a 1 = 100% and 0 = 0% loss.
type ipRings struct {
	Mu                  sync.Mutex
	Ip                  net.IP
	Stats100            *ring.Ring
	Stats1k             *ring.Ring
	Stats15             *ring.Ring
	Packetloss15        float64
	Packetloss100       float64
	Packetloss1000      float64
	TotalSent           int
	TotalLoss           int
	TotalReceived       int
	TotalDuplicates     int
	Avg1000LatencyNs    time.Duration
	Avg100LatencyNs     time.Duration
	Avg15LatencyNs      time.Duration
	Max1000LatencyNs    time.Duration
	Max100LatencyNs     time.Duration
	Max15LatencyNs      time.Duration
	Min1000LatencyNs    time.Duration
	Min100LatencyNs     time.Duration
	Min15LatencyNs      time.Duration
	Jitter1000LatencyNs time.Duration
	Jitter100LatencyNs  time.Duration
	Jitter15LatencyNs   time.Duration
	shutdown            chan bool
}

type RingStats struct {
	Hostname string
	Ips      []ipRings
}

var RingHosts map[string]*RingStats

/*
Add a new Ring Host for monitoring; we don't lock it since we aren't messing with the ring
We do DNS resolution and for each IP address we find we are going to init a stats ring for
our default packet windows for he last 15, 100 and 1k packets.
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
		newRing.Stats15 = ring.New(15)

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
		pinger.Timeout = time.Second * time.Duration(config.Config.ProbeTimeout)
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
This function is going to be called by the main loop to make sure we don't have any
ping packets that should be removed from the ring due to timing out. Normally we don't
remove packets anywhere else.
*/
func ringMaintance(s probing.Statistics, pIp *ipRings, host string, startTime time.Time) {

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
		err = ringAddStats(ping, pIp.Stats15)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" Host: " + hostname + " ---> 15 ring")
		}
	}

	pIp.Packetloss15 = genPacketloss(pIp.Stats15)
	pIp.Packetloss100 = genPacketloss(pIp.Stats100)
	pIp.Packetloss1000 = genPacketloss(pIp.Stats1k)

	pIp.Avg15LatencyNs = genAvgLatency(pIp.Stats15)
	pIp.Avg100LatencyNs = genAvgLatency(pIp.Stats100)
	pIp.Avg1000LatencyNs = genAvgLatency(pIp.Stats1k)

	pIp.Jitter15LatencyNs = genJitterLatency(pIp.Stats15)
	pIp.Jitter100LatencyNs = genJitterLatency(pIp.Stats100)
	pIp.Jitter1000LatencyNs = genJitterLatency(pIp.Stats1k)

	pIp.Max15LatencyNs = genMaxLatency(pIp.Stats15)
	pIp.Max100LatencyNs = genMaxLatency(pIp.Stats100)
	pIp.Max1000LatencyNs = genMaxLatency(pIp.Stats1k)

	pIp.Min15LatencyNs = genMinLatency(pIp.Stats15)
	pIp.Min100LatencyNs = genMinLatency(pIp.Stats100)
	pIp.Min1000LatencyNs = genMinLatency(pIp.Stats1k)
}

/*
Take a probing.Statistics and return an slice of pings.
If there are no pings in the out we still create blank packets
Also we don't have a super accurate sent time so we are just going add 1000ms to the start
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
		p.sent = startTime
		p.received = startTime.Add(p.rtts)
		p.replyReceived = true
		packets = append(packets, p)
	}

	if s.PacketLoss > 0 {
		droppedPackets := s.PacketsSent - s.PacketsRecv
		for x := 0; x < droppedPackets; x++ {
			var p ping
			/*
				We are taking the number of packets we got and adding the dropped packets at the end of the window
				Also adding the probe timeout to the sent time for each dropped packet.
			*/
			p.sent = startTime
			p.received = startTime.Add(time.Duration(config.Config.ProbeTimeout) * time.Second)
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

	// Find the oldest packet in the ring and check if we have open slots in the ring
	oldest := ringOldestPacket(stats)
	openSlots := ringOpenSlots(stats)

	var inserted bool

	// yes, we are going all the way around the ring + one element.
	for i := 0; i <= ringSize; i++ {
		// Checking that we have a ping packet in the ring
		switch v := stats.Value.(type) {
		case ping:
			// If we do not have open slots and the packet is the oldest packet in the ring then we replace it
			if !openSlots && stats.Value.(ping).sent == oldest.sent {
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

	/*
		If we get all the way around the ring and we haven't inserted the packet then we have a problem.
		We should never get here but if we do we need to know about it. So here is some debug info.
	*/
	if !inserted {
		fmt.Println("Value sent time  : " + stats.Value.(ping).sent.Format("2006-01-02T15:04:05.999999999Z07:00"))
		fmt.Println("Packet sent time : " + packet.sent.Format("2006-01-02T15:04:05.999999999Z07:00"))
		err := errors.New("unable to insert into ring")
		return err
	}

	return nil
}

/*
Find the oldest packet in the ring and return it.
*/
func ringOldestPacket(stats *ring.Ring) ping {
	var oldest ping
	ringSize := stats.Len()
	for i := 0; i < ringSize; i++ {
		switch v := stats.Value.(type) {
		case ping:
			if stats.Value.(ping).sent.Before(oldest.sent) || oldest.sent.IsZero() {
				oldest = stats.Value.(ping)
			}
		case int:
			fmt.Println(v)
		default:
		}
		stats = stats.Next()
	}

	return oldest
}

/*
Find if we have open slots in the ring for pings
*/
func ringOpenSlots(stats *ring.Ring) bool {
	ringSize := stats.Len()
	for i := 0; i < ringSize; i++ {
		switch v := stats.Value.(type) {
		case ping:
		case int:
			fmt.Println(v)
		default:
			return true
		}
		stats = stats.Next()
	}

	return false
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

/*
Recalculate current average latency from the long term statistics
*/
func genAvgLatency(ring *ring.Ring) time.Duration {
	var totalTime time.Duration
	var emptyPackets int

	ringSize := ring.Len()
	for i := 0; i < ringSize; i++ {
		switch v := ring.Value.(type) {
		case ping:
			if ring.Value.(ping).replyReceived {
				totalTime = v.rtts + totalTime
			} else {
				emptyPackets++
			}
		case int:
			fmt.Println(v)
		default:
			emptyPackets++
		}
		ring = ring.Next()
	}

	/*
		time.Duration is really a int64 of NanoSeconds so we can do this math here
		Need to make sure we don't divide by zero in the case we have 100% packet loss
	*/
	var avg int64
	if int64(ringSize-emptyPackets) != 0 {
		avg = int64(totalTime) / int64(ringSize-emptyPackets)
	}
	return time.Duration(avg)
}

/*
Return the max latency from the long term statistics
*/
func genMaxLatency(ring *ring.Ring) time.Duration {
	var maxLatency time.Duration
	ringSize := ring.Len()
	for i := 0; i < ringSize; i++ {
		switch v := ring.Value.(type) {
		case ping:
			if ring.Value.(ping).replyReceived {
				if v.rtts > maxLatency {
					maxLatency = v.rtts
				}
			}
		case int:
			fmt.Println(v)
		default:
		}
		ring = ring.Next()
	}

	return maxLatency
}

/*
Return the min latency from the long term statistics
*/
func genMinLatency(ring *ring.Ring) time.Duration {
	var minLatency time.Duration
	ringSize := ring.Len()
	for i := 0; i < ringSize; i++ {
		switch v := ring.Value.(type) {
		case ping:
			if ring.Value.(ping).replyReceived {
				// Set the first value as the min latency
				if v.rtts < minLatency || minLatency == 0 {
					minLatency = v.rtts
				}
			}
		case int:
			fmt.Println(v)
		default:
		}
		ring = ring.Next()
	}

	return minLatency
}

/*
Return the jitter latency from the long term statistics
*/
func genJitterLatency(ring *ring.Ring) time.Duration {
	var Jitter int64
	var absRtts []time.Duration
	var diffRtts []time.Duration

	// First we get all of the rtts that have a reply
	ringSize := ring.Len()
	for i := 0; i < ringSize; i++ {
		switch v := ring.Value.(type) {
		case ping:
			if ring.Value.(ping).replyReceived {
				absRtts = append(absRtts, v.rtts)
			}
		case int:
			fmt.Println(v)
		default:
		}
		ring = ring.Next()
	}

	// Then we calculate the differences between the rtts
	for i := 0; i < len(absRtts); i++ {
		if i != 0 {
			diffRtts = append(diffRtts, absRtts[i]-absRtts[i-1])
		}
	}

	// Calculate the average of the absolute values of the differences
	// here we take advantage of the fact that time.duration is really a int64 of NanoSeconds
	var totalDiff int64
	for _, v := range diffRtts {
		totalDiff = totalDiff + absInt(int64(v))
	}

	if len(diffRtts) != 0 {
		Jitter = totalDiff / int64(len(diffRtts))
	}

	if config.Config.Debug {
		fmt.Println("Jitter: ", Jitter)
	}

	return time.Duration(Jitter)
}

// Return the absolute value of an int64 number
func absInt(n int64) int64 {
	return int64(math.Abs(float64(n)))
}
