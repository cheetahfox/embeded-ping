/*
This is a small program to ping devices once a second and then ship the results off to a influxdb
database.

The main motivation here is to track long-term changes in link opperation.

Joshua Snyder 9/14/2022
*/

package main

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/sanity-io/litter"
)

// Ping result primitive
type ping struct {
	replyReceived bool
	dupReceived   bool
	rtts          time.Duration
}

var hostStats *list.List

func init() {
	hostStats = list.New()
}

func pings() probing.Statistics {
	pinger, err := probing.NewPinger("www.google.com")
	if err != nil {
		panic(err)
	}
	pinger.Count = 2
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
	}

	stats := pinger.Statistics()

	fmt.Println(stats.PacketsSent)

	return *stats
}

func parseStats(s probing.Statistics) error {
	if s.PacketsRecvDuplicates != 0 {
		return errors.New("dup packets: failing for now")
	}

	// create valid received ping packets
	for _, rtts := range s.Rtts {
		fmt.Println("create packet")
		var p ping
		p.rtts = rtts
		p.replyReceived = true

		hostStats.PushFront(p)

	}
	return nil
}

func main() {
	fmt.Println("Startup")

	parseStats(pings())

	litter.Dump(hostStats)

	for e := hostStats.Front(); e != nil; e = e.Next() {
		p := e.Value.(ping)

		fmt.Println(p.rtts)

	}

}
