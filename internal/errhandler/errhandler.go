// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package errhandler

import (
	"go.uber.org/zap"
)

// Handler is used to define the error handler.
type Handler struct {
	// Logger is used to log messages.
	Logger *zap.SugaredLogger

	// TODO: handle sentry
}

// Tag is used to tag the error handler.
func (h Handler) Tag(tagName string) Handler {
	return Handler{
		Logger: h.Logger.Named(tagName),
	}
}

// HandleError is used to handle an error.
func (h Handler) HandleError(err error) {
	// Handle if the error is nil.
	if err == nil {
		return
	}

	// Get the error stacktrace.
	h.Logger.Error(
		"internal server error",
		zap.Error(err),
	)
}
