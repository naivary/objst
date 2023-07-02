package objst

import "net/url"

type MetaKey string

const (
	MetaKeyCreatedAt    MetaKey = "createdAt"
	MetaKeyLastModified MetaKey = "lastModified"
	MetaKeyContentType  MetaKey = "contentType"
)

type Metadata map[string]string

func NewMetadata() Metadata {
	return Metadata(map[string]string{})
}

func (m Metadata) Set(k string, v string) {
	m[k] = v
}

func (m Metadata) Has(k string) bool {
	_, ok := m[k]
	return ok
}

func (m Metadata) Get(k string) string {
	return m[k]
}


func (m Metadata) Del(k string) {
	delete(m, k)
}

func (m Metadata) Encode() string {
	values := url.Values{}
	for k, v := range m {
		values.Set(k, v)
	}
	return values.Encode()
}
