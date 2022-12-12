/*
 * Copyright (c) 2022, Gideon Williams <gideon@gideonw.com>
 *
 * SPDX-License-Identifier: BSD-2-Clause
 */

package proto

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMessageReadWrite(t *testing.T) {
	var buf bytes.Buffer
	message := Message{Type: MessageTypeIdentify, Data: []byte("Some string of bytes")}
	err := message.Write(&buf)
	if err != nil {
		t.Fail()
	}

	var message2 Message
	err = message2.Read(&buf)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(message, message2) {
		t.Fail()
	}
}

func TestIdentifyEncodeDecode(t *testing.T) {
	identify := Identify{
		ID:       "foo",
		Capacity: 4,
	}
	message, err := identify.Encode()
	if err != nil {
		t.Fail()
	}
	if message.Type != MessageTypeIdentify {
		t.Fail()
	}
	var identify2 Identify
	err = identify2.Decode(message)
	if err != nil {
		t.Fail()
	}
	if !reflect.DeepEqual(identify, identify2) {
		t.Fail()
	}
}
