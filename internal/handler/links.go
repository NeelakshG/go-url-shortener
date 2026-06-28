package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NeelakshG/snippy/internal/model"
)

// Store is the interface our handlers depend on.
// Using an interface (not a concrete type) means we can swap in a fake store during tests.
type Store interface {
	// CreateLink now owns short-code generation internally — caller just passes the URL.
	CreateLink(ctx context.Context, longURL string) (*model.Link, error)
	GetLink(ctx context.Context, shortCode string) (*model.Link, error)
	IncrementClicks(ctx context.Context, shortCode string) error
	GetClickCount(ctx context.Context, shortCode string) (int64, error)
}

// Links groups all link-related handlers and their shared dependency (the store).
type Links struct {
	Store Store
}

// CreateLink handles POST /links.
// Reference shape for every handler in this project: decode → validate → store → respond.
func (h *Links) CreateLink(w http.ResponseWriter, r *http.Request) {
	// 1. Decode — pull the URL out of the JSON body
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 2. Validate — reject empty URLs
	// TODO (you): also reject non-http(s) URLs to prevent open-redirect attacks
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// 3. Store — short-code generation + collision handling lives inside CreateLink on the store
	link, err := h.Store.CreateLink(r.Context(), req.URL)
	if err != nil {
		http.Error(w, "failed to create link", http.StatusInternalServerError)
		return
	}

	// 4. Respond — 201 Created with the full link object as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

// Resolve handles GET /{code}.
// YOUR task: implement this handler following the same decode→validate→store→respond shape.
// Hint: "decode" here means extracting the {code} path parameter — use r.PathValue("code").
// Hint: "respond" means an HTTP 302 redirect, not a JSON body — use http.Redirect.
// Hint: also call IncrementClicks so the stats endpoint has data.
func (h *Links) Resolve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract short code from URL
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// 2. Look up link in DB
	link, err := h.Store.GetLink(ctx, code)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// 3. Track click (don’t block redirect if this fails)
	go func() {
		_ = h.Store.IncrementClicks(context.Background(), code)
	}()

	// 4. Redirect to original URL
	http.Redirect(w, r, link.LongURL, http.StatusFound)
}

// Stats handles GET /stats/{code}.
// YOUR task: look up the click count for a code and return it as JSON.
func (h *Links) Stats(w http.ResponseWriter, r *http.Request) {
    code := r.PathValue("code")

    count, err := h.Store.GetClickCount(r.Context(), code)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    // write JSON response here
	resp := struct {
    ShortCode string `json:"short_code"`
    Clicks    int64  `json:"clicks"`
}{
    ShortCode: code,
    Clicks:    count,
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(resp)
}
