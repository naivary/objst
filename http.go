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
	r.Get("/{id}", h.read)
	r.Post("/", h.create)
	r.Delete("/{id}", h.remove)
	h.bucket = b
	h.router = r
	return &h
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h HTTPHandler) read(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	obj, err := h.bucket.GetByID(id)
	if err != nil {
		msg := fmt.Sprintf("couldn't get the object with the id %s", id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if _, err := io.Copy(w, obj); err != nil {
		http.Error(w, "something went wrong while straming the data to the client", http.StatusInternalServerError)
		return
	}
}
func (h HTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	m := models.Object{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "something went wrong while decoding the data into the model", http.StatusBadRequest)
		return
	}
	obj := Object{}
	obj.FromModel(&m)
}

func (h HTTPHandler) remove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.bucket.DeleteByID(id); err != nil {
		msg := fmt.Sprintf("couldn't delete the object with the id %s", id)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
}
