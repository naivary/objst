package object

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Object struct {
	id       string
	name     string
	owner    string
	metadata url.Values
}

func New(name, owner string) *Object {
	o := &Object{
		id:       uuid.NewString(),
		name:     name,
		owner:    owner,
		metadata: url.Values{},
	}
	fmt.Println(time.Now().Unix())
	createdAt := strconv.FormatInt(time.Now().Unix(), 10)
	o.metadata.Add("createdAt", createdAt)
	return o
}
