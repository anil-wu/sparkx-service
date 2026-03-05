package storage

import (
	"context"
	"io"
	"time"
)

type ObjectStat struct {
	ContentType string
	SizeBytes   int64
}

type ObjectStore interface {
	PresignPutObject(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error)
	PresignGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, objectKey string) error
	StatObject(ctx context.Context, objectKey string) (*ObjectStat, error)
}

