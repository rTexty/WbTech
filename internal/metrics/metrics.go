// Package metrics provides application metrics collection and exposure.
package metrics

// Metrics defines the interface for application metrics.
type Metrics interface {
	// IncMessagesTotal increments the counter for processed Kafka messages.
	// status should be "success" or "error".
	IncMessagesTotal(status string)

	// SetResourceUp sets the availability status of a resource.
	// resource is the name (e.g., "database", "kafka"), up is 1 for available, 0 for unavailable.
	SetResourceUp(resource string, up float64)

	// IncHTTPRequests increments the HTTP request counter.
	IncHTTPRequests(method, path, status string)

	// ObserveHTTPDuration records the duration of an HTTP request.
	ObserveHTTPDuration(method, path string, seconds float64)
}
