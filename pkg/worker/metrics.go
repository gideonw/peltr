package worker

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobIDLabel         = "job_id"
	JobStatusCodeLabel = "status_code"
)

type Metrics interface {
	IncJobs()
	IncJobRequestCount(id string, status int)
	ObserveJobRequestDurations(id string, status int, duration time.Duration)
}

type metrics struct {
	Jobs                prometheus.Counter
	JobRequestCounts    *prometheus.CounterVec
	JobRequestDurations *prometheus.HistogramVec
}

func NewMetricsStore() Metrics {
	return &metrics{
		Jobs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "peltr_worker_jobs",
			Help: "The total number of jobs",
		}),
		JobRequestCounts: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "peltr_worker_job_requests",
			Help: "Job requests durations",
		}, []string{JobIDLabel, JobStatusCodeLabel}),
		JobRequestDurations: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "peltr_worker_job_request_ms",
			Help: "Job requests durations",
		}, []string{JobIDLabel, JobStatusCodeLabel}),
	}
}

func (m *metrics) IncJobs() {
	m.Jobs.Inc()
}

func (m *metrics) IncJobRequestCount(id string, status int) {
	m.JobRequestCounts.With(prometheus.Labels{JobIDLabel: id, JobStatusCodeLabel: fmt.Sprint(status)}).Inc()
}

func (m *metrics) ObserveJobRequestDurations(id string, status int, duration time.Duration) {
	m.JobRequestDurations.
		With(prometheus.Labels{JobIDLabel: id, JobStatusCodeLabel: fmt.Sprint(status)}).
		Observe(float64(duration.Milliseconds()))
}
