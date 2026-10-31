package dlq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/kafka"

	"github.com/IBM/sarama"
)

// idleTimeout is the longest wait for a message which is known to exist; the tests replace it.
var idleTimeout = 30 * time.Second

// backend is the Backend of Kafka.
type backend struct {
	client   sarama.Client
	consumer sarama.Consumer
	offsets  sarama.OffsetManager
	producer sarama.SyncProducer
	managers map[string]sarama.PartitionOffsetManager
}

// Open connects to Kafka. The position of the tool is saved in the consumer group named group, with the prefix of the
// topics if needed.
func Open(host string, port int, options kafka.Options, group string) (Backend, error) {
	config := options.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewClient([]string{fmt.Sprintf("%s:%d", host, port)}, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to kafka: %w", err)
	}

	b := &backend{client: client, managers: map[string]sarama.PartitionOffsetManager{}}
	if b.consumer, err = sarama.NewConsumerFromClient(client); err != nil {
		_ = b.Close()

		return nil, err
	}
	if b.offsets, err = sarama.NewOffsetManagerFromClient(group, client); err != nil {
		_ = b.Close()

		return nil, err
	}
	if b.producer, err = sarama.NewSyncProducerFromClient(client); err != nil {
		_ = b.Close()

		return nil, err
	}

	return b, nil
}

func (b *backend) Partitions(topic string) ([]int32, error) {
	return b.client.Partitions(topic)
}

// manager returns the manager of the position of the partition.
func (b *backend) manager(topic string, partition int32) (sarama.PartitionOffsetManager, error) {
	key := fmt.Sprintf("%s/%d", topic, partition)
	if pom, ok := b.managers[key]; ok {
		return pom, nil
	}

	pom, err := b.offsets.ManagePartition(topic, partition)
	if err != nil {
		return nil, err
	}
	b.managers[key] = pom

	return pom, nil
}

func (b *backend) Position(topic string, partition int32) (int64, int64, error) {
	oldest, err := b.client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return 0, 0, err
	}
	end, err := b.client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return 0, 0, err
	}

	pom, err := b.manager(topic, partition)
	if err != nil {
		return 0, 0, err
	}

	// The position is the oldest message if none is saved, or if the messages it was before were deleted.
	next, _ := pom.NextOffset()

	return max(next, oldest), end, nil
}

func (b *backend) Read(ctx context.Context, topic string, partition int32, from, to int64, fn func(*sarama.ConsumerMessage) error) error {
	partitionConsumer, err := b.consumer.ConsumePartition(topic, partition, from)
	if err != nil {
		return fmt.Errorf("unable to read %s[%d]: %w", topic, partition, err)
	}
	defer partitionConsumer.AsyncClose()

	for {
		select {
		case message, ok := <-partitionConsumer.Messages():
			if !ok {
				return errors.New("the connection to kafka closed")
			}
			if message.Offset >= to {
				return nil
			}
			if err := fn(message); err != nil {
				return err
			}
			if message.Offset+1 >= to {
				return nil
			}

		case err := <-partitionConsumer.Errors():
			return fmt.Errorf("unable to read %s[%d]: %w", topic, partition, err.Err)

		case <-time.After(idleTimeout):
			return fmt.Errorf("no message in %s[%d] since %v", topic, partition, idleTimeout)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *backend) Commit(topic string, partition int32, next int64) error {
	pom, err := b.manager(topic, partition)
	if err != nil {
		return err
	}

	pom.MarkOffset(next, "")
	b.offsets.Commit()

	return nil
}

func (b *backend) Publish(ctx context.Context, message *sarama.ProducerMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, _, err := b.producer.SendMessage(message)

	return err
}

func (b *backend) Close() error {
	var errs []error

	for _, pom := range b.managers {
		errs = append(errs, pom.Close())
	}
	if b.offsets != nil {
		errs = append(errs, b.offsets.Close())
	}
	if b.producer != nil {
		errs = append(errs, b.producer.Close())
	}
	if b.consumer != nil {
		errs = append(errs, b.consumer.Close())
	}

	return errors.Join(append(errs, b.client.Close())...)
}
