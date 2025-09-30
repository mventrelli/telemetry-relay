package observability

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Ingested    = promauto.NewCounter(prometheus.CounterOpts{Name: "telemetry_ingested_total", Help: "packets received"})
	ParseErrors = promauto.NewCounter(prometheus.CounterOpts{Name: "telemetry_parse_errors_total", Help: "json parse errors"})
	Forwarded   = promauto.NewCounter(prometheus.CounterOpts{Name: "telemetry_forwarded_total", Help: "packets forwarded"})
	ForwardErrs = promauto.NewCounter(prometheus.CounterOpts{Name: "telemetry_forward_errors_total", Help: "http forward errors"})
	UDPBytes    = promauto.NewCounter(prometheus.CounterOpts{Name: "telemetry_udp_bytes_total", Help: "udp payload bytes"})
	ForwardSec  = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "telemetry_forward_seconds",
		Help:    "HTTP forward latency",
		Buckets: prometheus.DefBuckets,
	})
)

func Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Handle("/metrics", promhttp.Handler())
	return r
}
