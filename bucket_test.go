package objst

import (
	"bytes"
	"errors"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestCreate(t *testing.T) {
	if err := tEnv.b.Create(tEnv.obj()); err != nil {
		t.Error(err)
		return
	}
}

func TestGetByID(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
	}
	oG, err := tEnv.b.GetByID(o.ID())
	if err != nil {
		t.Error(err)
	}
	if oG.ID() != o.ID() {
		t.Fatalf("id's aren't the same. Got: %s. Expected: %s", oG.ID(), o.ID())
	}
	if !oG.HasMetaKey(MetaKeyContentType) {
		t.Fatalf("object does not have the custom set meta data filed. Expected: %s. Got: %s", o.GetMetaKey(MetaKeyContentType), oG.GetMetaKey(MetaKeyContentType))
	}
	if !bytes.Equal(oG.Payload(), o.Payload()) {
		t.Fatalf("payload is not the same. Got: %s. Expected: %s", oG.Payload(), o.Payload())
	}
}

func TestDeleteByID(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
	}
	if err := tEnv.b.DeleteByID(o.ID()); err != nil {
		t.Error(err)
	}
	_, err := tEnv.b.GetByID(o.ID())
	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Fatalf("Key should be not found.")
	}
}

func TestDeleteByName(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	if err := tEnv.b.DeleteByName(o.Name(), o.Owner()); err != nil {
		t.Error(err)
		return
	}
	_, err := tEnv.b.GetByID(o.ID())
	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Fatalf("Key should be not found.")
	}
}

func TestDeleteNameAfterObjDelete(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	if err := tEnv.b.DeleteByID(o.ID()); err != nil {
		t.Error(err)
		return
	}
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
}

func TestNameDuplication(t *testing.T) {
	o1 := tEnv.obj()
	o2 := tEnv.obj()
	o1.meta.set(MetaKeyName, o2.Name())
	o1.meta.set(MetaKeyOwner, o2.Owner())
	objs := make([]*Object, 0, 2)
	objs = append(objs, o1, o2)
	if err := tEnv.b.BatchCreate(objs); err != nil {
		return
	}
	t.Fatal("should not create objects with the same name.")
}

func TestGetByName(t *testing.T) {
	o1 := tEnv.obj()
	if err := tEnv.b.Create(o1); err != nil {
		t.Error(err)
		return
	}
	oG, err := tEnv.b.GetByName(o1.Name(), o1.Owner())
	if err != nil {
		t.Error(err)
		return
	}
	if oG.Name() != o1.Name() {
		t.Fatalf("name should be equal. Got: %s. Expected: %s", oG.Name(), o1.Name())
	}
}

func TestImmutability(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	_, err := o.Write(tEnv.payload(10))
	if !errors.Is(err, ErrObjectIsImmutable) {
		t.Fatalf("object should not be mutable")
	}
}

func TestImmutabilityAfterGet(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	oG, err := tEnv.b.GetByID(o.ID())
	if err != nil {
		t.Error(err)
		return
	}
	_, err = oG.Write(tEnv.payload(5))
	if !errors.Is(err, ErrObjectIsImmutable) {
		t.Fatalf("object should be immutable after get")
	}
}

func TestInsertAfterRead(t *testing.T) {
	o := tEnv.obj()
	dst := make([]byte, 5)
	if _, err := o.Read(dst); err != nil {
		t.Error(err)
		return
	}
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}

	oG, err := tEnv.b.GetByID(o.ID())
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(o.Payload(), oG.Payload()) {
		t.Fatalf("the bytes should be equal after read anad retrieval. Got: %s. Expected: %s", oG.Payload(), o.Payload())
	}
}

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := tEnv.b.Create(tEnv.obj()); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkGet(b *testing.B) {
	objs := tEnv.nObj(b.N)
	for _, obj := range objs {
		if err := tEnv.b.Create(obj); err != nil {
			b.Error(err)
			return
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tEnv.b.GetByID(objs[i].ID()); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkBatchCreate(b *testing.B) {
	objs := tEnv.nObj(b.N)
	b.ResetTimer()
	if err := tEnv.b.BatchCreate(objs); err != nil {
		b.Error(err)
		return
	}
	b.ReportAllocs()
}
