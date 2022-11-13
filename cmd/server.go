package cmd

import (
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
	runtime := server.NewRuntime(m, sc.Port)

	err := runtime.Listen()
	if err != nil {
		return err
	}
	defer runtime.Close()

	go runtime.HandleConnections()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

	return nil
}
