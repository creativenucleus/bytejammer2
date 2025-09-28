package log

import (
	"fmt"
	"time"
)

type Log struct {
}

var GlobalLog = Log{}

type MsgLogData struct {
	Level   string
	Message string
}

func (l *Log) Log(level string, message string) {
	now := time.Now()
	fmt.Printf("LOG: (%s) %s\n", now.Format(time.RFC3339), message)
}
