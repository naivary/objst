package object

import (
	"bytes"
	"encoding/gob"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
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

func (o *Object) SetMeta(k, v string) {
	o.Meta.Set(k, v)
}

func (o *Object) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	gob.Register(url.Values{})
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
	if !o.Meta.Has("contentType") {
		return ErrContentTypeNotExist
	}
	return nil
}

// the default metadata inclused:
// createdAt: Unix Timestamp when the object is created
// lastModified: Unix Timestamp of the last modification
func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	o.Meta.Add("createdAt", t)
	o.Meta.Add("lastModified", t)
}
