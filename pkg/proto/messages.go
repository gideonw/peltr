/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
)

type MessageType uint8

const (
	MessageTypeHello MessageType = iota
	MessageTypeIdentify
	MessageTypeAssign
	MessageTypeAlive
	MessageTypeStatus
	MessageTypeAccept
)

// The Message type wraps all messages sent between workers and servers
type Message struct {
	Type MessageType
	Data []byte
}

// Read a Message from an io.Reader
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

// Write a Message to an io.Writer.
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

type ProtocolObject interface {
	Encode() (Message, error)
	Decode(Message) error
}

type (
	Identify struct {
		ID       string
		Capacity uint
	}

	Assign struct {
		Jobs []Job `json:"jobs"`
	}

	Status struct {
		// JobQueue of accepted jobs
		JobQueue []Job
		// ActiveJobs are the jobs currently being worked
		ActiveJobs []Job
		// [JobID]: [StatusCode]count
		Results map[string]map[int]int
	}
)

func encode(obj interface{}) ([]byte, error) {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}

func (id *Identify) Encode() (Message, error) {
	data, err := encode(id)
	if err != nil {
		return Message{}, err
	}
	return Message{Type: MessageTypeIdentify, Data: data}, nil
}

func (id *Identify) Decode(m Message) error {
	buf := bytes.NewBuffer(m.Data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(id)
	return err
}

func (assign *Assign) Encode() (Message, error) {
	data, err := encode(assign)
	if err != nil {
		return Message{}, err
	}
	return Message{Type: MessageTypeAssign, Data: data}, nil
}

func (assign *Assign) Decode(m Message) error {
	buf := bytes.NewBuffer(m.Data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(assign)
	return err
}

func (status *Status) Encode() (Message, error) {
	data, err := encode(status)
	if err != nil {
		return Message{}, err
	}
	return Message{Type: MessageTypeAssign, Data: data}, nil
}

func (status *Status) Decode(m Message) error {
	buf := bytes.NewBuffer(m.Data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(status)
	return err
}
