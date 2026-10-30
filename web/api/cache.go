package api

import "sync"

// Cache keeps the results of some API calls by key and locale. It is safe for concurrent use.
type Cache struct {
	mutex  sync.RWMutex
	values map[string]map[string]CachedValue // indexes : key > locale > get the value
}

// CachedValue is the cached result of an API call.
type CachedValue struct {
	Value  any
	Result Result
	Ok     bool
}

// NewCache returns an empty Cache.
func NewCache() (c *Cache) {
	c = new(Cache)
	c.values = map[string]map[string]CachedValue{}

	return
}

// Save saves the value of the key and the locale to the cache, replacing the previous one.
func (c *Cache) Save(key, locale string, value any, result Result, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.values[key]; !exists {
		c.values[key] = map[string]CachedValue{}
	}

	c.values[key][locale] = CachedValue{
		Value:  value,
		Result: result,
		Ok:     ok,
	}
}

// Get returns the cached value of the key and the locale, and whether there is one.
func (c *Cache) Get(key, locale string) (CachedValue, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.values[key][locale]

	return value, ok
}
