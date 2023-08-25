package mock

import "github.com/stretchr/testify/mock"

type Logger struct {
	mock.Mock
}

func (l *Logger) Info(v ...interface{}) {
	l.Called(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Called(v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Called(v...)
}
