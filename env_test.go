package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/naivary/objst/random"
)

var tEnv *testEnv

type testEnv struct {
	b           *Bucket
	ContentType string
	DataDir     string
	names       string
}

func newTestEnv() (*testEnv, error) {
	tEnv := testEnv{
		ContentType: "test/text",
		DataDir:     "/tmp/badger/objst",
		names:       "/tmp/badger/names",
	}
	b, err := NewBucket(nil)
	if err != nil {
		return nil, err
	}
	tEnv.b = b
	return &tEnv, nil
}

func (t testEnv) owner() string {
	return uuid.NewString()
}

func (t testEnv) name() string {
	return fmt.Sprintf("obj_name_%s", t.owner())
}

func (t testEnv) payload(n int) []byte {
	return []byte(random.String(n))
}

func (t testEnv) emptyObj() *Object {
	o := New(t.name(), t.owner())
	o.SetMeta(ContentType, t.ContentType)
	return o
}

func (t testEnv) obj() *Object {
	o := New(t.name(), t.owner())
	o.SetMeta(ContentType, t.ContentType)
	o.Write(t.payload(10))
	return o
}

func (t testEnv) nObj(n int) []*Object {
	objs := make([]*Object, 0, n)
	for i := 0; i < n; i++ {
		objs = append(objs, t.obj())
	}
	return objs
}

func (t testEnv) destroy() error {
	if err := t.b.store.Close(); err != nil {
		return err
	}
	if err := t.b.names.Close(); err != nil {
		return err
	}
	return os.RemoveAll("/tmp/badger")
}

func TestMain(t *testing.M) {
	te, err := newTestEnv()
	if err != nil {
		log.Fatal(err)
	}
	tEnv = te
	code := t.Run()
	if err := te.destroy(); err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}
