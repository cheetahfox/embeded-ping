package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Configuration struct {
	FiberConfig fiber.Config
}

type InfluxConfiguration struct {
	Bucket         string
	InfluxMaxError int
	InfluxdbServer string
	Org            string
	Token          string
}

// Set configuration options from Env values and setup the Fiber options
func FiberStartup() Configuration {
	var conf Configuration

	// Fiber Setup
	conf.FiberConfig = fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Emb-Ping",
		AppName:       "Embeded Ping v1.0.1",
		ReadTimeout:   (30 * time.Second),
	}

	return conf
}

// Get Hosts from Env and return them as a slice
func GetHosts() []string {
	var hosts []string

	// Get Hosts from Env
	hostsEnv := os.Getenv("HOSTS")
	if hostsEnv == "" {
		log.Fatal("Missing HOSTS Enviroment var")
	}

	// Split the string into a slice
	hosts = append(hosts, hostsEnv)

	return hosts
}

// Set configuration options from Env values and setup the Fiber options
func InfluxEnvStartup() InfluxConfiguration {
	var influxconf InfluxConfiguration

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

	influxconf.InfluxMaxError = 10
	influxerrors, err := strconv.Atoi(os.Getenv("DB_MAX_ERROR"))
	if err != nil {
		influxconf.InfluxMaxError = influxerrors
	}

	// Influxdb Settings
	influxconf.Token = os.Getenv("INFLUX_TOKEN")
	influxconf.Bucket = os.Getenv("INFLUX_BUCKET")
	influxconf.Org = os.Getenv("INFLUX_ORG")
	influxconf.InfluxdbServer = os.Getenv("INFLUX_SERVER")

	return influxconf
}
