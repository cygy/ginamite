// Package ratelimit is the base of the rate limiting plugins, in the subpackages mongodb and redis: they keep the
// counters of the rate limits of the API, shared by its instances. Without such a plugin, each instance of the API counts
// the requests in memory.
package ratelimit

import (
	"fmt"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/common/log"
)

// Register sets newLimiter as the limiters of the API of the server, which keep their counters in the storage with the
// name, such as "redis". It returns an error if a rate limiting plugin already registered its limiters: only one
// storage can count the requests.
func Register(s *ginamite.Server, store string, newLimiter request.NewRateLimiterFunc) error {
	if s.API.Hooks.NewRateLimiter != nil {
		return fmt.Errorf("two rate limiting plugins are loaded (%s and %s): keep only one", s.API.Hooks.RateLimiterStore, store)
	}

	s.API.Hooks.NewRateLimiter = newLimiter
	s.API.Hooks.RateLimiterStore = store
	log.Info("rate limiting plugin loaded", "store", store)

	return nil
}
