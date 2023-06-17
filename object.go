package objst

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/naivary/objst/models"
)

const (
	// Content type key to use for meta data.
	ContentType  = "contentType"
	createdAt    = "createdAt"
	lastModified = "lastModified"
)

type Object struct {
	// unique identifier
	id string
	// unique alias for the object
	name string
	// owner of the object.
	owner string
	// metadata of the object. The naming of the
	// of the keys follow the golang conventions
	// (e.g. camelCase).
	meta url.Values
	pl   *bytes.Buffer
	pos  int64
	// isMutable indicated if the object
	// can be mutated. An object is only mutable
	// if it isn't already inserted into the store
	// or wasn't retrieved from the store.
	isMutable bool
}

func New(name, owner string) *Object {
	o := &Object{
		id:        uuid.NewString(),
		name:      name,
		owner:     owner,
		meta:      url.Values{},
		pl:        new(bytes.Buffer),
		isMutable: true,
	}
	o.setDefaultMetadata()
	return o
}

func (o Object) ID() string {
	return o.id
}

func (o Object) Name() string {
	return o.name
}

func (o Object) Owner() string {
	return o.owner
}

func (o Object) Payload() []byte {
	return o.pl.Bytes()
}

// SetMeta will set the given key and
// value as a meta data key-pair, over-
// writing any key-pair which has been
// set before.
func (o *Object) SetMeta(k, v string) {
	o.meta.Set(k, v)
}

// GetMeta returns the corresponding value of the
// provided key. The bool is indicating if the value
// was retrieved successfully.
func (o *Object) GetMeta(k string) (string, bool) {
	return o.meta.Get(k), o.meta.Has(k)
}

func (o *Object) IsMetaExisting(k string) bool {
	return o.meta.Has(k)
}

func (o *Object) ToModel() *models.Object {
	return &models.Object{
		ID:      o.id,
		Name:    o.name,
		Owner:   o.owner,
		Meta:    o.meta,
		Payload: o.Payload(),
	}
}

func (o *Object) FromModel(m *models.Object) {
	o.id = m.ID
	o.meta = m.Meta
	o.owner = m.Owner
	o.name = m.Name
	o.pl = bytes.NewBuffer(m.Payload)
	o.isMutable = false
}

func (o *Object) Marshal() ([]byte, error) {
	if err := o.isValid(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(o.ToModel()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *Object) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	m := models.Object{}
	if err := gob.NewDecoder(r).Decode(&m); err != nil {
		return err
	}
	o.FromModel(&m)
	return nil
}

func (o Object) isValid() error {
	if !o.meta.Has(ContentType) {
		return ErrContentTypeNotExist
	}
	if len(o.pl.Bytes()) == 0 {
		return ErrEmptyPayload
	}
	return nil
}

func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	o.meta.Add(createdAt, t)
	o.meta.Add(lastModified, t)
}

// Write will write the data iff the object is mutable.
// Otherwise an ErrObjectIsImmutable will be returned.
// An object is mutable if it isn't inserted into the
// store or retrieved from the store.
func (o *Object) Write(p []byte) (int, error) {
	if !o.isMutable {
		return 0, ErrObjectIsImmutable
	}
	return o.pl.Write(p)
}

func (o *Object) WriteTo(w io.Writer) (int64, error) {
	buf := bytes.NewBuffer(o.Payload())
	defer buf.Reset()
	return buf.WriteTo(w)
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

// Reset resets the payload
func (o *Object) Reset() {
	o.pl.Reset()
}

func (o *Object) markAsImmutable() {
	o.isMutable = false
}
