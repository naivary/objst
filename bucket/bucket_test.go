package bucket

import (
	"errors"
	"fmt"
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

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		o := tObj()
		if err := tB.Create(o); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}
