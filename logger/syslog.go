package logger

import (
	"context"
	"log/slog"
	"log/syslog"
	"strings"
)

// SyslogHandler Implement interface slog.Handler
type SyslogHandler struct {
	writer *syslog.Writer
	level  slog.Level
}

func newSyslogHandler(server, tag string, level slog.Level, priority syslog.Priority) (*SyslogHandler, error) {
	network := "tcp"

	if server == "" || strings.HasPrefix(server, "/") {
		network = "unixgram"
	}

	writer, err := syslog.Dial(network, server, priority|syslog.LOG_MAIL, tag)
	if err != nil {
		return nil, err
	}

	return &SyslogHandler{writer: writer, level: level}, nil
}

func (h *SyslogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SyslogHandler) Handle(_ context.Context, r slog.Record) error {
	var err error
	switch r.Level {
	case slog.LevelDebug:
		err = h.writer.Debug(r.Message)
	case slog.LevelWarn:
		err = h.writer.Warning(r.Message)
	case slog.LevelError:
		err = h.writer.Err(r.Message)
	default:
		err = h.writer.Info(r.Message) // default Info
	}

	return err
}

func (h *SyslogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *SyslogHandler) WithGroup(_ string) slog.Handler {
	return h
}
