/*
This is a small program to ping devices once a second and then ship the results off to a influxdb
database.

The main motivation here is to track long-term changes in link operation.

Because we compute loss rates several different ways. First we do a pretty standard periodic ping.
I intend this to be samples of between 2-15 seconds. For these samples we store them normally into the tsdb.
But we also maintain stats on the last 100/1000 packets (should be configurable?). We then store these metrics
into the tsdb also. This can allow for long term changes to be monitored over time and not just periodic changes.

For example. lets say you start drop 5 packets every 45 seconds. This loss will only show up in a single sample.
But over the past 100/1000 packets this will be noticeable as a up tick in loss over a much longer period of time.

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

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/cheetahfox/longping/config"
	"github.com/cheetahfox/longping/influxdb"
	"github.com/cheetahfox/longping/router"
	"github.com/cheetahfox/longping/stats"
	"github.com/gofiber/fiber/v2"
	// "github.com/sanity-io/litter"
)

func main() {
	var appLevel = new(slog.LevelVar)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: appLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Startup")
	err := config.Startup()
	if err != nil {
		slog.Error("Error in config.Startup:" + err.Error())
		panic(err)
	}

	if config.Config.LogLevel != "" {
		switch config.Config.LogLevel {
		case "debug":
			appLevel.Set(slog.LevelDebug)
		case "info":
			appLevel.Set(slog.LevelInfo)
		case "warn":
			appLevel.Set(slog.LevelWarn)
		case "error":
			appLevel.Set(slog.LevelError)
		default:
			appLevel.Set(slog.LevelInfo)
		}
		slog.Info("Log level set to " + config.Config.LogLevel)
	}

	hosts := config.GetHosts()

	for _, host := range hosts {
		stats.InitHost(host)
		stats.RegisterRingHost(host)
	}

	// Always start the prometheus metrics and health checks
	longping := fiber.New(config.Config.FiberConfig)

	prometheus := fiberprometheus.New("longping")
	longping.Use(prometheus.Middleware)

	router.SetupRoutes(longping)

	// Start Fiber app in a separate goroutine
	go func() {
		if err := longping.Listen(":3000"); err != nil {
			slog.Error("Error starting Fiber app:" + err.Error())
			panic(err)
		}
	}()

	// Make influxdb optional
	if config.Config.InfluxEnabled {
		influx := config.InfluxEnvStartup()
		influxdb.NewInfluxConnection(influx)
		// need to wait for the influxdb to connect before we can start sending data.
		time.Sleep(time.Duration(time.Second * 1))
		influxdb.WriteRingMetrics(15)
	}

	slog.Debug("Startup successful: waiting for shutdown signal")

	// Listen for Sigint or SigTerm and exit if you get them.
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		slog.Debug("Received shutdown signal:" + sig.String())
		done <- true
	}()

	<-done
	fmt.Println("Shutting down...")
	if config.Config.InfluxEnabled {
		influxdb.DisconnectInflux()
	}
}
