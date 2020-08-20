package api

// Cache : cached results of some API calls.
type Cache struct {
	values map[string]map[string]CachedValue // indexes : key > locale > get the value
}

// CachedValue : cached result of a API call.
type CachedValue struct {
	Value  interface{}
	Result Result
	Ok     bool
}

// NewCache : returns a new 'Cache' struct.
func NewCache() (c *Cache) {
	c = new(Cache)
	c.values = map[string]map[string]CachedValue{}

	return
}

// Save : saves the value to the cache.
func (c *Cache) Save(key, locale string, value interface{}, result Result, ok bool) {
	if _, ok := c.values[key]; !ok {
		c.values[key] = map[string]CachedValue{}
	}

	c.values[key][locale] = CachedValue{
		Value:  value,
		Result: result,
		Ok:     ok,
	}
}

// Get : gets a cached value.
func (c *Cache) Get(key, locale string) (CachedValue, bool) {
	value, ok := c.values[key][locale]
	return value, ok
}
