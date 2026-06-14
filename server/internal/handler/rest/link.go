package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"shortify/server/internal/domain"
	"shortify/server/internal/middleware"
	"shortify/server/internal/service"

	"github.com/google/uuid"
)

type LinkHandler struct {
	links *service.LinkService
}

func NewLinkHandler(links *service.LinkService) *LinkHandler {
	return &LinkHandler{links: links}
}

type createLinkBody struct {
	URL string `json:"url"`
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}

	var body createLinkBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	link, err := h.links.Create(r.Context(), userID, body.URL)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, link)
}

func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}

	links, err := h.links.List(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, links)
}

func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}

	linkID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	if err := h.links.Delete(r.Context(), userID, linkID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, domain.ErrForbidden)
		return
	}

	linkID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, domain.ErrInvalidInput)
		return
	}

	total, recent, err := h.links.GetStats(r.Context(), userID, linkID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total_clicks": total,
		"recent":       recent,
	})
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		return
	}

	target, err := h.links.Resolve(r.Context(), code, r.RemoteAddr, r.UserAgent())
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}
