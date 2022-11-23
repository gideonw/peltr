package worker

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gideonw/peltr/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type WorkerRuntime interface {
	Connect() error
	Handle()
	Close()
}

type workerRuntime struct {
	metrics Metrics
	log     zerolog.Logger
	host    string
	port    int

	State    string
	Capacity uint
	ID       string
	conn     net.Conn
}

func NewRuntime(m Metrics, logger zerolog.Logger, host string, port int) WorkerRuntime {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return &workerRuntime{
		log:      logger,
		metrics:  m,
		host:     host,
		port:     port,
		State:    "new",
		Capacity: 10,
		ID:       id.String(),
		conn:     nil,
	}
}

func (wr *workerRuntime) Connect() error {
	retryCount := 3

	var err error

	for retryCount > 0 && wr.conn == nil {
		wr.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", wr.host, wr.port))
		if err != nil {
			wr.log.Error().Err(err)
			retryCount -= 1
		}
		time.Sleep(2 * time.Second)
	}
	wr.log.Info().Str("addr", wr.conn.RemoteAddr().String()).Msg("worker connected")

	return nil
}

func (wr *workerRuntime) Close() {
	wr.conn.Close()
}

func (wr *workerRuntime) Handle() {
	start := time.Now()

	for {
		b := make([]byte, 256)

		if time.Now().After(start.Add(1 * time.Second)) {
			start = time.Now()
			n, err := wr.conn.Read(b)
			if err == io.EOF {
				wr.log.Error().Err(err).Msg("disconnected")
				break
			} else if err != nil {
				wr.log.Error().Err(err).Msg("error reading from conn")
			}
			wr.log.Debug().Int("bytes", n).Msg("read from conn")

			command, msg := proto.ChompCommand(b)
			wr.log.Debug().Str("cmd", string(command)).Msg("chomp")
			wr.log.Trace().Bytes("msg", msg).Msg("chomp")
			switch string(command) {
			case string(proto.CommandHello):
				n, err := wr.conn.Write(proto.MakeMessageStruct(proto.CommandHello, proto.Identify{
					ID:       wr.ID,
					Capacity: wr.Capacity,
				}))
				if err != nil {
					wr.log.Error().Str("type", "hello").Err(err).Msg("error sending")
				}
				wr.log.Info().Str("type", "hello").Int("bytes", n).Msg("wrote")
			case string(proto.CommandPing):
				n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandPong, nil))
				if err != nil {
					wr.log.Error().Str("type", "pong").Err(err).Msg("error sending")
				}
				wr.log.Info().Str("type", "pong").Int("bytes", n).Msg("wrote")

			case string(proto.CommandAssign):
				// Parse Job
				job, err := proto.ParseAssign(msg)
				if err != nil {
					wr.log.Error().Str("type", "assign").Err(err).Msg("error parsing message")
				}
				go HandleJob(wr.log, job.Jobs[0])

				n, err := wr.conn.Write(proto.MakeMessageByte(proto.CommandWorking, nil))
				if err != nil {
					wr.log.Error().Str("type", "working").Err(err).Msg("error sending")
				}
				wr.log.Info().Str("type", "working").Int("bytes", n).Msg("wrote")

			default:
				wr.log.Error().Msgf("Unknown command '%s','%s'\n", command, msg)
				wr.log.Error().Msgf("wat '%v' '%v' '%v'", b, string(b), "hello")
			}
		}
	}
}
