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

func TestGet(t *testing.T) {
	o := tObj()
	if err := tB.Create(o); err != nil {
		t.Error(err)
	}
	oG, err := tB.Get(o.ID)
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
	_, err := tB.Get(o.ID)
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
	if len(objs) != 1 {
		t.Fatalf("at least one object should be contained. Got: %d", len(objs))
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
		if _, err := tB.Get(objs[i].ID); err != nil {
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
