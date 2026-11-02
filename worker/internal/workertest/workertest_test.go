package workertest

import (
	"testing"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
)

func TestCaptureQueue(t *testing.T) {
	q := CaptureQueue(t)
	q.Empty()

	q.Sender.DeleteUser(queue.TaskDeleteUser{UserID: "u"})

	m, ok := q.Next()
	if !ok || m.Topic != queue.TopicTaskUser || m.Type != queue.MessageTaskDeleteUser || m.Payload["user_id"] != "u" {
		t.Errorf("message = %+v, %v", m, ok)
	}

	if _, ok := q.Next(); ok {
		t.Error("only one message was sent")
	}
}

func TestHook(t *testing.T) {
	Hook.Reset()
	log.Info("hello")

	if entry := Hook.LastEntry(); entry == nil || entry.Message != "hello" {
		t.Errorf("entry = %+v", entry)
	}
}
