package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gideonw/peltr/pkg/proto"
)

func (r *runtime) HandleJob(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var j proto.Job

	err = json.Unmarshal(b, &j)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	r.Jobs = append(r.Jobs, j)
}

func (r *runtime) HandleListJobs(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := json.Marshal(r.Jobs)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = rw.Write(b)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (r *runtime) HandleListWorkers(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := json.Marshal(r.Workers)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = rw.Write(b)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
