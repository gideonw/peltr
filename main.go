package main

import (
	"encoding/gob"

	"github.com/gideonw/peltr/cmd/peltr"
	"github.com/gideonw/peltr/pkg/proto"
)

func init() {
	// init gob wire types
	gob.RegisterName("Assign", &proto.Assign{})
	gob.RegisterName("Job", &proto.Job{})
	gob.RegisterName("Update", &[]proto.Status{})
}

func main() {
	peltr.Execute()
}
