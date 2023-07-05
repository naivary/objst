package objst

import (
	"mime"
	"net/http"
	"path/filepath"
)

// getContentType returns the official mime-type for a given filename
// extension. Iff none is found the user defined content-type in
// the request form, if any, will be returned.
func (h *HTTPHandler) getContentType(filename string, r *http.Request) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = r.Form.Get(MetaKeyContentType.String())
	}
	if contentType == "" {
		return "", ErrUknownContentType
	}
	return contentType, nil
}
