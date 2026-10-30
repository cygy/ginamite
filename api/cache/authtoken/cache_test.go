package authtoken

import (
	"os"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

func TestCache_AddAndHas(t *testing.T) {
	c := newCache()
	c.add(map[string]time.Time{"a": time.Now().Add(time.Hour)})

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"known token", "a", true},
		{"unknown token", "b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.has(tt.token); got != tt.want {
				t.Errorf("has(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestCache_Remove(t *testing.T) {
	c := newCache()
	exp := time.Now().Add(time.Hour)
	c.add(map[string]time.Time{"a": exp, "b": exp, "c": exp})

	c.remove("a", "b", "missing")

	if c.has("a") || c.has("b") {
		t.Error("removed tokens are still present")
	}
	if !c.has("c") {
		t.Error("untouched token was removed")
	}
}

func TestCache_DeleteExpired(t *testing.T) {
	now := time.Now()
	c := newCache()
	c.add(map[string]time.Time{
		"expired": now.Add(-time.Minute),
		"valid":   now.Add(time.Minute),
	})

	if got := c.deleteExpired(now); got != 1 {
		t.Errorf("deleteExpired() = %d, want 1", got)
	}
	if c.has("expired") {
		t.Error("expired token still present")
	}
	if !c.has("valid") {
		t.Error("valid token was removed")
	}
	if got := c.deleteExpired(now); got != 0 {
		t.Errorf("second deleteExpired() = %d, want 0", got)
	}
}

func TestCache_RunDeleteExpired(t *testing.T) {
	c := newCache()
	c.add(map[string]time.Time{"expired": time.Now().Add(-time.Minute)})

	go c.runDeleteExpired(5 * time.Millisecond)

	deadline := time.Now().Add(2 * time.Second)
	for c.has("expired") {
		if time.Now().After(deadline) {
			t.Fatal("expired token was not removed by the background task")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := newCache()
	done := make(chan struct{})
	for range 4 {
		go func() {
			defer func() { done <- struct{}{} }()
			for range 200 {
				c.add(map[string]time.Time{"t": time.Now()})
				c.has("t")
				c.deleteExpired(time.Now().Add(time.Second))
				c.remove("t")
			}
		}()
	}
	for range 4 {
		<-done
	}
}

func TestNewCache(t *testing.T) {
	c := newCache()
	if c == nil || c.tokens == nil {
		t.Fatal("newCache() returned an unusable cache")
	}
	if c.has("anything") {
		t.Error("new cache is not empty")
	}
}
