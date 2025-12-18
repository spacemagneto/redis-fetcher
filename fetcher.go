package fetcher

import "context"

// Fetcher is a generic interface defining a contract for types that fetch data from a data source.
// It specifies a single method, Fetch, which retrieves tasks of type T and returns them as a slice.
// The generic type T allows implementations to handle any task type specified by the caller without restriction.
// This interface is designed for simplicity, providing direct access to fetched data without partitioning.
type Fetcher[T any] interface {
	// Fetch retrieves a collection of tasks of type T from a data source identified by the provided keys.
	// It returns a slice of T containing all fetched tasks and an error if the operation encounters a failure.
	// The context parameter enables cancellation and timeout management, while keys specify the data source location.
	Fetch(ctx context.Context, keys []string) ([]T, error)
}
