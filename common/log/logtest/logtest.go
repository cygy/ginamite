// Package logtest provides a logger collecting its records, for the tests of the packages which log.
package logtest

import (
	"context"
	"log/slog"
	"sync"
)

// Entry is a record collected by a Hook. The attributes are in Data by key, the groups being nested maps.
type Entry struct {
	Level   slog.Level
	Message string
	Data    map[string]any
}

// Hook collects the records of the logger returned by New.
type Hook struct {
	mu      sync.Mutex
	entries []Entry
}

// New returns a logger which discards its records, from the info level, and the hook collecting them. Every record
// carries the attribute server=test. The signed integers in Data are int and the unsigned ones uint, as slog stores them on 64 bits.
func New() (*slog.Logger, *Hook) {
	return NewWithLevel(slog.LevelInfo)
}

// NewWithLevel is like New with another minimum level.
func NewWithLevel(level slog.Level) (*slog.Logger, *Hook) {
	hook := &Hook{}

	return slog.New(&handler{hook: hook, level: level, attrs: map[string]any{}}).With("server", "test"), hook
}

// AllEntries returns the records collected, oldest first.
func (h *Hook) AllEntries() []Entry {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]Entry(nil), h.entries...)
}

// LastEntry returns the last record collected, or nil if there is none.
func (h *Hook) LastEntry() *Entry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 {
		return nil
	}

	entry := h.entries[len(h.entries)-1]

	return &entry
}

// Reset forgets the records collected.
func (h *Hook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = nil
}

// handler is the slog.Handler of the hook. The attributes added with With are kept in attrs, whose keys are already
// nested in the open groups.
type handler struct {
	hook   *Hook
	level  slog.Level
	attrs  map[string]any
	groups []string
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool { return level >= h.level }

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	data := map[string]any{}
	copyData(data, h.attrs)

	target := data
	for _, group := range h.groups {
		next, _ := target[group].(map[string]any)
		if next == nil {
			next = map[string]any{}
			target[group] = next
		}
		target = next
	}

	r.Attrs(func(a slog.Attr) bool {
		add(target, a)
		return true
	})

	h.hook.mu.Lock()
	defer h.hook.mu.Unlock()
	h.hook.entries = append(h.hook.entries, Entry{Level: r.Level, Message: r.Message, Data: data})

	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.clone()

	target := c.attrs
	for _, group := range c.groups {
		next, _ := target[group].(map[string]any)
		if next == nil {
			next = map[string]any{}
			target[group] = next
		}
		target = next
	}
	for _, a := range attrs {
		add(target, a)
	}

	return c
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	c := h.clone()
	c.groups = append(c.groups, name)

	return c
}

func (h *handler) clone() *handler {
	c := &handler{hook: h.hook, level: h.level, attrs: map[string]any{}, groups: append([]string(nil), h.groups...)}
	copyData(c.attrs, h.attrs)

	return c
}

// add sets an attribute in data, resolving its value and nesting its group.
func add(data map[string]any, a slog.Attr) {
	a.Value = a.Value.Resolve()
	switch a.Value.Kind() {
	case slog.KindInt64:
		data[a.Key] = int(a.Value.Int64())
		return
	case slog.KindUint64:
		data[a.Key] = uint(a.Value.Uint64())
		return
	case slog.KindGroup:
	default:
		data[a.Key] = a.Value.Any()
		return
	}

	nested := map[string]any{}
	for _, member := range a.Value.Group() {
		add(nested, member)
	}
	if a.Key == "" {
		copyData(data, nested)
		return
	}
	data[a.Key] = nested
}

// copyData copies src in dst, with a copy of the nested maps.
func copyData(dst, src map[string]any) {
	for k, v := range src {
		if nested, ok := v.(map[string]any); ok {
			c := map[string]any{}
			copyData(c, nested)
			v = c
		}
		dst[k] = v
	}
}

// Replace sets *logger, such as log.Main, to the logger of NewWithLevel and returns the hook collecting its records and
// the function restoring the previous logger.
func Replace(logger **slog.Logger, level slog.Level) (*Hook, func()) {
	previous := *logger

	var hook *Hook
	*logger, hook = NewWithLevel(level)

	return hook, func() { *logger = previous }
}
