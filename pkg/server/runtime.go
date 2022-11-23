package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/rs/zerolog"
)

type Runtime interface {
	Listen() error
	HandleConnections()
	ControlLoop()
	Close()
	AddWorker(conn net.Conn) *WorkerConnection
	HandleJob(rw http.ResponseWriter, req *http.Request)
	HandleListJobs(rw http.ResponseWriter, req *http.Request)
	HandleListWorkers(rw http.ResponseWriter, req *http.Request)
}

type runtime struct {
	metrics Metrics
	log     zerolog.Logger
	socket  *net.TCPListener
	port    int
	Workers []*WorkerConnection
	Jobs    []proto.Job
}

func NewRuntime(m Metrics, logger zerolog.Logger, port int) Runtime {
	return &runtime{
		metrics: m,
		log:     logger,
		socket:  nil,
		port:    port,
		Workers: []*WorkerConnection{},
		Jobs:    []proto.Job{},
	}
}

func (r *runtime) Listen() error {
	sock, err := net.ListenTCP("tcp", &net.TCPAddr{Port: r.port})
	if err != nil {
		return err
	}
	r.log.Info().Int("port", r.port).Msg("server listening")
	r.socket = sock

	return nil
}

func (r *runtime) HandleConnections() {

	for {
		conn, err := r.socket.Accept()
		if err != nil {
			r.log.Error().Err(err).Msg("error accepting connection")
			continue
		}
		wc := r.AddWorker(conn)
		go wc.Handle()
	}
}

func (r *runtime) ControlLoop() {
	for {
		if len(r.Jobs) <= 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		// Assign jobs to workers
		for i := range r.Workers {
			if r.Workers[i].State == "idle" {
				job := r.Jobs[0]
				success, err := r.Workers[i].AssignJob(job)
				if err != nil {
					r.log.Error().Err(err)
					continue
				}
				if !success {
					r.log.Info().Str("workerID", r.Workers[i].ID).Str("workerState", r.Workers[i].State).Msg("Unable to assign job")
				}
				r.Jobs = r.Jobs[0:]
			}
		}
	}
}

func (r *runtime) Close() {
	r.socket.Close()
}

func (r *runtime) AddWorker(conn net.Conn) *WorkerConnection {
	wc := NewWorkerConnection(r.log, conn)
	r.Workers = append(r.Workers, wc)

	r.metrics.IncConnections()

	return wc
}

func countWorkersInState(workers []*WorkerConnection, state string) int {
	count := 0
	for i := range workers {
		if workers[i].State == state {
			count += 1
		}
	}
	return count
}
