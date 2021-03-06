package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricNamespace = "opvic"
	metricSubsystem = "agent"
)

var (
	reconciliationErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "reconciliation_errors_total",
			Help:      "Number of reconciliation errors.",
		},
	)
	lastReconciliationTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "last_reconciliation_timestamp_seconds",
			Help:      "Timestamp of last successful reconciliation",
		},
	)
	reconciliationDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "reconciliation_duration_milliseconds",
			Help:      "Duration of last reconciliation",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(
		reconciliationErrorsTotal,
		lastReconciliationTimestamp,
		reconciliationDuration,
	)
}
