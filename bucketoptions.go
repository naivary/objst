package objst

import "github.com/dgraph-io/badger/v4"

type BucketOptions badger.Options

func NewDefaultBucketOptions() BucketOptions {
	return BucketOptions(badger.DefaultOptions(""))
}

func (b *BucketOptions) overwriteDataDir(dir string) {
	b.Dir = dir
	b.ValueDir = dir
}

func (b BucketOptions) toBadgerOpts() badger.Options {
	return badger.Options(b)
}
