package handler

import (
	"context"
	"encoding/json"
	"errors"
	"link-storage-service/internal/domain"
	"net/http"
	"strconv"
)

type linkService interface {
	Create(ctx context.Context, originalURL string) (*domain.Link, error)
	Get(ctx context.Context, shortCode string) (*domain.Link, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Link, error)
	Delete(ctx context.Context, shortCode string) error
	Stats(ctx context.Context, shortCode string) (*domain.Link, error)
}

type LinkHandler struct {
	svc linkService
}

func NewLinkHandler(svc linkService) *LinkHandler {
	return &LinkHandler{svc: svc}
}

func (h *LinkHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /links", h.create)
	mux.HandleFunc("GET /links", h.list)
	mux.HandleFunc("GET /links/{short_code}/stats", h.stats)
	mux.HandleFunc("GET /links/{short_code}", h.get)
	mux.HandleFunc("DELETE /links/{short_code}", h.delete)
}

// POST /links
func (h *LinkHandler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	link, err := h.svc.Create(r.Context(), req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"short_code": link.ShortCode})
}

// GET /links/{short_code}
func (h *LinkHandler) get(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")

	link, err := h.svc.Get(r.Context(), shortCode)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"url":    link.OriginalURL,
		"visits": link.Visits,
	})
}

// GET /links
func (h *LinkHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	links, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if links == nil {
		links = []*domain.Link{}
	}
	writeJSON(w, http.StatusOK, links)
}

// DELETE /links/{short_code}
func (h *LinkHandler) delete(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")

	if err := h.svc.Delete(r.Context(), shortCode); err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /links/{short_code}/stats
func (h *LinkHandler) stats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")

	link, err := h.svc.Stats(r.Context(), shortCode)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"short_code": link.ShortCode,
		"url":        link.OriginalURL,
		"visits":     link.Visits,
		"created_at": link.CreatedAt,
	})
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
