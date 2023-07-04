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
)

func (m MetaKey) String() string {
	return string(m)
}

type Metadata struct {
	data       map[MetaKey]string
	systemKeys []MetaKey
}

func NewMetadata() Metadata {
	return Metadata{
		data:       make(map[MetaKey]string),
		systemKeys: []MetaKey{MetaKeyID, MetaKeyCreatedAt},
	}
}

func (m Metadata) Set(k MetaKey, v string) {
	if slices.Contains(m.systemKeys, k) {
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
	delete(m.data, k)
}

func (m Metadata) Encode() string {
	values := url.Values{}
	for k, v := range m.data {
		values.Set(k.String(), v)
	}
	return values.Encode()
}
