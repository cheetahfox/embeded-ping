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
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cheetahfox/embeded-ping/stats"
	// "github.com/sanity-io/litter"
)

func printTotals(seconds int) {
	ticker := time.NewTicker(time.Second * time.Duration(seconds))
	for range ticker.C {
		for host := range stats.RingHosts {
			for index := 0; index < len(stats.RingHosts[host].Ips); index++ {
				stats.RingHosts[host].Ips[index].Mu.Lock()
				fmt.Println("PrintTotal Locked")
				fmt.Println("For host: " + host + " @---> " + stats.RingHosts[host].Ips[index].Ip.String())
				fmt.Printf("Total Packets send : %d \n", stats.RingHosts[host].Ips[index].TotalSent)
				fmt.Printf("Total Packets recv : %d \n", stats.RingHosts[host].Ips[index].TotalReceived)
				fmt.Printf("Total Packets loss : %d \n", stats.RingHosts[host].Ips[index].TotalLoss)
				stats.RingHosts[host].Ips[index].Mu.Unlock()
			}
		}
	}
}

func main() {
	fmt.Println("Startup")

	host := "cheetahfox.com"

	stats.InitHost(host)

	stats.RegisterRingHost(host)

	go printTotals(30)

	// Listen for Sigint or SigTerm and exit if you get them.
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
}
