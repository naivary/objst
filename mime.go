package objst

import "mime"

// AddExtensionType allows you to add an custom
// file extension e.g. `.test` and the associated
// Content-Type with the extension for example `text/plain`.
func AddExtensionType(ext string, typ string) error {
	return mime.AddExtensionType(ext, typ)
}
