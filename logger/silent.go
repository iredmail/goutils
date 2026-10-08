package logger

type silentLog struct{}

func (l *silentLog) Debug(string, ...interface{}) {}
func (l *silentLog) Info(string, ...interface{})  {}
func (l *silentLog) Warn(string, ...interface{})  {}
func (l *silentLog) Error(string, ...interface{}) {}
func (l *silentLog) Write([]byte) (int, error)    { return 0, nil }

func NewSilentLogger() LoggerWithWriter {
	return &silentLog{}
}
