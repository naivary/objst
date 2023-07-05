package objst

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

const (
	contentType     = "Content-type"
	applicationJSON = "application/json"
)

const defaultTimeout = 5 * time.Second

type CtxKey string

const (
	CtxKeyOwner CtxKey = "owner"
	CtxKeyReqID CtxKey = "reqid"
)

type objectModel struct {
	ID       string             `json:"id,omitempty"`
	Name     string             `json:"name,omitempty"`
	Owner    string             `json:"owner,omitempty"`
	Metadata map[MetaKey]string `json:"metadata,omitempty"`
}

type HTTPHandler struct {
	bucket *Bucket
	opts   HTTPHandlerOptions
}

func NewHTTPHandler(bucket *Bucket, opts HTTPHandlerOptions) *HTTPHandler {
	hl := HTTPHandler{}
	hl.opts = opts
	if opts.Handler == nil {
		hl.opts.Handler = hl.routes()
	}
	hl.bucket = bucket
	return &hl
}

// ServeHTTP implements http.Handler
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.opts.Handler.ServeHTTP(w, r)
}

func (h *HTTPHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.opts.IsAuthenticated)
	r.Use(requestID)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Timeout(defaultTimeout))

	r.Route("/objst", func(r chi.Router) {
		r.Route("/", func(r chi.Router) {
			r.Use(h.opts.IsAuthorized)
			r.Get("/read/{id}", h.Read)
			r.Get("/{id}", h.Get)
			r.Delete("/{id}", h.Remove)
		})
		r.Route("/upload", func(r chi.Router) {
			r.Use(assureOwner)
			r.Post("/", h.Upload)
		})
	})
	return r
}

// Get will return the object model witht he given
// payload iff any object is found.
func (h *HTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(CtxKeyReqID).(string)
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "id is an invalid uuid-v4", http.StatusBadRequest)
		return
	}
	obj, err := h.bucket.GetByID(id)
	if err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(obj.ToModel()); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) Upload(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(CtxKeyReqID).(string)
	if err := r.ParseMultipartForm(h.opts.MaxUploadSize); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "something went wrong while parsing the multipart form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile(h.opts.FormKey)
	if err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "couldn't get the file from the multipart form", http.StatusInternalServerError)
		return
	}
	owner := r.Context().Value(CtxKeyOwner).(string)
	obj, err := NewObject(header.Filename, owner)
	if err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(obj, file); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "couldn't copy the payload of the file into the object", http.StatusInternalServerError)
		return
	}
	if obj.GetMetaKey(MetaKeyContentType) == "" {
		contentType := r.Form.Get("contentType")
		if contentType == "" {
			msg := "contentType meta key was not set for the object"
			h.opts.Logger.ErrorCtx(r.Context(), msg, slog.String("req_id", reqID))
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		obj.SetMetaKey(MetaKeyContentType, contentType)
	}
	if err := h.bucket.Create(obj); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "something went wrong while creating the object", http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(obj.ToModel()); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "something went wrong while sending the object model", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) Read(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(CtxKeyReqID).(string)
	id := chi.URLParam(r, "id")
	if err := h.bucket.Read(id, w); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "something went wrong while streaming the object", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) Remove(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(CtxKeyReqID).(string)
	id := chi.URLParam(r, "id")
	if err := h.bucket.DeleteByID(id); err != nil {
		h.opts.Logger.ErrorCtx(r.Context(), err.Error(), slog.String("req_id", reqID))
		http.Error(w, "couldn't delete the object with the id: "+id, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
