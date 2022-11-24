package proto

var (
	CommandHello   = MakeCommand([]byte("hello"))
	CommandAssign  = MakeCommand([]byte("assign"))
	CommandWorking = MakeCommand([]byte("working"))
	CommandStatus  = MakeCommand([]byte("status"))
	CommandAccept  = MakeCommand([]byte("accept"))
)

func MakeCommand(cmd []byte) []byte {
	b := make([]byte, COMMAND_LEN)
	copy(b, []byte(cmd))
	return b
}

func ChompCommand(d []byte) (command, data []byte) {
	if len(d) == COMMAND_LEN+len(MESSAGE_TERMINATOR) {
		return d[0:COMMAND_LEN], d[0:0]
	}

	return d[0:COMMAND_LEN], d[COMMAND_LEN : len(d)-2]
}
