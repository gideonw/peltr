/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	IncConnections()
}

type metrics struct {
	Connections prometheus.Counter
	Jobs        prometheus.Counter
}

func NewMetricsStore() Metrics {
	return &metrics{
		Connections: promauto.NewCounter(prometheus.CounterOpts{
			Name: "peltr_server_connections",
			Help: "The total number of connections",
		}),
		Jobs: promauto.NewCounter(prometheus.CounterOpts{
			Name: "peltr_server_jobs",
			Help: "The total number of jobs",
		}),
	}
}

func (m *metrics) IncConnections() {
	m.Connections.Inc()
}
