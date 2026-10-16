package logger

type silentLog struct{}

func (l *silentLog) Debug(string, ...any)      {}
func (l *silentLog) Info(string, ...any)       {}
func (l *silentLog) Warn(string, ...any)       {}
func (l *silentLog) Error(string, ...any)      {}
func (l *silentLog) Write([]byte) (int, error) { return 0, nil }

func NewSilentLogger() LoggerWithWriter {
	return &silentLog{}
}
