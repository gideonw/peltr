/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package peltr

import (
	"os"
	"strings"

	"github.com/gideonw/peltr/cmd/peltr/server"
	"github.com/gideonw/peltr/cmd/peltr/test"
	"github.com/gideonw/peltr/cmd/peltr/worker"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "peltr",
	Short: "peltr is a cloud native load testing and testing tool",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
		initConfig(cmd.Root().PersistentFlags().Lookup("config").Value.String())
		initLogLevel()
		traceConfig()
	},
}

func init() {
	// Configure the common binary options
	rootCmd.PersistentFlags().CountP("verbose", "v", "-v for debug logs (-vv for trace)")
	rootCmd.PersistentFlags().Bool("local", true, "Configures the logger to print readable logs") //TODO: true until we have a config file format
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to the config file (default ./config.toml)")

	// Bind viper config to the root flags
	viper.BindPFlag("peltr.local", rootCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("peltr.verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	// Bind viper flags to ENV variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// Register commands on the root binary command
	rootCmd.AddCommand(server.Command)
	rootCmd.AddCommand(worker.Command)
	rootCmd.AddCommand(test.Command)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("root command failed")
		os.Exit(1)
	}
}
