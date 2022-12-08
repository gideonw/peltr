package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
)

type MessageType uint8

// The Message type wraps all messages sent between workers and servers
type Message struct {
	Type MessageType
	Data []byte
}

func (m *Message) Read(reader io.Reader) error {
	lengthPrefix := make([]byte, 4)
	err := binary.Read(reader, binary.LittleEndian, &lengthPrefix)
	if err != nil {
		return err
	}

	messageLength := binary.LittleEndian.Uint32(lengthPrefix)
	messageBuffer := make([]byte, messageLength)
	_, err = io.ReadFull(reader, messageBuffer)
	if err != nil {
		return err
	}

	m.Type = MessageType(messageBuffer[0])
	m.Data = messageBuffer[1:]

	return nil
}

func (m *Message) Write(writer io.Writer) error {
	// 1 byte for MessageType, and 4 bytes for message length
	b := make([]byte, len(m.Data)+1+4)
	binary.LittleEndian.PutUint32(b, uint32(len(m.Data)+1))
	b[4] = byte(m.Type)
	copy(b[5:], m.Data)

	_, err := writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

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
