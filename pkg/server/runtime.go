package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

type Runtime interface {
	Listen() error
	HandleConnections()
	Close()
	AddWorker(conn net.Conn) *WorkerConnection
	HandleJob(rw http.ResponseWriter, req *http.Request)
	HandleListJobs(rw http.ResponseWriter, req *http.Request)
	HandleListWorkers(rw http.ResponseWriter, req *http.Request)
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

func (r *runtime) HandleJob(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var j Job

	err = json.Unmarshal(b, &j)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	r.Jobs = append(r.Jobs, j)
}

func (r *runtime) HandleListJobs(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := json.Marshal(r.Jobs)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = rw.Write(b)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (r *runtime) HandleListWorkers(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := json.Marshal(r.Workers)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = rw.Write(b)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
