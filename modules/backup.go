package modules

import "context"

type Config struct {
}

type Database interface {
	Connect(ctx context.Context) error
	Backup(ctx context.Context) error
}
