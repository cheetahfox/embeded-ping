package stats

import (
	"container/list"
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

// Add a host to the longterm Stats
func InitHost(host string) {
	hostStats[host] = list.New()
	Hosts[host] = new(hostLongTerm)
}
