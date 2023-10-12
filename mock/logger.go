package mock

import (
	"github.com/mrsoftware/circuitbreaker"
	"github.com/stretchr/testify/mock"
)

var _ circuitbreaker.Logger = &Logger{}

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
