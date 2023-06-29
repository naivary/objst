package objst

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/naivary/objst/random"
	"golang.org/x/exp/slog"
)

const (
	nameDBDataDir  = "/tmp/badger/names"
	storeDBDataDir = "/tmp/badger/store"
)

// Bucket is the actual object storage
// containing all objects in a flat hierachy.
type Bucket struct {
	store *badger.DB
	names *badger.DB
}

// NewBucket will create a new object storage with the provided options.
func NewBucket(opts *badger.Options) (*Bucket, error) {
	store, err := badger.Open(*opts)
	if err != nil {
		return nil, err
	}
	names, err := badger.Open(badger.DefaultOptions(nameDBDataDir))
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
		e, err := b.createObjectEntry(obj)
		if err != nil {
			return err
		}
		return txn.SetEntry(e)
	})
	if err != nil {
		return err
	}
	return b.insertName(obj.Name(), obj.ID())
}

func (b Bucket) BatchCreate(objs []*Object) error {
	wb := b.store.NewWriteBatch()
	defer wb.Cancel()
	for _, obj := range objs {
		e, err := b.createObjectEntry(obj)
		if err != nil {
			return err
		}
		if err := wb.SetEntry(e); err != nil {
			return err
		}
		if err := b.insertName(obj.Name(), obj.ID()); err != nil {
			return err
		}
		obj.markAsImmutable()
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
	const prefetchSize = 10
	objs := make([]*Object, 0, prefetchSize)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				obj := &Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}
				if b.matchMetaOr(metas, obj) {
					objs = append(objs, obj)
					return nil
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
	const prefetchSize = 10
	objs := make([]*Object, 0, prefetchSize)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				obj := &Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}

				if b.matchMetaAnd(metas, obj) {
					objs = append(objs, obj)
					return nil
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

func (b Bucket) DeleteByID(id string) error {
	err := b.store.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	if err != nil {
		return err
	}
	return b.deleteName(id)
}

func (b Bucket) DeleteByName(name string) error {
	var id string
	b.names.View(func(txn *badger.Txn) error {
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
	return b.DeleteByID(id)
}

func (b Bucket) Shutdown() error {
	if err := b.store.Close(); err != nil {
		return err
	}
	return b.names.Close()
}

func (b Bucket) Health() error {
	owner := uuid.NewString()
	name := fmt.Sprintf("obj_name_%s", owner)
	obj := NewObject(name, owner)
	obj.SetMeta(ContentType, "text/test")
	if _, err := obj.Write([]byte(random.String(5))); err != nil {
		return err
	}
	if err := b.Create(obj); err != nil {
		return err
	}
	_, err := b.GetByID(obj.id)
	if err != nil {
		return err
	}
	if err := b.DeleteByID(obj.id); err != nil {
		return err
	}
	if err := b.store.DropAll(); err != nil {
		return err
	}
	return b.names.DropAll()
}

func (b Bucket) IsAuthorizedByName(owner string, name string) (*Object, error) {
	obj, err := b.GetByName(name)
	if err != nil {
		return nil, err
	}
	if obj.owner != owner {
		return nil, ErrUnauthorized
	}
	return obj, nil
}

func (b Bucket) IsAuthorizedByID(owner string, id string) (*Object, error) {
	obj, err := b.GetByID(id)
	if err != nil {
		return nil, err
	}
	if obj.owner != owner {
		return nil, ErrUnauthorized
	}
	return obj, nil
}

func (b Bucket) RunQuery(q *Query) ([]*Object, error) {
	if !q.isValid() {
		return nil, ErrInvalidQuery
	}

	if q.meta == nil {
		return b.GetByOwner(q.owner)
	}

	objs, err := b.GetByOwner(q.owner)
	if err != nil {
		return nil, err
	}
	return b.FilterByMeta(objs, q.meta, q.act), nil
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

func (b Bucket) deleteName(id string) error {
	return b.names.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				if string(val) == id {
					return txn.Delete(item.Key())
				}
				return nil
			})
		}
		return nil
	})
}

// createObjectEntry validates the object and creates a entry.
// Also the object will be marked as immutable.
func (b Bucket) createObjectEntry(obj *Object) (*badger.Entry, error) {
	if b.nameExists(obj.Name()) {
		return nil, fmt.Errorf("object with the name %s exists", obj.Name())
	}
	data, err := obj.Marshal()
	if err != nil {
		return nil, err
	}
	e := badger.NewEntry([]byte(obj.ID()), data)
	obj.markAsImmutable()
	return e, nil
}

// GetByOwner returns all the objects which the owner created
func (b Bucket) GetByOwner(owner string) ([]*Object, error) {
	const prefetchSize = 10
	objs := make([]*Object, 0, prefetchSize)
	err := b.store.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				obj := &Object{}
				if err := obj.Unmarshal(val); err != nil {
					return err
				}
				if obj.owner == owner {
					objs = append(objs, obj)
					return nil
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

func (b Bucket) FilterByMeta(objs []*Object, metas url.Values, act action) []*Object {
	res := make([]*Object, 0, len(objs))
	if act == Or {
		for _, obj := range objs {
			if b.matchMetaOr(metas, obj) {
				res = append(res, obj)
			}
		}
	}
	if act == And {
		for _, obj := range objs {
			if b.matchMetaAnd(metas, obj) {
				res = append(res, obj)
			}
		}
	}
	return res
}

func (b Bucket) matchMetaOr(metas url.Values, obj *Object) bool {
	for k, v := range metas {
		if obj.meta.Has(k) && obj.meta.Get(k) == v[0] {
			return true
		}
	}
	return false
}

func (b Bucket) matchMetaAnd(metas url.Values, obj *Object) bool {
	for k, v := range metas {
		if !(obj.meta.Has(k) && obj.meta.Get(k) == v[0]) {
			return false
		}
	}
	return true
}
