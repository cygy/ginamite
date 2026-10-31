package kafka

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// fakeAdmin records the topics it is asked to create.
type fakeAdmin struct {
	created map[string]*sarama.TopicDetail
	errs    map[string]error
	closed  bool
}

func (a *fakeAdmin) CreateTopic(topic string, detail *sarama.TopicDetail, _ bool) error {
	if err := a.errs[topic]; err != nil {
		return err
	}

	a.created[topic] = detail

	return nil
}

func (a *fakeAdmin) Close() error {
	a.closed = true

	return nil
}

func useAdmin(t *testing.T, admin *fakeAdmin, connectionErr error) *[]string {
	t.Helper()

	var addresses []string
	old := newTopicCreator
	newTopicCreator = func(a []string, _ *sarama.Config) (topicCreator, error) {
		addresses = a

		return admin, connectionErr
	}
	t.Cleanup(func() { newTopicCreator = old })

	return &addresses
}

func TestEnsureTopics(t *testing.T) {
	useHook(t)

	spec := TopicSpec{Partitions: 3, ReplicationFactor: 2, Retention: 48 * time.Hour, DeadLetterRetention: 24 * time.Hour}

	t.Run("creates the topics and their dead letters", func(t *testing.T) {
		admin := &fakeAdmin{created: map[string]*sarama.TopicDetail{}}
		addresses := useAdmin(t, admin, nil)

		if err := EnsureTopics("kafka", 9092, Options{}, "app_", []string{"mail", "task"}, spec); err != nil {
			t.Fatal(err)
		}

		if len(*addresses) != 1 || (*addresses)[0] != "kafka:9092" {
			t.Errorf("addresses = %v", *addresses)
		}
		if len(admin.created) != 4 || !admin.closed {
			t.Fatalf("created = %v, closed = %v", admin.created, admin.closed)
		}

		mail, dead := admin.created["app_mail"], admin.created["app_mail_dlq"]
		if mail == nil || mail.NumPartitions != 3 || mail.ReplicationFactor != 2 || *mail.ConfigEntries["retention.ms"] != "172800000" {
			t.Errorf("app_mail = %+v", mail)
		}
		if dead == nil || *dead.ConfigEntries["retention.ms"] != "86400000" {
			t.Errorf("app_mail_dlq = %+v", dead)
		}
	})

	t.Run("minimal settings", func(t *testing.T) {
		admin := &fakeAdmin{created: map[string]*sarama.TopicDetail{}}
		useAdmin(t, admin, nil)

		if err := EnsureTopics("kafka", 9092, Options{}, "", []string{"mail"}, TopicSpec{}); err != nil {
			t.Fatal(err)
		}

		if detail := admin.created["mail"]; detail.NumPartitions != 1 || detail.ReplicationFactor != 1 || detail.ConfigEntries != nil {
			t.Errorf("mail = %+v", detail)
		}
	})

	t.Run("existing topics are kept, the other failures are all returned", func(t *testing.T) {
		boom := errors.New("not authorized")
		admin := &fakeAdmin{created: map[string]*sarama.TopicDetail{}, errs: map[string]error{
			"app_mail":     &sarama.TopicError{Err: sarama.ErrTopicAlreadyExists},
			"app_task":     boom,
			"app_task_dlq": boom,
		}}
		useAdmin(t, admin, nil)

		err := EnsureTopics("kafka", 9092, Options{}, "app_", []string{"mail", "task"}, spec)

		if !errors.Is(err, boom) {
			t.Errorf("error = %v, want the failure to create app_task", err)
		}
		if _, ok := admin.created["app_mail_dlq"]; !ok || len(admin.created) != 1 {
			t.Errorf("created = %v, want only app_mail_dlq", admin.created)
		}
	})

	t.Run("connection failure", func(t *testing.T) {
		useAdmin(t, nil, errors.New("no broker"))

		if err := EnsureTopics("kafka", 9092, Options{}, "", []string{"mail"}, spec); err == nil {
			t.Error("error = nil")
		}
	})
}
