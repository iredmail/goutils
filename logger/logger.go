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

	level          string // log level: info, warn, error, debug
	maxSize        int
	rotateInterval string // rotate interval. e.g. `12h` (12 hours), `1d` (1 day), `1w` (1 week), `1m` (1 month)
	maxBackups     uint
	timeFormat     string
	compress       bool // compress rotated log file

	// Buffer size defaults to (8 * 1024).
	// Write to log file immediately if size is 0.
	bufferSize int
}

func NewStdoutLogger(options ...Option) (v Logger, err error) {
	l := newLog(options...)

	// custom log format
	logFormatter := genLogFormatter(l.timeFormat)

	h := handler.NewConsoleHandler(parseLevels(l.level))
	h.SetFormatter(logFormatter)
	l.sl.AddHandler(h)

	l.sl.DoNothingOnPanicFatal()

	return l, nil
}

func NewSyslogLogger(server, tag string, options ...Option) (logger Logger, err error) {
	l := newLog(options...)
	var syslogLevel syslog.Priority

	level := slog.LevelByName(l.level)
	if level == slog.DebugLevel {
		syslogLevel = syslog.LOG_DEBUG
		// 当 log level 为 debug 时开启 caller，方便快速定位打印日志位置
		// logTemplate = "{{datetime}} {{level}} {{message}} [{{caller}}]\n"
	} else {
		syslogLevel = syslog.LOG_INFO
		// logTemplate = "{{datetime}} {{level}} {{message}}\n"
	}

	// custom log format
	logFormatter := genLogFormatter(l.timeFormat)

	if len(server) == 0 {
		// Use local syslog socket by default.
		server = "/dev/log"
	}

	if strings.HasPrefix(server, "/") {
		h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, tag)
		if err != nil {
			return nil, err
		}
		h.SetFormatter(logFormatter)
		l.sl.AddHandler(h)
	} else {
		w, err := syslog.Dial("tcp", server, syslogLevel|syslog.LOG_MAIL, tag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, l.bufferSize, level)
		h.SetFormatter(logFormatter)
		l.sl.AddHandler(h)
	}

	l.sl.DoNothingOnPanicFatal()

	return l, nil
}

func NewFileLogger(pth string, options ...Option) (logger Logger, err error) {
	// enable compress
	l := newLog(WithCompress())
	for _, option := range options {
		option(&l)
	}

	// custom log format
	logFormatter := genLogFormatter(l.timeFormat)

	if l.maxSize > 0 {
		h, err := handlerRotateFile(l, pth)
		if err != nil {
			return nil, err
		}

		h.SetFormatter(logFormatter)
		l.sl.AddHandler(h)
	}

	if l.rotateInterval != "" {
		h, err := handlerRotateTime(l, pth)
		if err != nil {
			return nil, err
		}

		h.SetFormatter(logFormatter)
		l.sl.AddHandler(h)
	}

	l.sl.DoNothingOnPanicFatal()

	return l, nil
}

func newLog(options ...Option) logger {
	sl := slog.New()
	sl.ReportCaller = true
	sl.CallerSkip = 6
	l := logger{
		sl: sl,
	}

	for _, option := range options {
		option(&l)
	}

	return l
}

func genLogFormatter(timeFormat string) *slog.TextFormatter {
	logTemplate := "{{datetime}} {{level}} {{message}}\n"
	// custom log format
	logFormatter := slog.NewTextFormatter(logTemplate)
	logFormatter.EnableColor = false
	logFormatter.FullDisplay = true
	logFormatter.TimeFormat = timeFormat

	return logFormatter
}

func handlerRotateFile(log logger, logFile string) (*handler.SyncCloseHandler, error) {
	return handler.NewSizeRotateFileHandler(
		logFile,
		log.maxSize,
		handler.WithLogLevels(parseLevels(log.level)),
		handler.WithBuffSize(log.bufferSize),
		handler.WithBackupNum(log.maxBackups),
		handler.WithCompress(log.compress),
	)
}

// handlerRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func handlerRotateTime(log logger, logFile string) (*handler.SyncCloseHandler, error) {
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
		logFile,
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
