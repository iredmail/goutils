package logger

type Target string

const (
	TargetFile   Target = "file"
	TargetStdout Target = "stdout"
	TargetSyslog Target = "syslog"
)

type Option func(l *logger)

func WithTarget(target Target) Option {
	return func(l *logger) {
		l.target = target
	}
}

func WithLogFile(pth string) Option {
	return func(l *logger) {
		l.logFile = pth
	}
}

func WithSyslogServer(s string) Option {
	return func(l *logger) {
		l.syslogServer = s
	}
}

func WithSyslogTag(tag string) Option {
	return func(l *logger) {
		l.syslogTag = tag
	}
}

func WithLevel(level string) Option {
	switch level {
	case "info", "warn", "error", "debug":
		break
	default:
		level = "info"
	}

	return func(l *logger) {
		l.level = level
	}
}

func WithMaxSize(maxSize int) Option {
	if maxSize == 0 {
		maxSize = 500 * 1024 * 1024 // 500 MB
	}

	return func(l *logger) {
		l.maxSize = maxSize
	}
}

func WithRotateInterval(interval string) Option {
	return func(l *logger) {
		l.rotateInterval = interval
	}
}

func WithBufferSize(bufferSize int) Option {
	return func(l *logger) {
		l.bufferSize = bufferSize
	}
}

func WithMaxBackups(maxBackups uint) Option {
	return func(l *logger) {
		l.maxBackups = maxBackups
	}
}

func WithCompress(compress bool) Option {
	return func(l *logger) {
		l.compress = compress
	}
}

func WithTimeFormat(timeFormat string) Option {
	return func(l *logger) {
		l.timeFormat = timeFormat
	}
}
