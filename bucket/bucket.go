package bucket

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/naivary/objst/logger"
	"github.com/naivary/objst/object"
)

// Bucket is the actual object storage
// containing all objects in a flat hierachy.
type Bucket struct {
	store  *badger.DB
	logger *logger.Logger
}

func New(opts badger.Options) (*Bucket, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		store:  db,
		logger: logger.New(context.Background()),
	}
	return b, nil
}

func (b Bucket) Create(obj *object.Object) error {
	return b.store.Update(func(txn *badger.Txn) error {
		return nil
	})
}
