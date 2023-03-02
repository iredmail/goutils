package logger

func ConfigWithFile(logFilePath string) *Config {
	return &Config{
		logTarget:     "file",
		logFile:       logFilePath,
		logBufferSize: 0,
	}
}

func ConfigWithSys(logSyslogServer, logSyslogTag string) *Config {
	return &Config{
		logTarget:       "syslog",
		logSyslogServer: logSyslogServer,
		logSyslogTag:    logSyslogTag,
	}
}

type Config struct {
	logTarget         string
	logLevel          string
	logSyslogServer   string
	logSyslogTag      string
	logFile           string
	logMaxSize        int
	logRotateInterval string
	// Buffer size defaults to (8 * 1024).
	// Write to log file immediately if size is 0.
	logBufferSize int
	logMaxBackups int
	logTimeFormat string
	logCompress   bool // compress rotated log file
}

func (c *Config) SetTarget(target string) *Config {
	c.logTarget = target

	return c
}

func (c *Config) SetLevel(level string) *Config {
	c.logLevel = level

	return c
}

func (c *Config) SetMaxSize(maxSize int) *Config {
	c.logMaxSize = maxSize

	return c
}

func (c *Config) SetRotateInterval(rotateInterval string) *Config {
	c.logRotateInterval = rotateInterval

	return c
}

func (c *Config) SetBufferSize(bufferSize int) *Config {
	c.logBufferSize = bufferSize

	return c
}

func (c *Config) SetMaxBackups(maxBackups int) *Config {
	c.logMaxBackups = maxBackups

	return c
}

func (c *Config) SetCompress(compress bool) *Config {
	c.logCompress = compress

	return c
}

func (c *Config) SetTimeFormat(timeFormat string) *Config {
	c.logTimeFormat = timeFormat

	return c
}
