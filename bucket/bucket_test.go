package bucket

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/naivary/objst/logger"
)

func TestNew(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger")
	opts.Logger = logger.New(context.Background())
	b, err := New(opts)
	if err != nil {
		t.Error(err)
	}
	defer b.store.Close()
}
