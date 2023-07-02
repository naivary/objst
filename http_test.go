package objst

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const route = "objst"

func TestHTTPRead(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	target, err := url.JoinPath(tEnv.ts.URL, route, o.id)
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

func TestHTTPCreate(t *testing.T) {
	data := `
		{
			"name": "test_name",
			"owner": "daaa7824-2706-4587-9814-818a1d3d8953",
			"payload": [10, 12, 18, 128, 133],
			"metadata": {
				"foo": ["bar"],
				"contentType": ["text/plain"]
			}
		}	
	`
	target, err := url.JoinPath(tEnv.ts.URL, route)
	if err != nil {
		t.Error(err)
		return
	}
	body := bytes.NewReader([]byte(data))
	res, err := tEnv.ts.Client().Post(target, "application/json", body)
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
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("statuscode is not %d", http.StatusOK)
	}
}

func TestHTTPRemove(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	target, err := url.JoinPath(tEnv.ts.URL, route, o.id)
	if err != nil {
		t.Error(err)
		return
	}
	r, err := http.NewRequest(http.MethodDelete, target, nil)
	if err != nil {
		t.Error(err)
		return
	}
	res, err := tEnv.ts.Client().Do(r)
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("statuscode was not %d", http.StatusNoContent)
	}
}
