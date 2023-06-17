package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/naivary/objst/object"
	"github.com/naivary/objst/random"
	"golang.org/x/exp/slog"
)

const (
	// test content type for the objects
	tCt = "test/text"
)

var (
	tB *Bucket
)

func tObj() *object.Object {
	owner := uuid.NewString()
	name := fmt.Sprintf("obj_name_%s", owner)
	o := object.New(name, owner)
	o.SetMeta(object.ContentType, tCt)
	o.Write([]byte(random.String(10)))
	return o
}

func tNObj(n int) []*object.Object {
	objs := make([]*object.Object, 0, n)
	for i := 0; i < n; i++ {
		objs = append(objs, tObj())
	}
	return objs
}

func setup() error {
	opts := badger.DefaultOptions("/tmp/badger")
	b, err := New(opts)
	if err != nil {
		return err
	}
	tB = b
	return err
}

func destroy() error {
	if err := tB.store.Close(); err != nil {
		return err
	}
	return nil
}

func TestMain(t *testing.M) {
	if err := setup(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	code := t.Run()
	if err := destroy(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	os.Exit(code)
}

func TestCreate(t *testing.T) {
	if err := tB.Create(tObj()); err != nil {
		t.Error(err)
		return
	}
}

func TestGetByID(t *testing.T) {
	o := tObj()
	if err := tB.Create(o); err != nil {
		t.Error(err)
	}
	oG, err := tB.GetByID(o.ID)
	if err != nil {
		t.Error(err)
	}
	if oG.ID != o.ID {
		t.Fatalf("id's aren't the same. Got: %s. Expected: %s", oG.ID, o.ID)
	}
	if !oG.Meta.Has(object.ContentType) {
		t.Fatalf("object does not have the custom set meta data filed. Expected: %s. Got: %s", o.Meta.Get(object.ContentType), oG.Meta.Get(object.ContentType))
	}
	if !bytes.Equal(oG.Payload, o.Payload) {
		t.Fatalf("payload is not the same. Got: %s. Expected: %s", oG.Payload, o.Payload)
	}
}

func TestDelete(t *testing.T) {
	o := tObj()
	if err := tB.Create(o); err != nil {
		t.Error(err)
	}
	if err := tB.Delete(o.ID); err != nil {
		t.Error(err)
	}
	_, err := tB.GetByID(o.ID)
	if !errors.Is(err, badger.ErrKeyNotFound) {
		t.Fatalf("Key should be not found.")
	}
}

func TestGetByMetasOr(t *testing.T) {
	o := tObj()
	if err := tB.Create(o); err != nil {
		t.Error(err)
		return
	}
	v := url.Values{}
	v.Set(object.ContentType, tCt)
	objs, err := tB.GetByMetasOr(v)
	if err != nil {
		t.Error(err)
		return
	}
	if len(objs) == 0 {
		t.Fatalf("at least one object should be contained. Got: %d", len(objs))
	}
}
func TestNameDuplication(t *testing.T) {
	o1 := tObj()
	o2 := tObj()
	o1.Name = o2.Name
	objs := make([]*object.Object, 0, 2)
	objs = append(objs, o1, o2)
	if err := tB.BatchCreate(objs); err != nil {
		t.Log(err)
		return
	}
	t.Fatal("should not create objects with the same name.")
}

func TestGetByMetasAnd(t *testing.T) {
	o1 := tObj()
	o1.SetMeta("foo", "bar")
	o1.SetMeta("bymetasand", "bymetasand")
	v := url.Values{}
	v.Set(object.ContentType, tCt)
	v.Set("foo", "bar")
	v.Set("bymetasand", "bymetasand")
	if err := tB.Create(o1); err != nil {
		t.Error(err)
		return
	}
	objs, err := tB.GetByMetasAnd(v)
	if err != nil {
		t.Error(err)
		return
	}

	if len(objs) != 1 {
		t.Fatalf("only one object should be returned. Got: %d", len(objs))
	}

}

func TestGetByName(t *testing.T) {
	o1 := tObj()
	if err := tB.Create(o1); err != nil {
		t.Error(err)
		return
	}
	oG, err := tB.GetByName(o1.Name)
	if err != nil {
		t.Error(err)
		return
	}
	if oG.Name != o1.Name {
		t.Fatalf("name should be equal. Got: %s. Expected: %s", oG.Name, o1.Name)
	}
}

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := tB.Create(tObj()); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkGet(b *testing.B) {
	objs := tNObj(b.N)
	for _, obj := range objs {
		if err := tB.Create(obj); err != nil {
			b.Error(err)
			return
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tB.GetByID(objs[i].ID); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkBatchCreate(b *testing.B) {
	objs := tNObj(b.N)
	b.ResetTimer()
	if err := tB.BatchCreate(objs); err != nil {
		b.Error(err)
		return
	}
	b.ReportAllocs()
}
