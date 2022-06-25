package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

func recordMetrics(delay time.Duration) {
	temp := 10
	go func() {
		for {
			opsProcessed.Inc()
			temperature.Set(float64(temp))
			time.Sleep(delay)
			temp += 1
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
	temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_temperature",
		Help: "The current room temperature",
	})
)

func main() {

	if os.Args[1] == "scrape" {
		recordMetrics(2 * time.Second)

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	} else if os.Args[1] == "scrape2" {
		recordMetrics(60 * time.Second)

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	} else if os.Args[1] == "push" {
		for i := 0; i < 10; i++ {
			opsProcessed.Inc()
			time.Sleep(1 * time.Second)
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(opsProcessed)

		pusher := push.New("http://localhost:8888", "db_backup").Gatherer(registry)
		pusher.Format(expfmt.FmtOpenMetrics)
		if err := pusher.Add(); err != nil {
			fmt.Println("Could not push to Pushgateway:", err)
		}
	} else if os.Args[1] == "push2" {
		registry := prometheus.NewRegistry()
		registry.MustRegister(opsProcessed)
		pusher := push.New("http://localhost:8888", "db_backup").Gatherer(registry)
		pusher.Format(expfmt.FmtOpenMetrics)

		for i := 0; i < 10; i++ {
			opsProcessed.Inc()
			time.Sleep(1 * time.Second)
			if err := pusher.Push(); err != nil {
				fmt.Println("Could not push to Pushgateway:", err)
			}
		}

	}
}
