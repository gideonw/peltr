package cmd

import (
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type WorkerCommand struct {
	Port        int    `short:"p" long:"port" description:"Port to connect to for worker instructions" default:"8000"`
	MetricsPort int    `short:"m" long:"metrics-port" description:"Port to listen for prom metrics" default:"8010"`
	ServerHost  string `short:"h" long:"host" description:"Server to connect to" default:"localhost"`
}

var WorkerCmd WorkerCommand

func (sc *WorkerCommand) Execute(args []string) error {
	logger := configLog("worker")

	m := worker.NewMetricsStore()
	runtime := worker.NewRuntime(m, logger, sc.ServerHost, sc.Port)

	err := runtime.Connect()
	if err != nil {
		return err
	}
	defer runtime.Close()

	go runtime.Handle()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", sc.MetricsPort), nil)

	return nil
}
