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

NOTE: The owner of an object has to be a valid uuid-v4. The uuid can be created using the
[google/uuid](https://github.com/google/uuid) package. It is used internally for testing
and promises the most resilient results in production use.

The object struct has implemented many useful interfaces which allow you to use it as a
usual file. For example an object can be passed to any function which accepts an `io.Reader`,
`io.Writer`, `io.WriterTo` or `io.ReaderFrom`.

```golang
func main() {
  obj, err := objst.NewObject("name", "owner")
  if err != nil {
    panic(err)
  }
  buf := new(bytes.Buffer)
  if _, err := buf.ReadFrom(obj); err != nil {
    panic(err)
  }
  if _, err := buf.WriteTo(obj); err != nil {
    return err
  }
}
```

### Metadata

The most powerful feature of `objst` is the use of meta data. Meta data are custom key=value
pairs which will be associated with object and used for querying purposes. Setting a key-value
on an object can be done using the `obj.SetMetaKey` function. Some meta data is managed directly
by `objst` and cannot be manipulated by you. For example `objst.MetaKeyID` or `objst.MetaKeyCreatedAt`
can not be set using `obj.SetMetaKey`. The key of the meta data is of type `objst.MetaKey` and the value
is a string.

```golang
func main() {
  obj, err := objst.NewObject("name", "owner")
  if err != nil {
    panic(err)
  }
  // set the meta data contentType
  obj.SetMetaKey(objst.MetaKeyContentType, "text/plain")

  // get the meta data id
  id := obj.GetMetaKey(objst.MetaKeyID)
}
```

There are some helper function implemented for the object struct e.g `ID()` or `Owner()` which will return the
meta data in a convenient way. Calling `ID()` is the same as `obj.GetMetaKey(objst.MetaKeyID)`.

### Queries

`objst.NewQuery` allows you to get multiple or one object at once in a convenient way. For example
you can get all the objects which have the meta data `foo=bar` in the following way:

```golang
func main() {
  opts := objst.NewDefaultBucketOptions()
  bucket, err := objst.NewBucket(opts)
  if err != nil {
    panic(err)
  }
  // Create a query with the parameter foo=bar and the `Get` operation
  // which is the default.
  q := objst.NewQuery().Param("foo", "bar").Operation(objst.OperationGet)

  objs, err := bucket.Execute(q)
  if err != nil {
    panic(err)
  }
}
```

The query is smart engough to figure out if only one record will be fetched or multiple. This allows you
to use queries to fetch one record:

```golang
func main() {
  opts := objst.NewDefaultBucketOptions()
  bucket, err := objst.NewBucket(opts)
  if err != nil {
    panic(err)
  }

  // Get one object for the owner `owner` and name `name`.
  q := objst.NewQuery().Owner("owner").Name("name")

  objs, err := bucket.Execute(q)
  if err != nil {
    panic(err)
  }

  // or fetch by id
  q = objst.NewQuery().ID("id")
  objs, err := bucket.Execute(q)
  if err != nil {
    panic(err)
  }
}
```

### HTTP Handler

objst delivers a default `HTTPHandler` to serve public a bucket over http

```golang
func main() {
  opts := objst.NewDefaultBucketOptions()
  bucket, err := objst.NewBucket(opts)
  if err != nil {
    panic(err)
  }

  // the handler options allow you to set different
  // parameters to modify the behavior of the handler.
  // For the different options available see:
  // https://pkg.go.dev/github.com/naivary/objst#HTTPHandlerOptions
  handlerOpts := objst.DefaultHTTPHandlerOptions()
  hl := objst.NewHTTPHandler(handlerOpts)
}
```

All endpoints require authentication some require authorization. You can specify
how authorization or authentication is implemented by setting the `IsAuthorized`
and `IsAuthenticated` middleware in the handler's options. The endpoints requiring
authorization expected the owner uuid in the request context with the key
`objst.CtxKeyOwner`. So don't forget to set the key otherwise the endpoints
will not be able to serve the data. By default `IsAuthenticated` and `IsAuthorized` will
allow all incoming request assigning some random owner to the request context.

The different endpoints are as follow:

1. GET /objst/{id}: Get the object as a model without the payload
2. GET /objst/read/{id}: Read the payload of the object
3. DELETE /objst/{id}: Delete the object
4. POST /objst/upload: Upload a file to the object storage
