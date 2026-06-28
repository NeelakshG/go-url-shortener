package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NeelakshG/snippy/internal/model"
)

type fakeStore struct {
	CreateLinkFn      func(ctx context.Context, longURL string) (*model.Link, error)
	GetLinkFn         func(ctx context.Context, shortCode string) (*model.Link, error)
	IncrementClicksFn func(ctx context.Context, shortCode string) error
	GetClickCountFn   func(ctx context.Context, shortCode string) (int64, error)
}

func (f *fakeStore) CreateLink(ctx context.Context, longURL string) (*model.Link, error) {
	return f.CreateLinkFn(ctx, longURL)
}

func (f *fakeStore) GetLink(ctx context.Context, shortCode string) (*model.Link, error) {
	return f.GetLinkFn(ctx, shortCode)
}

func (f *fakeStore) IncrementClicks(ctx context.Context, shortCode string) error {
	return f.IncrementClicksFn(ctx, shortCode)
}

func (f *fakeStore) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	return f.GetClickCountFn(ctx, shortCode)
}

func TestCreateLink(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		storeResp      *model.Link
		storeErr       error
		expectedStatus int
	}{
		{
			name: "happy path",
			body: `{"url":"https://google.com"}`,
			storeResp: &model.Link{
				ShortCode: "abc123",
				LongURL:   "https://google.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing url field",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "store error",
			body:           `{"url":"https://google.com"}`,
			storeErr:       errors.New("db failure"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{
				CreateLinkFn: func(ctx context.Context, url string) (*model.Link, error) {
					return tt.storeResp, tt.storeErr
				},
			}

			h := &Links{Store: store}

			req := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.CreateLink(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	store := &fakeStore{
		GetLinkFn: func(ctx context.Context, code string) (*model.Link, error) {
			return &model.Link{
				ShortCode: code,
				LongURL:   "https://google.com",
			}, nil
		},
		IncrementClicksFn: func(ctx context.Context, code string) error {
			return nil
		},
	}

	h := &Links{Store: store}

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req.SetPathValue("code", "abc123")

	w := httptest.NewRecorder()

	h.Resolve(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}

	if loc := w.Header().Get("Location"); loc != "https://google.com" {
		t.Fatalf("expected redirect to google.com, got %s", loc)
	}
}

func TestStats(t *testing.T) {
	store := &fakeStore{
		GetClickCountFn: func(ctx context.Context, code string) (int64, error) {
			return 42, nil
		},
	}

	h := &Links{Store: store}

	req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
	req.SetPathValue("code", "abc123")

	w := httptest.NewRecorder()

	h.Stats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if resp["short_code"] != "abc123" {
		t.Fatalf("expected short_code abc123, got %v", resp["short_code"])
	}

	if resp["clicks"] != float64(42) {
		t.Fatalf("expected clicks 42, got %v", resp["clicks"])
	}
}
