package objst

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

const (
	basePath = "/var/lib/objst"
	dataDir  = "data"
	nameDir  = "name"
	metaDir  = "meta"
)

type Bucket struct {
	// store persists the objects and the
	// actual data the client will interact with.
	payload *badger.DB
	// names is a helper db, storing the different names
	// of the objects. It assures a quick and easy way to
	// check if a names exists, without unmarshaling the
	// objects.
	names *badger.DB

	meta *badger.DB

	uniqueBasePath string
}

// NewBucket will create a new object storage with the provided options.
// The `Dir` option will be overwritten by the application to have
// a gurantee about the data path.
func NewBucket(opts badger.Options) (*Bucket, error) {
	uniqueBasePath := filepath.Join(basePath, uuid.NewString())
	payloadDataDir := filepath.Join(uniqueBasePath, dataDir)
	opts.Dir = payloadDataDir
	opts.ValueDir = payloadDataDir
	payload, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	nameDataDir := filepath.Join(uniqueBasePath, nameDir)
	names, err := badger.Open(badger.DefaultOptions(nameDataDir))
	if err != nil {
		return nil, err
	}
	metaDataDir := filepath.Join(uniqueBasePath, metaDir)
	meta, err := badger.Open(badger.DefaultOptions(metaDataDir))
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		payload:        payload,
		names:          names,
		meta:           meta,
		uniqueBasePath: uniqueBasePath,
	}
	return b, nil
}

func (b Bucket) Get(q Query) ([]*Object, error) {
	return nil, nil
}

// Create inserts the given object into the storage.
// If you have to create multiple objects use
// `BatchCreate` which is more performant than
// multiple calls to Create.
func (b Bucket) Create(obj *Object) error {
	err := b.payload.Update(func(txn *badger.Txn) error {
		e, err := b.createObjectEntry(obj)
		if err != nil {
			return err
		}
		if err := txn.SetEntry(e); err != nil {
			return err
		}
		obj.markAsImmutable()
		return nil
	})
	if err != nil {
		return err
	}
	if err := b.insertName(obj.Name(), obj.Owner(), obj.ID()); err != nil {
		return err
	}
	return b.insertMeta(obj.ID(), obj.meta)
}

// BatchCreate inserts multiple objects in an efficient way.
func (b Bucket) BatchCreate(objs []*Object) error {
	wb := b.payload.NewWriteBatch()
	defer wb.Cancel()
	for _, obj := range objs {
		e, err := b.createObjectEntry(obj)
		if err != nil {
			return err
		}
		if err := wb.SetEntry(e); err != nil {
			return err
		}
		if err := b.insertName(obj.Name(), obj.Owner(), obj.ID()); err != nil {
			return err
		}
		if err := b.insertMeta(obj.ID(), obj.meta); err != nil {
			return err
		}
		obj.markAsImmutable()
	}
	return wb.Flush()
}

func (b Bucket) GetByID(id string) (*Object, error) {
	return b.composeObjectByID(id)
}

func (b Bucket) GetByName(name, owner string) (*Object, error) {
	var id string
	err := b.names.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(b.nameFormat(name, owner)))
		if err != nil {
			return err
		}
		data := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(data); err != nil {
			return err
		}
		id = string(data)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b.GetByID(id)
}

func Get(meta Metadata, act action, limit int) ([]*Object, error) {
	return nil, nil
}

func (b Bucket) GetByMeta(meta Metadata, act action) ([]*Object, error) {
	const prefetchSize = 10
	objs := make([]*Object, 0, prefetchSize)
	err := b.meta.View(func(txn *badger.Txn) error {
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
				if act == Or {
					if b.matchMetaOr(meta, obj) {
						objs = append(objs, obj)
						return nil
					}
				} else if act == And {
					if b.matchMetaAnd(meta, obj) {
						objs = append(objs, obj)
						return nil
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

func (b Bucket) DeleteByID(id string) error {
	err := b.payload.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	if err != nil {
		return err
	}
	if err := b.deleteName(id); err != nil {
		return err
	}
	return b.deleteMeta(id)
}

func (b Bucket) DeleteByName(name, owner string) error {
	var id string
	b.names.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(b.nameFormat(name, owner)))
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
	if err := b.payload.Close(); err != nil {
		return err
	}
	if err := b.meta.Close(); err != nil {
		return err
	}
	return b.names.Close()
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
	return b.FilterByMeta(objs, *q.meta, q.act), nil
}

func (b Bucket) nameExists(name, owner string) bool {
	err := b.names.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(b.nameFormat(name, owner)))
		return err
	})
	return !errors.Is(err, badger.ErrKeyNotFound)
}

func (b Bucket) insertName(name, owner, id string) error {
	return b.names.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(b.nameFormat(name, owner)), []byte(id))
	})
}

func (b Bucket) insertMeta(id string, meta *Metadata) error {
	return b.meta.Update(func(txn *badger.Txn) error {
		data, err := meta.Marshal()
		if err != nil {
			return err
		}
		return txn.Set([]byte(id), data)
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

func (b Bucket) deleteMeta(id string) error {
	return b.meta.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (b Bucket) nameFormat(name, owner string) string {
	// choosing the name format as <name>_<owner> allows
	// to have unique names in the context of a owner e.g.
	// owner 1 can have name_1 and owner 2 can have name_2.
	return fmt.Sprintf("%s_%s", name, owner)
}

// createObjectEntry validates the object and creates a entry.
// Also the object will be marked as immutable.
func (b Bucket) createObjectEntry(obj *Object) (*badger.Entry, error) {
	if b.nameExists(obj.Name(), obj.Owner()) {
		return nil, fmt.Errorf("object with the name %s for the owner %s exists", obj.Name(), obj.Owner())
	}
	data, err := obj.Marshal()
	if err != nil {
		return nil, err
	}
	e := badger.NewEntry([]byte(obj.ID()), data)
	return e, nil
}

// GetByOwner returns all the objects which the owner created
func (b Bucket) GetByOwner(owner string) ([]*Object, error) {
	const prefetchSize = 10
	objs := make([]*Object, 0, prefetchSize)
	err := b.meta.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				m := NewMetadata()
				if err := m.Unmarshal(val); err != nil {
					return err
				}
				if m.Get(MetaKeyOwner) != owner {
					return nil
				}
				obj, err := b.composeObject(m)
				if err != nil {
					return err
				}
				objs = append(objs, obj)
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

func (b Bucket) composeObject(meta *Metadata) (*Object, error) {
	id := meta.Get(MetaKeyID)
	obj := &Object{
		meta: meta,
	}
	err := b.payload.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		dst := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(dst); err != nil {
			return err
		}
		return obj.Unmarshal(dst)
	})
	return obj, err
}

func (b Bucket) composeObjectByID(id string) (*Object, error) {
	m := NewMetadata()
	err := b.meta.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		dst := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(dst); err != nil {
			return err
		}
		return m.Unmarshal(dst)
	})
	if err != nil {
		return nil, err
	}
	return b.composeObject(m)
}

func (b Bucket) FilterByMeta(objs []*Object, meta Metadata, act action) []*Object {
	res := make([]*Object, 0, len(objs))
	if act == Or {
		for _, obj := range objs {
			if b.matchMetaOr(meta, obj) {
				res = append(res, obj)
			}
		}
	}
	if act == And {
		for _, obj := range objs {
			if b.matchMetaAnd(meta, obj) {
				res = append(res, obj)
			}
		}
	}
	return res
}

func (b Bucket) matchMetaOr(meta Metadata, obj *Object) bool {
	for k := range meta.data {
		if obj.HasMetaKey(k) {
			return true
		}
	}
	return false
}

func (b Bucket) matchMetaAnd(meta Metadata, obj *Object) bool {
	for k := range meta.data {
		if !obj.HasMetaKey(k) {
			return false
		}
	}
	return true
}
