// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package logger

import (
	"os"
	"strings"
	"testing"
)

// Logger is used to define the logger interface.
type Logger interface {
	// Debug is used to log a debug message. If waiter is not nil, it will put three
	// dots at the end of the message, and then put the text at the end of the message
	// when the waiter is done.
	Debug(message string, waiter chan string)

	// Info is used to log a info message. If waiter is not nil, it will put three
	// dots at the end of the message, and then put the text at the end of the message
	// when the waiter is done.
	Info(message string, waiter chan string)

	// Warn is used to log a warn message. If waiter is not nil, it will put three
	// dots at the end of the message, and then put the text at the end of the message
	// when the waiter is done.
	Warn(message string, waiter chan string)

	// Error is used to log a error message. If waiter is not nil, it will put three
	// dots at the end of the message, and then put the text at the end of the message
	// when the waiter is done.
	Error(message string, waiter chan string)

	// Tag is used to tag the logger with a tag.
	Tag(tag string) Logger
}

type testingLogger struct {
	t    *testing.T
	tags []string
}

func (l testingLogger) Debug(message string, waiter chan string) {
	if waiter != nil {
		message += "... " + <-waiter
	}
	tags := "[DEBUG] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}
	l.t.Log(tags + message)
}

func (l testingLogger) Info(message string, waiter chan string) {
	if waiter != nil {
		message += "... " + <-waiter
	}
	tags := "[INFO] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}
	l.t.Log(tags + message)
}

func (l testingLogger) Warn(message string, waiter chan string) {
	if waiter != nil {
		message += "... " + <-waiter
	}
	tags := "[WARN] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}
	l.t.Log(tags + message)
}

func (l testingLogger) Error(message string, waiter chan string) {
	if waiter != nil {
		message += "... " + <-waiter
	}
	tags := "[ERROR] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}
	l.t.Log(tags + message)
}

func (l testingLogger) Tag(tag string) Logger {
	tagsLen := len(l.tags)
	tags := make([]string, tagsLen+1)
	copy(tags, l.tags)
	tags[tagsLen] = tag

	return testingLogger{
		t:    l.t,
		tags: tags,
	}
}

var _ Logger = testingLogger{}

// NewTestingLogger creates a new testing logger.
func NewTestingLogger(t *testing.T) Logger {
	return testingLogger{
		t:    t,
		tags: []string{},
	}
}

type loggingItem struct {
	message string
	err     bool
	waiter  chan string
}

type stdoutLogger struct {
	tags []string
	ch   chan loggingItem
}

var debugEnv = os.Getenv("DEBUG")

func truth(s string) bool {
	s = strings.ToLower(s)
	return s == "1" || s == "true" || s == "yes"
}

func (l stdoutLogger) Debug(message string, waiter chan string) {
	if !truth(debugEnv) {
		return
	}

	tags := "[DEBUG] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}

	select {
	case l.ch <- loggingItem{
		message: tags + message,
		waiter:  waiter,
	}:
	default:
	}
}

func (l stdoutLogger) Info(message string, waiter chan string) {
	tags := "[INFO] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}

	select {
	case l.ch <- loggingItem{
		message: tags + message,
		waiter:  waiter,
	}:
	default:
	}
}

func (l stdoutLogger) Warn(message string, waiter chan string) {
	tags := "[WARN] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}

	select {
	case l.ch <- loggingItem{
		message: tags + message,
		waiter:  waiter,
	}:
	default:
	}
}

func (l stdoutLogger) Error(message string, waiter chan string) {
	tags := "[ERROR] "
	for _, tag := range l.tags {
		tags += "[" + tag + "] "
	}

	select {
	case l.ch <- loggingItem{
		message: tags + message,
		err:     true,
		waiter:  waiter,
	}:
	default:
	}
}

func (l stdoutLogger) Tag(tag string) Logger {
	tagsLen := len(l.tags)
	tags := make([]string, tagsLen+1)
	copy(tags, l.tags)
	tags[tagsLen] = tag

	return stdoutLogger{
		tags: tags,
		ch:   l.ch,
	}
}

var _ Logger = stdoutLogger{}

// NewStdLogger creates a new stdout/stderr logger.
func NewStdLogger() Logger {
	ch := make(chan loggingItem)

	go func() {
		for {
			item := <-ch
			if item.waiter == nil {
				item.message += "\n"
			} else {
				item.message += "... "
			}

			if item.err {
				_, _ = os.Stderr.WriteString(item.message)
			} else {
				_, _ = os.Stdout.WriteString(item.message)
			}

			if item.waiter != nil {
				if item.err {
					_, _ = os.Stderr.WriteString(<-item.waiter + "\n")
				} else {
					_, _ = os.Stdout.WriteString(<-item.waiter + "\n")
				}
			}
		}
	}()

	return stdoutLogger{
		tags: []string{},
		ch:   ch,
	}
}
