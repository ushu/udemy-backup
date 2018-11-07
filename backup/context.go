package backup

import (
	"context"

	"github.com/ushu/udemy-backup/client"
)

type ContextKey string

var (
	ClientKey ContextKey = "github.com/ushu/udemy-backup/backup.ClientKey"
)

func SetClient(ctx context.Context, c *client.Client) context.Context {
	return context.WithValue(ctx, ClientKey, c)
}

func GetClient(ctx context.Context) *client.Client {
	return ctx.Value(ClientKey).(*client.Client)
}
