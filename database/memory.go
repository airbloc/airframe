package database

import (
	"bytes"
	"context"
	"github.com/airbloc/airframe/auth"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type InMemoryDatabase struct {
	objects map[string]map[string]*Object
}

func NewInMemoryDatabase() (Database, error) {
	return &InMemoryDatabase{
		objects: make(map[string]map[string]*Object),
	}, nil
}

func (imdb *InMemoryDatabase) Get(ctx context.Context, typ, id string) (*Object, error) {
	if objects, ok := imdb.objects[typ]; ok {
		if obj, ok := objects[id]; ok {
			return obj, nil
		}
	}
	return nil, ErrNotExists
}

func (imdb *InMemoryDatabase) Exists(ctx context.Context, typ, id string) (bool, error) {
	if objects, ok := imdb.objects[typ]; ok {
		if _, ok := objects[id]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (imdb *InMemoryDatabase) Query(ctx context.Context, typ string, q *Query, skip, limit int) (results []*Object, err error) {
	results = []*Object{}

	objects := imdb.objects[typ]
	if objects == nil {
		return
	}
	skipped := 0
	for _, obj := range objects {
		for _, op := range q.Conditions {
			if compare(obj, op) {
				if skipped == skip {
					results = append(results, obj)
				} else {
					skipped++
				}
				if limit > 0 && len(results) == limit {
					return
				}
			}
		}
	}
	return
}

func compare(obj *Object, op Operator) bool {
	fieldVal := obj.Data[op.Field]
	switch op.Type {
	case OpEquals:
		return fieldVal == op.Operand
	case OpGreaterThan:
		return fieldVal.(int) > op.Operand.(int)
	case OpGreaterThanOrEqual:
		return fieldVal.(int) >= op.Operand.(int)
	case OpLessThan:
		return fieldVal.(int) < op.Operand.(int)
	case OpLessThanOrEqual:
		return fieldVal.(int) <= op.Operand.(int)
	case OpContains:
		if fieldValStr, ok := fieldVal.(string); ok {
			return strings.Contains(fieldValStr, op.Operand.(string))
		}
		if fieldValBytes, ok := fieldVal.([]byte); ok {
			return bytes.Contains(fieldValBytes, op.Operand.([]byte))
		}
		for _, elem := range fieldVal.([]interface{}) {
			if elem == op.Operand {
				return true
			}
		}
	}
	return false
}

func (imdb *InMemoryDatabase) Put(ctx context.Context, typ, id string, data Payload, signature []byte) (*PutResult, error) {
	if strings.Contains(id, "/") {
		return nil, ErrInvalidID
	}
	obj, err := imdb.Get(ctx, typ, id)
	if err == ErrNotExists {
		// create new
		obj = &Object{
			ID:   id,
			Type: typ,
			Data: data,

			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}
		if obj.Owner, err = auth.GetSigner(typ, id, data, signature); err != nil {
			return nil, errors.Wrap(err, "invalid signature")
		}
		if _, collectionExists := imdb.objects[typ]; !collectionExists {
			imdb.objects[typ] = make(map[string]*Object)
		}
		imdb.objects[typ][id] = obj
		return &PutResult{
			FeeUsed: 0,
			Created: true,
		}, nil

	} else if obj != nil {
		// update object
		signer, err := auth.GetSigner(typ, id, data, signature)
		if err != nil {
			return nil, errors.Wrap(err, "failed to recover signature")
		}

		// object owners can only update the existing object.
		if !bytes.Equal(signer[:], obj.Owner[:]) {
			return nil, ErrNotAuthorized
		}
		obj.Data = data
		obj.LastUpdatedAt = time.Now()

		return &PutResult{
			FeeUsed: 0,
			Created: false,
		}, nil
	} else {
		return nil, errors.Wrap(err, "error while checking existence")
	}
}
