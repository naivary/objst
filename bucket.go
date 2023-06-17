package main

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/dgraph-io/badger/v4"
	"golang.org/x/exp/slog"
)

// Bucket is the actual object storage
// containing all objects in a flat hierachy.
type Bucket struct {
	store *badger.DB
	names *badger.DB
}

func NewBucket(opts badger.Options) (*Bucket, error) {
	store, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	names, err := badger.Open(badger.DefaultOptions("/tmp/badger/names"))
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		store: store,
		names: names,
	}
	go b.gc()
	return b, nil
}

func (b Bucket) Create(obj *Object) error {
	err := b.store.Update(func(txn *badger.Txn) error {
		if b.nameExists(obj.Name) {
			return fmt.Errorf("object with the name %s already exists", obj.Name)
		}
		data, err := obj.Marshal()
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(obj.ID), data)
		return txn.SetEntry(e)
	})
	if err != nil {
		return err
	}
	return b.insertName(obj.Name, obj.ID)
}

func (b Bucket) BatchCreate(objs []*Object) error {
	wb := b.store.NewWriteBatch()
	defer wb.Cancel()
	for _, obj := range objs {
		if b.nameExists(obj.Name) {
			return fmt.Errorf("object with the name %s exists", obj.Name)
		}
		data, err := obj.Marshal()
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(obj.ID), data)
		if err := wb.SetEntry(e); err != nil {
			return err
		}
		if err := b.insertName(obj.Name, obj.ID); err != nil {
			return err
		}
	}
	return wb.Flush()
}

func (b Bucket) GetByID(id string) (*Object, error) {
	var obj Object
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

func (b Bucket) GetByName(name string) (*Object, error) {
	var id string
	err := b.names.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}
		dst := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(dst); err != nil {
			return err
		}
		id = string(dst)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b.GetByID(id)
}

// GetByMetasOr gets all objects which include at least
// one of the metas provided (logical or)
func (b Bucket) GetByMetasOr(metas url.Values) ([]*Object, error) {
	objs := make([]*Object, 0, 10)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				obj := &Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}
				for k, v := range metas {
					if obj.Meta.Has(k) && obj.Meta.Get(k) == v[0] {
						objs = append(objs, obj)
						break
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return objs, err
}

func (b Bucket) GetByMetasAnd(metas url.Values) ([]*Object, error) {
	objs := make([]*Object, 0, 10)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				var count int
				obj := &Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}
				for k, v := range metas {
					if obj.Meta.Has(k) && obj.Meta.Get(k) == v[0] {
						count++
						continue
					}
					break
				}
				if len(metas) == count {
					objs = append(objs, obj)
				}
				return nil
			})
			if err != nil {
				return err
			}
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

func (b Bucket) nameExists(name string) bool {
	err := b.names.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(name))
		return err
	})
	return !errors.Is(err, badger.ErrKeyNotFound)
}

func (b Bucket) insertName(name, id string) error {
	return b.names.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(name), []byte(id))
	})
}
