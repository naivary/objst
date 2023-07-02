package objst

import (
	"bytes"
	"io"
	"net/http/httptest"
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
	w := httptest.NewRecorder()
	if _, err := io.Copy(w, res.Body); err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(w.Body.Bytes(), o.Payload()) {
		t.Fatalf("Payload is not the same. Got: %s. Expected: %s", w.Body.String(), o.Payload())
	}
}
