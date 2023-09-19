package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
)

// NewTextHandler creates a new handler that writes human-readable log records to the given output.
func NewTextHandler(output io.Writer) slog.Handler {
	return &textHandler{output: output}
}

// textHandler is a slog.Handler that writes human-readable log records to an output.
type textHandler struct {
	output io.Writer
}

// Enabled returns true for all levels as we handle the cutoff ourselves using reactive variables and the ability to
// set loggers to nil.
func (t *textHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle writes the log record to the output.
func (t *textHandler) Handle(_ context.Context, r slog.Record) error {
	var namespace string
	fieldsBuffer := new(bytes.Buffer)

	fieldCount := r.NumAttrs() - 1
	if fieldCount > 0 {
		fieldsBuffer.WriteString("(")
	}

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == namespaceKey {
			namespace = attr.Value.Any().(string)
		} else {
			fieldsBuffer.WriteString(attr.String())
			fieldsBuffer.WriteString(" ")
		}

		return true
	})

	if fieldCount > 0 {
		fieldsBuffer.Truncate(fieldsBuffer.Len() - 1)
		fieldsBuffer.WriteString(")")
	}

	fmt.Fprintf(t.output, "%s\t%-7s\t%s\t%s %s\n", r.Time.Format("2006/01/02 15:04:05"), LevelName(r.Level), namespace, r.Message, fieldsBuffer.String())

	return nil
}

// WithAttrs is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (t *textHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not supported")
}

// WithGroup is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (t *textHandler) WithGroup(_ string) slog.Handler {
	panic("not supported")
}
