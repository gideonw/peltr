package worker

import "net"

type WorkerRuntime interface {
}

type workerRuntime struct {
	conn *net.TCPConn
}
