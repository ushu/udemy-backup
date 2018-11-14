package backup

import (
	"context"
)

type key int

const (
	workerPoolKey key = iota
)

func NewContext(ctx context.Context, workerPool *Pool) context.Context {
	return context.WithValue(ctx, workerPoolKey, workerPool)
}

func FromContext(ctx context.Context) (*Pool, bool) {
	pool, ok := ctx.Value(workerPoolKey).(*Pool)
	return pool, ok
}
