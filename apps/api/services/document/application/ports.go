package application

import (
	"context"
	"time"

	"github.com/Briyantama/stery-kit/go/objstore"
)

// ObjectStore is the part of object storage the document use cases need.
type ObjectStore interface {
	PresignPut(ctx context.Context, key, contentType string, size int64, ttl time.Duration) (objstore.PresignedRequest, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}
