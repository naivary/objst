package bucket

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/naivary/objst/object"
	"golang.org/x/exp/slog"
)

var (
	tB     *Bucket
	tOwner string
	tName  string
)

func setup() error {
	opts := badger.DefaultOptions("/tmp/badger")
	b, err := New(opts)
	if err != nil {
		return err
	}
	tB = b
	tOwner = uuid.NewString()
	tName = fmt.Sprintf("obj_name_%s", tOwner)
	return nil
}

func destroy() error {
	if err := tB.store.Close(); err != nil {
		return err
	}
	return nil
}

func TestMain(t *testing.M) {
	// setup
	if err := setup(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	// run
	code := t.Run()
	// cleanup
	if err := destroy(); err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	os.Exit(code)
}

func TestCreate(t *testing.T) {
	o := object.New(tName, tOwner)
	o.SetMeta(object.ContentType, "html/text")
	if err := tB.Create(o); err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	o := object.New(tName, tOwner)
	o.SetMeta(object.ContentType, "html/text")
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
}

func TestDelete(t *testing.T) {
	o := object.New(tName, tOwner)
	o.SetMeta(object.ContentType, "html/text")
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
		o := object.New(uuid.NewString(), tOwner)
		o.SetMeta(object.ContentType, "html/text")
		if err := tB.Create(o); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}
