package peltr

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func initConfig(configFile string) {
	log := viper.Get("logger").(zerolog.Logger)

	// config Read
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath("/etc/peltr")
	viper.AddConfigPath("/usr/local/etc/peltr")
	viper.AddConfigPath("$HOME/.peltr")
	viper.AddConfigPath(".")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Info().Msg("No config file found, using defaults as a base")
	} else if err != nil {
		log.Error().Msg("Error loading config file")
	}

	log.Info().Str("file", viper.ConfigFileUsed()).Msg("loaded config from file")
}

func initLogLevel() {
	level := viper.GetInt("peltr.verbose")
	switch clamp(2, level) {
	case 2:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func initLogging() {
	var writer io.Writer

	writer = os.Stderr
	if viper.GetBool("peltr.local") {
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
}

func traceConfig() {
	log := viper.Get("logger").(zerolog.Logger)

	for _, v := range viper.AllKeys() {
		if v == "logger" {
			continue
		}
		log.Trace().Msgf("%s=%v", v, viper.Get(v))
	}
}

func clamp(clamp, a int) int {
	if a >= clamp {
		return clamp
	}
	return a
}
