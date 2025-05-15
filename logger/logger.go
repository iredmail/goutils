package logger

import (
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DeRuina/timberjack"
)

func NewStdoutLogger(opts ...Option) (LoggerWithWriter, error) {
	l := &logger{
		w:     os.Stdout,
		level: "info",
	}

	for _, opt := range opts {
		opt(l)
	}

	level, _, err := l.ParseLevel()
	if err != nil {
		return nil, err
	}

	l.sl = slog.New(&CustomHandler{
		w:          os.Stdout,
		level:      level,
		timeFormat: l.timeFormat,
	})

	return l, nil
}

func NewFileLogger(pth string, opts ...Option) (LoggerWithWriter, error) {
	l := &logger{
		level: "info",
	}

	for _, opt := range opts {
		opt(l)
	}

	tj := &timberjack.Logger{
		Filename:   pth,
		MaxSize:    l.maxSize,
		MaxBackups: int(l.maxBackups),
		Compress:   l.compress,
	}
	if l.maxSize == 0 {
		tj.MaxSize = 300 //MB
	}

	if l.rotateInterval != "" {
		if len(l.rotateInterval) < 2 {
			return nil, fmt.Errorf("invalid rotate interval: %s", l.rotateInterval)
		}

		lastChar := l.rotateInterval[len(l.rotateInterval)-1]
		lowerLastChar := strings.ToLower(string(lastChar))

		switch lowerLastChar {
		case "w", "d":
			// time.ParseDuration() 不支持 w、d，因此需要转换成 h。
			prefix, err := strconv.Atoi(l.rotateInterval[:len(l.rotateInterval)-1])
			if err != nil {
				return nil, err
			}

			if lowerLastChar == "w" {
				l.rotateInterval = fmt.Sprintf("%dh", prefix*7*24)
			} else {
				l.rotateInterval = fmt.Sprintf("%dh", prefix*24)
			}
		case "h", "m", "s":
			break
		default:
			return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
		}

		var err error
		tj.RotationInterval, err = time.ParseDuration(l.rotateInterval)
		if err != nil {
			return nil, err
		}
	}

	level, _, err := l.ParseLevel()
	if err != nil {
		return nil, err
	}

	l.w = tj
	l.sl = slog.New(&CustomHandler{
		w:          tj,
		level:      level,
		timeFormat: l.timeFormat,
	})

	return l, nil
}

func NewSyslogLogger(server, tag string, options ...Option) (LoggerWithWriter, error) {
	l := &logger{
		level: "info",
	}

	for _, opt := range options {
		opt(l)
	}

	level, priority, err := l.ParseLevel()
	if err != nil {
		return nil, err
	}

	h, err := newSyslogHandler(server, tag, level, priority)
	if err != nil {
		return nil, err
	}

	l.w = h.writer
	l.sl = slog.New(h)

	return l, nil
}

type logger struct {
	sl *slog.Logger
	w  io.Writer

	level          string // log level: info, warn, error, debug
	maxSize        int
	rotateInterval string // rotate interval. e.g. `12h` (12 hours), `1d` (1 day), `1w` (1 week), `1m` (1 month)
	maxBackups     uint
	timeFormat     string
	compress       bool // compress rotated log file
	filePerm       os.FileMode

	// Buffer size defaults to (8 * 1024).
	// Write to log file immediately if size is 0.
	bufferSize int
}

func (l *logger) ParseLevel() (level slog.Level, priority syslog.Priority, err error) {
	l.level = strings.ToLower(l.level)
	switch l.level {
	case "debug":
		level = slog.LevelDebug
		priority = syslog.LOG_DEBUG
	case "error":
		level = slog.LevelError
		priority = syslog.LOG_ERR
	case "warn":
		level = slog.LevelWarn
		priority = syslog.LOG_WARNING
	case "info":
		level = slog.LevelInfo
		priority = syslog.LOG_INFO
	default:
		err = fmt.Errorf("invalid level: %s", l.level)
	}

	return
}

func (l *logger) formatTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && l.timeFormat != "" {
		t := a.Value.Time()
		a.Value = slog.StringValue(t.Format(l.timeFormat))
	}

	return a
}

//
// 实现 Logger 接口。
//

func (l *logger) Info(msg string, args ...interface{}) {
	l.sl.Info(fmt.Sprintf(msg, args...))
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.sl.Error(fmt.Sprintf(msg, args...))
}

func (l *logger) Warn(msg string, args ...interface{}) {
	l.sl.Warn(fmt.Sprintf(msg, args...))
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.sl.Debug(fmt.Sprintf(msg, args...))
}

// Write Only supported file target mode.
func (l *logger) Write(p []byte) (int, error) {
	return l.w.Write(p)
}
