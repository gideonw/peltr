package server

import "net"

type Runtime interface {
	AddWorker(conn net.Conn) *WorkerConnection
}

type runtime struct {
	metrics Metrics
	Workers []*WorkerConnection
	Jobs    []Job
}

func NewRuntime(m Metrics) Runtime {
	return &runtime{
		metrics: m,
		Workers: []*WorkerConnection{},
		Jobs:    []Job{},
	}
}

func (r *runtime) AddWorker(conn net.Conn) *WorkerConnection {
	wc := NewWorkerConnection(conn)
	r.Workers = append(r.Workers, wc)

	r.metrics.IncConnections()

	return wc
}
