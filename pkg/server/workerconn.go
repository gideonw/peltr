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
)

var (
	PING_INTERVAL = 1 * time.Second
)

type WorkerConnection struct {
	Conn     net.Conn
	ID       string
	Capacity uint
	/// State
	// new
	// hello
	// named
	// idle
	// idle-ping
	// working
	State       string
	LastSeen    time.Time
	AssignedJob string
	Debug       bool
	// TODO handles to provide state to the runtime
}

// NewWorkerConnection handles the connection and state for a worker connection
func NewWorkerConnection(conn net.Conn) *WorkerConnection {
	return &WorkerConnection{
		ID:       "",
		Conn:     conn,
		Capacity: 0,
		State:    "new",
		Debug:    true,
	}
}

func (wc *WorkerConnection) AssignJob() (bool, error) {
	if wc.State != "idle" {
		return false, nil
	}

	return true, nil
}

func (wc *WorkerConnection) Handle() {
	fmt.Println(wc.Conn.RemoteAddr(), wc.Conn.LocalAddr())
	lastPing := time.Now().Add(2 * time.Second)
	for {
		var err error
		data := make([]byte, 256)

		// Write to the client depending on state
		switch wc.State {
		case "new":
			err = wc.sendHello()
		case "hello":
			fmt.Println("Hello worker ", wc.ID)
			wc.updateState("idle")
			continue
		case "idle":
			// Loop while idle to wait for a job or ping
			for {
				if wc.AssignedJob != "" {
					// TODO: Job logic
					break
				} else if time.Now().Sub(lastPing) > PING_INTERVAL {
					err = wc.sendPing()
					wc.updateState("idle-ping")
					break
				}
			}
		}
		// Check for write errors
		if errors.Is(err, syscall.EPIPE) {
			fmt.Println("Connection closed", err)
			return
		} else if err != nil {
			fmt.Println(err)
			return
		}

		// Read from the client
		n, err := wc.Conn.Read(data)
		if wc.Debug {
			fmt.Printf("Read %d bytes\n", n)
		}
		if errors.Is(err, syscall.EPIPE) {
			fmt.Println("Connection closed", err)
			return
		} else if err == io.EOF {
			fmt.Println("Connection closed. EOF", err)
			return
		} else if err != nil {
			fmt.Println(err)
			return
		}

		command, message := proto.ChompCommand(data)
		switch 0 {
		case bytes.Compare(command, proto.CommandHello):
			if wc.State != "new" {
				fmt.Printf("Expected 'new' state, got %s. Disconnecting", wc.State)
				wc.Conn.Close()
				wc.updateState("closed")
				return
			}

			data, err := proto.ParseIdentify(message)
			if err != nil {
				fmt.Println(err)
				continue
			}

			wc.ID = data.ID
			wc.Capacity = data.Capacity

			wc.updateState("hello")
		case bytes.Compare(command, proto.CommandPong):
			if wc.State != "idle-ping" {
				fmt.Printf("Expected 'idle-ping' state, got %s. Disconnecting", wc.State)
				wc.Conn.Close()
				return
			}

			lastPing = time.Now()
			wc.updateState("idle")
		}
	}

}

func (wc *WorkerConnection) updateState(state string) {
	if wc.Debug {
		fmt.Println("sc:", state)
	}
	wc.LastSeen = time.Now()
	wc.State = state
}

func (wc *WorkerConnection) sendHello() error {
	if wc.Debug {
		fmt.Println("Sending Hello")
	}

	n, err := wc.Conn.Write(proto.MakeMessageByte(proto.CommandHello, nil))
	if wc.Debug {
		fmt.Printf("Wrote %d bytes\n", n)
	}

	return err
}

func (wc *WorkerConnection) sendPing() error {
	if wc.Debug {
		fmt.Println("Sending Ping")
	}

	n, err := wc.Conn.Write(proto.MakeMessageByte(proto.CommandPing, nil))
	if wc.Debug {
		fmt.Printf("Wrote %d bytes\n", n)
	}

	return err
}
