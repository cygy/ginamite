package queue

import (
	"strings"
	"testing"
)

func TestMessage_Decode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      bool
		wantType  string
		wantError bool
	}{
		{"valid", `{"type":"t","payload":{"a":1}}`, true, "t", false},
		{"invalid json", `{`, false, "", true},
		{"empty", ``, false, "", true},
		{"wrong type", `{"type":3}`, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			var m Message
			if got := m.Decode([]byte(tt.input)); got != tt.want {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
			if tt.want && m.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", m.Type, tt.wantType)
			}
			if logged := len(hook.AllEntries()) > 0; logged != tt.wantError {
				t.Errorf("error logged = %v, want %v", logged, tt.wantError)
			}
		})
	}
}

func TestParseMessageAndPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		kind, payload, ok := ParseMessageAndPayload([]byte(`{"type":"t","payload":{"user_id":"u"}}`), "topic")
		if !ok || kind != "t" || string(payload) != `{"user_id":"u"}` {
			t.Errorf("got %q, %s, %v", kind, payload, ok)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		hook.Reset()

		kind, payload, ok := ParseMessageAndPayload([]byte(`nope`), "topic")
		if ok || kind != "" || payload != nil {
			t.Errorf("got %q, %s, %v", kind, payload, ok)
		}
		if hook.LastEntry() == nil || hook.LastEntry().Data["topic"] != "topic" {
			t.Errorf("the failure must be logged with its topic: %+v", hook.LastEntry())
		}
	})
}

func TestParseGroupAbilityAndPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ability, payload, ok := ParseGroupAbilityAndPayload([]byte(`{"ability":"admin","payload":[1,2]}`), "topic")
		if !ok || ability != "admin" || string(payload) != `[1,2]` {
			t.Errorf("got %q, %s, %v", ability, payload, ok)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		hook.Reset()

		ability, payload, ok := ParseGroupAbilityAndPayload([]byte(`nope`), "topic")
		if ok || ability != "" || payload != nil {
			t.Errorf("got %q, %s, %v", ability, payload, ok)
		}
		if hook.LastEntry() == nil {
			t.Error("the failure must be logged")
		}
	})
}

func TestUnmarshalPayload(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var task TaskDeleteUser
		if err := UnmarshalPayload([]byte(`{"user_id":"u"}`), &task); err != nil || task.UserID != "u" {
			t.Errorf("got %+v, %v", task, err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		hook.Reset()

		var task TaskDeleteUser
		if err := UnmarshalPayload([]byte(`{`), &task); err == nil {
			t.Error("expected an error")
		}
		if hook.LastEntry() == nil {
			t.Error("the failure must be logged")
		}
	})
}

func TestSendMessage_UnmarshalablePayload(t *testing.T) {
	q, next := capture(t)
	hook.Reset()

	q.sendMessage("type", make(chan int), TopicMail)

	entry := hook.LastEntry()
	if entry == nil || !strings.Contains(entry.Message, "unable to marshal") {
		t.Errorf("entry = %+v", entry)
	}

	// Nothing was sent: the next message is the one of a valid payload.
	if got := next(func() { q.sendMessage("type", "ok", TopicMail) }); got.message.Payload != "ok" {
		t.Errorf("got %+v", got)
	}
}
