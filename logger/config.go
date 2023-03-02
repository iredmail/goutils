package logger

func ConfigWithFile(logFilePath string) *Config {
	return &Config{
		LogTarget:     "file",
		LogFile:       logFilePath,
		LogBufferSize: 0,
	}
}

func ConfigWithSys(logSyslogServer, logSyslogTag string) *Config {
	return &Config{
		LogTarget:       "syslog",
		LogSyslogServer: logSyslogServer,
		LogSyslogTag:    logSyslogTag,
	}
}

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

func (c *Config) SetLevel(level string) {
	c.LogLevel = level
}

func (c *Config) SetMaxSize(maxSize int) {
	c.LogMaxSize = maxSize
}

func (c *Config) SetRotateInterval(rotateInterval string) {
	c.LogRotateInterval = rotateInterval
}

func (c *Config) SetBufferSize(bufferSize int) {
	c.LogBufferSize = bufferSize
}

func (c *Config) SetMaxBackups(maxBackups int) {
	c.LogMaxBackups = maxBackups
}

func (c *Config) SetCompress(compress bool) {
	c.LogCompress = compress
}

func (c *Config) SetTimeFormat(timeFormat string) {
	c.LogTimeFormat = timeFormat
}
