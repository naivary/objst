package objst

import (
	"fmt"
)

type action int

const (
	// logical `Or` relationship
	Or action = iota + 1

	// logical `And` relationship
	And
)

type identifier int

const (
	id identifier = iota + 1

	name
)

type operation int

const (
	OperationDelete = iota + 1

	OperationGet
)

type Query struct {
	params *Metadata
	// logical action of the meta datas
	act action

	op operation

	singleEntryIdentitifer identifier
}

func NewQuery() *Query {
	return &Query{
		params: NewMetadata(),
		act:    Or,
		op:     OperationGet,
	}
}

func (q *Query) Owner(owner string) *Query {
	q.params.set(MetaKeyOwner, owner)
	return q
}

func (q *Query) ID(id string) *Query {
	q.params.set(MetaKeyID, id)
	return q
}

func (q *Query) Name(name string) *Query {
	q.params.set(MetaKeyName, name)
	return q
}

// Action sets the logical connection
// between the params and the meta data
// of all compared objects.
func (q *Query) Action(act action) *Query {
	q.act = act
	return q
}

// Param sets a given key value pair as a parameter
// of the query.
func (q *Query) Param(k MetaKey, v string) *Query {
	q.params.Set(k, v)
	return q
}

func (q *Query) Operation(op operation) *Query {
	q.op = op
	return q
}

func (q *Query) isValid() error {
	if len(q.params.data) <= 0 {
		return ErrEmptyQuery
	}
	if !isValidUUID(q.params.Get(MetaKeyOwner)) {
		return fmt.Errorf("invalid uuid for the field `owner`: %s", q.params.Get(MetaKeyOwner))
	}
	if !isValidUUID(q.params.Get(MetaKeyID)) {
		return fmt.Errorf("invalid uuid for the field `id`: %s", q.params.Get(MetaKeyID))
	}
	if q.params.Get(MetaKeyName) != "" && q.params.Get(MetaKeyOwner) == "" {
		return ErrNameOwnerCtxMissing
	}
	return nil
}

func (q *Query) isSingleEntry() bool {
	if q.params.Get(MetaKeyID) != "" {
		q.singleEntryIdentitifer = id
		return true
	}
	if q.params.Get(MetaKeyName) != "" {
		q.singleEntryIdentitifer = name
		return true
	}
	return false
}

func (q *Query) isIDIdentifier() bool {
	return q.singleEntryIdentitifer == id
}
