package main

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// PrettyTextHandler is a custom slog.Handler that outputs formatted text
// For values implementing encoding.TextMarshaler, it outputs the formatted text directly
// instead of escaping it
type PrettyTextHandler struct {
	opts slog.HandlerOptions
	out  io.Writer
}

// NewPrettyTextHandler creates a new PrettyTextHandler
func NewPrettyTextHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyTextHandler {
	h := &PrettyTextHandler{
		out: out,
	}
	if opts != nil {
		h.opts = *opts
	}
	return h
}

// Enabled implements slog.Handler
func (h *PrettyTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle implements slog.Handler
func (h *PrettyTextHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf strings.Builder

	// Write time
	if !r.Time.IsZero() {
		buf.WriteString(r.Time.Format("2006-01-02T15:04:05.000-07:00"))
		buf.WriteString(" ")
	}

	// Write level
	buf.WriteString(r.Level.String())
	buf.WriteString(" ")

	// Write message
	buf.WriteString(r.Message)

	// Write attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		h.appendAttr(&buf, a)
		return true
	})

	buf.WriteString("\n")

	_, err := h.out.Write([]byte(buf.String()))
	return err
}

// WithAttrs implements slog.Handler
func (h *PrettyTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we don't support WithAttrs in this implementation
	return h
}

// WithGroup implements slog.Handler
func (h *PrettyTextHandler) WithGroup(name string) slog.Handler {
	// For simplicity, we don't support WithGroup in this implementation
	return h
}

// appendAttr appends an attribute to the buffer
func (h *PrettyTextHandler) appendAttr(buf *strings.Builder, a slog.Attr) {
	// Handle the value
	switch v := a.Value.Any().(type) {
	case encoding.TextMarshaler:
		// For TextMarshaler, output the formatted text directly
		if text, err := v.MarshalText(); err == nil {
			buf.WriteString(string(text))
			return
		}
	case fmt.Stringer:
		// For Stringer, output the string directly
		buf.WriteString(v.String())
		return
	case string:
		buf.WriteString(v)
		return
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		fmt.Fprintf(buf, "%d", v)
		return
	case float32, float64:
		fmt.Fprintf(buf, "%g", v)
		return
	case bool:
		fmt.Fprintf(buf, "%t", v)
		return
	case time.Duration:
		buf.WriteString(v.String())
		return
	case time.Time:
		buf.WriteString(v.Format(time.RFC3339))
		return
	}

	// For other types, use default formatting
	buf.WriteString(a.Value.String())
}
