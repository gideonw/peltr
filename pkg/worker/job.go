/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package worker

import (
	"fmt"
	"net/http"
	"time"

	fossil "github.com/dburkart/fossil/api"
	"github.com/gideonw/peltr/pkg/proto"
	"github.com/rs/zerolog"
)

type JobWorker struct {
	log     zerolog.Logger
	metrics Metrics

	Done    bool
	Results map[int]int
	Job     proto.Job
	client  fossil.Client
}

func NewJobWorker(log zerolog.Logger, metrics Metrics, j proto.Job, client fossil.Client) JobWorker {
	return JobWorker{
		log:     log,
		metrics: metrics,
		Done:    false,
		Results: make(map[int]int),
		Job:     j,
		client:  client,
	}
}

func (jw *JobWorker) HandleJob() {
	tick := (time.Duration(jw.Job.Duration) * time.Second) / time.Duration(jw.Job.Req)
	jw.log.Debug().Dur("tick", tick).Msg("status")
	ticker := time.NewTicker(tick)

	count := 0
	errcount := 0
	for range ticker.C {
		if count+errcount >= jw.Job.Req {
			ticker.Stop()
			e := jw.log.Debug()
			for k, v := range jw.Results {
				e.Int(fmt.Sprint(k), v)
			}
			e.Msg("job complete")
			ticker.Stop()
			break
		}

		// Make request
		code, r, dur, err := makeRequest(jw.Job.URL)
		if err != nil {
			jw.log.Error().Err(err).Msg("making request")
			errcount += 1
		}
		jw.Results[code] += r
		count += r
		jw.log.Debug().Int("count", count).Int("code", code).Dur("ms", dur).Msg("status")
		jw.client.Append(fmt.Sprintf("/%s/%d", jw.Job.ID, count), []byte(fmt.Sprint(code)))
	}

	jw.Done = true
}

func makeRequest(url string) (int, int, time.Duration, error) {
	start := time.Now()
	resp, err := http.Get(url)
	dur := time.Now().Sub(start)
	if err != nil {
		return 0, 0, dur, err
	}

	return resp.StatusCode, 1, dur, nil
}
