package broker

import "sync"

// Deduplicator remembers the messages a group handled, to ignore a message delivered again: Kafka delivers a message
// at least once, and again if the process stops after the handling and before the commit.
//
// A memory implementation, see NewMemoryDeduplicator, covers the redeliveries within a process, such as the ones of
// a rebalance. An implementation backed by a database, shared by the workers, also covers a restart. The messages
// are remembered after being handled, so a crash in between still handles a message twice: the handlers should
// stay idempotent. The implementations must be safe for concurrent use.
type Deduplicator interface {
	// Handled reports whether the group already handled the message.
	Handled(group, id string) bool
	// MarkHandled remembers that the group handled the message.
	MarkHandled(group, id string)
}

// MemoryDeduplicator is a Deduplicator remembering the last messages in memory.
type MemoryDeduplicator struct {
	mu   sync.Mutex
	seen map[string]struct{}
	ring []string
	next int
}

// NewMemoryDeduplicator returns a Deduplicator remembering the last size messages, 1 at least.
func NewMemoryDeduplicator(size int) *MemoryDeduplicator {
	return &MemoryDeduplicator{seen: map[string]struct{}{}, ring: make([]string, max(size, 1))}
}

// Handled reports whether the group already handled the message.
func (d *MemoryDeduplicator) Handled(group, id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.seen[group+"/"+id]

	return ok
}

// MarkHandled remembers that the group handled the message, and forgets the oldest one when it is full.
func (d *MemoryDeduplicator) MarkHandled(group, id string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := group + "/" + id
	if _, ok := d.seen[key]; ok {
		return
	}

	if old := d.ring[d.next]; old != "" {
		delete(d.seen, old)
	}
	d.ring[d.next] = key
	d.seen[key] = struct{}{}
	d.next = (d.next + 1) % len(d.ring)
}
