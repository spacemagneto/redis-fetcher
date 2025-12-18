package fetcher

import "errors"

// ErrEmptyRedisClient is returned when attempting to create a cache without providing a Redis client.
// The Redis client is mandatory for all cache operations — construction fails if it is missing.
var ErrEmptyRedisClient = errors.New("redis client is empty")
