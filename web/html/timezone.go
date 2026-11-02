package html

import (
	"sync"
	"time"
)

var (
	locations      = map[string]*time.Location{}
	locationsMutex sync.RWMutex
)

// timeInTimezone returns t in the timezone tz, and an error if the timezone is unknown, in which case t is unchanged.
// The locations are cached.
func timeInTimezone(t time.Time, tz string) (time.Time, error) {
	locationsMutex.RLock()
	location, ok := locations[tz]
	locationsMutex.RUnlock()

	if !ok {
		var err error
		if location, err = time.LoadLocation(tz); err != nil {
			return t, err
		}

		locationsMutex.Lock()
		locations[tz] = location
		locationsMutex.Unlock()
	}

	return t.In(location), nil
}
