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

type Logger slog.SugaredLogger

type Config struct {
	LogTarget         string `json:"log_target"`
	LogLevel          string `json:"log_level"`
	LogSyslogServer   string `json:"log_syslog_server"`
	LogSysTag         string `json:"log_sys_tag"`
	LogPath           string `json:"log_path"`
	LogMaxSize        int    `json:"log_max_size"`
	LogRotateInterval string `json:"log_rotate_interval"`
	LogBufferSize     int    `json:"log_buffer_size"` // 指定缓冲区的大小，注意：默认为 8 * 1024。当为 0 时将会即时写入日志文件
	LogMaxBackups     int    `json:"log_max_backups"`
	LogTimeFormat     string `json:"log_time_format"`
	LogWithCompress   bool   `json:"log_with_compress"` // 是否开启 gzip 压缩
}

func New(c *Config) (*Logger, error) {
	var logTemplate string
	var syslogLevel syslog.Priority
	// 当 Logger level 为 debug 时开启 caller，方便快速定位打印日志位置
	level := slog.LevelByName(c.LogLevel)
	if level == slog.DebugLevel {
		syslogLevel = syslog.LOG_DEBUG
		logTemplate = "{{datetime}} {{level}} {{message}} [{{caller}}]\n"
	} else {
		syslogLevel = syslog.LOG_INFO
		logTemplate = "{{datetime}} {{level}} {{message}}\n"
	}

	// 自定义 Logger formatter
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
	case "file":
		var h *handler.SyncCloseHandler
		var err error
		if c.LogMaxSize > 0 {
			h, err = handlerRotateFile(c)
			if err != nil {
				return nil, err
			}
		} else {
			h, err = handlerRotateTime(c)
			if err != nil {
				return nil, err
			}
		}

		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	case "syslog":
		if strings.HasPrefix(c.LogSyslogServer, "/") {
			h, err := handler.NewSysLogHandler(syslogLevel|syslog.LOG_MAIL, c.LogSysTag)
			if err != nil {
				return nil, err
			}
			h.SetFormatter(logFormatter)
			l.AddHandler(h)

			break
		}

		w, err := syslog.Dial("tcp", c.LogSyslogServer, syslogLevel|syslog.LOG_MAIL, c.LogSysTag)
		if err != nil {
			return nil, err
		}
		h := handler.NewBufferedHandler(w, c.LogBufferSize, level)
		h.SetFormatter(logFormatter)
		l.AddHandler(h)
	}

	l.Level = level
	l.DoNothingOnPanicFatal()

	return (*Logger)(l), nil
}

func handlerRotateFile(c *Config) (*handler.SyncCloseHandler, error) {
	return handler.NewSizeRotateFileHandler(
		c.LogPath,
		c.LogMaxSize,
		handler.WithBuffSize(c.LogBufferSize),
		handler.WithCompress(c.LogWithCompress),
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
		c.LogPath,
		rotatefile.RotateTime(rotateIntervalDuration.Seconds()),
		handler.WithBuffSize(c.LogBufferSize),
		handler.WithCompress(c.LogWithCompress),
		handler.WithLogLevels(slog.AllLevels),
	)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.Debugf(format, args...)
}
