package proto_test

import (
	"bytes"
	"testing"

	"github.com/gideonw/peltr/pkg/proto"
)

func TestChompCommand(t *testing.T) {
	buf := make([]byte, 256)
	testMessage := proto.MakeMessageString(proto.CommandHello, "1290w0wefj0w-w-fwfn-,100")
	copy(buf, testMessage)

	a, b := proto.ChompCommand(testMessage)
	if string(a) != string(proto.CommandHello) {
		t.Errorf("'%s' != hello", a)
	}
	if string(b) != "1290w0wefj0w-w-fwfn-,100" {
		t.Errorf("'%s' != 1290w0wefj0w-w-fwfn-,100", b)
	}

	buf = make([]byte, 256)
	testMessage = proto.MakeMessageString(proto.CommandHello, "")
	copy(buf, testMessage)

	a, b = proto.ChompCommand(testMessage)
	if string(a) != string(proto.CommandHello) {
		t.Errorf("'%s' != hello", a)
	}
	if string(b) != "" {
		t.Errorf("'%s' != ''", b)
	}

	buf = make([]byte, 256)
	testMessage = proto.MakeMessageString(proto.CommandPong, "")
	copy(buf, testMessage)

	a, b = proto.ChompCommand(testMessage)
	if string(a) != string(proto.CommandPong) {
		t.Errorf("'%s' != pong", a)
	}
	if string(b) != "" {
		t.Errorf("'%s' != ''", b)
	}

	buf = make([]byte, 256)
	testMessage = proto.MakeMessageString(proto.CommandPing, "")
	copy(buf, testMessage)

	a, b = proto.ChompCommand(testMessage)
	if string(a) != string(proto.CommandPing) {
		t.Errorf("'%s' != ping", a)
	}
	if string(b) != "" {
		t.Errorf("'%s' != ''", b)
	}

}

func TestMakeMessageByte(t *testing.T) {
	testData := []byte("")

	msg := proto.MakeMessageByte(proto.CommandHello, testData)
	cmp := []byte{'h', 'e', 'l', 'l', 'o', 0, 0, 0, '\n', '\r'}
	if bytes.Compare(msg, cmp) != 0 {
		t.Errorf("'%v'!='%v'", msg, cmp)
	}

	testData = []byte("idaa,10")

	msg = proto.MakeMessageByte(proto.CommandHello, testData)
	cmp = []byte{'h', 'e', 'l', 'l', 'o', 0, 0, 0, 'i', 'd', 'a', 'a', ',', '1', '0', '\n', '\r'}
	if bytes.Compare(msg, cmp) != 0 {
		t.Errorf("'%v'!=\n'%v'", msg, cmp)
	}
}
