package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Configuration struct {
	FiberConfig    fiber.Config
	Bucket         string
	InfluxdbServer string
	Org            string
	Token          string
	InfluxMaxError int
}

// Set configuration options from Env values and setup the Fiber options
func Startup() Configuration {
	var conf Configuration

	requiredEnvVars := []string{
		"INFLUX_SERVER", // Influxdb server url including port number
		"INFLUX_TOKEN",  // Influx Token
		"INFLUX_BUCKET", // Influx bucket
		"INFLUX_ORG",    // Influx ord
		"DB_MAX_ERROR",
	}

	// Check if the Required Enviromental varibles are set exit if they aren't.
	for index := range requiredEnvVars {
		if os.Getenv(requiredEnvVars[index]) == "" {
			log.Fatalf("Missing %s Enviroment var \n", requiredEnvVars[index])
		}
	}

	conf.InfluxMaxError = 10
	influxerrors, err := strconv.Atoi(os.Getenv("DB_MAX_ERROR"))
	if err != nil {
		conf.InfluxMaxError = influxerrors
	}

	// Fiber Setup
	conf.FiberConfig = fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Emb-Ping",
		AppName:       "Embeded Ping v0.01",
		ReadTimeout:   (30 * time.Second),
	}

	// Influxdb Settings
	conf.Token = os.Getenv("INFLUX_TOKEN")
	conf.Bucket = os.Getenv("INFLUX_BUCKET")
	conf.Org = os.Getenv("INFLUX_ORG")
	conf.InfluxdbServer = os.Getenv("INFLUX_SERVER")

	return conf
}
