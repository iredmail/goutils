package logger

import (
	"context"
	"fmt"
	"log/slog"
	"log/syslog"
)

// SyslogHandler Implement interface slog.Handler
type SyslogHandler struct {
	writer *syslog.Writer
	attrs  []slog.Attr
}

func newSyslogHandler(server, tag string, priority syslog.Priority) (*SyslogHandler, error) {
	writer, err := syslog.Dial("tcp", server, priority|syslog.LOG_MAIL, tag)
	if err != nil {
		return nil, err
	}

	return &SyslogHandler{writer: writer}, nil
}

func (h *SyslogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (h *SyslogHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})
	if len(attrs) > 0 {
		msg += " " + fmt.Sprint(attrs)
	}

	var err error
	switch r.Level {
	case slog.LevelDebug:
		err = h.writer.Debug(msg)
	case slog.LevelWarn:
		err = h.writer.Warning(msg)
	case slog.LevelError:
		err = h.writer.Err(msg)
	default:
		err = h.writer.Info(msg) // default Info
	}

	return err
}

func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(newHandler.attrs, attrs...)

	return &newHandler
}

func (h *SyslogHandler) WithGroup(name string) slog.Handler {
	return h
}
