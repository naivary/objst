package objst

type action int

const (
	// localgical Or relationship
	Or action = iota + 1

	// localgical And relationship
	And
)

type Query struct {
	meta *Metadata
	// logical action of the meta datas
	act action
}

func NewQuery() *Query {
	return &Query{
		meta: nil,
		act:  Or,
	}
}

func (q *Query) WithMeta(meta *Metadata) *Query {
	q.meta = meta
	return q
}

func (q *Query) WithOwner(owner string) *Query {
	q.meta.set(MetaKeyOwner, owner)
	return q
}

func (q *Query) WithAction(act action) *Query {
	q.act = act
	return q
}

func (q *Query) WithMetaPair(k MetaKey, v string) {
	q.meta.Set(k, v)
}
