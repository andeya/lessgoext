package xorm

import (
	"fmt"

	"github.com/go-xorm/core"

	log "github.com/lessgo/lessgo/logs"
	"github.com/lessgo/lessgo/logs/logs"
)

type ILogger struct {
	BeeLogger *logs.BeeLogger
	level     core.LogLevel
	showSQL   bool
}

func newILogger(channelLen int64, l int, filename string) *ILogger {
	tl := logs.NewLogger(channelLen)
	tl.SetLogFuncCallDepth(3)
	tl.AddAdapter("console", "")
	tl.AddAdapter("file", `{"filename":"`+LOG_FOLDER+filename+`.xorm.log"}`)
	tl.SetLevel(log.ExchangeLevel(l))
	return &ILogger{
		BeeLogger: tl,
		level:     level(l),
	}
}

func level(l int) core.LogLevel {
	return core.LogLevel(log.ExchangeLevel(l))
}

func (i *ILogger) Debug(v ...interface{}) {
	i.BeeLogger.Debug(fmt.Sprintln(v...))
	return
}

func (i *ILogger) Debugf(format string, v ...interface{}) {
	i.BeeLogger.Debug(format, v...)
	return
}

func (i *ILogger) Error(v ...interface{}) {
	i.BeeLogger.Error(fmt.Sprintln(v...))
}

func (i *ILogger) Errorf(format string, v ...interface{}) {
	i.BeeLogger.Error(format, v...)
}

func (i *ILogger) Info(v ...interface{}) {
	i.BeeLogger.Info(fmt.Sprintln(v...))
}

func (i *ILogger) Infof(format string, v ...interface{}) {
	i.BeeLogger.Info(format, v...)
}

func (i *ILogger) Warn(v ...interface{}) {
	i.BeeLogger.Warn(fmt.Sprintln(v...))
}
func (i *ILogger) Warnf(format string, v ...interface{}) {
	i.BeeLogger.Warn(format, v...)
}

func (i *ILogger) Level() core.LogLevel {
	return i.level
}

func (i *ILogger) SetLevel(l core.LogLevel) {
	i.level = level(int(l))
	i.BeeLogger.SetLevel(int(i.level))
}

func (i *ILogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		i.showSQL = true
		return
	}
	i.showSQL = show[0]
}

func (i *ILogger) IsShowSQL() bool {
	return i.showSQL
}
