package object

import (
	"bytes"
	"encoding/gob"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	// Content type key to use
	// for meta data.
	ContentType  = "contentType"
	createdAt    = "createdAt"
	lastModified = "lastModified"
)

type Object struct {
	// unique identifier
	ID string
	// unique alias for the object
	Name  string
	Owner string
	// metadata of the object. The naming of the
	// of the keys follow the golang conventions
	// (e.g. camelCase).
	Meta url.Values
}

func New(name, owner string) *Object {
	o := &Object{
		ID:    uuid.NewString(),
		Name:  name,
		Owner: owner,
		Meta:  url.Values{},
	}
	o.setDefaultMetadata()
	return o
}

// SetMeta will set the given key and
// value as a meta data key-pair, over-
// writing any key-pair which has been
// set before.
func (o *Object) SetMeta(k, v string) {
	o.Meta.Set(k, v)
}

func (o *Object) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *Object) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	return gob.NewDecoder(r).Decode(&o)
}

func (o Object) IsValid() error {
	if !o.Meta.Has(ContentType) {
		return ErrContentTypeNotExist
	}
	return nil
}

// the default metadata inclused:
// createdAt: Unix Timestamp when the object is created
// lastModified: Unix Timestamp of the last modification
func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	o.Meta.Add(createdAt, t)
	o.Meta.Add(lastModified, t)
}
