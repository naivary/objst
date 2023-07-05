## objst

objst is an emmedable object storage written in golang. It can also be used to serve public
http traffic with a default http handler. It is based upon BadgerDB, a highly performant key-value store.

## Docs

The following section will explain the basic concepts on how to use objst and what some design choices.

### Bucket

A bucket is representing an object storage. It allows you to interact with the underlying
object storage in an easy manner. To create a simple bucket you can use the following code snippet:

```golang
func main() {
  opts := objst.NewDefaultBucketOptions()
  bucket, err := objst.NewBucket(opts)
  if err != nil {
    panic(err)
  }
  // use the bucket for different operations
}
```

### Object

An object is the main abstraction in objst to represent different payload with some metadata.

An object can be created using the `NewObject` function:

```golang
func main() {
  obj, err := objst.NewObject("name", "owner")
  if err != nil {
    panic(err)
  }
}
```

NOTE: The owner of an object has to be a valid uuid-v4. The uuid can be created using the google package
[google/uuid](https://github.com/google/uuid). It is used internally for testing and promises the most
resilient results in production use.

The object struct has implemented many useful interfaces which allow you to use it as a
usual file. For example an object can be passed to any function which accepts an `io.Reader`:

```golang
func main() {
  obj, err := objst.NewObject("name", "owner")
  if err != nil {
    panic(err)
  }

}
```

By passing the object to a function accepting an `io.Reader` the `Read([]byte) (int, error)` function will be called
but the payload of the object will not be lost. Still the behavior of the read operation is as you would expeted from
any other read operation in golang. The returned error of `NewObject` function can be ignored if you can assure that name and owner
will not be empty strings.

### Metadata

### Queries

### HTTP Handler
