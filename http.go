package objst

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/naivary/objst/models"
	"golang.org/x/exp/slog"
)

type HTTPHandlerOptions struct {
	maxUploadSize int64
	formKeyFile   string
}

type HTTPHandler struct {
	router chi.Router
	bucket *Bucket
	logger *slog.Logger
	opts   HTTPHandlerOptions
}

func NewHTTPHandler(b *Bucket) *HTTPHandler {
	h := HTTPHandler{}
	r := chi.NewRouter()
	r.Route("/objst", func(r chi.Router) {
		r.Get("/read/{id}", h.read)
		r.Get("/{id}", h.get)
		r.Post("/", h.create)
		r.Delete("/{id}", h.remove)
		r.Route("/upload", func(r chi.Router) {
			r.Use(h.assureContentType)
			r.Post("/", h.upload)
		})
	})
	h.bucket = b
	h.router = r
	h.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &h
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *HTTPHandler) assureContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		contentType := r.Form.Get("contentType")
		if contentType == "" {
			http.Error(w, "missing contentType in request form", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), ContentTypeMetaKey, contentType)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (h *HTTPHandler) get(w http.ResponseWriter, r *http.Request) {}

func (h *HTTPHandler) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(h.opts.maxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile(h.opts.formKeyFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	owner, ok := r.Context().Value("owner").(string)
	if !ok {
		http.Error(w, "owner in request context is not a string", http.StatusInternalServerError)
		return
	}
	obj, err := NewObject(header.Filename, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(obj, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := r.Context().Value(ContentTypeMetaKey).(string)
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
