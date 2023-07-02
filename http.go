package objst

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naivary/objst/models"
)

type HTTPHandler struct {
	router chi.Router
	bucket *Bucket
}

func NewHTTPHandler(b *Bucket) *HTTPHandler {
	h := HTTPHandler{}
	r := chi.NewRouter()
	r.Route("/objst", func(r chi.Router) {
		r.Get("/{id}", h.read)
		r.Post("/", h.create)
		r.Delete("/{id}", h.remove)
	})
	h.bucket = b
	h.router = r
	return &h
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
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
}

func (h *HTTPHandler) Serve(addr string) error {
	return http.ListenAndServe(addr, h.router)
}
