package stats

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/sanity-io/litter"
)

type hostLongTerm struct {
	mu           sync.Mutex
	Drop1000p    float64
	Latency1000p time.Duration
	Received1000 int
	Drop100p     float64
	Latency100p  time.Duration
	Received100  int
}

var Hosts map[string]*hostLongTerm
var hostStats map[string]*list.List

type ping struct {
	rtts          time.Duration
	received      time.Time
	sent          time.Time
	replyReceived bool
}

/*
Package Init:
	Init the package External data structs
*/
func init() {
	hostStats = make(map[string]*list.List)
	Hosts = make(map[string]*hostLongTerm)
	RingHosts = make(map[string]*RingStats)
}

// Debugging func
func GetRawStats(host string) {
	fmt.Println("Stats for host: ", host)
	for e := hostStats[host].Front(); e != nil; e = e.Next() {
		p := e.Value.(ping)

		if p.replyReceived {
			fmt.Println(p.rtts)

		}
	}
}

// Add a host to the longterm Stats
func InitHost(host string) {
	hostStats[host] = list.New()
	Hosts[host] = new(hostLongTerm)
}

// Create new ping primatives and append them to the longterm data.
func ParseStats(s probing.Statistics, host string) error {
	if s.PacketsRecvDuplicates != 0 {
		return errors.New("dup packets: failing for now")
	}

	// create valid received ping packets
	for _, rtts := range s.Rtts {
		var p ping
		p.rtts = rtts
		p.replyReceived = true
		p.received = time.Now()
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

// Trim off the oldest ping's from the hostStats so it no longer than len
func tripStats(len int, host string) error {
	hlen := hostStats[host].Len()

	// Do nothing if we are under the size we need
	if hlen <= len {
		return nil
	}

	// Trim the list
	for hlen > len {
		e := hostStats[host].Back()
		hostStats[host].Remove(e)
		hlen--
	}

	// check that we trimmed to the right size
	if hlen != len {
		return errors.New("stats trimmed to the wrong size")
	}

	return nil
}

func updateStats(host string) error {
	h := &Hosts[host].mu

	h.Lock()

	litter.Dump(hostStats[host])

	var count int
	var avg100Latency, avg1000Latency time.Duration
	// generate stats for the last 100 packets
	for e := hostStats[host].Front(); count <= 100; e = e.Next() {
		count++
		litter.Dump(e)

		p := e.Value.(ping)
		litter.Dump(p.rtts)
		/*
			if e.Value.(ping).replyReceived {
				r100++
			}
		*/
		avg100Latency = p.rtts + avg100Latency
	}

	Hosts[host].Latency100p = avg1000Latency / time.Duration(time.Second*100)

	h.Unlock()
	return nil
}
