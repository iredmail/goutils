package logger

type Option func(l *logger)

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

// WithMaxSize sets the maximum size of the log file in megabytes.
// If it's 0 or not set, defaults to 300 (MB).
func WithMaxSize(maxSize int) Option {
	if maxSize == 0 {
		maxSize = 300 * 1024 * 1024 // 100 MB
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

func WithMaxBackups(maxBackups uint) Option {
	return func(l *logger) {
		l.maxBackups = maxBackups
	}
}

func WithCompress() Option {
	return func(l *logger) {
		l.compress = true
	}
}

func WithTimeFormat(timeFormat string) Option {
	return func(l *logger) {
		l.timeFormat = timeFormat
	}
}
