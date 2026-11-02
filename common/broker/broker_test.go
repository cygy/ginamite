package broker

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPermanent(t *testing.T) {
	if Permanent(nil) != nil {
		t.Error("Permanent(nil) must be nil")
	}

	cause := errors.New("cause")
	err := Permanent(cause)
	if err.Error() != "cause" || !errors.Is(err, cause) {
		t.Errorf("err = %v", err)
	}
	if !IsPermanent(err) || !IsPermanent(fmt.Errorf("wrapped: %w", err)) {
		t.Error("a permanent error, wrapped or not, must be reported")
	}
	if IsPermanent(cause) || IsPermanent(nil) {
		t.Error("an ordinary error is not permanent")
	}
}

func TestUniqueGroup(t *testing.T) {
	a, b := UniqueGroup("cache"), UniqueGroup("cache")
	if !strings.HasPrefix(a, "cache_") || a == b {
		t.Errorf("groups = %q, %q", a, b)
	}
}

func TestMemory(t *testing.T) {
	m := NewMemory()

	var got []string
	m.Start([]MessageConsumer{
		{Topic: "mail", Action: func(v []byte) error { got = append(got, string(v)); return nil }},
		{Topic: "mail", Action: func([]byte) error { return errors.New("failed") }},
		{Topic: "tasks", Action: func([]byte) error { panic("boom") }},
	})

	m.SendMessageToTopic("a", "mail")
	m.SendKeyedMessageToTopic("user-1", "b", "mail")
	m.SendMessageToTopic("c", "tasks")
	m.SendMessageToTopic("d", "unknown")

	sent := m.Sent()
	if len(sent) != 4 || sent[1] != (Sent{Topic: "mail", Key: "user-1", Message: "b"}) {
		t.Errorf("sent = %+v", sent)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("handled = %v", got)
	}
	errs := m.Errors()
	if len(errs) != 3 || errs[2].Error() != "panic: boom" {
		t.Errorf("errors = %v", errs)
	}

	m.Close()
	m.SendMessageToTopic("e", "mail")
	if len(m.Sent()) != 4 {
		t.Error("a closed broker must ignore the messages")
	}
}
