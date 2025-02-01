package logger

import (
	"os"
)

type Option func(l *logger)

func WithLevel(level string) Option {
	switch level {
	case "info", "warn", "error", "debug":
		break
	default:
		level = "info"
	}

	return func(l *logger) {
		l.level = level
	}
}

func WithMaxSize(maxSize int) Option {
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // 100 MB
	}

	return func(l *logger) {
		l.maxSize = maxSize
	}
}

func WithRotateInterval(interval string) Option {
	return func(l *logger) {
		l.rotateInterval = interval
	}
}

func WithBufferSize(bufferSize int) Option {
	return func(l *logger) {
		l.bufferSize = bufferSize
	}
}

func WithMaxBackups(maxBackups uint) Option {
	return func(l *logger) {
		l.maxBackups = maxBackups
	}
}

func WithCompress() Option {
	return func(l *logger) {
		l.compress = true
	}
}

func WithTimeFormat(timeFormat string) Option {
	return func(l *logger) {
		l.timeFormat = timeFormat
	}
}

func WithFilePerm(perm os.FileMode) Option {
	return func(l *logger) {
		l.filePerm = perm
	}
}

func WithPrintAfterInitialized() Option {
	return func(l *logger) {
		l.printAfterInitialized = true
	}
}
