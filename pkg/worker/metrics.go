package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	IncJobs()
}

type metrics struct {
	Jobs prometheus.Counter
}

func NewMetricsStore() Metrics {
	return &metrics{
		Jobs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "peltr_server_jobs",
			Help: "The total number of jobs",
		}),
	}
}

func (m *metrics) IncJobs() {
	m.Jobs.Inc()
}
