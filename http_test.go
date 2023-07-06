package objst

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
)

const route = "objst"

func TestHTTPRead(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	target, err := url.JoinPath(tEnv.ts.URL, route, "read", o.ID())
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
				"foo": "bar",
				"contentType": "text/plain"
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
	target, err := url.JoinPath(tEnv.ts.URL, route, o.ID())
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

func TestHTTPGet(t *testing.T) {
	o := tEnv.obj()
	if err := tEnv.b.Create(o); err != nil {
		t.Error(err)
		return
	}
	target, err := url.JoinPath(tEnv.ts.URL, route, o.ID())
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
	m := objectModel{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		t.Error(err)
		return
	}
	data1, err := json.Marshal(&m)
	if err != nil {
		t.Error(err)
		return
	}
	data2, err := json.Marshal(o.ToModel())
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(data1, data2) {
		t.Fatalf("models are not equal. Got: %v. Expected: %v", m, o.ToModel())
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("statuscode is not %d", http.StatusOK)
	}
}

func TestHTTPUpload(t *testing.T) {
	target, err := url.JoinPath(tEnv.ts.URL, route, "upload")
	if err != nil {
		t.Error(err)
		return
	}
	r, err := tEnv.newUploadRequest(target, nil, tEnv.h.opts.FormKey, "testdata/images/large.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	// injectOwner is a potential middleware to authorize and intject the
	// owner into the middleware
	injectOwner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), CtxKeyOwner, uuid.NewString())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	hl := injectOwner(tEnv.h)
	w := httptest.NewRecorder()
	hl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("statuscode is not %d. Got: %d", http.StatusOK, w.Code)
	}
}

func TestHTTPUploadUknownMimeType(t *testing.T) {
	target, err := url.JoinPath(tEnv.ts.URL, route, "upload")
	if err != nil {
		t.Error(err)
		return
	}
	opts := map[string]string{
		"contentType": "text/plain",
	}
	r, err := tEnv.newUploadRequest(target, opts, tEnv.h.opts.FormKey, "testdata/files/unofficial.testtype")
	if err != nil {
		t.Error(err)
		return
	}
	// injectOwner is a potential middleware to authorize and intject the
	// owner into the middleware
	injectOwner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), CtxKeyOwner, uuid.NewString())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	hl := injectOwner(tEnv.h)
	w := httptest.NewRecorder()
	hl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("statuscode is not %d. Got: %d. Res: %v", http.StatusOK, w.Code, w.Body)
	}
}

func TestHTTPUploadUknownMimeTypeAndEmptyCt(t *testing.T) {
	target, err := url.JoinPath(tEnv.ts.URL, route, "upload")
	if err != nil {
		t.Error(err)
		return
	}
	r, err := tEnv.newUploadRequest(target, nil, tEnv.h.opts.FormKey, "testdata/files/unofficial.testtype")
	if err != nil {
		t.Error(err)
		return
	}
	// injectOwner is a potential middleware to authorize and intject the
	// owner into the middleware
	injectOwner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), CtxKeyOwner, uuid.NewString())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	hl := injectOwner(tEnv.h)
	w := httptest.NewRecorder()
	hl.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("statuscode is not as expected. Got: %d. Expected: %d", w.Code, http.StatusOK)
	}
}
