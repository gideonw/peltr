package server

import (
	"bytes"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/rs/zerolog"
)

var (
	PING_INTERVAL = 1 * time.Second
)

type WorkerConnection struct {
	log      zerolog.Logger
	Conn     net.Conn
	ID       string
	Capacity uint
	// State
	// - new
	// - hello
	// - alive
	State    string
	LastSeen time.Time
	// Server assigned jobs that have not been accepted or sent
	JobQueue []proto.Job
	// Server assigned jobs that have not been accepted
	AssignJobQueue []proto.Job
	// Jobs accepted by the worker
	AcceptedJobs []proto.Job
	// TODO handles to provide state to the runtime
}

// NewWorkerConnection handles the connection and state for a worker connection
func NewWorkerConnection(logger zerolog.Logger, conn net.Conn) *WorkerConnection {
	return &WorkerConnection{
		log:      logger,
		ID:       "",
		Conn:     conn,
		Capacity: 0,
		State:    "new",
	}
}

func (wc *WorkerConnection) AssignJob(job proto.Job) {
	wc.JobQueue = append(wc.JobQueue, job)
	wc.updateState("accept")
}

func (wc *WorkerConnection) Handle() {
	wc.log.Info().Str("remote", wc.Conn.RemoteAddr().String()).Str("local", wc.Conn.LocalAddr().String()).Msg("handling connection")
	wc.LastSeen = time.Now().Add(2 * time.Second)
	for {
		var err error
		data := make([]byte, 4096)

		// Write to the client depending on state
		wc.log.Debug().Str("state", wc.State).Msg("process state")
		switch wc.State {
		case "new":
			err = wc.sendHello()
		case "hello":
			wc.updateState("alive")
			continue
		case "alive":
			for {
				if len(wc.JobQueue) > 0 {
					err = wc.sendAssign()
					wc.updateState("accept")
					break
				} else if time.Since(wc.LastSeen) > PING_INTERVAL {
					err = wc.sendAlive()
					wc.updateState("alive")
					break
				}
			}
		case "accept":
			wc.log.Info().Msg("waiting for next status for accept")
			// wc.updateState("alive")
		}

		// Check for write errors
		if errors.Is(err, syscall.EPIPE) {
			wc.log.Error().Err(err).Msg("EPIPE Connection closed")
			return
		} else if err != nil {
			wc.log.Error().Err(err).Msg("Connection error")
			return
		}

		// Read from the client
		n, err := wc.Conn.Read(data)
		wc.log.Debug().Int("bytes", n).Msg("read")
		if errors.Is(err, syscall.EPIPE) {
			wc.log.Error().Err(err).Msg("EPIPE Connection closed")
			return
		} else if err == io.EOF {
			wc.log.Error().Err(err).Msg("EOF Connection closed")
			return
		} else if err != nil {
			wc.log.Error().Err(err).Msg("Connection error")
			return
		}

		command, message := proto.ChompCommand(data)
		switch 0 {
		case bytes.Compare(command, proto.CommandHello):
			wc.log.Info().Str("cmd", "hello").Msg("identify")
			if wc.State != "new" {
				wc.log.Error().Msgf("Expected 'new' state, got %s. Disconnecting", wc.State)
				wc.Conn.Close()
				wc.updateState("closed")
				return
			}

			data, err := proto.ParseIdentify(message)
			if err != nil {
				wc.log.Error().Err(err)
				continue
			}

			wc.ID = data.ID
			wc.log = wc.log.With().Str("id", wc.ID).Logger()
			wc.Capacity = data.Capacity

			wc.updateState("hello")
		case bytes.Compare(command, proto.CommandStatus):
			wc.log.Info().Str("cmd", "status").Msg("sync jobs")
			wc.syncJobs(message)
			wc.updateState("alive")
		case bytes.Compare(command, proto.CommandAccept):
			wc.log.Info().Str("cmd", "accept").Msg("sync jobs")
			wc.syncJobs(message)
			wc.updateState("alive")
		}
	}
}

func (wc *WorkerConnection) updateState(state string) {
	wc.log.Debug().Str("state", state).Msg("state change")
	wc.LastSeen = time.Now()
	wc.State = state
}

func (wc *WorkerConnection) sendHello() error {
	wc.log.Debug().Str("type", "hello").Msg("send")
	n, err := wc.Conn.Write(proto.MakeMessageByte(proto.CommandHello, nil))
	wc.log.Debug().Int("bytes", n).Msg("write")

	return err
}

func (wc *WorkerConnection) sendAlive() error {
	wc.log.Debug().Str("type", "alive").Msg("send")
	n, err := wc.Conn.Write(proto.MakeMessageByte(proto.CommandAlive, nil))
	wc.log.Debug().Int("bytes", n).Msg("write")

	return err
}

func (wc *WorkerConnection) sendAssign() error {
	wc.log.Debug().Str("type", "assign").Msg("send")
	n, err := wc.Conn.Write(proto.MakeMessageStruct(proto.CommandAssign, proto.Assign{Jobs: wc.JobQueue}))
	if err != nil {
		return err
	}
	wc.log.Debug().Int("bytes", n).Msg("write")

	wc.AssignJobQueue = wc.JobQueue[:]
	wc.JobQueue = wc.JobQueue[0:0]

	return nil
}

func (wc *WorkerConnection) syncJobs(b []byte) error {
	status, err := proto.ParseStatus(b)
	if err != nil {
		return err
	}

	for i := range status.ActiveJobs {
		found := false
		foundID := ""
		for j := range wc.AcceptedJobs {
			found = wc.AcceptedJobs[j].ID == status.ActiveJobs[i].ID
			if found {
				foundID = status.ActiveJobs[i].ID
				break
			}
		}
		if found {
			for i := range wc.AssignJobQueue {
				if foundID == wc.AssignJobQueue[i].ID {
					job := wc.AssignJobQueue[i]
					wc.AssignJobQueue = append(wc.AssignJobQueue[0:i], wc.AssignJobQueue[min(i+1, len(wc.AssignJobQueue)):]...)
					wc.AcceptedJobs = append(wc.AcceptedJobs, job)
				}
			}
		}
	}

	return nil
}

func min(a, b int) int {
	if a >= b {
		return b
	}
	return a
}
