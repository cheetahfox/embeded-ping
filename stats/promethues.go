package stats

import (
	"fmt"
	"time"

	"github.com/cheetahfox/longping/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	probing "github.com/prometheus-community/pro-bing"
)

// Register all of the metrics for Prometheus
var (
	// Prometheus metrics

	apiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Total number of API requests",
	}, []string{"method", "endpoint", "status"})

	// Host metrics
	TotalSent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_sent",
			Help: "Total number of packets sent",
		},
		[]string{"hostname", "ip_address"},
	)
	TotalReceived = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_received",
			Help: "Total number of packets received",
		},
		[]string{"hostname", "ip_address"},
	)
	TotalLoss = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_loss",
			Help: "Total number of packets lost",
		},
		[]string{"hostname", "ip_address"},
	)
	TotalDuplicates = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_duplicates",
			Help: "Total number of duplicate packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Avg1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_1000_latency_ns",
			Help: "Average latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Jitter1000Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_1000_ns",
			Help: "Jitter in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Max1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_1000_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Min1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_1000_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Packetloss1000 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_1000",
			Help: "Packet loss for the last 1000 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Avg100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_100_latency_ns",
			Help: "Average latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Jitter100Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_100_ns",
			Help: "Jitter in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Max100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_100_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Min100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_100_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Packetloss100 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_100",
			Help: "Packet loss for the last 100 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Avg15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_15_latency_ns",
			Help: "Average latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Jitter15Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_15_ns",
			Help: "Jitter in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Max15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_15_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Min15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_15_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	Packetloss15 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_15",
			Help: "Packet loss for the last 15 packets",
		},
		[]string{"hostname", "ip_address"},
	)
	PingLatencyNs = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:                            "ping_latency_ns",
		Help:                            "Histogram of ping latency in nanoseconds",
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	}, []string{"hostname", "ip_address"},
	)
)

// updatedHistogramMetrics updates the histogram metrics with the latest ping latency
func updatedHistogramMetrics(hostname string, s probing.Statistics) {
	// Update the histogram with the latest ping latency
	if config.Config.Debug {
		fmt.Printf("Updating histogram for %s with latency %d ns\n", hostname, s.Addr)
	}

	// loop through the RTTs this covers cases where there are multiple RTTs
	for _, rtts := range s.Rtts {
		PingLatencyNs.WithLabelValues(hostname, s.Addr).Observe(float64(rtts.Nanoseconds()))
	}
}

// I hate how this is just a big list of metrics that need to be updated
func prometheusUpdateMetrics(hostname string, pIp *ipRings) {
	// Update the metrics with the values from the ipRings struct
	if config.Config.Debug {
		sd, err := TotalSent.GetMetricWith(prometheus.Labels{"hostname": hostname, "ip_address": pIp.Ip.String()})
		if err != nil {
			fmt.Println("Error getting TotalSent metric:", err)
		}
		fmt.Printf("Total-Sent pings for %s\n", hostname)
		spew.Dump(sd)
	}
	TotalSent.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.TotalSent))
	TotalReceived.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.TotalReceived))
	TotalLoss.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.TotalLoss))
	TotalDuplicates.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.TotalDuplicates))
	// 1000 packet ring
	Avg1000LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Avg1000LatencyNs))
	Jitter1000Ns.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Jitter1000LatencyNs))
	Max1000LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Max1000LatencyNs))
	Min1000LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Min1000LatencyNs))
	Packetloss1000.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Packetloss1000))
	// 100 packet ring
	Avg100LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Avg100LatencyNs))
	Jitter100Ns.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Jitter100LatencyNs))
	Max100LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Max100LatencyNs))
	Min100LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Min100LatencyNs))
	Packetloss100.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Packetloss100))
	// 15 packet ring
	Avg15LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Avg15LatencyNs))
	Jitter15Ns.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Jitter15LatencyNs))
	Max15LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Max15LatencyNs))
	Min15LatencyNs.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Min15LatencyNs))
	Packetloss15.WithLabelValues(hostname, pIp.Ip.String()).Set(float64(pIp.Packetloss15))

	//TotalPingsSent.WithLabelValues(hostname, pIp.Ip.String()).Inc()

	if config.Config.Debug {
		fmt.Printf("Updated Prometheus metrics for %s\n", hostname)
	}
}
