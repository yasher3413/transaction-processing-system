package processor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsConsumedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_consumed_total",
			Help: "Total number of events consumed",
		},
		[]string{"event_type", "status"},
	)

	workerProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_processing_duration_seconds",
			Help:    "Worker processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)

	retryCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_retries_total",
			Help: "Total number of retries",
		},
		[]string{"event_type"},
	)

	DLQMessagesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dlq_messages_total",
			Help: "Total number of messages sent to DLQ",
		},
	)

	RetryCounter = retryCounter
)
