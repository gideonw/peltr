package server

import (
	"fmt"
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
	HandleListJobQueue(rw http.ResponseWriter, req *http.Request)
	HandleListWorkers(rw http.ResponseWriter, req *http.Request)
}

type runtime struct {
	metrics      Metrics
	log          zerolog.Logger
	socket       *net.TCPListener
	port         int
	Workers      []*WorkerConnection
	JobQueue     []proto.Job
	AssignedJobs []proto.Job
}

func NewRuntime(m Metrics, logger zerolog.Logger, port int) Runtime {
	return &runtime{
		metrics:      m,
		log:          logger,
		socket:       nil,
		port:         port,
		Workers:      []*WorkerConnection{},
		JobQueue:     []proto.Job{},
		AssignedJobs: []proto.Job{},
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
		r.metrics.IncConnections()
		wc := r.AddWorker(conn)
		go wc.Handle()
	}
}

func (r *runtime) ControlLoop() {
	for {
		if len(r.JobQueue) <= 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Assign jobs to workers
		for i := range r.Workers {
			if len(r.JobQueue) == 0 {
				break
			}
			if r.Workers[i].State == "alive" {
				job := r.JobQueue[0]
				r.Workers[i].AssignJob(job)
				r.AssignedJobs = append(r.AssignedJobs, job)
				r.JobQueue = r.JobQueue[1:]
				r.log.Debug().Func(func(e *zerolog.Event) {
					l := e.Int("jobQueue", len(r.JobQueue))
					for i := range r.JobQueue {
						l.Str(fmt.Sprint(i), r.JobQueue[i].ID)
					}
				}).Msg("jobQueue items")

				r.log.Info().
					Str("workerID", r.Workers[i].ID).
					Str("workerState", r.Workers[i].State).
					Str("JobID", job.ID).
					Msg("Assigned job")
			}
		}

		r.Workers = shiftSlice(r.Workers)
	}
}

func shiftSlice[V *WorkerConnection](s []V) []V {
	if len(s) <= 1 {
		return s
	}
	temp := s[0]
	ret := s[1:]
	ret = append(ret, temp)
	return ret
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
