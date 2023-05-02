package influxdb

import (
	"fmt"
	"net"
	"time"

	"github.com/cheetahfox/embeded-ping/config"
	"github.com/cheetahfox/embeded-ping/stats"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

/*
Creates a new

*/
func NewInfluxConnection(config config.InfluxConfiguration) {
	err := ConnectInflux(config)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteRingMetrics(frequency int) {
	ticker := time.NewTicker(time.Second * time.Duration(frequency))
	for range ticker.C {
		for host := range stats.RingHosts {
			for index := 0; index < len(stats.RingHosts[host].Ips); index++ {
				hn := stats.RingHosts[host].Hostname
				ip := stats.RingHosts[host].Ips[index].Ip

				stats.RingHosts[host].Ips[index].Mu.Lock()
				writeInflux("longping", hn, ip, "Total Packets Sent", float64(stats.RingHosts[host].Ips[index].TotalSent))
				writeInflux("longping", hn, ip, "Total Packets Revc", float64(stats.RingHosts[host].Ips[index].TotalReceived))
				writeInflux("longping", hn, ip, "Total Packets Loss", float64(stats.RingHosts[host].Ips[index].TotalLoss))
				writeInflux("longping", hn, ip, "100 Packet loss", stats.RingHosts[host].Ips[index].Packetloss100)
				writeInflux("longping", hn, ip, "1k Packet loss", stats.RingHosts[host].Ips[index].Packetloss1000)
				writeInflux("longping", hn, ip, "100 Packet Latency", float64(stats.RingHosts[host].Ips[index].Avg100LatencyNs.Nanoseconds()))
				writeInflux("longping", hn, ip, "1k Packet Latency", float64(stats.RingHosts[host].Ips[index].Avg1000LatencyNs.Nanoseconds()))
				stats.RingHosts[host].Ips[index].Mu.Unlock()
				fmt.Println("--- Done Write ---")
			}
		}
	}
}

func writeInflux(measure string, host string, ip net.IP, metric string, value float64) {
	s := fmt.Sprintf("%f", value)
	fmt.Println("Writing point --->  Measure: " + measure + " Host: " + host + " Metric: " + metric + " Value: " + s)

	p := influxdb2.NewPointWithMeasurement(measure)

	p.AddTag("Host", host)
	p.AddTag("Ip", ip.String())
	p.SetTime(time.Now())

	p.AddField(metric, value)

	DbWrite.WritePoint(p)
}
