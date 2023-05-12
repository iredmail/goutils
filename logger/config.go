package logger

import (
	"time"
)

const (
	TargetFile   string = "file"
	TargetStdout string = "stdout"
	TargetSyslog string = "syslog"
)

func ConfigWithFile(logFilePath string) *Config {
	return &Config{
		target:     TargetFile,
		logFile:    logFilePath,
		level:      "info",
		bufferSize: 0,
		timeFormat: time.DateTime,
		compress:   true,
	}
}

func ConfigWithSyslog(logSyslogServer, logSyslogTag string) *Config {
	return &Config{
		target:       TargetSyslog,
		level:        "info",
		syslogServer: logSyslogServer,
		syslogTag:    logSyslogTag,
		timeFormat:   time.DateTime,
	}
}

type Config struct {
	target         string // log target: file, syslog
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

// SetTarget 控制将 log 记录到哪里。
// 允许的值：stdout, file, syslog。其余值按 `file` 处理。
func (c *Config) SetTarget(target string) *Config {
	switch target {
	case TargetStdout, TargetSyslog, TargetFile:
		c.target = target
	default:
		c.target = TargetFile
	}

	return c
}

// SetLevel 设置 log level。
// 允许的值：info, warn, error, debug。其余值按 `info` 处理。
func (c *Config) SetLevel(level string) *Config {
	switch level {
	case "info", "warn", "error", "debug":
		c.level = level
	default:
		c.level = "info"
	}

	return c
}

// SetMaxSize 设置 log 文件的最大大小（单位 bytes），达到该值则触发 rotate。
func (c *Config) SetMaxSize(maxSize int) *Config {
	if maxSize > 0 {
		c.maxSize = maxSize
	} else {
		c.maxSize = 500 * 1024 * 1024 // 500 MB
	}

	return c
}

// SetRotateInterval 设置 log 文件的 rotate 时间间隔。
// 示例：`12h` (12 hours), `1d` (1 day), `1w` (1 week), `1m` (1 month)。
func (c *Config) SetRotateInterval(interval string) *Config {
	c.rotateInterval = interval

	return c
}

// SetBufferSize 设置 log 内容的缓冲大小，达到该值后再写入 log target。
// 设置为 0 表示不做缓冲，小于 0 的值按 0 处理。
func (c *Config) SetBufferSize(bufferSize int) *Config {
	if bufferSize >= 0 {
		c.bufferSize = bufferSize
	} else {
		c.bufferSize = 0
	}

	return c
}

// SetMaxBackups 设置 rotate 后的 log 文件数量。
func (c *Config) SetMaxBackups(maxBackups uint) *Config {
	c.maxBackups = maxBackups

	return c
}

// SetCompress 设置是否压缩 rotate 后的 log 文件。默认为 false。
func (c *Config) SetCompress(compress bool) *Config {
	c.compress = compress

	return c
}

// SetTimeFormat 设置 log 里记录的时间格式。默认为 time.RFC3339。
func (c *Config) SetTimeFormat(timeFormat string) *Config {
	c.timeFormat = timeFormat

	return c
}
