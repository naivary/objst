## objst

objst is an emmedable object storage written in golang. It is based upon
BadgerDB, a highly performant key-value store.

# Documentation

## Create a Bucket

A bucket is the actual object storage containing all the data. An example can be:

```golang
func main() {
  opts := badger.DefaultOptions("/tmp/objst")
  b, err := objst.NewBucket(&opts)
  if err != nil {
    panic(err)
  }
  // use the bucket stored in `b` for example by creating a new object
  obj := objst.NewObject("name of object", "owner of object")
  if err := b.Create(obj); err != nil {
    panic(err)
  }
}
```
