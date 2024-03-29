// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package log 日志初始化
package log

import (
	beegolog "github.com/liangdas/mqant/log/beego"
)

var beego *beegolog.BeeLogger


// InitLog 初始化日志
func InitLog(debug bool, ProcessID string, Logdir string, settings map[string]interface{}) {
	beego = NewBeegoLogger(debug, ProcessID, Logdir, settings)
}

// LogBeego LogBeego
func LogBeego() *beegolog.BeeLogger {
	if beego == nil {
		beego = beegolog.NewLogger()
	}
	return beego
}

// Debug Debug
func Debug(format string, a ...interface{}) {
	LogBeego().Debug(nil, format, a...)
}

// Info Info
func Info(format string, a ...interface{}) {
	LogBeego().Info(nil, format, a...)
}

// Error Error
func Error(format string, a ...interface{}) {
	LogBeego().Error(nil, format, a...)
}

// Warning Warning
func Warning(format string, a ...interface{}) {
	LogBeego().Warning(nil, format, a...)
}


/*-------------------------应用层不使用---------------------------*/

// TDebug TDebug
func TDebug(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Debug(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Debug(nil, format, a...)
	}
}

// TInfo TInfo
func TInfo(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Info(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Info(nil, format, a...)
	}
}

// TError TError
func TError(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Error(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Error(nil, format, a...)
	}
}

// TWarning TWarning
func TWarning(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Warning(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Warning(nil, format, a...)
	}
}

// Close Close
func Close() {
	LogBeego().Close()
}
