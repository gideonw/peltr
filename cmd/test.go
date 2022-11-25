package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/google/uuid"
)

type testCommand struct {
	Duplicates  int    `short:"n" log:"duplicates" default:"1"`
	Req         int    `short:"t" log:"req" default:"100"`
	Rate        int    `short:"r" log:"rate" default:"100"`
	Concurrency int    `short:"c" log:"concurrency" default:"10"`
	Duration    int    `short:"s" log:"duration" default:"10"`
	Host        string `short:"H" log:"host"`
}

var TestCmd testCommand

func (tc *testCommand) Execute(args []string) error {
	log := configLog("test")

	if len(args) == 0 {
		log.Panic().Int("args", len(args)).Msg("Endpoint needed")
		return fmt.Errorf("missing arg: endpoint needed")
	}

	log.Info().
		Int("dup", tc.Duplicates).
		Int("req", tc.Req).
		Int("rate", tc.Rate).
		Int("con", tc.Concurrency).
		Int("dur", tc.Duration).
		Str("host", tc.Host).
		Str("endpoint", args[0]).
		Msg("test settings")

	for i := 0; i < tc.Duplicates; i++ {
		uuid, _ := uuid.NewRandom()
		b, err := json.Marshal(proto.Job{
			ID:          uuid.String(),
			URL:         args[0],
			Req:         tc.Req,
			Concurrency: tc.Concurrency,
			Duration:    tc.Duration,
			Rate:        tc.Rate,
		})
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer(b)
		resp, err := http.Post(tc.Host, "application/json", buf)
		if err != nil {
			log.Error().Err(err).Str("id", uuid.String()).Msg("error making POST")
		}
		log.Info().
			Str("status", resp.Status).
			Int("dup", tc.Duplicates).
			Int("req", tc.Req).
			Int("rate", tc.Rate).
			Int("con", tc.Concurrency).
			Int("dur", tc.Duration).
			Str("host", tc.Host).
			Str("endpoint", args[0]).
			Msg("sending job via POST")
	}

	log.Info().Msg("complete")
	return nil
}
