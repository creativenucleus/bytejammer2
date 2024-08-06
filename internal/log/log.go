package log

import (
	"fmt"

	"github.com/creativenucleus/bytejammer2/internal/message"
)

type Log struct {
	message.MsgSender
}

var GlobalLog = Log{}

type MsgLogData struct {
	Level   string
	Message string
}

func (l *Log) Send(msg *message.Msg) {
	switch msg.Type {
	case message.MsgTypeLog:
		data, ok := msg.Data.(MsgLogData)
		if !ok {
			fmt.Printf("UNHANDLED")
			return
		}
		fmt.Printf("LOG: %s %s\n", data.Level, data.Message)
	default:
		fmt.Printf("LOG: Unhandled message type [%s] - ignored\n", msg.Type)
	}
}
