package logger

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
	"golang.org/x/exp/slices"
)

type log slog.SugaredLogger

type Config struct {
	LogLevel     slog.Level
	LogDir       string
	FileName     string
	BufferSize   int // 指定缓冲区的大小，注意：默认为 8 * 1024。当为 0 时将会即时写入日志文件
	MaxBackups   int
	TimeFormat   string
	WithCompress bool // 是否开启 gzip 压缩
}

func NewRotateFile(c *Config, maxSize int) (*log, error) {
	pth := filepath.Join(c.LogDir, c.FileName)
	h, err := handler.NewSizeRotateFileHandler(
		pth,
		maxSize,
		handler.WithBuffSize(c.BufferSize),
		handler.WithCompress(c.WithCompress),
		handler.WithLogLevels(slog.AllLevels),
	)

	if err != nil {
		return nil, err
	}

	var logTemplate string
	// 当 log level 为 debug 时开启 caller，方便快速定位打印日志位置
	if c.LogLevel == slog.DebugLevel {
		logTemplate = "{{datetime}} {{level}} [{{caller}}] {{message}}\n"
	} else {
		logTemplate = "{{datetime}} {{level}} {{message}}\n"
	}

	// 自定义 log formatter
	logFormatter := slog.NewTextFormatter(logTemplate)
	logFormatter.TimeFormat = c.TimeFormat
	h.SetFormatter(logFormatter)

	l := slog.NewStdLogger().Configure(func(sl *slog.SugaredLogger) {
		sl.ReportCaller = true
		sl.CallerSkip = 6
	})
	l.Config(func(sl *slog.SugaredLogger) {
		f := sl.Formatter.(*slog.TextFormatter)
		f.TimeFormat = c.TimeFormat
		f.SetTemplate(logTemplate)
		f.FullDisplay = true
		f.EnableColor = false
	})

	l.Level = c.LogLevel
	l.AddHandler(h)
	l.DoNothingOnPanicFatal()

	return (*log)(l), nil
}

// NewRotateTime
// rotateInterval: 1w, 1d, 1h, 1m, 1s
func NewRotateTime(c *Config, rotateInterval string) (*log, error) {
	if len(rotateInterval) == 0 {
		return nil, errors.New("empty rotate interval")
	}

	lastChar := rotateInterval[len(rotateInterval)-1]
	lowerLastChar := strings.ToLower(string(lastChar))
	if !slices.Contains([]string{"w", "d", "h", "m", "s"}, lowerLastChar) {
		return nil, fmt.Errorf("unsuppored rotate interval type: %s", lowerLastChar)
	}

	rotateIntervalDuration, err := time.ParseDuration(rotateInterval)
	if err != nil {
		return nil, err
	}

	pth := filepath.Join(c.LogDir, c.FileName)
	h, err := handler.NewTimeRotateFileHandler(
		pth,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithBuffSize(c.BufferSize),
		handler.WithCompress(c.WithCompress),
		handler.WithLogLevels(slog.AllLevels),
	)

	if err != nil {
		return nil, err
	}

	var logTemplate string
	// 当 log level 为 debug 时开启 caller，方便快速定位打印日志位置
	if c.LogLevel == slog.DebugLevel {
		logTemplate = "{{datetime}} {{level}} [{{caller}}] {{message}}\n"
	} else {
		logTemplate = "{{datetime}} {{level}} {{message}}\n"
	}

	// 自定义 log formatter
	logFormatter := slog.NewTextFormatter(logTemplate)
	logFormatter.TimeFormat = c.TimeFormat
	h.SetFormatter(logFormatter)

	l := slog.NewStdLogger().Configure(func(sl *slog.SugaredLogger) {
		sl.ReportCaller = true
		sl.CallerSkip = 6
	})
	l.Config(func(sl *slog.SugaredLogger) {
		f := sl.Formatter.(*slog.TextFormatter)
		f.TimeFormat = c.TimeFormat
		f.SetTemplate(logTemplate)
		f.FullDisplay = true
		f.EnableColor = false
	})

	l.Level = c.LogLevel
	l.AddHandler(h)
	l.DoNothingOnPanicFatal()

	return (*log)(l), nil
}

func (l *log) Info(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l *log) Warn(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *log) Error(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func (l *log) Debug(format string, args ...interface{}) {
	l.Debugf(format, args...)
}
