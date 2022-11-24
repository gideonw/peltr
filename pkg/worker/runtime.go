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
	// start := time.Now()

	for wr.State != "closed" {
		b := make([]byte, 256)

		_, err := wr.read(b)
		if err != nil {
			break
		}

		command, msg := proto.ChompCommand(b)
		wr.log.Debug().Str("cmd", string(command)).Msg("chomp")
		wr.log.Trace().Bytes("msg", msg).Msg("chomp")

		wr.processInput(command, msg)

		wr.processState()

		wr.scheduler()
	}
}

func (wr *workerRuntime) read(b []byte) (int, error) {
	n, err := wr.conn.Read(b)
	if err == io.EOF {
		wr.log.Error().Err(err).Msg("disconnected")
		wr.State = "closed"
		return n, err
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
	jw := NewJobWorker(wr.log, job)
	wr.Workers = append(wr.Workers, jw)

	// Start the job handler
	go jw.HandleJob()
}

func (wr *workerRuntime) processInput(cmd, msg []byte) {
	switch 0 {
	case bytes.Compare(cmd, proto.CommandHello):
		wr.updateState("identify")
	case bytes.Compare(cmd, proto.CommandPing):
		wr.updateState("ping")
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
		wr.updateState("idle")
	case "idle":
	case "ping":
		err := wr.sendPong()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
		}
	case "assign":
		// go HandleJob(wr.log, job.Jobs[0])
		err := wr.sendWorking()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
		}
	case "working":
		err := wr.sendUpdate()
		if err != nil {
			log.Error().Err(err).Msg("error sending")
		}
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

func (wr *workerRuntime) sendPong() error {
	n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandPong, nil))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "pong").Int("bytes", n).Msg("wrote")
	return nil
}

func (wr *workerRuntime) sendWorking() error {
	n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandWorking, nil))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "working").Int("bytes", n).Msg("wrote")

	return nil
}

func (wr *workerRuntime) sendUpdate() error {
	if len(wr.Workers) == 0 {
		return fmt.Errorf("no current job")
	}
	updates := []proto.Update{}
	for i := range wr.Workers {
		if wr.Workers[i].Done {
			updates = append(updates, proto.Update{
				ID:      wr.Workers[i].Job.ID,
				Results: wr.Workers[i].Results,
			})
		}
	}
	n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandUpdate, updates))
	if err != nil {
		return err
	}
	wr.log.Info().Str("type", "working").Int("bytes", n).Msg("wrote")

	return nil
}

func (wr *workerRuntime) updateState(s string) {
	wr.log.Debug().Str("state", s).Msg("state change")
	wr.State = s
}
