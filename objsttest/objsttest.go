package objsttest

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/naivary/objst"
	"github.com/naivary/objst/random"
)

type Env struct {
	b           *objst.Bucket
	ts          *httptest.Server
	h           *objst.HTTPHandler
	ContentType string
}

func NewEnv() (*Env, error) {
	tEnv := Env{
		ContentType: "test/text",
	}
	opts := objst.NewDefaultBucketOptions()
	// turn of default loggin of badger
	opts.Logger = nil
	b, err := objst.NewBucket(opts)
	if err != nil {
		return nil, err
	}
	tEnv.b = b
	tEnv.h = objst.NewHTTPHandler(b, objst.DefaultHTTPHandlerOptions())
	tEnv.ts = httptest.NewServer(tEnv.h)
	mime.AddExtensionType(".test", "text/plain")
	return &tEnv, nil
}

func (e Env) Owner() string {
	return uuid.NewString()
}

// Name returns a random Name in the format
// obj_name_<id>.test
func (e Env) Name() string {
	return fmt.Sprintf("obj_name_%s.test", e.Owner())
}

func (e Env) Payload(n int) []byte {
	return []byte(random.String(n))
}

func (e Env) EmptyObj() *objst.Object {
	o, _ := objst.NewObject(e.Name(), e.Owner())
	o.SetMetaKey(objst.MetaKeyContentType, e.ContentType)
	return o
}

func (e Env) Obj() *objst.Object {
	o, _ := objst.NewObject(e.Name(), e.Owner())
	o.Write(e.Payload(10))
	return o
}

func (e Env) NObj(n int) []*objst.Object {
	objs := make([]*objst.Object, 0, n)
	for i := 0; i < n; i++ {
		objs = append(objs, e.Obj())
	}
	return objs
}

func (e Env) NewUploadRequest(url string, params map[string]string, formKey string, path string) (*http.Request, error) {
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

func (e Env) Destroy() error {
	if err := e.b.Shutdown(); err != nil {
		return err
	}
	if err := os.RemoveAll(e.b.BasePath); err != nil {
		return err
	}
	e.ts.Close()
	return nil
}
