package go_logger

import (
	"log"
	"runtime"
)

// Basic is a logger that allows you to set go's logger
// Println will be used for logging values
type Basic struct {
	l *log.Logger
}

// NewBasic creates a logger fashioned from Go's default logging apparatus
// This allows you to use the new logging interface, but use the default logger instead of a file
// This is very useful for just outputting log entries while manually testing
func NewBasic(l *log.Logger) *Basic {
	return &Basic{
		l: l,
	}
}

// Does nothing
func (n *Basic) Log(tag string, msg Msg, skip int) {
	// no-op
	_, file, line, _ := runtime.Caller(skip + 1)
	n.l.Printf(`%s %s:%d %s`, tag, file, line, msg)
}

// Does nothing
// @return nil
func (n *Basic) ReOpen() error {
	// no-op
	return nil
}

// Does nothing
// @return nil
func (n *Basic) Close() error {
	// no-op
	return nil
}
