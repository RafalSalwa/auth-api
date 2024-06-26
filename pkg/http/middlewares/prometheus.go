package middlewares

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	PrometheusMiddleware struct {
		request *prometheus.CounterVec
		latency *prometheus.HistogramVec
	}
	responseWriterDelegator struct {
		http.ResponseWriter
		status      int
		written     int64
		wroteHeader bool
	}
)

const (
	requestName = "http_requests_total"
	latencyName = "http_request_duration_seconds"
)

var (
	dflBuckets = []float64{0.3, 1.0, 2.5, 5.0}
)

func NewPrometheusMiddleware() *PrometheusMiddleware {
	var prometheusMiddleware PrometheusMiddleware

	prometheusMiddleware.request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: requestName,
			Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
		},
		[]string{"code", "method", "path"},
	)

	if err := prometheus.Register(prometheusMiddleware.request); err != nil {
		log.Println("prometheusMiddleware.request was not registered:", err)
	}

	prometheusMiddleware.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    latencyName,
		Help:    "How long it took to process the request, partitioned by status code, method and HTTP path.",
		Buckets: dflBuckets,
	},
		[]string{"code", "method", "path"},
	)

	if err := prometheus.Register(prometheusMiddleware.latency); err != nil {
		log.Println("prometheusMiddleware.latency was not registered:", err)
	}

	return &prometheusMiddleware
}

func (p *PrometheusMiddleware) Prometheus() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			delegate := &responseWriterDelegator{ResponseWriter: w}
			rw := delegate

			next.ServeHTTP(rw, r) // call original

			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()

			code := sanitizeCode(delegate.status)
			method := sanitizeMethod(r.Method)

			go p.request.WithLabelValues(
				code,
				method,
				path,
			).Inc()

			go p.latency.WithLabelValues(
				code,
				method,
				path,
			).Observe(float64(time.Since(begin)) / float64(time.Second))
		})
	}
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func sanitizeMethod(m string) string {
	return strings.ToLower(m)
}

func sanitizeCode(s int) string {
	return strconv.Itoa(s)
}
