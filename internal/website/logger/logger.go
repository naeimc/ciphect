package logger

import "github.com/naeimc/ciphect/logging"

var Logger *logging.Logger

func Printf(level logging.Level, format string, a ...any) {
	if Logger != nil {
		Logger.Printf(level, format, a...)
	}
}

func Print(level logging.Level, a any) {
	if Logger != nil {
		Logger.Print(level, a)
	}
}
