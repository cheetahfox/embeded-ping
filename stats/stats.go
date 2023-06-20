package stats

import (
	"container/list"
	"fmt"
	"sync"
	"time"
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
