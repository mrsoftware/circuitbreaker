package circuitbreaker

import (
	"encoding/json"
	"fmt"
	"io"
)

// Logger is what a logger look like in circuit breaker.
type Logger interface {
	Info(v ...interface{})
	Error(v ...interface{})
	Warn(v ...interface{})
}

const (
	// OutPutTypeJSON is json type of output enum.
	OutPutTypeJSON = "json"

	// OutPutTypeSimple is simple text type of output enum.
	OutPutTypeSimple = "simple"
)

type logType struct {
	status string
}

// nolint:gochecknoglobals
var (
	logTypeError   = logType{status: "error"}
	logTypeWarning = logType{status: "warning"}
	logTypeInfo    = logType{status: "info"}
)

// IOLogger is io logger and use any io.Writer.
type IOLogger struct {
	writer     io.Writer
	outPutType string
}

// NewIOLogger create new instance of  IOLogger.
func NewIOLogger(writer io.Writer, outPutType string) *IOLogger {
	return &IOLogger{writer, outPutType}
}

// Info log with info logTypeInfo.
func (i *IOLogger) Info(v ...interface{}) {
	_ = i.formatToWriter(append(v, logTypeInfo))
}

// Error log with logTypeError.
func (i *IOLogger) Error(v ...interface{}) {
	_ = i.formatToWriter(append(v, logTypeError))
}

// Warn log with logTypeWarning.
func (i *IOLogger) Warn(v ...interface{}) {
	_ = i.formatToWriter(append(v, logTypeWarning))
}

func (i *IOLogger) formatToWriter(v []interface{}) error { // nolint:varnamelen
	if i.outPutType == OutPutTypeSimple {
		_, err := fmt.Fprintln(i.writer, v...)

		return err // nolint:wrapcheck
	}

	return json.NewEncoder(i.writer).Encode(v) // nolint:wrapcheck
}
