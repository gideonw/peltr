package worker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/rs/zerolog"
)

type JobPool struct {
}

func HandleJob(log zerolog.Logger, j proto.Job) {
	// done := make(chan bool)
	ticker := time.NewTicker(10 * time.Millisecond)

	stats := make(map[int]int)
	count := 0
	errcount := 0
	for range ticker.C {
		if count+errcount >= j.Req {
			ticker.Stop()
			e := log.Debug()
			for k, v := range stats {
				e.Int(fmt.Sprint(k), v)
			}
			e.Msg("job complete")
			break
		}

		// Make request
		code, r, err := makeRequest(j.URL)
		if err != nil {
			log.Error().Err(err).Msg("making request")
			errcount += 1
		}
		stats[code] += r
		count += r
		log.Debug().Int("count", count).Msg("status")
	}
}

func makeRequest(url string) (int, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}

	return resp.StatusCode, 1, nil
}
