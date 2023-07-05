package objst

import (
	"bytes"
	"io"
	"regexp"

	"github.com/google/uuid"
)

const (
	objectNamePattern = "^([a-zA-Z0-9_.\\/-]+)(\\.[a-z]+)$"
)

type Object struct {
	meta *Metadata
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
		meta:      NewMetadata(),
		pl:        new(bytes.Buffer),
		isMutable: true,
	}
	o.meta.set(MetaKeyID, uuid.NewString())
	o.meta.set(MetaKeyName, name)
	o.meta.set(MetaKeyOwner, owner)
	return o, nil
}

func (o Object) ID() string {
	return o.meta.Get(MetaKeyID)
}

func (o Object) Name() string {
	return o.meta.Get(MetaKeyName)
}

func (o Object) Owner() string {
	return o.meta.Get(MetaKeyOwner)
}

func (o Object) Payload() []byte {
	return o.pl.Bytes()
}

// SetMeta will set the given key and
// value as a meta data key-pair, over-
// writing any key-pair which has been
// set before.
func (o *Object) SetMetaKey(k MetaKey, v string) {
	o.meta.Set(k, v)
}

// GetMeta returns the corresponding value of the
// provided key. The bool is indicating if the value
// was retrieved successfully.
func (o *Object) GetMetaKey(k MetaKey) string {
	return o.meta.Get(k)
}

// HasMetaKey check if the meta data of the
// object contains the given key.
func (o *Object) HasMetaKey(k MetaKey) bool {
	return o.meta.Has(k)
}

func (o *Object) BinaryMarshaler() ([]byte, error) {
	return o.Marshal()
}

func (o *Object) BinaryUnmarshaler(data []byte) error {
	return o.Unmarshal(data)
}

func (o *Object) Marshal() ([]byte, error) {
	return o.Payload(), nil
}

func (o *Object) Unmarshal(data []byte) error {
	o.pl = bytes.NewBuffer(data)
	return nil
}

func (o Object) isValid() error {
	if !o.HasMetaKey(MetaKeyContentType) {
		return ErrContentTypeNotExist
	}
	if len(o.pl.Bytes()) == 0 {
		return ErrEmptyPayload
	}
	if ok, _ := regexp.MatchString(objectNamePattern, o.Name()); !ok {
		return ErrInvalidNamePattern
	}
	return nil
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

func (o *Object) ToModel() *objectModel {
	return &objectModel{
		ID:       o.ID(),
		Name:     o.Name(),
		Owner:    o.Owner(),
		Metadata: o.meta.UserDefinedPairs(),
	}
}
