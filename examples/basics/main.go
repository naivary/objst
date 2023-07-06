package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/naivary/objst"
)

func main() {
	// create a new bucket wiht default options
	opts := objst.NewDefaultBucketOptions()
	bucket, err := objst.NewBucket(opts)
	if err != nil {
		panic(err)
	}

	// create a new object and write the
	// payload of the file `test.txt` to it.
	owner := uuid.NewString()
	// naming the object test.txt will set the
	// objst.MetaKeyContentType automatically.
	obj, err := objst.NewObject("test.txt", owner)
	if err != nil {
		panic(err)
	}
	file, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	// use the io.ReadFrom interface of the object
	// to read from an io.Reader e.g. `os.File`
	if _, err := obj.ReadFrom(file); err != nil {
		panic(err)
	}

	// check whats inside the object
	fmt.Printf("The payload of the object is: %s\n", obj.Payload())

	// insert the object into the object storage
	if err := bucket.Create(obj); err != nil {
		panic(err)
	}

	// get the object back from the object storage.
	// Now the object is immutable.
	obj, err = bucket.GetByID(obj.ID())
	if err != nil {
		panic(err)
	}

	_, err = obj.Write([]byte("some more test data"))
	if errors.Is(err, objst.ErrObjectIsImmutable) {
		fmt.Println("object can not be written to because it is immutable")
	}

	err = bucket.DeleteByID(obj.ID())
	if err != nil {
		panic(err)
	}
}
