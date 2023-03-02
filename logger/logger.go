package logger

import (
	"errors"
	"fmt"
	"log/syslog"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
	"golang.org/x/exp/slices"
)

type Config struct {
	LogTarget         string `json:"log_target"`
	LogLevel          string `json:"log_level"`
	LogSyslogServer   string `json:"log_syslog_server"`
	LogSyslogTag      string `json:"log_syslog_tag"`
	LogFile           string `json:"log_file"`
	LogMaxSize        int    `json:"log_max_size"`
	LogRotateInterval string `json:"log_rotate_interval"`
	// Buffer size defaults to (8 * 1024).
	// Write to log file immediately if size is 0.
	LogBufferSize int    `json:"log_buffer_size"`
	LogMaxBackups int    `json:"log_max_backups"`
	LogTimeFormat string `json:"log_time_format"`
	LogCompress   bool   `json:"log_compress"` // compress rotated log file
}

func New(c *Config) (logger slog.SLogger, err error) {
	var logTemplate string
	var syslogLevel syslog.Priority

	level := slog.LevelByName(c.LogLevel)
	if level == slog.DebugLevel {
		syslogLevel = syslog.LOG_DEBUG
		// 当 log level 为 debug 时开启 caller，方便快速定位打印日志位置
		// logTemplate = "{{datetime}} {{level}} {{message}} [{{caller}}]\n"
	} else {
		syslogLevel = syslog.LOG_INFO
		// logTemplate = "{{datetime}} {{level}} {{message}}\n"
	}
	logTemplate = "{{datetime}} {{level}} {{message}}\n"

	// custom log format
	logFormatter := slog.NewTextFormatter(logTemplate)
	logFormatter.TimeFormat = c.LogTimeFormat
	l := slog.NewStdLogger().Configure(func(sl *slog.SugaredLogger) {
		sl.ReportCaller = true
		sl.CallerSkip = 6
	})
	l.Config(func(sl *slog.SugaredLogger) {
		f := sl.Formatter.(*slog.TextFormatter)
		f.TimeFormat = c.LogTimeFormat
		f.SetTemplate(logTemplate)
		f.FullDisplay = true
		f.EnableColor = false
	})

	switch c.LogTarget {
	// case "stdout":
	// 	h := handler.NewConsoleHandler([]slog.Level{level})
	// 	h.SetFormatter(logFormatter)
	// 	l.AddHandler(h)
	case "file":
		if c.LogMaxSize > 0 {
			h, err := handlerRotateFile(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}

		if c.LogRotateInterval != "" {
			h, err := handlerRotateTime(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}
	case "syslog":
		if len(c.LogSyslogServer) == 0 {
			// Use local syslog socket by default.
			c.LogSyslogServer = "/dev/log"
		}

		if strings.HasPrefix(c.LogSyslogServer, "/") {
			h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, c.LogSyslogTag)
			if err != nil {
				return nil, err
			}
			h.SetFormatter(logFormatter)
			l.AddHandler(h)

			break
		}

		w, err := syslog.Dial("tcp", c.LogSyslogServer, syslogLevel|syslog.LOG_MAIL, c.LogSyslogTag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, c.LogBufferSize, level)
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	}

	l.Level = level
	l.DoNothingOnPanicFatal()
	logger = l

	return
}

func handlerRotateFile(c *Config) (*handler.SyncCloseHandler, error) {
	return handler.NewSizeRotateFileHandler(
		c.LogFile,
		c.LogMaxSize,
		handler.WithBuffSize(c.LogBufferSize),
		handler.WithCompress(c.LogCompress),
		handler.WithLogLevels(slog.AllLevels),
	)
}

// handlerRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func handlerRotateTime(c *Config) (*handler.SyncCloseHandler, error) {
	if len(c.LogRotateInterval) == 0 {
		return nil, errors.New("empty rotate interval")
	}

	lastChar := c.LogRotateInterval[len(c.LogRotateInterval)-1]
	lowerLastChar := strings.ToLower(string(lastChar))
	if !slices.Contains([]string{"w", "d", "h", "m", "s"}, lowerLastChar) {
		return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
	}

	rotateIntervalDuration, err := time.ParseDuration(c.LogRotateInterval)
	if err != nil {
		return nil, err
	}

	return handler.NewTimeRotateFileHandler(
		c.LogFile,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithBuffSize(c.LogBufferSize),
		handler.WithCompress(c.LogCompress),
		handler.WithLogLevels(slog.AllLevels),
	)
}
