package objst

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/naivary/objst/models"
)

const (
	ContentTypeMetaKey  = "contentType"
	createdAtMetaKey    = "createdAt"
	lastModifiedMetaKey = "lastModified"
)

type Object struct {
	// unique object identifier
	id string
	// unique alias for the object
	name string
	// owner of the object. Intenally an uuid is used
	// but it can be every type of unique string identifier.
	owner string
	// metadata of the object. The naming of the
	// of the keys follow the golang conventions
	// (e.g. camelCase).
	meta url.Values
	// payload of the object
	pl *bytes.Buffer
	// current reading psotion
	pos int64
	// An object is only mutable if it
	// isn't already inserted into the store
	// or wasn't retrieved from the store.
	isMutable bool
}

func NewObject(name, owner string) (*Object, error) {
	if owner == "" || name == "" {
		return nil, ErrMustIncludeOwnerAndName
	}
	o := &Object{
		id:        uuid.NewString(),
		name:      name,
		owner:     owner,
		meta:      url.Values{},
		pl:        new(bytes.Buffer),
		isMutable: true,
	}
	o.setDefaultMetadata()
	return o, nil
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
	// dont allow the user to
	// overwrite default metadata
	if o.isDefaultMetadata(k) {
		return
	}
	o.meta.Set(k, v)
}

// isDefaultMetadata checks if the given key `k`
// is a default metadata. `isDefaultMetadata` should
// always be before a metadata will be set to not overwrite
// any default metadatas.
func (o *Object) isDefaultMetadata(k string) bool {
	switch k {
	case lastModifiedMetaKey:
		return true
	case createdAtMetaKey:
		return true
	default:
		return false
	}
}

// GetMeta returns the corresponding value of the
// provided key. The bool is indicating if the value
// was retrieved successfully.
func (o *Object) GetMeta(k string) (string, bool) {
	return o.meta.Get(k), o.meta.Has(k)
}

// HasMetaKey check if the meta data of the
// object contains the given key.
func (o *Object) HasMetaKey(k string) bool {
	return o.meta.Has(k)
}

// ToModel returns a object which only
// contains primitiv value types for serialization.
func (o *Object) ToModel() *models.Object {
	return &models.Object{
		ID:      o.id,
		Name:    o.name,
		Owner:   o.owner,
		Meta:    o.meta,
		Payload: o.Payload(),
	}
}

func (o *Object) fromModel(m *models.Object) {
	o.id = m.ID
	o.meta = m.Meta
	o.owner = m.Owner
	o.name = m.Name
	o.pl = bytes.NewBuffer(m.Payload)
	o.isMutable = false
}

func (o *Object) BinaryMarshaler() ([]byte, error) {
	return o.Marshal()
}

func (o *Object) BinaryUnmarshaler(data []byte) error {
	return o.Unmarshal(data)
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
	o.fromModel(&m)
	return nil
}

func (o Object) isValid() error {
	const namePattern = "^[a-zA-Z0-9_.-]+$"
	if !o.HasMetaKey(ContentTypeMetaKey) {
		return ErrContentTypeNotExist
	}
	if len(o.pl.Bytes()) == 0 {
		return ErrEmptyPayload
	}
	if ok, _ := regexp.MatchString(namePattern, o.name); !ok {
		return ErrInvalidNamePattern
	}
	return nil
}

func (o *Object) setDefaultMetadata() {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	// o.SetMeta can't be used here because system defaults cannot
	// be overwritten using o.SetMeta.
	o.meta.Add(createdAtMetaKey, t)
	o.meta.Add(lastModifiedMetaKey, t)
}

// Write will write the data iff the object is mutable.
// Otherwise an ErrObjectIsImmutable will be returned.
// An object is mutable if it isn't inserted or retrieved
// from the store.
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

func FromModel(m *models.Object) (*Object, error) {
	obj, err := NewObject(m.Name, m.Owner)
	if err != nil {
		return nil, err
	}
	if _, err := obj.Write(m.Payload); err != nil {
		return nil, err
	}
	for k, v := range m.Meta {
		obj.SetMeta(k, v[0])
	}
	return obj, nil
}
