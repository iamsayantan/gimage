package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	// chiware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/iamsayantan/gimage"
)

// WebHandler is the interface for request handlers.
type WebHandler interface {
	Route() chi.Router
}

// Server holds the dependencies for an http webserver.
type Server struct {
	resizer  *gimage.Resizer
	uploader *gimage.S3Uploader

	router chi.Router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// NewServer returns an HTTP server implementation.
func NewServer(resizer *gimage.Resizer, uploader *gimage.S3Uploader) *Server {
	srv := &Server{
		resizer:  resizer,
		uploader: uploader,
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	})

	r := chi.NewRouter()
	// r.Use(chiware.AllowContentType("application/json", "application/x-www-form-urlencoded"))
	r.Use(corsHandler.Handler)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Method("GET", "/", http.HandlerFunc(rootHandler))

	// Upload route handlers
	r.Route("/upload", func(r chi.Router) {
		uploadHandler := NewUploadHandler(srv.resizer, srv.uploader)
		r.Mount("/v1", uploadHandler.Route())
	})

	srv.router = r
	return srv
}

// rootHandler is the handler for the root route [/]. Just a basic welcome message.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Message string `json:"message"`
	}{Message: "Welcome to Gimage."}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// sendResponse marshalls the given struct and writes to the response writer.
func sendResponse(w http.ResponseWriter, statusCode int, message string, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	responsePayload := struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}{Message: message, Data: v}
	resp, _ := json.Marshal(responsePayload)

	w.WriteHeader(statusCode)
	_, _ = w.Write(resp)
}

// sendError works same as sendResponse but this works for errors only. Need a way to DRY it up.
func sendError(w http.ResponseWriter, errorCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	responsePayload := struct {
		Message string `json:"message"`
	}{Message: errorMessage}
	resp, _ := json.Marshal(responsePayload)

	w.WriteHeader(errorCode)
	_, _ = w.Write(resp)
}
