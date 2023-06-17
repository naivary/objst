package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/naivary/objst/random"
)

var tEnv *testEnv

type testEnv struct {
	b           *Bucket
	ContentType string
	DataDir     string
}

func newTestEnv() (*testEnv, error) {
	tEnv := testEnv{
		ContentType: "test/text",
		DataDir:     "/tmp/badger",
	}

	b, err := NewBucket(badger.DefaultOptions(tEnv.DataDir))
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
	return os.RemoveAll(t.DataDir)
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
