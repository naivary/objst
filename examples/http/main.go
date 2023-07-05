package main

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/naivary/objst"
	"golang.org/x/exp/slog"
)

func main() {
	opts := objst.NewDefaultBucketOptions()
	bucket, err := objst.NewBucket(opts)
	if err != nil {
		panic(err)
	}
	// serve the created bucket over http
	handler := objst.NewHTTPHandler(bucket, objst.DefaultHTTPHandlerOptions())
	ow := injectOwner(handler)
	slog.Info("http server running!", "addr", "localhost:8080")
	if err := http.ListenAndServe(":8080", ow); err != nil {
		panic(err)
	}
}

func injectOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner := uuid.NewString()
		ctx := context.WithValue(r.Context(), objst.CtxKeyOwner, owner)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
