package influxdb

import (
	"fmt"
	"net"
	"time"

	"github.com/cheetahfox/longping/config"
	"github.com/cheetahfox/longping/stats"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

/*
Creates a new InfluxDB connection and stores it in the global variable DbWrite
*/
func NewInfluxConnection(config config.InfluxConfiguration) {
	err := ConnectInflux(config)
	if err != nil {
		fmt.Println(err)
	}
}

// This function will write the metrics to InfluxDB every X seconds
func WriteRingMetrics(frequency int) {
	ticker := time.NewTicker(time.Second * time.Duration(frequency))
	var start time.Time
	for range ticker.C {
		for host := range stats.RingHosts {
			for index := 0; index < len(stats.RingHosts[host].Ips); index++ {
				hn := stats.RingHosts[host].Hostname
				ip := stats.RingHosts[host].Ips[index].Ip
				if config.Config.Debug {
					start = time.Now()
				}

				writeInflux("longping", hn, ip, "Total Packets Sent", float64(stats.RingHosts[host].Ips[index].TotalSent))
				writeInflux("longping", hn, ip, "Total Packets Revc", float64(stats.RingHosts[host].Ips[index].TotalReceived))
				writeInflux("longping", hn, ip, "Total Packets Loss", float64(stats.RingHosts[host].Ips[index].TotalLoss))

				writeInflux("longping", hn, ip, "15 Packet loss", stats.RingHosts[host].Ips[index].Packetloss15)
				writeInflux("longping", hn, ip, "100 Packet loss", stats.RingHosts[host].Ips[index].Packetloss100)
				writeInflux("longping", hn, ip, "1k Packet loss", stats.RingHosts[host].Ips[index].Packetloss1000)

				writeInflux("longping", hn, ip, "15 Packet Latency", float64(stats.RingHosts[host].Ips[index].Avg15LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "100 Packet Latency", float64(stats.RingHosts[host].Ips[index].Avg100LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "1k Packet Latency", float64(stats.RingHosts[host].Ips[index].Avg1000LatencyNs.Nanoseconds()))

				writeInflux("longping", hn, ip, "15 Packet Max Latency", float64(stats.RingHosts[host].Ips[index].Max15LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "100 Packet Max Latency", float64(stats.RingHosts[host].Ips[index].Max100LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "1k Packet Max Latency", float64(stats.RingHosts[host].Ips[index].Max1000LatencyNs.Nanoseconds()))

				writeInflux("longping", hn, ip, "15 Packet Min Latency", float64(stats.RingHosts[host].Ips[index].Min15LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "100 Packet Min Latency", float64(stats.RingHosts[host].Ips[index].Min100LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "1k Packet Min Latency", float64(stats.RingHosts[host].Ips[index].Min1000LatencyNs.Nanoseconds()))

				writeInflux("longping", hn, ip, "15 Packet Jitter", float64(stats.RingHosts[host].Ips[index].Jitter15LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "100 Packet Jitter", float64(stats.RingHosts[host].Ips[index].Jitter100LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "1k Packet Jitter", float64(stats.RingHosts[host].Ips[index].Jitter1000LatencyNs.Nanoseconds()))

				if config.Config.Debug {
					elapsed := time.Since(start)
					fmt.Println("Time to write to InfluxDB: " + elapsed.String())
				}
			}
		}
	}
}

func writeInflux(measure string, host string, ip net.IP, metric string, value float64) {
	if config.Config.Debug {
		s := fmt.Sprintf("%f", value)
		fmt.Println("Writing point --->  Measure: " + measure + " Host: " + host + " Metric: " + metric + " Value: " + s)
	}

	p := influxdb2.NewPointWithMeasurement(measure)

	p.AddTag("Host", host)
	p.AddTag("Ip", ip.String())
	p.SetTime(time.Now())

	p.AddField(metric, value)

	DbWrite.WritePoint(p)
}
