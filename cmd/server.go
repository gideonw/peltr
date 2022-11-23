package cmd

import (
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerCommand struct {
	Port        int `short:"p" long:"port" description:"Port to listen for workers" default:"8000"`
	MetricsPort int `short:"m" long:"metrics-port" description:"Port to listen for job results" default:"8010"`
}

var ServerCmd ServerCommand

func (sc *ServerCommand) Execute(args []string) error {
	logger := configLog("server")

	m := server.NewMetricsStore()
	runtime := server.NewRuntime(m, logger, sc.Port)

	err := runtime.Listen()
	if err != nil {
		return err
	}
	defer runtime.Close()

	go runtime.HandleConnections()
	go runtime.ControlLoop()

	http.HandleFunc("/workers", runtime.HandleListWorkers)
	http.HandleFunc("/jobs", runtime.HandleListJobs)
	http.HandleFunc("/job", runtime.HandleJob)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", sc.MetricsPort), nil)

	return nil
}
