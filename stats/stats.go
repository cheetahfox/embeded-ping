package stats

import (
	"fmt"

	probing "github.com/prometheus-community/pro-bing"
)

type hostLongTerm struct {
	Drop1000p    float64
	Drop100p     float64
	Latency1000p float64
	Latency100p  float64
	Send100p     int
	Send1000p    int
	Receive1000p int
	Receive100p  int
}

var Hosts map[string]hostLongTerm

func UpdateStats() {

}

func pings() {
	pinger, err := probing.NewPinger("www.google.com")
	if err != nil {
		panic(err)
	}
	pinger.Count = 5

	stats := pinger.Statistics()

	fmt.Println(stats.PacketsSent)
}
