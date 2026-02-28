package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics implements Metrics interface using Prometheus.
type PrometheusMetrics struct {
	messagesTotal   *prometheus.CounterVec
	resourceUp      *prometheus.GaugeVec
	httpRequests    *prometheus.CounterVec
	httpDuration    *prometheus.HistogramVec
}

// NewPrometheus creates a new PrometheusMetrics instance with all metrics registered.
func NewPrometheus() *PrometheusMetrics {
	return &PrometheusMetrics{
		messagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_messages_total",
				Help: "Total number of processed Kafka messages",
			},
			[]string{"status"},
		),
		resourceUp: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "resource_up",
				Help: "Resource availability status (1 = up, 0 = down)",
			},
			[]string{"resource"},
		),
		httpRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

// IncMessagesTotal increments the Kafka messages counter.
func (p *PrometheusMetrics) IncMessagesTotal(status string) {
	p.messagesTotal.WithLabelValues(status).Inc()
}

// SetResourceUp sets the availability gauge for a resource.
func (p *PrometheusMetrics) SetResourceUp(resource string, up float64) {
	p.resourceUp.WithLabelValues(resource).Set(up)
}

// IncHTTPRequests increments the HTTP requests counter.
func (p *PrometheusMetrics) IncHTTPRequests(method, path, status string) {
	p.httpRequests.WithLabelValues(method, path, status).Inc()
}

// ObserveHTTPDuration records HTTP request duration.
func (p *PrometheusMetrics) ObserveHTTPDuration(method, path string, seconds float64) {
	p.httpDuration.WithLabelValues(method, path).Observe(seconds)
}
