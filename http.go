package objst

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/naivary/objst/models"
	"golang.org/x/exp/slog"
)

type CtxKey string

const (
	CtxKeyOwner CtxKey = "owner"
)

var (
	ErrMissingOwner = errors.New("missing owner in the request context")
)

type HTTPHandlerOptions struct {
	// MaxUploadSize is limiting the size of a file
	// which can be uploaded using the /objst/upload
	// endpoint. Default: 32 MB.
	MaxUploadSize int64
	// FormKey is the key to access the file
	// in the multipart form. Default: "file"
	FormKey string
	// IsAuthorized is the middleware used to authorize
	// the incoming request. By default no authorization
	// checks will be done.
	IsAuthorized func(http.Handler) http.Handler
}

type HTTPHandler struct {
	router chi.Router
	bucket *Bucket
	logger *slog.Logger
	opts   HTTPHandlerOptions
}

func DefaultHTTPHandlerOptions() HTTPHandlerOptions {
	opts := HTTPHandlerOptions{}
	// ~33 MB
	opts.MaxUploadSize = 32 << 20
	opts.FormKey = "file"
	opts.IsAuthorized = isAuthorized
	return opts
}

func NewHTTPHandler(b *Bucket, opts HTTPHandlerOptions) *HTTPHandler {
	h := HTTPHandler{}
	r := chi.NewRouter()
	h.opts = opts
	r.Route("/objst", func(r chi.Router) {
		r.Post("/", h.create)
		// authorization needed routes
		r.Route("/", func(r chi.Router) {
			r.Use(h.opts.IsAuthorized)
			r.Get("/read/{id}", h.read)
			r.Get("/{id}", h.get)
			r.Delete("/{id}", h.remove)
		})
		r.Route("/upload", func(r chi.Router) {
			r.Use(h.assureOwner)
			r.Post("/", h.upload)
		})
	})
	h.bucket = b
	h.router = r
	h.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &h
}

// isAuthorized is the default authorization checker which accepts all requests.
func isAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP implements http.Handler
func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Context().Value(CtxKeyOwner))
	h.router.ServeHTTP(w, r)
}

func (h *HTTPHandler) assureOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := r.Context().Value(CtxKeyOwner).(string)
		if !ok {
			http.Error(w, "owner in request context is not a string", http.StatusInternalServerError)
			return
		}
		if owner == "" {
			http.Error(w, ErrMissingOwner.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *HTTPHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	obj, err := h.bucket.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(obj.ToModel()); err != nil {
		http.Error(w, "something went wrong while send the object", http.StatusInternalServerError)
		return
	}
}

// TODO: allow custom content type to be passed in the form. If set
// the will take precedence.
func (h *HTTPHandler) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.opts.MaxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile(h.opts.FormKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ext := strings.Split(header.Filename, ".")[1]
	owner := r.Context().Value(CtxKeyOwner).(string)
	obj, err := NewObject(header.Filename, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(obj, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := mime.TypeByExtension(ext)
	obj.SetMeta(ContentTypeMetaKey, contentType)
	if err := h.bucket.Create(obj); err != nil {
		http.Error(w, "something went wrong while creating the object", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) read(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	obj, err := h.bucket.GetByID(id)
	if err != nil {
		msg := fmt.Sprintf("something went wrong while getting the object. ID: %s", id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if _, err := io.Copy(w, obj); err != nil {
		http.Error(w, "something went wrong while streaming the object", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	m := models.Object{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "something went wrong while decoding the data into the model", http.StatusBadRequest)
		return
	}
	obj, err := FromModel(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.bucket.Create(obj); err != nil {
		http.Error(w, "something went wrong while creating the object", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(obj.ToModel()); err != nil {
		http.Error(w, "couldn't send the object back", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.bucket.DeleteByID(id); err != nil {
		msg := fmt.Sprintf("couldn't delete the object with the id %s", id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) Serve(addr string) error {
	return http.ListenAndServe(addr, h.router)
}
