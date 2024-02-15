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

	"github.com/cheetahfox/embeded-ping/config"
	"github.com/cheetahfox/embeded-ping/influxdb"
	"github.com/cheetahfox/embeded-ping/stats"
	"github.com/labstack/gommon/log"
	// "github.com/sanity-io/litter"
)

func printTotals(seconds int) {
	ticker := time.NewTicker(time.Second * time.Duration(seconds))
	for range ticker.C {
		for host := range stats.RingHosts {
			for index := 0; index < len(stats.RingHosts[host].Ips); index++ {
				stats.RingHosts[host].Ips[index].Mu.Lock()
				fmt.Println("For host: " + host + " @---> " + stats.RingHosts[host].Ips[index].Ip.String())
				fmt.Printf("Total Packets send : %d \n", stats.RingHosts[host].Ips[index].TotalSent)
				fmt.Printf("Total Packets recv : %d \n", stats.RingHosts[host].Ips[index].TotalReceived)
				fmt.Printf("Total Packets loss : %d \n", stats.RingHosts[host].Ips[index].TotalLoss)
				fmt.Printf("100 Packets loss : %f \n", stats.RingHosts[host].Ips[index].Packetloss100)
				fmt.Printf("1k Packets loss : %f \n", stats.RingHosts[host].Ips[index].Packetloss1000)
				fmt.Println("---")
				fmt.Println("100 Packet Latency : " + stats.RingHosts[host].Ips[index].Avg100LatencyNs.String())
				fmt.Println("1k Packet Latency  : " + stats.RingHosts[host].Ips[index].Avg1000LatencyNs.String())
				stats.RingHosts[host].Ips[index].Mu.Unlock()
			}
		}
	}
}

func main() {
	fmt.Println("Startup")
	err := config.Startup()
	if err != nil {
		log.Fatal(err)
	}

	hosts := config.GetHosts()

	for _, host := range hosts {
		stats.InitHost(host)
		stats.RegisterRingHost(host)
	}

	influx := config.InfluxEnvStartup()
	influxdb.NewInfluxConnection(influx)

	// go printTotals(30)
	// What a hack for now... we need to wait for the influxdb to connect before we can start sending data.
	time.Sleep(time.Duration(time.Second * 1))
	fmt.Println("Startup sleeping")
	influxdb.WriteRingMetrics(15)

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
	fmt.Println("Shutting down...")
	influxdb.DisconnectInflux()
}
