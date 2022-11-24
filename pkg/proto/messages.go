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

type Update struct {
	// ID job id for the stats
	ID      string
	Results map[int]int
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

func ParseUpdate(b []byte) ([]Update, error) {
	var ret []Update
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&ret)
	return ret, err
}
