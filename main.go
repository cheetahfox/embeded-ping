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
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cheetahfox/embeded-ping/config"
	"github.com/cheetahfox/embeded-ping/influxdb"
	"github.com/cheetahfox/embeded-ping/stats"
	// "github.com/sanity-io/litter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("Startup")
	err := config.Startup()
	if err != nil {
		logger.Error("Error in config.Startup", err)
		panic(err)
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
