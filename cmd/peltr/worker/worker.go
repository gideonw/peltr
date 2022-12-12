package worker

import (
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Command = &cobra.Command{
	Use:   "worker",
	Short: "Worker that processes jobs assigned by the server",

	Run: func(cmd *cobra.Command, args []string) {
		log := viper.Get("logger").(zerolog.Logger)

		m := worker.NewMetricsStore()
		runtime := worker.NewRuntime(m, log, viper.GetString("peltr.host"), viper.GetInt("peltr.port"))

		err := runtime.Connect()
		if err != nil {
			log.Fatal().Err(err).Msg("error connecting to server")
			return
		}
		defer runtime.Close()

		go runtime.Handle()

		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("peltr.prom-http")), nil)
		if err != nil {
			log.Fatal().Err(err).Msg("error starting prom-http")
			return
		}
	},
}

func init() {
	// Flags for this command
	Command.Flags().IntP("concurrency", "j", 100, "Server port to communicate over")
	Command.Flags().StringP("host", "H", "localhost:8000", "Server host to connect to")
	Command.Flags().IntP("prom-http", "m", 8010, "Set the port for /metrics")

	// Bind flags to viper
	viper.BindPFlag("peltr.host", Command.Flags().Lookup("host"))
	viper.BindPFlag("peltr.prom-http", Command.Flags().Lookup("prom-http"))
	viper.BindPFlag("peltr.worker.concurrency", Command.Flags().Lookup("concurrency"))
}
