package bucket

import (
	"os"

	"github.com/dgraph-io/badger/v4"
	"golang.org/x/exp/slog"
)

type Bucket struct {
	store  *badger.DB
	logger *slog.Logger
}

func New(opts badger.Options) (*Bucket, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		store:  db,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
	return b, nil
}
