package objst

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

type CtxKey string

const (
	CtxKeyOwner CtxKey = "owner"
)

type objectModel struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Owner    string             `json:"owner"`
	Metadata map[MetaKey]string `json:"metadata"`
}

type HTTPHandler struct {
	bucket *Bucket
	logger *slog.Logger
	opts   HTTPHandlerOptions
}

func NewHTTPHandler(bucket *Bucket, opts HTTPHandlerOptions) *HTTPHandler {
	hl := HTTPHandler{}
	hl.opts = opts
	if opts.Handler == nil {
		hl.opts.Handler = hl.routes()
	}
	hl.bucket = bucket
	hl.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &hl
}

// ServeHTTP implements http.Handler
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.opts.Handler.ServeHTTP(w, r)
}

func (h *HTTPHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.opts.IsAuthenticated)
	r.Route("/objst", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(h.opts.IsAuthorized)
			r.Get("/read/{id}", h.Read)
			r.Get("/{id}", h.Get)
			r.Delete("/{id}", h.Remove)
		})
		r.Route("/upload", func(r chi.Router) {
			r.Use(h.assureOwner)
			r.Post("/", h.Upload)
		})
	})
	return r
}

func (h *HTTPHandler) assureOwner(next http.Handler) http.Handler {
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

// Get will return the object model witht he given
// payload iff any object is found.
func (h *HTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := NewQuery()
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
func (h *HTTPHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.opts.MaxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile(h.opts.FormKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	contentType, err := h.getContentType(header.Filename, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	obj.SetMetaKey(MetaKeyContentType, contentType)
	if err := h.bucket.Create(obj); err != nil {
		http.Error(w, "something went wrong while creating the object", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(obj.ToModel()); err != nil {
		http.Error(w, "something went wrong while sending the object model", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) Read(w http.ResponseWriter, r *http.Request) {
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

func (h *HTTPHandler) Remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.bucket.DeleteByID(id); err != nil {
		msg := fmt.Sprintf("couldn't delete the object with the id %s", id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
