## objst

objst is an emmedable object storage written in golang. It is based upon
BadgerDB, a highly performant key-value store.

## Documentation

### Create a Bucket

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

For the full documentaiton of the bucket api see the [godocs](https://pkg.go.dev/github.com/naivary/objst)

### Object

An object is the main abstraction in objst. An object can contain any payload in a `[]byte` format.
Multiple helpful interfaces are implemented by object e.g. `io.Writer`, `io.Reader` or `io.WriteTo`
and many more which allows you to integrate the object as a regular "file". The following examples will
show some easy examples how to use objects and interact with the bucket/

Create a single object with the name "name" and owner "owner". The owner of an object can be any uniquely
identifiable string. Internal objst uses uuid which is the recommend way to use objst.

```golang
func main() {
  // create a single object
  obj := objst.NewObject("name", "owner")
  // every object has to contain a contentType meta data key
  obj.SetMeta(objst.ContentTypeMetaKey, "image/jpeg")
  // every object has to contain some payload
  if _, err := obj.Write([]byte("some random data")); err != nil {
    panic(err)
  }
  // insert the object to the object storage
  if err := bucket.Create(obj); err != nil {
    panic(err)
  }
}
```

Create multiple objects using a batch writer. Using a `BatchCreate` for multiple objects
instead of calling `Create` multiple times has extreme performance benefits.

```golang
  // create multiple objects at once
  objs := []*objst.Object{...}
  if err := bucket.BatchCreate(objs); err != nil {
    panic(err)
  }

```

Retrieve an object which is inserted in the object storage by name or id. Every object which is
inserted once in the object storage will be marked as immutable. By retrieving an object
you can only use read Operations upon the retrieved object. Only the meta data of the object can be changed.

```golang

func main() {
  // get an object by id
  obj, err := bucket.GetByID("d0f351d6-0fa7-4409-9d6c-46719f657016")
  if err != nil {
    panic(err)
  }

  // get an object by name.
  obj, err := bucket.GetByName("name", "owner")
  if err != nil {
    panic(err)
  }
}
```

Some useful object operations

```golang
func main() {
  obj, err := objst.NewObject("owner", "name")
  if err != nil {
    panic(err)
  }

  // write some payload to the object
  if _, err := obj.Write([]byte("foo")); err != nil {
    panic(err)
  }

  // read the payload from the object
  data := make([]byte, 0, 3)
  if _, err := obj.Read(data); err != nil {
    panic(err)
  }

  // write the data to another object
  obj2, err := objst.NewObject("owner2", "name2")
  if _, err := obj.WriteTo(obj2); err != nil {
    panic(err)
  }
}
```
