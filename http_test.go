package objst

import (
	"net/url"
	"testing"
)

func TestHTTPRead(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	target, err := url.JoinPath(tEnv.ts.URL, o.id)
	if err != nil {
		t.Error(err)
		return
	}
	res, err := tEnv.ts.Client().Get(target)
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()
	t.Log(res)
}
