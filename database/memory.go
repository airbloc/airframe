package database

import (
	"github.com/pkg/errors"
	"strings"
	"time"
)

type InMemoryDatabase struct {
	objects map[string]*Object
}

func NewInMemoryDatabase() (Database, error) {
	return &InMemoryDatabase{
		objects: map[string]*Object{},
	}, nil
}

func (imdb *InMemoryDatabase) Get(uri string) (*Object, error) {
	pathSegs := strings.Split(uri, "/")
	if len(pathSegs) < 3 {
		return nil, errors.New("invalid URI")
	}
	typ, id := pathSegs[1], pathSegs[2]

	if obj, ok := imdb.objects[typ+"/"+id]; !ok {
		return nil, ErrNotExists
	} else {
		return obj, nil
	}
}

func (imdb *InMemoryDatabase) Exists(typ, id string) (bool, error) {
	_, exists := imdb.objects[typ+"/"+id]
	return exists, nil
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
		obj := imdb.objects[typ+"/"+id]

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
	imdb.objects[typ+"/"+id] = obj
	return &PutResult{
		FeeUsed: 0,
		Created: true,
	}, nil
}
