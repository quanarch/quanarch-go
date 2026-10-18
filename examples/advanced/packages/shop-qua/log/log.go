// Package log writes the quan's log lines; the quanar writes them under the quan's name.
package log

import shopqua "shop-qua"

func Info(message string, data any)  { shopqua.Core().Log().Info(message, data) }
func Warn(message string, data any)  { shopqua.Core().Log().Warn(message, data) }
func Error(message string, data any) { shopqua.Core().Log().Error(message, data) }
