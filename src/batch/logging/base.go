package logging

import "log"

type Logger struct {
	Log log.Logger
}

// 信息打印
func (lg *Logger) Infoln(v ...interface{}) {
	lg.Log.Println(v...)
}
