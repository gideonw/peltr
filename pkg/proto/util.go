package proto

var (
	COMMAND_LEN        = 8
	MESSAGE_TERMINATOR = []byte("\n\r")
	CommandHello       = MakeCommand([]byte("hello"))
	CommandPing        = MakeCommand([]byte("ping"))
	CommandPong        = MakeCommand([]byte("pong"))
)

func MakeCommand(cmd []byte) []byte {
	b := make([]byte, COMMAND_LEN)
	copy(b, []byte(cmd))
	return b
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

func ChompCommand(d []byte) (command, data []byte) {
	if len(d) == COMMAND_LEN+len(MESSAGE_TERMINATOR) {
		return d[0:COMMAND_LEN], d[0:0]
	}

	return d[0:COMMAND_LEN], d[COMMAND_LEN : len(d)-2]
}
