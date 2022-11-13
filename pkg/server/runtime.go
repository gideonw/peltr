package server

import (
	"fmt"
	"net"
)

type Runtime interface {
	Listen() error
	HandleConnections()
	Close()
	AddWorker(conn net.Conn) *WorkerConnection
}

type runtime struct {
	metrics Metrics
	socket  *net.TCPListener
	port    int
	Workers []*WorkerConnection
	Jobs    []Job
}

func NewRuntime(m Metrics, port int) Runtime {
	return &runtime{
		metrics: m,
		socket:  nil,
		port:    port,
		Workers: []*WorkerConnection{},
		Jobs:    []Job{},
	}
}

func (r *runtime) Listen() error {
	sock, err := net.ListenTCP("tcp", &net.TCPAddr{Port: r.port})
	if err != nil {
		return err
	}
	fmt.Println("Server listening on ", r.port)
	r.socket = sock

	return nil
}

func (r *runtime) HandleConnections() {
	for {
		conn, err := r.socket.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}
		wc := r.AddWorker(conn)
		go wc.Handle()
	}
}

func (r *runtime) Close() {
	r.socket.Close()
}

func (r *runtime) AddWorker(conn net.Conn) *WorkerConnection {
	wc := NewWorkerConnection(conn)
	r.Workers = append(r.Workers, wc)

	r.metrics.IncConnections()

	return wc
}
