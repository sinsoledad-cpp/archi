package logger

var _ Logger = (*NopLogger)(nil)

type NopLogger struct {
}

func NewNopLogger() Logger {
	return &NopLogger{}
}

func (l *NopLogger) Debug(msg string, args ...Field) {
}

func (l *NopLogger) Info(msg string, args ...Field) {}

func (l *NopLogger) Warn(msg string, args ...Field) {}

func (l *NopLogger) Error(msg string, args ...Field) {}
