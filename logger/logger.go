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

func New(c *Config) (logger slog.SLogger, err error) {
	var logTemplate string
	var syslogLevel syslog.Priority

	level := slog.LevelByName(c.logLevel)
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
	logFormatter.TimeFormat = c.logTimeFormat
	l := slog.NewStdLogger().Configure(func(sl *slog.SugaredLogger) {
		sl.ReportCaller = true
		sl.CallerSkip = 6
	})
	l.Config(func(sl *slog.SugaredLogger) {
		f := sl.Formatter.(*slog.TextFormatter)
		f.TimeFormat = c.logTimeFormat
		f.SetTemplate(logTemplate)
		f.FullDisplay = true
		f.EnableColor = false
	})

	switch c.logTarget {
	case "stdout":
		h := handler.NewConsoleHandler([]slog.Level{level})
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	case "file":
		if c.logMaxSize > 0 {
			h, err := handlerRotateFile(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}

		if c.logRotateInterval != "" {
			h, err := handlerRotateTime(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}
	case "syslog":
		if len(c.logSyslogServer) == 0 {
			// Use local syslog socket by default.
			c.logSyslogServer = "/dev/log"
		}

		if strings.HasPrefix(c.logSyslogServer, "/") {
			h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, c.logSyslogTag)
			if err != nil {
				return nil, err
			}
			h.SetFormatter(logFormatter)
			l.AddHandler(h)

			break
		}

		w, err := syslog.Dial("tcp", c.logSyslogServer, syslogLevel|syslog.LOG_MAIL, c.logSyslogTag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, c.logBufferSize, level)
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
		c.logFile,
		c.logMaxSize,
		handler.WithBuffSize(c.logBufferSize),
		handler.WithCompress(c.logCompress),
		handler.WithLogLevels(slog.AllLevels),
	)
}

// handlerRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func handlerRotateTime(c *Config) (*handler.SyncCloseHandler, error) {
	if len(c.logRotateInterval) == 0 {
		return nil, errors.New("empty rotate interval")
	}

	lastChar := c.logRotateInterval[len(c.logRotateInterval)-1]
	lowerLastChar := strings.ToLower(string(lastChar))
	if !slices.Contains([]string{"w", "d", "h", "m", "s"}, lowerLastChar) {
		return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
	}

	rotateIntervalDuration, err := time.ParseDuration(c.logRotateInterval)
	if err != nil {
		return nil, err
	}

	return handler.NewTimeRotateFileHandler(
		c.logFile,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithBuffSize(c.logBufferSize),
		handler.WithCompress(c.logCompress),
		handler.WithLogLevels(slog.AllLevels),
	)
}
