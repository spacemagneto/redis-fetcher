package fetcher

import "errors"

// ErrEmptyRedisClient is returned when attempting to create a fetcher without providing a Redis client.
// The Redis client is mandatory for all fetcher operations — construction fails if it is missing.
var ErrEmptyRedisClient = errors.New("redis client is empty")
