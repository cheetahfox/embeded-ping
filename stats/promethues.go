package stats

import (
	"fmt"

	"github.com/cheetahfox/embeded-ping/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Register all of the metrics for Prometheus
var (
	// Prometheus metrics

	apiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_requests_total",
		Help: "Total number of API requests",
	}, []string{"method", "endpoint", "status"})

	TotalPingsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "total_pings_sent",
		Help: "Total number of pings sent",
	}, []string{"method", "endpoint", "status"})

	// Host metrics
	TotalSent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_sent",
			Help: "Total number of packets sent",
		},
		[]string{"hostname", "ip address"},
	)
	TotalReceived = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_received",
			Help: "Total number of packets received",
		},
		[]string{"hostname", "ip address"},
	)
	TotalLoss = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_loss",
			Help: "Total number of packets lost",
		},
		[]string{"hostname", "ip address"},
	)
	TotalDuplicates = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "total_duplicates",
			Help: "Total number of duplicate packets",
		},
		[]string{"hostname", "ip address"},
	)
	Avg1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_1000_latency_ns",
			Help: "Average latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Jitter1000Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_1000_ns",
			Help: "Jitter in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Max1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_1000_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Min1000LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_1000_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 1000 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Packetloss1000 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_1000",
			Help: "Packet loss for the last 1000 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Avg100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_100_latency_ns",
			Help: "Average latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Jitter100Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_100_ns",
			Help: "Jitter in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Max100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_100_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Min100LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_100_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 100 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Packetloss100 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_100",
			Help: "Packet loss for the last 100 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Avg15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "avg_15_latency_ns",
			Help: "Average latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Jitter15Ns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jitter_15_ns",
			Help: "Jitter in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Max15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_15_latency_ns",
			Help: "Maximum latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Min15LatencyNs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_15_latency_ns",
			Help: "Minimum latency in nanoseconds for the last 15 packets",
		},
		[]string{"hostname", "ip address"},
	)
	Packetloss15 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "packetloss_15",
			Help: "Packet loss for the last 15 packets",
		},
		[]string{"hostname", "ip address"},
	)
)

func init() {
	// Register the metrics with Prometheus
	/*
		if err := prometheus.Register(TotalSent); err != nil {
			fmt.Println("Error registering TotalSent metric:", err)
		}
	*/

	prometheus.MustRegister(TotalSent)
	prometheus.MustRegister(TotalPingsSent)

	if err := prometheus.Register(TotalReceived); err != nil {
		fmt.Println("Error registering TotalReceived metric:", err)
	}
	if err := prometheus.Register(TotalLoss); err != nil {
		fmt.Println("Error registering TotalLoss metric:", err)
	}
	if err := prometheus.Register(TotalDuplicates); err != nil {
		fmt.Println("Error registering TotalDuplicates metric:", err)
	}
	if err := prometheus.Register(Avg1000LatencyNs); err != nil {
		fmt.Println("Error registering Avg1000LatencyNs metric:", err)
	}
	if err := prometheus.Register(Jitter1000Ns); err != nil {
		fmt.Println("Error registering Jitter1000Ns metric:", err)
	}
	if err := prometheus.Register(Max1000LatencyNs); err != nil {
		fmt.Println("Error registering Max1000LatencyNs metric:", err)
	}
	if err := prometheus.Register(Min1000LatencyNs); err != nil {
		fmt.Println("Error registering Min1000LatencyNs metric:", err)
	}
	if err := prometheus.Register(Packetloss1000); err != nil {
		fmt.Println("Error registering Packetloss1000 metric:", err)
	}
	if err := prometheus.Register(Avg100LatencyNs); err != nil {
		fmt.Println("Error registering Avg100LatencyNs metric:", err)
	}
	if err := prometheus.Register(Jitter100Ns); err != nil {
		fmt.Println("Error registering Jitter100Ns metric:", err)
	}
	if err := prometheus.Register(Max100LatencyNs); err != nil {
		fmt.Println("Error registering Max100LatencyNs metric:", err)
	}
	if err := prometheus.Register(Min100LatencyNs); err != nil {
		fmt.Println("Error registering Min100LatencyNs metric:", err)
	}
	if err := prometheus.Register(Packetloss100); err != nil {
		fmt.Println("Error registering Packetloss100 metric:", err)
	}
	if err := prometheus.Register(Avg15LatencyNs); err != nil {
		fmt.Println("Error registering Avg15LatencyNs metric:", err)
	}
	if err := prometheus.Register(Jitter15Ns); err != nil {
		fmt.Println("Error registering Jitter15Ns metric:", err)
	}
	if err := prometheus.Register(Max15LatencyNs); err != nil {
		fmt.Println("Error registering Max15LatencyNs metric:", err)
	}
	if err := prometheus.Register(Min15LatencyNs); err != nil {
		fmt.Println("Error registering Min15LatencyNs metric:", err)
	}
	if err := prometheus.Register(Packetloss15); err != nil {
		fmt.Println("Error registering Packetloss15 metric:", err)
	}
}

// I hate how this is just a big list of metrics that need to be updated
func prometheusUpdateMetrics(hostname string, pIp *ipRings) {
	// Update the metrics with the values from the ipRings struct
	if config.Config.Debug {
		sd, err := TotalSent.GetMetricWith(prometheus.Labels{"hostname": hostname})
		if err != nil {
			fmt.Println("Error getting TotalSent metric:", err)
		}
		fmt.Printf("Total-Sent pings for %s\n", hostname)
		spew.Dump(sd)
	}
	TotalSent.WithLabelValues(hostname).Set(float64(pIp.TotalSent))
	TotalReceived.WithLabelValues(hostname).Set(float64(pIp.TotalReceived))
	TotalLoss.WithLabelValues(hostname).Set(float64(pIp.TotalLoss))
	TotalDuplicates.WithLabelValues(hostname).Set(float64(pIp.TotalDuplicates))
	// 1000 packet ring
	Avg1000LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Avg1000LatencyNs))
	Jitter1000Ns.WithLabelValues(hostname).Set(float64(pIp.Jitter1000LatencyNs))
	Max1000LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Max1000LatencyNs))
	Min1000LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Min1000LatencyNs))
	Packetloss1000.WithLabelValues(hostname).Set(float64(pIp.Packetloss1000))
	// 100 packet ring
	Avg100LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Avg100LatencyNs))
	Jitter100Ns.WithLabelValues(hostname).Set(float64(pIp.Jitter100LatencyNs))
	Max100LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Max100LatencyNs))
	Min100LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Min100LatencyNs))
	Packetloss100.WithLabelValues(hostname).Set(float64(pIp.Packetloss100))
	// 15 packet ring
	Avg15LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Avg15LatencyNs))
	Jitter15Ns.WithLabelValues(hostname).Set(float64(pIp.Jitter15LatencyNs))
	Max15LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Max15LatencyNs))
	Min15LatencyNs.WithLabelValues(hostname).Set(float64(pIp.Min15LatencyNs))
	Packetloss15.WithLabelValues(hostname).Set(float64(pIp.Packetloss15))

	TotalPingsSent.WithLabelValues(hostname, pIp.Ip.String()).Inc()

	if config.Config.Debug {
		fmt.Printf("Updated Prometheus metrics for %s\n", hostname)
	}
}
