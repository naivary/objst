package objst

import (
	"context"
	"fmt"
	"net/http"

	"github.com/naivary/objst/random"
)

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), CtxKeyReqID, fmt.Sprintf("%s/%s", r.Host, random.ID(5)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func assureOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := r.Context().Value(CtxKeyOwner).(string)
		if !ok {
			http.Error(w, "owner in request context is not a string", http.StatusBadRequest)
			return
		}
		if owner == "" {
			http.Error(w, ErrMissingOwner.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
