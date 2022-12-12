/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Command = &cobra.Command{
	Use:   "test",
	Short: "Creates and submits test jobs to the server",

	Run: func(cmd *cobra.Command, args []string) {
		log := viper.Get("logger").(zerolog.Logger)

		if len(args) == 0 {
			log.Panic().Err(fmt.Errorf("missing arg: endpoint needed")).Int("args", len(args)).Msg("Endpoint needed")
			return
		}

		log.Info().
			Int("num", viper.GetInt("number")).
			Int("req", viper.GetInt("req")).
			Int("rate", viper.GetInt("rate")).
			Int("con", viper.GetInt("concurrency")).
			Int("dur", viper.GetInt("duration")).
			Str("host", viper.GetString("host")).
			Str("endpoint", args[0]).
			Msg("test settings")

		for i := 0; i < viper.GetInt("number"); i++ {
			uuid, _ := uuid.NewRandom()
			b, err := json.Marshal(proto.Job{
				ID:          uuid.String(),
				URL:         args[0],
				Req:         viper.GetInt("req"),
				Concurrency: viper.GetInt("concurrency"),
				Duration:    viper.GetInt("duration"),
				Rate:        viper.GetInt("rate"),
			})
			if err != nil {
				log.Error().Err(err).Str("id", uuid.String()).Msg("error making json payload")
				return
			}
			buf := bytes.NewBuffer(b)
			resp, err := http.Post(viper.GetString("host"), "application/json", buf)
			if err != nil {
				log.Error().Err(err).Str("id", uuid.String()).Msg("error making POST")
			}
			log.Info().
				Str("status", resp.Status).
				Int("num", viper.GetInt("number")).
				Int("req", viper.GetInt("req")).
				Int("rate", viper.GetInt("rate")).
				Int("con", viper.GetInt("concurrency")).
				Int("dur", viper.GetInt("duration")).
				Str("host", viper.GetString("host")).
				Str("endpoint", args[0]).
				Msg("sending job via POST")
		}

		log.Info().Msg("complete")
	},
}

func init() {
	// Flags for this command
	Command.Flags().IntP("number", "n", 1, "")
	Command.Flags().IntP("req", "t", 100, "")
	Command.Flags().IntP("rate", "r", 100, "")
	Command.Flags().IntP("concurrency", "c", 10, "")
	Command.Flags().IntP("duration", "s", 10, "")
	Command.Flags().StringP("host", "H", "", "")

	// Bind flags to viper
	viper.BindPFlag("number", Command.Flags().Lookup("number"))
	viper.BindPFlag("req", Command.Flags().Lookup("req"))
	viper.BindPFlag("rate", Command.Flags().Lookup("rate"))
	viper.BindPFlag("concurrency", Command.Flags().Lookup("concurrency"))
	viper.BindPFlag("duration", Command.Flags().Lookup("duration"))
	viper.BindPFlag("host", Command.Flags().Lookup("host"))
}
