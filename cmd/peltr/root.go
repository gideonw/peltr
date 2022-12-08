package peltr

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/gideonw/peltr/cmd/peltr/server"
	"github.com/gideonw/peltr/cmd/peltr/test"
	"github.com/gideonw/peltr/cmd/peltr/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "peltr",
	Short: "peltr is a cloud native load testing and testing tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger := initLogging()
		for _, v := range viper.AllKeys() {
			logger.Trace().Interface(v, viper.Get(v)).Send()
		}
	},
}

func init() {
	// Configure the common binary options
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().CountP("verbose", "v", "-v for debug logs (-vv for trace)")
	rootCmd.PersistentFlags().Bool("local", true, "Configures the logger to print readable logs") //TODO: true until we have a config file format

	// Bind viper config to the root flags
	viper.BindPFlag("local", rootCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Bind viper flags to ENV variables
	viper.SetEnvPrefix("PELTR")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Register commands on the root binary command
	rootCmd.AddCommand(server.Command)
	rootCmd.AddCommand(worker.Command)
	rootCmd.AddCommand(test.Command)
}

func initConfig() {
	// config Read
}

func initLogging() zerolog.Logger {
	level := viper.GetInt("verbose")
	switch clamp(2, level) {
	case 2:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var writer io.Writer

	writer = os.Stderr
	if viper.GetBool("local") {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(writer).
		With().
		Timestamp().
		Caller().
		Logger()

	viper.Set("logger", logger)
	return logger
}

func clamp(clamp, a int) int {
	if a >= clamp {
		return clamp
	}
	return a
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("root command failed")
		os.Exit(1)
	}
}
