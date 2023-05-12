package logger

import (
	"fmt"
	"log/syslog"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
)

type logger struct {
	sl *slog.Logger

	target         Target // log target: file, syslog
	level          string // log level: info, warn, error, debug
	syslogServer   string
	syslogTag      string
	logFile        string
	maxSize        int
	rotateInterval string // rotate interval. e.g. `12h` (12 hours), `1d` (1 day), `1w` (1 week), `1m` (1 month)
	maxBackups     uint
	timeFormat     string
	compress       bool // compress rotated log file

	// Buffer size defaults to (8 * 1024).
	// Write to log file immediately if size is 0.
	bufferSize int
}

func New(options ...Option) (v Logger, err error) {
	var logTemplate string
	var syslogLevel syslog.Priority

	l := slog.New()
	l.ReportCaller = true
	l.CallerSkip = 6
	log := logger{sl: l}
	for _, option := range options {
		option(&log)
	}

	level := slog.LevelByName(log.level)
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
	logFormatter.EnableColor = false
	logFormatter.FullDisplay = true
	logFormatter.TimeFormat = log.timeFormat

	switch log.target {
	case TargetStdout:
		h := handler.NewConsoleHandler([]slog.Level{level})
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	case TargetFile:
		if log.maxSize > 0 {
			h, err := handlerRotateFile(log)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}

		if log.rotateInterval != "" {
			h, err := handlerRotateTime(log)
			if err != nil {
				return nil, err
			}

			h.SetFormatter(logFormatter)
			l.AddHandler(h)
		}
	case TargetSyslog:
		if len(log.syslogServer) == 0 {
			// Use local syslog socket by default.
			log.syslogServer = "/dev/log"
		}

		if strings.HasPrefix(log.syslogServer, "/") {
			h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, log.syslogTag)
			if err != nil {
				return nil, err
			}
			h.SetFormatter(logFormatter)
			l.AddHandler(h)

			break
		}

		w, err := syslog.Dial("tcp", log.syslogServer, syslogLevel|syslog.LOG_MAIL, log.syslogTag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, log.bufferSize, level)
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	}

	l.DoNothingOnPanicFatal()
	v = log

	return
}

func handlerRotateFile(log logger) (*handler.SyncCloseHandler, error) {
	return handler.NewSizeRotateFileHandler(
		log.logFile,
		log.maxSize,
		handler.WithLogLevels(parseLevels(log.level)),
		handler.WithBuffSize(log.bufferSize),
		handler.WithBackupNum(log.maxBackups),
		handler.WithCompress(log.compress),
	)
}

// handlerRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func handlerRotateTime(log logger) (*handler.SyncCloseHandler, error) {
	if len(log.rotateInterval) < 2 {
		return nil, fmt.Errorf("invalid rotate interval: %s", log.rotateInterval)
	}

	lastChar := log.rotateInterval[len(log.rotateInterval)-1]
	lowerLastChar := strings.ToLower(string(lastChar))

	switch lowerLastChar {
	case "w", "d":
		// time.ParseDuration() 不支持 w、d，因此需要转换成 h。
		prefix, err := strconv.Atoi(log.rotateInterval[:len(log.rotateInterval)-1])
		if err != nil {
			return nil, err
		}

		if lowerLastChar == "w" {
			log.rotateInterval = fmt.Sprintf("%dh", prefix*7*24)
		} else {
			log.rotateInterval = fmt.Sprintf("%dh", prefix*24)
		}
	case "h", "m", "s":
		break
	default:
		return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
	}

	rotateIntervalDuration, err := time.ParseDuration(log.rotateInterval)
	if err != nil {
		return nil, err
	}

	return handler.NewTimeRotateFileHandler(
		log.logFile,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithLogLevels(parseLevels(log.level)),
		handler.WithBuffSize(log.bufferSize),
		handler.WithBackupNum(log.maxBackups),
		handler.WithCompress(log.compress),
	)
}

func parseLevels(level string) []slog.Level {
	var levels []slog.Level
	switch strings.ToLower(level) {
	case "debug":
		levels = append(levels, slog.InfoLevel, slog.WarnLevel, slog.ErrorLevel, slog.DebugLevel)
	default:
		levels = append(levels, slog.InfoLevel, slog.WarnLevel, slog.ErrorLevel)
	}

	return levels
}

//
// 实现 Logger 接口。
//

func (l logger) Info(msg string, args ...interface{}) {
	l.sl.Infof(msg, args...)
}

func (l logger) Error(msg string, args ...interface{}) {
	l.sl.Errorf(msg, args...)
}

func (l logger) Warn(msg string, args ...interface{}) {
	l.sl.Warnf(msg, args...)
}

func (l logger) Debug(msg string, args ...interface{}) {
	l.sl.Debugf(msg, args...)
}
