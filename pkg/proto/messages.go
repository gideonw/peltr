package proto

import (
	"bytes"
	"encoding/gob"
)

type Identify struct {
	ID       string
	Capacity uint
}

type Assign struct {
	Jobs []Job `json:"jobs"`
}

type Status struct {
	// JobQueue of accepted jobs
	JobQueue []Job
	// ActiveJobs are the jobs currently being worked
	ActiveJobs []Job
	// [JobID]: [StatusCode]count
	Results map[string]map[int]int
}

func ParseIdentify(b []byte) (Identify, error) {
	var ret Identify
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ret)
	return ret, err
}

func ParseAssign(b []byte) (Assign, error) {
	var ret Assign
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ret)
	return ret, err
}

func ParseStatus(b []byte) (Status, error) {
	var ret Status
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ret)
	return ret, err
}
