package proto

import (
	"bytes"
	"encoding/gob"
)

var (
	COMMAND_LEN        = 8
	MESSAGE_TERMINATOR = []byte("\n\r")
)

func MakeMessageStruct(command []byte, data interface{}) []byte {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	enc.Encode(data)
	return MakeMessageByte(command, b.Bytes())
}

func MakeMessageString(command []byte, data string) []byte {
	return MakeMessageByte(command, []byte(data))
}

func MakeMessageByte(command, data []byte) []byte {
	b := make([]byte, COMMAND_LEN+len(MESSAGE_TERMINATOR)+len(data))
	copy(b, command[:COMMAND_LEN])
	copy(b[COMMAND_LEN:], data)
	copy(b[COMMAND_LEN+len(data):], MESSAGE_TERMINATOR)

	return b
}
