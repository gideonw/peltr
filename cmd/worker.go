package cmd

import (
	"net/http"

	"github.com/gideonw/peltr/pkg/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type WorkerCommand struct {
	Port     int `short:"p" long:"port" description:"Port to connect to for worker instructions" default:"8001"`
	DataPort int `short:"d" long:"data-port" description:"Port to send job results to" default:"8002"`
}

var WorkerCmd WorkerCommand

func (sc *WorkerCommand) Execute(args []string) error {
	m := worker.NewMetricsStore()
	runtime := worker.NewRuntime(m, sc.Port)

	err := runtime.Connect()
	if err != nil {
		return err
	}
	defer runtime.Close()

	go runtime.Handle()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2113", nil)

	return nil
}
