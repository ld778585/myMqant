package log

import (
	beegolog "github.com/liangdas/mqant/log/beego"
	mqanttools "github.com/liangdas/mqant/utils"
)

var bi *beegolog.BeeLogger

// InitBI 初始化BI日志
func InitBI(debug bool, ProcessID string, Logdir string, settings map[string]interface{}) {
	bi = NewBeegoLogger(debug, ProcessID, Logdir, settings)
}

// BiBeego BiBeego
func BiBeego() *beegolog.BeeLogger {
	return bi
}

// CreateRootTrace CreateRootTrace
func CreateRootTrace() TraceSpan {
	return &TraceSpanImp{
		Trace: mqanttools.GenerateID().String(),
		Span:  mqanttools.GenerateID().String(),
	}
}

// CreateTrace CreateTrace
func CreateTrace(trace, span string) TraceSpan {
	return &TraceSpanImp{
		Trace: trace,
		Span:  span,
	}
}

// BiReport BiReport
func BiReport(msg string) {
	//gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
	l := BiBeego()
	if l != nil {
		l.BiReport(msg)
	}
}
