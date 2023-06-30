package objst

import (
	"net/url"
)

type action int

const (
	Or action = iota + 1
	And
)

type Query struct {
	meta url.Values
	// logical action of the meta datas
	act   action
	owner string
}

// NewQuery returns an empty query
// with some default values.
// The Query follows the following rules:
func NewQuery(owner string) *Query {
	return &Query{
		meta:  nil,
		act:   Or,
		owner: owner,
	}
}

func (q *Query) WithOwner(owner string) *Query {
	q.owner = owner
	return q
}

func (q *Query) WithMeta(meta url.Values, act action) *Query {
	q.meta = meta
	q.act = act
	return q
}

func (q *Query) WithAction(act action) *Query {
	q.act = act
	return q
}

func (q Query) isValid() bool {
	return q.owner != ""
}