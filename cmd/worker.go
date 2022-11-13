package cmd

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
)

type WorkerCommand struct {
	Port     int `short:"p" long:"port" description:"Port to connect to for worker instructions" default:"8001"`
	DataPort int `short:"d" long:"data-port" description:"Port to send job results to" default:"8002"`
}

var WorkerCmd WorkerCommand

func (sc *WorkerCommand) Execute(args []string) error {
	retryCount := 3

	var conn net.Conn
	var err error

	for retryCount > 0 && conn == nil {
		conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sc.Port))
		if err != nil {
			fmt.Println(err)
			retryCount -= 1
		}
		time.Sleep(2 * time.Second)
	}
	defer conn.Close()
	fmt.Println("Worker connected on ", conn.RemoteAddr())

	HandleServerProt(conn)

	return nil
}

func HandleServerProt(conn net.Conn) {
	start := time.Now()

	for {
		b := make([]byte, 256)

		if time.Now().After(start.Add(1 * time.Second)) {
			start = time.Now()
			n, err := conn.Read(b)
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
				n, err := conn.Write(proto.MakeMessageString(proto.CommandHello, "a,10"))
				if err != nil {
					fmt.Println("Error sending hello", err)
				}
				fmt.Printf("Wrote %d b\n", n)
			case string(proto.CommandPing):
				n, err := conn.Write(proto.MakeMessageByte(proto.CommandPong, nil))
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
