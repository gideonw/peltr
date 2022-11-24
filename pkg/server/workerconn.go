package server

import (
	"bytes"
	"errors"
	"fmt"
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
	/// State
	// new
	// hello
	// status
	State       string
	LastSeen    time.Time
	AssignedJob proto.Job
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

func (wc *WorkerConnection) AssignJob(job proto.Job) (bool, error) {
	wc.AssignedJob = job

	return true, nil
}

func (wc *WorkerConnection) Handle() {
	wc.log.Info().Str("remote", wc.Conn.RemoteAddr().String()).Str("local", wc.Conn.LocalAddr().String()).Msg("handling connection")
	lastPing := time.Now().Add(2 * time.Second)
	for {
		var err error
		data := make([]byte, 256)

		// Write to the client depending on state
		wc.log.Debug().Str("state", wc.State).Msg("process state")
		switch wc.State {
		case "new":
			err = wc.sendHello()
		case "hello":
			wc.updateState("idle")
			continue
		case "idle":
			// Loop while idle to wait for a job or ping
			for {
				if wc.AssignedJob.ID != "" {
					// TODO: Job logic
					err = wc.sendAssign()
					break
				} else if time.Now().Sub(lastPing) > PING_INTERVAL {
					err = wc.sendPing()
					wc.updateState("idle-ping")
					break
				}
			}
		case "working":
			wc.log.Info().Msg("working")
		case "done":
			fmt.Println("done")
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
		case bytes.Compare(command, proto.CommandPong):
			if wc.State != "idle-ping" {
				wc.log.Error().Msgf("Expected 'idle-ping' state, got %s. Disconnecting", wc.State)
				wc.Conn.Close()
				return
			}

			lastPing = time.Now()
			wc.updateState("idle")
		case bytes.Compare(command, proto.CommandWorking):
			wc.updateState("working")
		case bytes.Compare(command, proto.CommandUpdate):
			wc.updateState("working")
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

func (wc *WorkerConnection) sendPing() error {
	wc.log.Debug().Str("type", "ping").Msg("send")
	n, err := wc.Conn.Write(proto.MakeMessageByte(proto.CommandPing, nil))
	wc.log.Debug().Int("bytes", n).Msg("write")

	return err
}

func (wc *WorkerConnection) sendAssign() error {
	wc.log.Debug().Str("type", "assign").Msg("send")
	n, err := wc.Conn.Write(proto.MakeMessageStruct(proto.CommandAssign, proto.Assign{Jobs: []proto.Job{wc.AssignedJob}}))
	if err != nil {
		return err
	}
	wc.log.Debug().Int("bytes", n).Msg("write")

	return nil
}
