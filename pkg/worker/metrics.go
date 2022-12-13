/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobIDLabel         = "job_id"
	JobStatusCodeLabel = "status_code"
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
			Name: "peltr_worker_jobs",
			Help: "The total number of jobs",
		}),
	}
}

func (m *metrics) IncJobs() {
	m.Jobs.Inc()
}
