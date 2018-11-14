package config

import (
	"context"
)

type key int

const (
	configKey key = iota
)

func NewContext(ctx context.Context, workerPool *Config) context.Context {
	return context.WithValue(ctx, configKey, workerPool)
}

func FromContext(ctx context.Context) (*Config, bool) {
	pool, ok := ctx.Value(configKey).(*Config)
	return pool, ok
}
