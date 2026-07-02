package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
)

// ANSI color codes
const (
	ansiReset     = "\033[0m"
	ansiDim       = "\033[2m"
	ansiGrey      = "\033[90m"
	ansiLightGrey = "\033[37m"
	ansiCyan      = "\033[36m"
	ansiBlue      = "\033[34m"
	ansiYellow    = "\033[33m"
	ansiRed       = "\033[31m"
)

// prettyHandler is a custom slog.Handler that formats log output with colors,
// mirroring the previous zap-based console format: a dim timestamp, a colored
// padded level, a dim caller (pkg/file.go:line), the message, and structured
// fields as grey key=value pairs. Error-level entries append a stack trace,
// matching the old zap AddStacktrace(ErrorLevel) behavior.
type prettyHandler struct {
	level  slog.Leveler
	w      io.Writer
	mu     *sync.Mutex
	attrs  []slog.Attr
	prefix string // group prefix applied to attr keys ("" when no active group)
}

func newPrettyHandler(w io.Writer, level slog.Leveler) *prettyHandler {
	return &prettyHandler{level: level, w: w, mu: &sync.Mutex{}}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var sb strings.Builder

	// Timestamp (dim), UTC millisecond precision.
	ts := r.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	sb.WriteString(ansiDim)
	sb.WriteString(ts)
	sb.WriteString(ansiReset)
	sb.WriteString("    ")

	// Level (colored, padded to 5 chars).
	sb.WriteString(levelColor(r.Level))
	sb.WriteString(levelLabel(r.Level))
	sb.WriteString(ansiReset)
	sb.WriteString("    ")

	// Caller (dim), rendered as the last two path segments: pkg/file.go:line.
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			sb.WriteString(ansiDim)
			sb.WriteString(trimCaller(f.File))
			sb.WriteByte(':')
			sb.WriteString(itoa(f.Line))
			sb.WriteString(ansiReset)
			sb.WriteString("    ")
		}
	}

	// Message (normal).
	sb.WriteString(r.Message)

	// Structured fields (preset via With + per-record) as grey key=value pairs.
	parts := make([]string, 0, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		parts = appendAttr(parts, h.prefix, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		parts = appendAttr(parts, h.prefix, a)
		return true
	})
	if len(parts) > 0 {
		sb.WriteString("    ")
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString(ansiReset)
	}

	// Stack trace on error and above (mirrors zap AddStacktrace(ErrorLevel)).
	if r.Level >= slog.LevelError {
		if stack := captureStack(); stack != "" {
			sb.WriteByte('\n')
			sb.WriteString(ansiDim)
			sb.WriteString(stack)
			sb.WriteString(ansiReset)
		}
	}

	sb.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, sb.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	clone.attrs = append(clone.attrs, h.attrs...)
	clone.attrs = append(clone.attrs, attrs...)
	return &clone
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.prefix = h.prefix + name + "."
	return &clone
}

func appendAttr(parts []string, prefix string, a slog.Attr) []string {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return parts
	}
	return append(parts, fmt.Sprintf("%s%s%s=%s%v", ansiGrey, prefix, a.Key, ansiLightGrey, a.Value.Any()))
}

func levelColor(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return ansiCyan
	case level < slog.LevelWarn:
		return ansiBlue
	case level < slog.LevelError:
		return ansiYellow
	default:
		return ansiRed
	}
}

// levelLabel returns the level name padded to 5 characters (DEBUG/INFO /WARN /ERROR).
func levelLabel(level slog.Level) string {
	s := level.String()
	for len(s) < 5 {
		s += " "
	}
	return s
}

// trimCaller returns the last two path segments of a file path (pkg/file.go),
// matching zap's Caller.TrimmedPath().
func trimCaller(file string) string {
	idx := strings.LastIndexByte(file, '/')
	if idx < 0 {
		return file
	}
	prev := strings.LastIndexByte(file[:idx], '/')
	if prev < 0 {
		return file
	}
	return file[prev+1:]
}

// captureStack renders the current goroutine's stack, skipping the slog and
// logger internal frames so the trace starts at the actual log call site.
func captureStack() string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	for {
		f, more := frames.Next()
		if !strings.Contains(f.Function, "log/slog.") &&
			!strings.Contains(f.File, "internal/logger/") {
			fmt.Fprintf(&sb, "%s\n\t%s:%d\n", f.Function, f.File, f.Line)
		}
		if !more {
			break
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
