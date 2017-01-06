package gorm

import (
	"fmt"

	log "github.com/henrylee2cn/lessgo/logs"
	"github.com/henrylee2cn/lessgo/logs/logs"
)

type ILogger struct {
	BeeLogger *logs.BeeLogger
}

func newILogger(channelLen int64, l int, filename string) *ILogger {
	tl := logs.NewLogger(channelLen)
	tl.SetLogFuncCallDepth(3)
	tl.AddAdapter("console", "")
	tl.AddAdapter("file", `{"filename":"`+LOG_FOLDER+filename+`.gorm.log"}`)
	tl.SetLevel(log.ExchangeLevel(l))
	return &ILogger{
		BeeLogger: tl,
	}
}

func (i *ILogger) Print(v ...interface{}) {
	i.BeeLogger.Info(fmt.Sprintln(v...))
	return
}
