package fetcher

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// The script defaultExtractCommand is a Lua script that interacts with Redis to fetch tasks from a Redis list.
// It uses the LPOP command to pop tasks from the list until a specified maximum number of tasks max_tasks are fetched,
// or the list is empty, whichever comes first. The Lua script ensures efficient retrieval of tasks while respecting the max limit.
var defaultExtractCommand = redis.NewScript(`
local key = KEYS[1]
local max_tasks = tonumber(ARGV[1])
local tasks = {}

for i = 1, max_tasks do
	local task = redis.call('LPOP', key)
	if not task then
		break
	end
	table.insert(tasks, task)
end

return tasks
`)

// defaultTaskSize defines the maximum number of tasks to be fetched in a single operation.
// This constant ensures that the system will not try to fetch an unreasonably large number of tasks from Redis at once.
const defaultTaskSize = 1000

// RedisFetcher struct provides a redis-backed mechanism for extracting tasks of type T.
// It encapsulates the redis client, a Lua script used for extraction, a transcoder for decoding,
// and a configurable batch size that controls how many tasks are retrieved per operation.
// All fields are configured during construction and are not modified afterward.
type RedisFetcher[T any] struct {
	transcoder     Transcoder[T]
	rdb            redis.UniversalClient
	extractCommand *redis.Script
	size           int
}

// NewRedisFetcher function constructs a fully configured RedisFetcher instance.
// It applies all provided functional options, validates required dependencies,
// and initializes default values for any optional configuration not explicitly set.
// The function returns an error only when mandatory configuration is missing.
func NewRedisFetcher[T any](opts ...options[T]) (*RedisFetcher[T], error) {
	fetcher := &RedisFetcher[T]{}

	for _, opt := range opts {
		opt(fetcher)
	}

	if fetcher.rdb == nil {
		return nil, ErrEmptyRedisClient
	}

	if fetcher.extractCommand == nil {
		fetcher.extractCommand = defaultExtractCommand
	}

	if fetcher.size <= 0 {
		fetcher.size = defaultTaskSize
	}

	if fetcher.transcoder == nil {
		fetcher.transcoder = &defaultTranscoder[T]{}
	}

	return fetcher, nil
}

// Fetch is a method on the RedisFetcher struct that retrieves a list of tasks from Redis based on the provided keys.
// It executes a Lua script using the Redis client to fetch up to a maximum number of tasks from the Redis list.
// The method returns a slice of tasks of type T and an error if any occurred during the operation.
func (f *RedisFetcher[T]) Fetch(ctx context.Context, keys []string) ([]T, error) {
	// Run the Redis Lua script using the provided context, Redis client universal client,
	// and the specified keys, along with the maxTask limit as an argument.
	result, err := f.extractCommand.Run(ctx, f.rdb, keys, f.size).Result()
	// Check if an error occurred during the script execution.
	if err != nil {
		return nil, err
	}

	// Create an empty slice tasks of type T using the make function.
	// T is a generic type, so the actual type of tasks will be determined at runtime.
	// The make function initializes the slice with an initial length of 0, meaning it starts empty.
	// The capacity of the slice will be dynamic, growing as elements are appended to it.
	// This slice will store the tasks fetched and unmarshalled from Redis.
	tasks := make([]T, 0)
	// Check if the result from Redis is a slice of empty interfaces and contains elements.
	// This ensures that the result is in the expected format of a list and is not empty.
	if results, ok := result.([]interface{}); ok && len(results) > 0 {
		// Iterate over each task in the results slice.
		for _, task := range results {
			// Check if the task is of type string.
			// The ok variable indicates whether the type assertion was successful.
			// If successful, the task is stored in the value variable for further processing.
			if value, ok := task.(string); ok {
				// Attempt to unmarshal the task (which is in string format) into the out variable of type T.
				// The task is expected to be in JSON format as a string, so json.Unmarshal is used to decode it.
				res, decodeErr := f.transcoder.Decode(value)
				if decodeErr != nil {
					// If unmarshalling fails, log the error and continue to the next task.
					// This ensures that one failed task does not interrupt the processing of other tasks.
					continue
				}

				// If unmarshalling is successful, append the unmarshalled task to the tasks slice.
				// The task is now an instance of type T, and can be used further in the application.
				tasks = append(tasks, res)
			}
		}
	}

	// After all tasks are processed, the tasks slice will contain all successfully unmarshalled tasks.
	// If no valid tasks were found or unmarshalled, the tasks slice will be empty, which is valid.
	return tasks, nil
}
