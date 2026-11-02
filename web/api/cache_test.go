package api

import (
	"fmt"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	cache := NewCache()

	if _, ok := cache.Get("key", "en-us"); ok {
		t.Error("an empty cache has no value")
	}

	result := Result{StatusCode: 200}
	cache.Save("key", "en-us", "value", result, true)
	cache.Save("key", "fr-fr", "valeur", result, false)

	tests := []struct {
		key, locale string
		want        any
		wantOk      bool
		found       bool
	}{
		{"key", "en-us", "value", true, true},
		{"key", "fr-fr", "valeur", false, true},
		{"key", "de-de", nil, false, false},
		{"other", "en-us", nil, false, false},
	}
	for _, tt := range tests {
		got, found := cache.Get(tt.key, tt.locale)

		if found != tt.found || got.Value != tt.want || got.Ok != tt.wantOk {
			t.Errorf("Get(%q, %q) = %+v, %v", tt.key, tt.locale, got, found)
		}
	}

	cache.Save("key", "en-us", "new", result, true)
	if got, _ := cache.Get("key", "en-us"); got.Value != "new" {
		t.Errorf("Value = %v, want it replaced", got.Value)
	}
}

func TestCache_Concurrent(t *testing.T) {
	cache := NewCache()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i%5)
			cache.Save(key, "en-us", i, Result{}, true)
			cache.Get(key, "en-us")
		}()
	}
	wg.Wait()
}
