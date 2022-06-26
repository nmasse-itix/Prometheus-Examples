package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	protov1 "github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"google.golang.org/protobuf/proto"
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
	} else if os.Args[1] == "remote-write" {
		u, _ := url.Parse("http://localhost:9090/api/v1/write")
		remoteConfig := remote.ClientConfig{
			URL:              &promConfig.URL{URL: u},
			Timeout:          model.Duration(10 * time.Second),
			RetryOnRateLimit: true,
		}

		client, err := remote.NewWriteClient("test-app", &remoteConfig)
		if err != nil {
			panic(err)
		}

		val := 1.0
		for {
			ts := prompb.TimeSeries{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "myapp_highfreq_data"},
				},
				Samples: []prompb.Sample{
					{Timestamp: timestamp.FromTime(time.Now()), Value: val},
				},
			}

			reqv1 := prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{ts},
			}
			reqv2 := protov1.MessageV2(&reqv1)

			if buf, err := proto.Marshal(reqv2); err != nil {
				panic(err)
			} else {
				encoded := snappy.Encode(nil, buf) // this call can panic
				if err = client.Store(context.Background(), encoded); err != nil {
					panic(err)
				}
			}
			time.Sleep(10 * time.Millisecond)
			val += 0.01
		}
	}
}
