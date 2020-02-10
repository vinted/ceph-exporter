package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	bindAddr      = flag.String("telemetry.addr", ":9353", "host:port for ceph exporter")
	asokPath      = flag.String("asok.path", "/var/run/ceph", "path to ceph admin socket direcotry")
	queryInterval = flag.Int("query.interval", 15, "How often should daemon read asok metrics (in seconds0")
	logLevel      = flag.String("log.level", "info", "Logging level")
)

func main() {

	flag.Parse()

	switch *logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	go CollectTimer(*queryInterval)

	ceph := newCephCollector()
	prometheus.MustRegister(ceph)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
    <head><title>Ceph Exporter</title></head>
    <body>
    <h1>Ceph Exporter</h1>
    <p><a href='metrics'>Metrics</a></p>
    </body>
    </html>`))
		if err != nil {
			log.Error("HTTP write failed: ", err)
		}
	})
	log.Info("Listening on: ", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))

}
