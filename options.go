package fetcher

import "github.com/redis/go-redis/v9"

// options type defines the functional options pattern used to configure a RedisFetcher instance.
type options[T any] func(c *RedisFetcher[T])

// WithClient option assigns the redis client used by the RedisFetcher to communicate with redis.
// This client is responsible for executing commands and Lua scripts against the redis instance.
// Providing a valid redis client is required for the fetcher to function correctly.
// The option stores the client reference directly on the RedisFetcher instance.
func WithClient[T any](rdb redis.UniversalClient) options[T] {
	return func(r *RedisFetcher[T]) {
		r.rdb = rdb
	}
}

// WithTranscoder option configures the transcoder used to decode extracted task data.
// The transcoder is responsible for transforming raw redis data into the target task type.
// Providing a custom transcoder allows callers to control deserialization behavior.
// The configured transcoder is stored on the RedisFetcher for later use during extraction.
func WithTranscoder[T any](t Transcoder[T]) options[T] {
	return func(r *RedisFetcher[T]) {
		r.transcoder = t
	}
}

// WithScript option specifies the Lua script used to extract tasks from redis.
// If no script is provided through this option, the RedisFetcher falls back to its default script.
// This option allows callers to customize extraction logic without modifying the fetcher itself.
// The script reference is stored and later executed during task retrieval.
func WithScript[T any](src *redis.Script) options[T] {
	return func(r *RedisFetcher[T]) {
		r.extractCommand = src
	}
}

// WithTaskSize option configures the maximum number of tasks extracted from redis in a single operation.
// If this option is not provided, the RedisFetcher uses its internal default task size of 1000.
// This option allows callers to control batch size based on workload or performance characteristics.
// The configured size is stored on the RedisFetcher and used during extraction.
func WithTaskSize[T any](size int) options[T] {
	return func(r *RedisFetcher[T]) {
		r.size = size
	}
}
