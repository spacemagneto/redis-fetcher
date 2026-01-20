# redis-fetcher

A lightweight, generic Go library for efficiently fetching and decoding tasks from Redis lists using Lua scripts.

The RedisFetcher uses an atomic LPOP approach via Lua to ensure high-performance batch retrieval while respecting defined limits, making it ideal for distributed task processing systems.

## Overview

`golang-fetcher` is a lightweight, generic library that makes it easy and efficient to **fetch batches of tasks** from **Redis lists** in a single atomic operation.

It is especially useful for **distributed task queues** and **worker systems** that need to reliably pop multiple tasks from Redis queues without race conditions.

### Features

- **Generic** – Works with any task type `T` via Go generics.
- **Atomic batch fetching** – Uses a Redis Lua script to safely `LPOP` up to N tasks in one atomic operation.
- **Configurable** – Functional options for Redis client, batch size, custom Lua script, and transcoder.
- **Built-in JSON transcoder** – Default JSON encoding/decoding; easy to replace with custom formats (e.g., MessagePack, Protobuf).
- **Simple interface** – Implements a clean `Fetcher[T]` interface for easy dependency injection and testing.
- **Graceful error handling** – Failed decodes skip individual tasks without aborting the entire batch.
## Installation

```bash
go get github.com/spacemagneto/redis-fetcher
```

## Quick Start

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spacemagneto/redis-fetcher"
)

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	fetcher, err := fetcher.NewRedisFetcher[Task](
		fetcher.WithClient[Task](rdb),
		fetcher.WithTaskSize(100),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		task := Task{ID: i, Name: fmt.Sprintf("task-%d", i)}
		data, _ := json.Marshal(task)
		_ = rdb.RPush(ctx, "my-tasks", string(data)).Err()
	}

	// Fetch up to 1000 tasks from the list "my-tasks"
	tasks, err := fetcher.Fetch(ctx, []string{"my-tasks"})
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tasks {
		fmt.Printf("- %d: %s\n", t.ID, t.Name)
	}
}
```

## Usage
1. Basic usage (JSON by default).
```go
fetcher, err := fetcher.NewRedisFetcher[YourTask](
    fetcher.WithClient(redisClient),
)
```
2. Custom batch size.
```go
fetcher, err := fetcher.NewRedisFetcher[YourTask](
	fetcher.WithClient(redisClient),
	fetcher.WithTaskSize(500), // fetch up to 500 tasks at once
)
```
3. Custom transcoder (Protobuf, MessagePack, custom format...).
```go
type ProtobufTranscoder[T any] struct{}

func (t *ProtobufTranscoder[T]) Decode(data string) (T, error) { ... }

fetcher, err := fetcher.NewRedisFetcher[YourTask](
	fetcher.WithClient(redisClient),
	fetcher.WithTranscoder(&ProtobufTranscoder[YourTask]{}),
)
```
4. Custom Lua extraction script.
```go
var customScript = redis.NewScript(`
    -- your custom logic here
    local key = KEYS[1]
    local max = tonumber(ARGV[1])
    ...
`)

fetcher, err := fetcher.NewRedisFetcher[YourTask](
	fetcher.WithClient(redisClient),
	fetcher.WithScript(customScript),
)
```

# Interface
```go
type Fetcher[T any] interface {
	Fetch(ctx context.Context, keys []string) ([]T, error)
}
```




-------------------------------------------------
# License

This package is licensed under the Apache License, Version 2.0. See the LICENSE file for details.