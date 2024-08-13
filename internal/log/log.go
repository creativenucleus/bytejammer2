package log

import (
	"fmt"
)

type Log struct {
}

var GlobalLog = Log{}

type MsgLogData struct {
	Level   string
	Message string
}

func (l *Log) Log(level string, message string) {
	fmt.Printf("LOG: %s %s\n", level, message)
}
