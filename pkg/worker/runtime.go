package worker

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type WorkerRuntime interface {
	Connect() error
	Handle()
	Close()
}

type workerRuntime struct {
	metrics Metrics
	log     zerolog.Logger
	host    string
	port    int

	Capacity uint
	ID       string
	conn     net.Conn

	State string

	JobQueue []proto.Job
	Workers  []JobWorker
}

func NewRuntime(m Metrics, logger zerolog.Logger, host string, port int) WorkerRuntime {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return &workerRuntime{
		log:      logger,
		metrics:  m,
		host:     host,
		port:     port,
		State:    "new",
		Capacity: 10,
		ID:       id.String(),
		conn:     nil,
	}
}

func (wr *workerRuntime) Connect() error {
	retryCount := 3

	var err error

	for retryCount > 0 && wr.conn == nil {
		wr.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", wr.host, wr.port))
		if err != nil {
			wr.log.Error().Err(err)
			retryCount -= 1
		}
		time.Sleep(2 * time.Second)
	}
	wr.log.Info().Str("addr", wr.conn.RemoteAddr().String()).Msg("worker connected")

	return nil
}

func (wr *workerRuntime) Close() {
	wr.conn.Close()
}

func (wr *workerRuntime) Handle() {
	buffer := make([]byte, 4096)
	for wr.State != "closed" {

		n, err := wr.read(buffer)
		if err != nil {
			break
		}
		if n != 0 {
			b := make([]byte, n)
			copy(b, buffer[:n])
			command, msg := proto.ChompCommand(b)
			wr.log.Debug().Str("cmd", string(command)).Msg("chomp")
			wr.log.Trace().Bytes("msg", msg).Msg("chomp")

			wr.processInput(command, msg)
		}

		wr.processState()

		wr.scheduler()
	}
}

func (wr *workerRuntime) read(b []byte) (int, error) {
	err := wr.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		wr.log.Error().Err(err).Msg("unable to set read deadline")
		return 0, nil
	}
	n, err := wr.conn.Read(b)
	if err == io.EOF {
		wr.log.Error().Err(err).Msg("disconnected")
		wr.State = "closed"
		return n, err
	} else if err, ok := err.(net.Error); ok && err.Timeout() {
		// timeout is fine since we have a deadline and don't want to block forever
		wr.log.Error().Err(err).Msg("error reading from conn")
		return 0, nil
	} else if err != nil {
		wr.log.Error().Err(err).Msg("error reading from conn")
		return n, err
	}
	wr.log.Debug().Int("bytes", n).Msg("read from conn")

	return n, nil
}

func (wr *workerRuntime) scheduler() {
	if len(wr.Workers) >= int(wr.Capacity) {
		wr.log.Debug().Msg("worker at capacity")
		return
	}

	if len(wr.JobQueue) <= 0 {
		wr.log.Debug().Msg("no assigned jobs waiting")
		return
	}

	// FIFO take the head of the job queue and create a JobWorker
	job := wr.JobQueue[0]
	wr.JobQueue = wr.JobQueue[1:]

	// create the worker and keep track of it
	jw := NewJobWorker(wr.log, wr.metrics, job)
	wr.Workers = append(wr.Workers, jw)

	// metrics
	wr.metrics.IncJobs()

	// Start the job handler
	go jw.HandleJob()
}

func (wr *workerRuntime) processInput(cmd, msg []byte) {
	switch 0 {
	case bytes.Compare(cmd, proto.CommandHello):
		wr.updateState("identify")
	case bytes.Compare(cmd, proto.CommandAlive):
		wr.updateState("alive")
	case bytes.Compare(cmd, proto.CommandAssign):
		// Parse Job
		job, err := proto.ParseAssign(msg)
		if err != nil {
			wr.log.Error().Str("type", "assign").Err(err).Msg("error parsing message")
		}
		wr.JobQueue = append(wr.JobQueue, job.Jobs...)

		wr.updateState("assign")
	default:
		wr.log.Error().Msgf("Unknown command '%s','%s'\n", cmd, msg)
	}
}

func (wr *workerRuntime) processState() {
	log := wr.log.With().Str("state", wr.State).Logger()
	switch wr.State {
	case "new":
	case "identify":
		err := wr.sendIdentify()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
			wr.updateState("new")
		}
		wr.updateState("alive")

	case "alive":
		err := wr.sendStatus()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
		}
		wr.updateState("alive")
	case "assign":
		err := wr.sendAccept()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
		}
		wr.updateState("alive")
	}
}

func (wr *workerRuntime) sendIdentify() error {
	n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandHello, proto.Identify{
		ID:       wr.ID,
		Capacity: wr.Capacity,
	}))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "hello").Int("bytes", n).Msg("wrote")
	return nil
}

func (wr *workerRuntime) sendStatus() error {
	n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandStatus, wr.compileStatus()))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "status").Int("bytes", n).Msg("wrote")

	return nil
}

func (wr *workerRuntime) sendAccept() error {
	n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandAccept, wr.compileStatus()))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "status").Int("bytes", n).Msg("wrote")

	return nil
}

func (wr *workerRuntime) compileStatus() proto.Status {
	status := proto.Status{
		JobQueue:   wr.JobQueue,
		ActiveJobs: []proto.Job{},
		Results:    make(map[string]map[int]int),
	}
	for i := range wr.Workers {
		status.ActiveJobs = append(status.ActiveJobs, wr.Workers[i].Job)
		if wr.Workers[i].Done {
			status.Results[wr.Workers[i].Job.ID] = wr.Workers[i].Results
		}
	}
	return status
}

func (wr *workerRuntime) updateState(s string) {
	wr.log.Debug().Str("state", s).Msg("state change")
	wr.State = s
}
