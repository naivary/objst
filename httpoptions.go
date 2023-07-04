package objst

import "net/http"

type HTTPHandlerOptions struct {
	// MaxUploadSize is limiting the size of a file
	// which can be uploaded using the /objst/upload
	// endpoint. Default: 32 MB.
	MaxUploadSize int64

	// FormKey is the key to access the file
	// in the multipart form. Default: "file"
	FormKey string

	// IsAuthorized is the middleware used to validate
	// if the incoming request is considered authorized.
	// By default all request will be considered authorized.
	// This middleware is further usedj to inject the `CtxKeyOwner`
	// key into the request context to set the owner.
	IsAuthorized func(http.Handler) http.Handler

	// IsAuthenticated is the middleware used to validate
	// if the incoming request is considered authenticated.
	// By default all request will be considered authenticated.
	IsAuthenticated func(http.Handler) http.Handler

	// Handler is used to serve public http traffic.
	// By default if the handler is nil it will be replaced
	// by the default handler.
	Handler http.Handler
}

func DefaultHTTPHandlerOptions() HTTPHandlerOptions {
	const (
		mib32   = 2 << 24
		formKey = "file"
	)
	opts := HTTPHandlerOptions{}

	opts.MaxUploadSize = mib32
	opts.FormKey = formKey
	opts.IsAuthorized = isAuthorized
	opts.IsAuthenticated = isAuthenticated
	opts.Handler = nil
	return opts
}

// isAuthorized is the default authorization middleware which accepts all requests.
func isAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// isAuthenticated is the default authentication middlware which accepts all requests.
func isAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
