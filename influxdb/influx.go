package influxdb

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cheetahfox/Iot-local-midware/health"
	"github.com/cheetahfox/embeded-ping/config"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var DbWrite api.WriteAPI
var dbclient influxdb2.Client

func ConnectInflux(config config.InfluxConfiguration) error {
	// Check if the Influxdb server is valid
	if !dnsCheck(config.InfluxdbServer) {
		return fmt.Errorf("Influxdb server %s is not valid", config.InfluxdbServer)
	}

	dbclient = influxdb2.NewClient(config.InfluxdbServer, config.Token)
	dbhealth, err := dbclient.Health(context.Background())
	if (err != nil) && dbhealth.Status == domain.HealthCheckStatusFail {
		return err
	}
	DbWrite = dbclient.WriteAPI(config.Org, config.Bucket)
	errorsCh := DbWrite.Errors()

	// Catch any write errors
	go func() {
		var errorCount int
		for err := range errorsCh {
			fmt.Printf("Influx write error: %s\n", err.Error())
			errorCount++

			/*
				check if the Influxdb database is healthy after seeing a error
				I don't really like this logic I need to come up with something
				simpler and more reliable.
			*/
			if !DbHealthCheck(time.Duration(errorCount) * time.Second) {
				health.InfluxReady = false
				fmt.Println("unhealthy Influxdb")
				if errorCount > config.InfluxMaxError {
					fmt.Println("Maximum Influx error count reached!")
					DisconnectInflux()
					time.Sleep(time.Duration(5) * time.Second)
					os.Exit(1)
				}
			} else {
				// Reset error count if the database is healthy
				errorCount = 0
				health.InfluxReady = true
			}
		}
	}()
	fmt.Printf("Connected to Influxdb %s\n", config.InfluxdbServer)
	health.InfluxReady = true

	return nil
}

func DbHealthCheck(sleepTime time.Duration) bool {
	time.Sleep(sleepTime)
	dbhealth, err := dbclient.Health(context.Background())
	if (err != nil) || dbhealth.Status == domain.HealthCheckStatusFail {
		return false
	}
	return true
}

// Check for valid DNS
func dnsCheck(host string) bool {
	_, err := net.LookupHost(host)
	if err != nil {
		return false
	}
	return true
}

func DisconnectInflux() {
	health.InfluxReady = false
	dbclient.Close()
}
