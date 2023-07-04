package objst

import (
	"bytes"
	"encoding/gob"
	"net/url"

	"golang.org/x/exp/slices"
)

type MetaKey string

const (
	MetaKeyCreatedAt   MetaKey = "createdAt"
	MetaKeyContentType MetaKey = "contentType"
	MetaKeyName        MetaKey = "name"
	MetaKeyID          MetaKey = "id"
	MetaKeyOwner       MetaKey = "owner"
)

func (m MetaKey) String() string {
	return string(m)
}

type Metadata struct {
	// the actual metadata
	data map[MetaKey]string
	// systemKeys contains all the keys
	// which will be managed by objst.
	systemKeys []MetaKey
}

func NewMetadata() *Metadata {
	return &Metadata{
		data:       make(map[MetaKey]string),
		systemKeys: []MetaKey{MetaKeyID, MetaKeyCreatedAt, MetaKeyName, MetaKeyOwner},
	}
}

// Set will insert the given key value pair
// iff it isn't a systemKey like MetaKeyID or
// MetaKeyCreatedAt.
func (m Metadata) Set(k MetaKey, v string) {
	if m.isSystemMetaKey(k) {
		return
	}
	if v == "" {
		return
	}
	m.data[k] = v
}

// Is checks if the value of the given MetaKey
// is equal to `v`.
func (m Metadata) Is(k MetaKey, v string) bool {
	return m.Get(k) == v
}

func (m Metadata) Has(k MetaKey) bool {
	_, ok := m.data[k]
	return ok
}

func (m Metadata) Get(k MetaKey) string {
	return m.data[k]
}

func (m Metadata) Del(k MetaKey) {
	if m.isSystemMetaKey(k) {
		return
	}
	delete(m.data, k)
}

func (m Metadata) Encode() string {
	values := url.Values{}
	for k, v := range m.data {
		values.Set(k.String(), v)
	}
	return values.Encode()
}

func (m Metadata) Merge(mp map[MetaKey]string) {
	for k, v := range mp {
		m.Set(k, v)
	}
}

func (m Metadata) isSystemMetaKey(k MetaKey) bool {
	return slices.Contains(m.systemKeys, k)
}

// set is intended for internal usage where
// SystemMetaKeys can be set.
func (m Metadata) set(k MetaKey, v string) {
	m.data[k] = v
}

func (m Metadata) UserDefinedPairs() map[MetaKey]string {
	res := make(map[MetaKey]string)
	for k, v := range m.data {
		if !m.isSystemMetaKey(k) {
			res[k] = v
		}
	}
	return res
}

func (m Metadata) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(m.data)
	return buf.Bytes(), err
}

func (m *Metadata) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)
	return gob.NewDecoder(r).Decode(&m.data)
}

func (m *Metadata) Compare(md *Metadata, act action) bool {
	if act == Or {
		return m.or(md)
	}
	return m.and(md)
}

func (m Metadata) or(md *Metadata) bool {
	for k := range m.data {
		if md.Has(k) {
			return true
		}
	}
	return false
}

func (m Metadata) and(md *Metadata) bool {
	for k := range m.data {
		if !md.Has(k) {
			return false
		}
	}
	return true
}
