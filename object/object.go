package object

import (
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Object struct {
	// unique identifier
	id string
	// unique alias for the object
	name  string
	owner string
	// metadata of the object. The naming of the
	// of the keys follow the golang conventions
	// (e.g. camelCase).
	metadata url.Values
}

func New(name, owner string) *Object {
	o := &Object{
		id:       uuid.NewString(),
		name:     name,
		owner:    owner,
		metadata: url.Values{},
	}
	o.setDefaultMetadata()
	return o
}

// the default metadata inclused:
// createdAt: Unix Timestamp when the object is created
// lastModified: Unix Timestamp of the last modification
func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	o.metadata.Add("createdAt", t)
	o.metadata.Add("lastModified", t)
}

func (o *Object) SetMetadata(k, v string) {
	o.metadata.Set(k, v)
}

func (o Object) isValid() error {
	if !o.metadata.Has("contentType") {
		return ErrContentTypeNotExist
	}
	return nil
}
