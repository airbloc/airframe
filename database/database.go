package database

import (
	"context"
	"github.com/airbloc/airframe/auth"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrInvalidID = errors.New("invalid ID format.")
	ErrNotExists = errors.New("given object not exists.")

	// ErrNotAuthorized is raised when given signature mismatched with the object owner's one.
	ErrNotAuthorized = errors.New("you're not authorized to update the object.")
)

// Payload is a shorthand of `map[string]interface{}`.
type Payload map[string]interface{}

type Database interface {
	Get(ctx context.Context, typ, id string) (*Object, error)
	Exists(ctx context.Context, typ, id string) (bool, error)
	Query(ctx context.Context, typ string, query *Query, skip, limit int) ([]*Object, error)
	Put(ctx context.Context, typ, id string, data Payload, signature []byte) (*PutResult, error)
}

type Object struct {
	ID   string
	Type string `dynamo:"-"`
	Data Payload

	Owner auth.PublicKey

	// timestamps
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

type PutResult struct {
	FeeUsed uint64
	Created bool
}
