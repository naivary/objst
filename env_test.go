package objst

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/naivary/objst/random"
)

var tEnv *testEnv

type testEnv struct {
	b           *Bucket
	ts          *httptest.Server
	h           *HTTPHandler
	ContentType string
}

func newTestEnv() (*testEnv, error) {
	tEnv := testEnv{
		ContentType: "test/text",
	}
	opts := badger.DefaultOptions("")
	// turn of default loggin of badger
	opts.Logger = nil
	b, err := NewBucket(opts)
	if err != nil {
		return nil, err
	}
	tEnv.b = b
	h := NewHTTPHandler(b, DefaultHTTPHandlerOptions())
	tEnv.h = h
	tEnv.ts = httptest.NewServer(h)
	return &tEnv, nil
}

func (t testEnv) owner() string {
	return uuid.NewString()
}

func (t testEnv) id() string {
	return uuid.NewString()
}

func (t testEnv) name() string {
	return fmt.Sprintf("obj_name_%s", t.owner())
}

func (t testEnv) payload(n int) []byte {
	return []byte(random.String(n))
}

func (t testEnv) emptyObj() *Object {
	o, _ := NewObject(t.name(), t.owner())
	o.SetMeta(ContentTypeMetaKey, t.ContentType)
	return o
}

func (t testEnv) obj() *Object {
	o, _ := NewObject(t.name(), t.owner())
	o.SetMeta(ContentTypeMetaKey, t.ContentType)
	o.Write(t.payload(10))
	return o
}

func (t testEnv) nObj(n int) []*Object {
	objs := make([]*Object, 0, n)
	for i := 0; i < n; i++ {
		objs = append(objs, t.obj())
	}
	return objs
}

func (t testEnv) newUploadRequest(url string, params map[string]string, formKey string, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	multiFile, err := w.CreateFormFile(formKey, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(multiFile, file); err != nil {
		return nil, err
	}
	for k, v := range params {
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func (t testEnv) destroy() error {
	if err := t.b.store.Close(); err != nil {
		return err
	}
	if err := t.b.names.Close(); err != nil {
		return err
	}
	if err := os.RemoveAll("/tmp/badger"); err != nil {
		return err
	}
	t.ts.Close()
	return nil
}

func TestMain(t *testing.M) {
	te, err := newTestEnv()
	if err != nil {
		log.Fatal(err)
	}
	tEnv = te
	code := t.Run()
	if err := te.destroy(); err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}
