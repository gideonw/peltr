package cmd

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gideonw/peltr/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerCommand struct {
	Port     int `short:"p" long:"port" description:"Port to listen for workers" default:"8001"`
	DataPort int `short:"d" long:"data-port" description:"Port to listen for job results" default:"8002"`
}

var ServerCmd ServerCommand

func (sc *ServerCommand) Execute(args []string) error {
	m := server.NewMetricsStore()
	runtime := server.NewRuntime(m)

	http.Handle("/metrics", promhttp.Handler())
	go func() { http.ListenAndServe(":2112", nil) }()

	sock, err := net.ListenTCP("tcp", &net.TCPAddr{Port: sc.Port})
	if err != nil {
		return err
	}
	defer func() {
		err := sock.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("Server listening on ", sc.Port)

	for {
		conn, err := sock.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}
		wc := runtime.AddWorker(conn)
		go wc.Handle()
	}
}
