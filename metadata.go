package objst

import (
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

func NewMetadata() Metadata {
	return Metadata{
		data:       make(map[MetaKey]string),
		systemKeys: []MetaKey{MetaKeyID, MetaKeyCreatedAt},
	}
}

// Set will insert the given key value pair
// iff it isn't a systemKey like MetaKeyID or
// MetaKeyCreatedAt.
func (m Metadata) Set(k MetaKey, v string) {
	if m.isSystemMetaKey(k) {
		return
	}
	m.data[k] = v
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
// SystemMetaKeys can be set
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
