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
		runtime := worker.NewRuntime(m, log, viper.GetString("host"), viper.GetInt("port"))

		err := runtime.Connect()
		if err != nil {
			log.Fatal().Err(err).Msg("error connecting to server")
			return
		}
		defer runtime.Close()

		go runtime.Handle()

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("prom-http")), nil)
	},
}

func init() {
	// Flags for this command
	Command.Flags().IntP("port", "p", 8000, "Server port to communicate over (-p8000)")
	Command.Flags().StringP("host", "H", "localhost", "Server host to connect to -Hlocalhost")
	Command.Flags().IntP("prom-http", "m", 8010, "Set the port for /metrics is bound to (-m8010)")

	// Bind flags to viper
	viper.BindPFlag("port", Command.Flags().Lookup("port"))
	viper.BindPFlag("host", Command.Flags().Lookup("host"))
	viper.BindPFlag("prom-http", Command.Flags().Lookup("prom-http"))
}
