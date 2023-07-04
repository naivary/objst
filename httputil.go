package objst

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
)

func (h *HTTPHandler) getContentType(filename string, r *http.Request) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = r.Form.Get(string(MetaKeyContentType))
	}
	if contentType == "" {
		return "", errors.New("content type of the file is not an official mime-type and no contentType key could be found in the form")
	}
	return contentType, nil
}
