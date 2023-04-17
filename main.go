/*
This is a small program to ping devices once a second and then ship the results off to a influxdb
database.

The main motivation here is to track long-term changes in link opperation.

Because we compute loss rates several different ways. First we do a pretty standard perodic ping.
I intend this to be samples of between 2-15 seconds. For these samples we store them normally into the tsdb.
But we also maintain stats on the last 100/1000 packets (should be configurable?). We then store these metrics
into the tsdb also. This can allow for long term changes to be monitored over time and not just perodic changes.

For example. lets say you start drop 5 packets every 45 seconds. This loss will only show up in a single sample.
But over the past 100/1000 packets this will be noticable as a up tick in loss over a much longer period of time.

Joshua Snyder 9/14/2022
*/

package main

import (
	"container/ring"
	"fmt"

	"github.com/cheetahfox/embeded-ping/stats"
	probing "github.com/prometheus-community/pro-bing"
	// "github.com/sanity-io/litter"
)

func pings(count int, host string) (probing.Statistics, string) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		panic(err)
	}
	pinger.Count = count
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
	}

	stats := pinger.Statistics()
	return *stats, host
}

func main() {
	fmt.Println("Startup")

	ringStats := ring.New(100)

	fmt.Println(ringStats.Len())

	host := "www.google.com"

	stats.InitHost(host)

	err := stats.ParseStats(pings(10, host))
	if err != nil {
		fmt.Println("Error parsing pings")
		fmt.Println(err)
	}

	// litter.Dump(hostStats)

	stats.GetRawStats(host)
}
