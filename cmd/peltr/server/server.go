package server

import (
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Command = &cobra.Command{
	Use:   "server",
	Short: "Server that accepts and assigns jobs to workers",

	Run: func(cmd *cobra.Command, args []string) {
		log := viper.Get("logger").(zerolog.Logger)

		m := server.NewMetricsStore()
		runtime := server.NewRuntime(m, log, viper.GetInt("port"))

		err := runtime.Listen()
		if err != nil {
			log.Fatal().Err(err).Msg("Error listening on server port")
			return
		}
		defer runtime.Close()

		go runtime.HandleConnections()
		go runtime.ControlLoop()

		http.HandleFunc("/workers", runtime.HandleListWorkers)
		http.HandleFunc("/jobs", runtime.HandleListJobQueue)
		http.HandleFunc("/job", runtime.HandleJob)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("prom-http")), nil)

	},
}

func init() {
	// Flags for this command
	Command.Flags().IntP("port", "p", 8000, "Database server port for client connections (-p8000)")
	Command.Flags().Int("prom-http", 8010, "Set the port for /metrics is bound to (-m8010)")

	// Bind flags to viper
	viper.BindPFlag("port", Command.Flags().Lookup("port"))
	viper.BindPFlag("prom-http", Command.Flags().Lookup("prom-http"))
}
