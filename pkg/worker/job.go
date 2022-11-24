package worker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/rs/zerolog"
)

type JobWorker struct {
	log zerolog.Logger

	Done    bool
	Results map[int]int
	Job     proto.Job
}

func NewJobWorker(log zerolog.Logger, j proto.Job) JobWorker {
	return JobWorker{
		log:     log,
		Done:    false,
		Results: make(map[int]int),
		Job:     j,
	}
}

func (jw *JobWorker) HandleJob() {
	// done := make(chan bool)
	ticker := time.NewTicker(10 * time.Millisecond)

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
