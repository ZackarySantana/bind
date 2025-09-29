package bind

import "log/slog"

type options struct {
	logger *slog.Logger
	level  int

	testOnly bool
}

type Option func(*options)

// WithLogger sets the logger to use for debug messages.
func WithLogger(logger *slog.Logger) func(*options) {
	return func(o *options) {
		o.logger = logger
	}
}

// WithLevel sets the binding level to use. By default this is 1.
func WithLevel(level int) func(*options) {
	return func(o *options) {
		o.level = level
	}
}

// WithTestOnly makes all suppliers match the struct tag `test-only`.
func WithTestOnly() func(*options) {
	return func(o *options) {
		o.testOnly = true
	}
}
