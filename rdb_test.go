package fetcher

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type TestTask struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

func TestFetcher(t *testing.T) {
	// Create a new background context for the operation.
	// This context is typically used when no cancellation, timeout, or specific context values are needed.
	ctx := context.Background()

	redisAddress := os.Getenv("REDIS_ADDRESS")

	// Retrieve the Redis cluster client from the container.
	// NewUniversalClient() is a method that obtains an instance of the Redis client
	// which is used to interact with the Redis cluster in the test environment. This client is necessary
	// for performing operations like setting and retrieving data in Redis.
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{redisAddress}})
	// Ensure that the Redis cluster client is closed when the test function completes.
	// defer ensures that closeFn() is called at the end of the test function,
	// releasing any resources associated with the Redis client. This is important to avoid resource leaks
	// and to properly clean up the Redis client after the test has finished.
	defer rdb.Close()

	// Perform a health check by pinging the Redis server using the provided context.
	// This ensures that the connection to the Redis server is active and functional.
	err := rdb.Ping(ctx).Err()
	// Assert that no error occurred during the ping operation.
	// If an error is returned, it indicates an issue with the Redis connection.
	assert.NoError(t, err, "Expected Redis server to respond to ping without errors")
	// Assert that the Redis client is not nil.
	// This check ensures that the Redis client has been properly initialized and is ready to be used.
	// If `rdb` is nil, it indicates that the Redis client could not be retrieved from the container,
	// which would mean the setup is incomplete or incorrect.
	assert.NotNil(t, rdb, "Expected the Redis client to be initialized, but got nil")

	transcoder := &defaultTranscoder[TestTask]{}

	// Create a new fetcher instance for the TestTask type using the provided Redis client.
	// This ensures that the fetcher is initialized with the necessary dependencies for performing operations.
	fetcher, err := NewRedisFetcher[TestTask](WithClient[TestTask](rdb), WithTranscoder[TestTask](transcoder))
	assert.NoError(t, err)
	// Assert that the fetcher instance is successfully created and not nil.
	// This verifies that the New function has properly initialized the fetcher,
	// ensuring that it is ready to perform its intended operations.
	assert.NotNil(t, fetcher, "Expected fetcher instance to be initialized and not nil")

	// SuccessFetch verifies the behavior of the `Fetch` method when retrieving tasks from Redis.
	// This test ensures that the method correctly fetches tasks, processes them as expected, and
	// validates the task count returned by the Redis query. The goal is to confirm that the `Fetch`
	// method behaves correctly when interacting with Redis and handles tasks appropriately.
	t.Run("SuccessFetch", func(t *testing.T) {
		// Define the Redis key where the test tasks will be stored.
		// This key is used as an identifier to store and later retrieve tasks from Redis.
		testKey := "fetcher.domain.com::test_tasks"
		// Define a set of test tasks with IDs and associated data.
		// These tasks are designed to simulate a real-world task fetching scenario.
		testTasks := []TestTask{{ID: 1, Data: "task1"}, {ID: 2, Data: "task2"}, {ID: 3, Data: "task3"}}

		// Push each test task into the Redis list, simulating adding tasks to the Redis store.
		// Each task is marshaled into JSON format before being pushed to Redis.
		// This mimics how tasks would typically be stored for later processing.
		for _, task := range testTasks {
			// Marshal the task into JSON format.
			// This step converts the task structure into a byte slice, which is the format required for Redis storage.
			// If this operation fails, the test will fail due to the lack of task serialization.
			taskJSON, _ := transcoder.Encode(task)
			// Push the marshaled task into the Redis list at the given testKey.
			// The RPush operation appends the task to the list in Redis, simulating task storage in the cache.
			// If this operation fails, it indicates that the task was not successfully added to Redis.
			err = rdb.RPush(ctx, testKey, taskJSON).Err()
			// Assert that no error occurred during the RPush operation.
			// This verifies that the task was successfully pushed to Redis.
			// If an error is encountered, it suggests an issue with interacting with Redis.
			assert.NoError(t, err, "Failed to push task into Redis")
		}

		// Call the Fetch function to retrieve tasks from Redis using the defined key.
		// This simulates fetching tasks from Redis, where the `Fetch` method should return
		// the tasks that were previously pushed to the Redis list.
		fetchedTasks, fetchErr := fetcher.Fetch(ctx, []string{testKey})
		// Assert that no error occurred while fetching tasks.
		// If an error is encountered, it suggests a problem with retrieving tasks from Redis.
		assert.NoError(t, fetchErr, "Failed to fetch tasks")
		// Assert that the number of fetched tasks matches the number of test tasks pushed to Redis.
		// This confirms that the fetch operation correctly retrieves all the tasks stored in Redis.
		assert.Len(t, fetchedTasks, len(testTasks), "Fetched task count mismatch")
	})

	// EmptyList verifies the behavior of the Fetch method when the Redis list is empty.
	// This test ensures that the fetcher handles the case where no tasks are present in the list.
	// It checks that the method correctly returns an empty result without any errors, which is important
	// to verify that the system can gracefully handle scenarios with no data in the source without causing failures or unexpected behavior.
	t.Run("EmptyList", func(t *testing.T) {
		// Define the Redis key where the test tasks will be stored.
		// This key is used as an identifier to store and later retrieve tasks from Redis.
		testKey := "fetcher.domain.com::empty_list"

		// Fetch tasks from Redis using the defined testKey.
		// Since the list is expected to be empty, this operation should return no tasks.
		fetchedEmptyTasks, fetchErr := fetcher.Fetch(ctx, []string{testKey})
		// Assert that no error occurred during the fetching operation.
		// If an error occurs, it indicates a failure to retrieve tasks from Redis.
		assert.NoError(t, fetchErr, "Failed to fetch empty tasks")
		// Assert that the fetched tasks slice is empty.
		// This ensures that the fetcher correctly handles the scenario where there are no tasks in the Redis list.
		// The length of the fetched tasks slice should be zero.
		assert.Len(t, fetchedEmptyTasks, 0, "Empty task list")
	})
}
