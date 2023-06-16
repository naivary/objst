package object

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	// Content type key to use for meta data.
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
	Meta    url.Values
	Payload []byte
	pl      *bytes.Buffer
	pos     int64
}

func New(name, owner string) *Object {
	o := &Object{
		ID:    uuid.NewString(),
		Name:  name,
		Owner: owner,
		Meta:  url.Values{},
		pl:    new(bytes.Buffer),
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
	o.Payload = o.pl.Bytes()
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
	if len(o.pl.Bytes()) == 0 {
		return ErrEmptyPayload
	}
	return nil
}

// the default metadata includes:
// createdAt: Unix Timestamp when the object is created
// lastModified: Unix Timestamp of the last modification
func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	o.Meta.Add(createdAt, t)
	o.Meta.Add(lastModified, t)
}

// TODO:(naivary) tmp file for big writes
func (o *Object) Write(p []byte) (int, error) {
	return o.pl.Write(p)
}

func (o *Object) ReadFrom(r io.Reader) (int64, error) {
	return o.pl.ReadFrom(r)
}

func (o *Object) Read(b []byte) (int, error) {
	n, err := bytes.NewReader(o.pl.Bytes()[o.pos:]).Read(b)
	if err != nil {
		return 0, err
	}
	o.pos += int64(n)
	return n, nil
}
