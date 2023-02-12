package stats

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type hostLongTerm struct {
	mu           sync.Mutex
	Drop1000p    float64
	Drop100p     float64
	Latency1000p float64
	Latency100p  float64
}

var Hosts map[string]hostLongTerm

type ping struct {
	rtts          time.Duration
	replyReceived bool
}

var hostStats map[string]*list.List

func init() {
	hostStats = make(map[string]*list.List)
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
	return nil
}

// Trim off the oldest ping's from the hostStats so it no longer than len
func tripStats(len int, host string) error {
	hlen := hostStats[host].Len()

	// Do nothing if we are under the size we need
	if hlen <= len {
		return nil
	}

	for hlen > len {
		e := hostStats[host].Back()
		hostStats[host].Remove(e)
		hlen--
	}

	if hlen != len {
		return errors.New("stats trimmed to the wrong size")
	}

	return nil
}

func updateStats(host string) error {

	return nil
}
