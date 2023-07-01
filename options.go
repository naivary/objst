package objst

import "github.com/dgraph-io/badger/v4"

func DefaultOptions() badger.Options {
	return badger.DefaultOptions("")
}
