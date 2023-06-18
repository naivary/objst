package objst

import (
	"bytes"
	"errors"
	"net/url"
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
	oG, err := tEnv.b.GetByID(o.id)
	if err != nil {
		t.Error(err)
	}
	if oG.id != o.id {
		t.Fatalf("id's aren't the same. Got: %s. Expected: %s", oG.id, o.id)
	}
	if !oG.meta.Has(ContentType) {
		t.Fatalf("object does not have the custom set meta data filed. Expected: %s. Got: %s", o.meta.Get(ContentType), oG.meta.Get(ContentType))
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
	if err := tEnv.b.DeleteByID(o.id); err != nil {
		t.Error(err)
	}
	_, err := tEnv.b.GetByID(o.id)
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
	if err := tEnv.b.DeleteByName(o.name); err != nil {
		t.Error(err)
		return
	}
	_, err := tEnv.b.GetByID(o.id)
	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Fatalf("Key should be not found.")
	}
}

func TestGetByMetasOr(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	v := url.Values{}
	v.Set(ContentType, tEnv.ContentType)
	objs, err := tEnv.b.GetByMetasOr(v)
	if err != nil {
		t.Error(err)
		return
	}
	if len(objs) == 0 {
		t.Fatalf("at least one object should be contained. Got: %d", len(objs))
	}
}
func TestNameDuplication(t *testing.T) {
	o1 := tEnv.obj()
	o2 := tEnv.obj()
	o1.name = o2.name
	objs := make([]*Object, 0, 2)
	objs = append(objs, o1, o2)
	if err := tEnv.b.BatchCreate(objs); err != nil {
		return
	}
	t.Fatal("should not create objects with the same name.")
}

func TestGetByMetasAnd(t *testing.T) {
	o1 := tEnv.obj()
	o1.SetMeta("foo", "bar")
	o1.SetMeta("bymetasand", "bymetasand")
	v := url.Values{}
	v.Set(ContentType, tEnv.ContentType)
	v.Set("foo", "bar")
	v.Set("bymetasand", "bymetasand")
	if err := tEnv.b.Create(o1); err != nil {
		t.Error(err)
		return
	}
	objs, err := tEnv.b.GetByMetasAnd(v)
	if err != nil {
		t.Error(err)
		return
	}

	if len(objs) != 1 {
		t.Fatalf("only one object should be returned. Got: %d", len(objs))
	}

}

func TestGetByName(t *testing.T) {
	o1 := tEnv.obj()
	if err := tEnv.b.Create(o1); err != nil {
		t.Error(err)
		return
	}
	oG, err := tEnv.b.GetByName(o1.name)
	if err != nil {
		t.Error(err)
		return
	}
	if oG.name != o1.name {
		t.Fatalf("name should be equal. Got: %s. Expected: %s", oG.name, o1.name)
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
	oG, err := tEnv.b.GetByID(o.id)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = oG.Write(tEnv.payload(5))
	if !errors.Is(err, ErrObjectIsImmutable) {
		t.Fatalf("object should be immutable after get")
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
		if _, err := tEnv.b.GetByID(objs[i].id); err != nil {
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
