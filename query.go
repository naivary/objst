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
	act   action
	owner string
}

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

func (q *Query) WithMeta(meta *Metadata) *Query {
	q.meta = meta
	return q
}

func (q *Query) WithAction(act action) *Query {
	q.act = act
	return q
}

func (q Query) isValid() bool {
	return q.owner != ""
}
