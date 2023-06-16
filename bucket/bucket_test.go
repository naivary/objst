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

func setup() (*Bucket, error) {
	opts := badger.DefaultOptions("/tmp/badger")
	return New(opts)
}

func destroy() {
	tB.store.Close()
}

func TestMain(t *testing.M) {
	// setup
	b, err := setup()
	if err != nil {
		slog.Error("something went wrong while setting up the test", slog.String("msg", err.Error()))
		return
	}
	tB = b
	tOwner = uuid.NewString()
	tName = fmt.Sprintf("obj_name_%s", tOwner)
	// run
	code := t.Run()
	// cleanup
	destroy()
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
