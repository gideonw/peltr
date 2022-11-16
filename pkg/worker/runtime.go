package worker

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/google/uuid"
)

type WorkerRuntime interface {
	Connect() error
	Handle()
	Close()
}

type workerRuntime struct {
	metrics Metrics
	port    int

	State    string
	Capacity uint
	ID       string
	conn     net.Conn
}

func NewRuntime(m Metrics, port int) WorkerRuntime {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return &workerRuntime{
		metrics:  m,
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
		wr.conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", wr.port))
		if err != nil {
			fmt.Println(err)
			retryCount -= 1
		}
		time.Sleep(2 * time.Second)
	}
	fmt.Println("Worker connected on ", wr.conn.RemoteAddr())

	return nil
}

func (wr *workerRuntime) Close() {
	wr.conn.Close()
}

func (wr *workerRuntime) Handle() {
	start := time.Now()

	for {
		b := make([]byte, 256)

		if time.Now().After(start.Add(1 * time.Second)) {
			start = time.Now()
			n, err := wr.conn.Read(b)
			if err == io.EOF {
				fmt.Println("Disconnected", err)
				break
			} else if err != nil {
				fmt.Println("Error reading from conn", err)
			}
			fmt.Println("Read from conn", n, string(b))
			command, msg := proto.ChompCommand(b)
			switch string(command) {
			case string(proto.CommandHello):
				n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandHello, proto.Identify{
					ID:       wr.ID,
					Capacity: wr.Capacity,
				}))
				if err != nil {
					fmt.Println("Error sending hello", err)
				}
				fmt.Printf("Wrote %d b\n", n)
			case string(proto.CommandPing):
				n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandPong, nil))
				if err != nil {
					fmt.Println("Error sending pong", err)
				}
				fmt.Printf("Wrote %d b\n", n)

			case string(proto.CommandAssign):
				n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandWorking, nil))
				if err != nil {
					fmt.Println("Error sending pong", err)
				}
				fmt.Printf("Wrote %d b\n", n)

			default:
				fmt.Printf("Unknown command '%s','%s'\n", command, msg)
				fmt.Printf("wat '%v' '%v' '%v'", b, string(b), "hello")
			}
		}
	}
}
