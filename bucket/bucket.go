package bucket

import (
	"net/url"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/naivary/objst/object"
	"golang.org/x/exp/slog"
)

// Bucket is the actual object storage
// containing all objects in a flat hierachy.
type Bucket struct {
	store *badger.DB
}

func New(opts badger.Options) (*Bucket, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		store: db,
	}
	go b.gc()
	return b, nil
}

func (b Bucket) Create(obj *object.Object) error {
	return b.store.Update(func(txn *badger.Txn) error {
		data, err := obj.Marshal()
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(obj.ID), data)
		return txn.SetEntry(e)
	})
}

func (b Bucket) BatchCreate(objs []*object.Object) error {
	wb := b.store.NewWriteBatch()
	defer wb.Cancel()
	for _, obj := range objs {
		data, err := obj.Marshal()
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(obj.ID), data)
		if err := wb.SetEntry(e); err != nil {
			return err
		}
	}
	return wb.Flush()
}

func (b Bucket) Get(id string) (*object.Object, error) {
	var obj object.Object
	err := b.store.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		data := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(data); err != nil {
			return err
		}
		return obj.Unmarshal(data)
	})
	return &obj, err
}

func (b Bucket) GetByMetasOr(metas url.Values) ([]*object.Object, error) {
	objs := make([]*object.Object, 0, 10)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			return it.Item().Value(func(val []byte) error {
				obj := &object.Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}
				for k, v := range metas {
					// its save to assume that v[0] exists because
					// metadata will only be replaced and not appended
					// in a object
					if obj.Meta.Has(k) && obj.Meta.Get(k) == v[0] {
						objs = append(objs, obj)
						return nil
					}
				}
				return nil
			})
		}
		return nil
	})
	return objs, err
}

func (b Bucket) Delete(id string) error {
	return b.store.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

// gc garbace collects every 10 minutes
// the values of the key value store.
func (b Bucket) gc() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		if err := b.store.Close(); err != nil {
			slog.Error("something went wrong", slog.String("msg", err.Error()))
			return
		}
		ticker.Stop()
		if err := b.store.RunValueLogGC(0.7); err != nil {
			slog.Error("something went wrong", slog.String("msg", err.Error()))
			return
		}
	}
}
