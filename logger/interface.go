package logger

import (
	"io"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type LoggerWithWriter interface {
	Logger
	io.Writer
}
