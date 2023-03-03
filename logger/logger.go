package logger

import (
	"errors"
	"fmt"
	"log/syslog"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
)

func New(c *Config) (logger slog.SLogger, err error) {
	var logTemplate string
	var syslogLevel syslog.Priority

	level := slog.LevelByName(c.level)
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
	logFormatter.TimeFormat = c.timeFormat

	l := slog.NewStdLogger()
	l.ReportCaller = true
	l.CallerSkip = 6

	l.Config(func(sl *slog.SugaredLogger) {
		f := sl.Formatter.(*slog.TextFormatter)
		f.TimeFormat = c.timeFormat
		f.SetTemplate(logTemplate)
		f.FullDisplay = true
		f.EnableColor = false
	})

	switch c.target {
	case "stdout":
		h := handler.NewConsoleHandler([]slog.Level{level})
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	case "file":
		if c.maxSize > 0 {
			h, err := handlerRotateFile(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}

		if c.rotateInterval != "" {
			h, err := handlerRotateTime(c)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}
	case "syslog":
		if len(c.syslogServer) == 0 {
			// Use local syslog socket by default.
			c.syslogServer = "/dev/log"
		}

		if strings.HasPrefix(c.syslogServer, "/") {
			h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, c.syslogTag)
			if err != nil {
				return nil, err
			}
			h.SetFormatter(logFormatter)
			l.AddHandler(h)

			break
		}

		w, err := syslog.Dial("tcp", c.syslogServer, syslogLevel|syslog.LOG_MAIL, c.syslogTag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, c.bufferSize, level)
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
		c.maxSize,
		handler.WithBuffSize(c.bufferSize),
		handler.WithBackupNum(c.maxBackups),
		handler.WithCompress(c.compress),
		handler.WithLogLevels(slog.AllLevels),
	)
}

// handlerRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func handlerRotateTime(c *Config) (*handler.SyncCloseHandler, error) {
	if len(c.rotateInterval) == 0 {
		return nil, errors.New("empty rotate interval")
	}

	if len(c.rotateInterval) < 2 {
		return nil, fmt.Errorf("invalid rotate interval: %s", c.rotateInterval)
	}

	lastChar := c.rotateInterval[len(c.rotateInterval)-1]
	lowerLastChar := strings.ToLower(string(lastChar))
	switch lowerLastChar {
	case "w", "d":
		prefix, err := strconv.Atoi(c.rotateInterval[:len(c.rotateInterval)-1])
		if err != nil {
			return nil, err
		}

		if lowerLastChar == "w" {
			c.rotateInterval = fmt.Sprintf("%dh", prefix*7*24)
		} else {
			c.rotateInterval = fmt.Sprintf("%dh", prefix*24)
		}
	case "h", "m", "s":
		break
	default:
		return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
	}

	rotateIntervalDuration, err := time.ParseDuration(c.rotateInterval)
	if err != nil {
		return nil, err
	}

	return handler.NewTimeRotateFileHandler(
		c.logFile,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithBuffSize(c.bufferSize),
		handler.WithBackupNum(c.maxBackups),
		handler.WithCompress(c.compress),
		handler.WithLogLevels(slog.AllLevels),
	)
}
