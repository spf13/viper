package viper

import (
	"fmt"

	jww "github.com/spf13/jwalterweatherman"
)

// Logger is a unified interface for various logging use cases and practices, including:
//   - leveled logging
//   - structured logging
type Logger interface {
	// Trace logs a Trace event.
	//
	// Even more fine-grained information than Debug events.
	// Loggers not supporting this level should fall back to Debug.
	Trace(msg string, keyvals ...interface{})

	// Debug logs a Debug event.
	//
	// A verbose series of information events.
	// They are useful when debugging the system.
	Debug(msg string, keyvals ...interface{})

	// Info logs an Info event.
	//
	// General information about what's happening inside the system.
	Info(msg string, keyvals ...interface{})

	// Warn logs a Warn(ing) event.
	//
	// Non-critical events that should be looked at.
	Warn(msg string, keyvals ...interface{})

	// Error logs an Error event.
	//
	// Critical events that require immediate attention.
	// Loggers commonly provide Fatal and Panic levels above Error level,
	// but exiting and panicing is out of scope for a logging library.
	Error(msg string, keyvals ...interface{})

	// SetName sets a name for logger, the logger name will be the prefix of all the log entry.
	SetName(string)
}

type jwwLogger struct {
	name string
}

func (j *jwwLogger) Trace(msg string, keyvals ...interface{}) {
	msg = fmt.Sprintf("%s | %s", j.name, msg)
	jww.TRACE.Printf(jwwLogMessage(msg, keyvals...))
}

func (j *jwwLogger) Debug(msg string, keyvals ...interface{}) {
	msg = fmt.Sprintf("%s | %s", j.name, msg)
	jww.DEBUG.Printf(jwwLogMessage(msg, keyvals...))
}

func (j *jwwLogger) Info(msg string, keyvals ...interface{}) {
	msg = fmt.Sprintf("%s | %s", j.name, msg)
	jww.INFO.Printf(jwwLogMessage(msg, keyvals...))
}

func (j *jwwLogger) Warn(msg string, keyvals ...interface{}) {
	msg = fmt.Sprintf("%s | %s", j.name, msg)
	jww.WARN.Printf(jwwLogMessage(msg, keyvals...))
}

func (j *jwwLogger) Error(msg string, keyvals ...interface{}) {
	msg = fmt.Sprintf("%s | %s", j.name, msg)
	jww.ERROR.Printf(jwwLogMessage(msg, keyvals...))
}

func (j *jwwLogger) SetName(name string) {
	j.name = name
}

func jwwLogMessage(msg string, keyvals ...interface{}) string {
	out := msg

	if len(keyvals) > 0 && len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	for i := 0; i <= len(keyvals)-2; i += 2 {
		out = fmt.Sprintf("%s %v=%v", out, keyvals[i], keyvals[i+1])
	}

	return out
}
