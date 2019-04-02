package database

import (
	"bytes"
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

func (imdb *InMemoryDatabase) Get(typ, id string) (*Object, error) {
	if objects, ok := imdb.objects[typ]; ok {
		if obj, ok := objects[id]; ok {
			return obj, nil
		}
	}
	return nil, ErrNotExists
}

func (imdb *InMemoryDatabase) Exists(typ, id string) (bool, error) {
	if objects, ok := imdb.objects[typ]; ok {
		if _, ok := objects[id]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (imdb *InMemoryDatabase) Query(typ string, q *Query, skip, limit int) (results []*Object, err error) {
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

func (imdb *InMemoryDatabase) Put(typ, id string, data Payload, signature []byte) (*PutResult, error) {
	if strings.Contains(id, "/") {
		return nil, ErrInvalidID
	}

	exists, err := imdb.Exists(typ, id)
	if err != nil {
		return nil, errors.Wrap(err, "error while checking existence")
	}

	if exists {
		// update object
		obj := imdb.objects[typ][id]

		tmpObj := &Object{
			Type:  typ,
			ID:    id,
			Data:  data,
			Owner: obj.Owner,
		}
		if !IsOwner(tmpObj, signature) {
			return nil, ErrNotAuthorized
		}
		obj.Data = data
		obj.LastUpdatedAt = time.Now()

		return &PutResult{
			FeeUsed: 0,
			Created: false,
		}, nil
	}

	// create new
	obj := &Object{
		ID:   id,
		Type: typ,
		Data: data,

		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
	if obj.Owner, err = GetOwnerFromSignature(obj, signature); err != nil {
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
}
