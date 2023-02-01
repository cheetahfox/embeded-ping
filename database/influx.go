package database

import (
	"context"
	"fmt"
	"log"

	"github.com/cheetahfox/Iot-local-midware/health"
	"github.com/cheetahfox/embeded-ping/config"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var DbWrite api.WriteAPI
var dbclient influxdb2.Client

func ConnectInflux(config config.Configuration) {
	dbclient = influxdb2.NewClient(config.InfluxdbServer, config.Token)
	dbhealth, err := dbclient.Health(context.Background())
	if (err != nil) && dbhealth.Status == domain.HealthCheckStatusPass {
		log.Panic(err)
	}
	DbWrite = dbclient.WriteAPI(config.Org, config.Bucket)
	errorsCh := DbWrite.Errors()
	// Catch any write errors
	go func() {
		var errorCount int
		for err := range errorsCh {
			fmt.Printf("Influx write error: %s\n", err.Error())
			errorCount++
			// max number of errors before we set health state to false.
			if errorCount > config.InfluxMaxError {
				fmt.Println("Maximum Influx error count reached!")
				health.InfluxReady = false
			}
		}
	}()
	fmt.Printf("Connected to Influxdb %s\n", config.InfluxdbServer)
	health.InfluxReady = true
}

func DisconnectInflux() {
	health.InfluxReady = false
	dbclient.Close()
}
