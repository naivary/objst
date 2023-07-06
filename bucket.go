package objst

import (
	"errors"
	"fmt"
	"io"
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
func NewBucket(opts BucketOptions) (*Bucket, error) {
	uniqueBasePath := filepath.Join(basePath, uuid.NewString())
	payloadDataDir := filepath.Join(uniqueBasePath, dataDir)
	opts.overwriteDataDir(payloadDataDir)
	payload, err := badger.Open(opts.toBadgerOpts())
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

func (b Bucket) Execute(q *Query) ([]*Object, error) {
	if err := q.isValid(); err != nil {
		return nil, err
	}
	// empty object array for operation which
	// do not return any objects
	var defRes []*Object
	name := q.params.Get(MetaKeyName)
	id := q.params.Get(MetaKeyID)
	owner := q.params.Get(MetaKeyOwner)
	// check if it is a multi query
	if q.op == OperationGet && !q.isSingleEntry() {
		return b.Get(q)
	} else if q.op == OperationGet && q.isSingleEntry() {
		// check which kind of single get it is
		objs := make([]*Object, 0, 1)
		var (
			obj *Object
			err error
		)
		if q.isIDIdentifier() {
			obj, err = b.GetByID(id)
		} else {
			obj, err = b.GetByName(name, owner)
		}
		objs = append(objs, obj)
		return objs, err

	}

	if q.op == OperationDelete && !q.isSingleEntry() {
		err := b.Delete(q)
		return defRes, err
	}

	if q.op == OperationDelete && q.isSingleEntry() {
		if q.isIDIdentifier() {
			err := b.DeleteByID(id)
			return defRes, err
		}
		err := b.DeleteByName(name, owner)
		return defRes, err
	}

	return nil, nil
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

func (b Bucket) Get(q *Query) ([]*Object, error) {
	ids, err := b.getMatchingIDs(q)
	if err != nil {
		return nil, err
	}
	return b.idsToObjs(ids)
}

// Create inserts the given object into the storage.
// If you have to create multiple objects use
// `BatchCreate` which is more performant than
// multiple calls to Create.
func (b Bucket) Create(obj *Object) error {
	e, err := b.createObjectEntry(obj)
	if err != nil {
		return err
	}
	if err := b.insertPayload(string(e.Key), e.Value); err != nil {
		return err
	}
	if err := b.insertName(obj.Name(), obj.Owner(), obj.ID()); err != nil {
		return err
	}
	if err := b.insertMeta(obj.ID(), obj.meta); err != nil {
		return err
	}
	obj.markAsImmutable()
	return nil
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

func (b Bucket) Delete(q *Query) error {
	ids, err := b.getMatchingIDs(q)
	if err != nil {
		return nil
	}
	for _, id := range ids {
		if err := b.DeleteByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (b Bucket) GetMeta(id string) (*Metadata, error) {
	meta := NewMetadata()
	err := b.meta.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		dst := make([]byte, item.ValueSize())
		if _, err := item.ValueCopy(dst); err != nil {
			return err
		}
		return meta.Unmarshal(dst)
	})
	return meta, err
}

func (b Bucket) DeleteByID(id string) error {
	meta, err := b.GetMeta(id)
	if err != nil {
		return err
	}
	if err := b.deletePayload(id); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err := b.deleteName(meta.Get(MetaKeyName), meta.Get(MetaKeyOwner)); err != nil {
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

func (b Bucket) Read(id string, w io.Writer) error {
	obj, err := b.GetByID(id)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, obj)
	return err
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

func (b Bucket) getMatchingIDs(q *Query) ([]string, error) {
	const prefetchSize = 10
	ids := make([]string, 0, prefetchSize)
	err := b.meta.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				meta := NewMetadata()
				if err := meta.Unmarshal(val); err != nil {
					return err
				}
				if meta.Compare(q.params, q.act) {
					dst := make([]byte, it.Item().KeySize())
					it.Item().KeyCopy(dst)
					ids = append(ids, string(dst))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return ids, err
}

func (b Bucket) idsToObjs(ids []string) ([]*Object, error) {
	objs := make([]*Object, 0, len(ids))
	for _, id := range ids {
		obj, err := b.composeObjectByID(id)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (b Bucket) isNameExisting(name, owner string) bool {
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

func (b Bucket) insertPayload(id string, pl []byte) error {
	return b.payload.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), pl)
	})
}

func (b Bucket) deleteName(name, owner string) error {
	return b.names.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(b.nameFormat(name, owner)))
	})
}

func (b Bucket) deleteMeta(id string) error {
	return b.meta.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (b Bucket) deletePayload(id string) error {
	return b.payload.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (b Bucket) nameFormat(name, owner string) string {
	// choosing the name format as <name>_<owner> allows
	// to have unique names in the context of a owner e.g.
	// owner 1 can have foo_1 and owner 2 can have foo_2
	// without having a duplication error.
	return fmt.Sprintf("%s_%s", name, owner)
}

// createObjectEntry validates the object and creates a entry.
// Also the object will be marked as immutable.
func (b Bucket) createObjectEntry(obj *Object) (*badger.Entry, error) {
	if err := obj.isValid(); err != nil {
		return nil, err
	}
	if b.isNameExisting(obj.Name(), obj.Owner()) {
		return nil, fmt.Errorf("object with the name %s for the owner %s exists", obj.Name(), obj.Owner())
	}
	data, err := obj.Marshal()
	if err != nil {
		return nil, err
	}
	e := badger.NewEntry([]byte(obj.ID()), data)
	return e, nil
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
	meta, err := b.GetMeta(id)
	if err != nil {
		return nil, err
	}
	return b.composeObject(meta)
}
