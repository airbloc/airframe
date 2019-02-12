package database

import (
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

type Database interface {
	Get(uri string) (*Object, error)
	Exists(typ, id string) (bool, error)
	Put(typ, id, data string, signature []byte) (*PutResult, error)
}

type Object struct {
	ID   string
	Type string
	Data string

	Owner PublicKey

	// timestamps
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

type PutResult struct {
	FeeUsed uint64
	Created bool
}
