package metrics

import (
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// requestKey identifies the requests counted together. The route is the pattern of the route, not the path, so the
// number of series is bounded.
type requestKey struct {
	method, route string
	status        int
}

type requestTotal struct {
	count   uint64
	seconds float64
}

// Requests counts the HTTP requests handled by a gin engine, and their duration, by method, route and status.
type Requests struct {
	mu     sync.Mutex
	totals map[requestKey]*requestTotal
}

// NewRequests returns a collector without any request.
func NewRequests() *Requests {
	return &Requests{totals: map[requestKey]*requestTotal{}}
}

// Observe counts a request which lasted d. An unknown route, empty, is counted as "unmatched".
func (r *Requests) Observe(method, route string, status int, d time.Duration) {
	if route == "" {
		route = "unmatched"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := requestKey{method, route, status}
	total := r.totals[key]
	if total == nil {
		total = &requestTotal{}
		r.totals[key] = total
	}
	total.count++
	total.seconds += d.Seconds()
}

// Middleware counts the requests which go through it.
func (r *Requests) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		r.Observe(c.Request.Method, c.FullPath(), c.Writer.Status(), time.Since(start))
	}
}

// Collect writes http_requests_total and http_request_duration_seconds_total.
func (r *Requests) Collect(w *Writer) {
	r.mu.Lock()
	keys := make([]requestKey, 0, len(r.totals))
	totals := make(map[requestKey]requestTotal, len(r.totals))
	for key, total := range r.totals {
		keys = append(keys, key)
		totals[key] = *total
	}
	r.mu.Unlock()

	slices.SortFunc(keys, func(a, b requestKey) int {
		switch {
		case a.route != b.route:
			return compare(a.route, b.route)
		case a.method != b.method:
			return compare(a.method, b.method)
		}

		return a.status - b.status
	})

	for _, key := range keys {
		w.Counter("http_requests_total", "HTTP requests handled.", float64(totals[key].count),
			"method", key.method, "route", key.route, "status", strconv.Itoa(key.status))
	}
	for _, key := range keys {
		w.Counter("http_request_duration_seconds_total", "Time spent handling the HTTP requests.", totals[key].seconds,
			"method", key.method, "route", key.route, "status", strconv.Itoa(key.status))
	}
}

func compare(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}

	return 0
}
