package model

// RateLimitCollection is the name of the collection of the counters of the rate limits, shared by the instances of the
// API. A document counts the requests of a client in a window of time and expires when the window ends.
const RateLimitCollection = "rateLimits"
