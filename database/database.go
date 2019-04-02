package database

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

var (
	// ErrInvalidURI is returned when given URI query format is invalid.
	ErrInvalidURI = errors.New("invalid URI.")

	ErrInvalidID = errors.New("invalid ID format.")
	ErrNotExists = errors.New("given object not exists.")

	// ErrNotAuthorized is raised when given signature mismatched with the object owner's one.
	ErrNotAuthorized = errors.New("you're not authorized to update the object.")
)

// PublicKey is 33-byte compressed ECDSA public key.
type PublicKey [33]byte

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
	Type string
	Data Payload

	Owner PublicKey

	// timestamps
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

type PutResult struct {
	FeeUsed uint64
	Created bool
}
