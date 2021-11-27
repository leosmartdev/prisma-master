package main

import (
	"net/http"
	"strings"
)

type Handler struct {
	http.Handler
	handlers map[string]http.Handler
}

func NewHandler() *Handler {
	return &Handler{
		handlers: make(map[string]http.Handler),
	}
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// check exact match first
	if handler, ok := h.handlers[request.RequestURI]; ok {
		handler.ServeHTTP(response, request)
		return
	}
	for path, handler := range h.handlers {
		if strings.HasPrefix(request.RequestURI, path) {
			handler.ServeHTTP(response, request)
			break
		}
	}
}

func (h *Handler) Add(path string, handler http.Handler) {
	h.handlers[path] = handler
}
