package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("expected non-nil client")
	}
	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}
}

func TestSearchDocAuthor(t *testing.T) {
	tests := []struct {
		name       string
		authorName []string
		expected   string
	}{
		{"with authors", []string{"Andy Hunt", "Dave Thomas"}, "Andy Hunt"},
		{"empty authors", []string{}, "Unknown Author"},
		{"nil authors", nil, "Unknown Author"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := SearchDoc{AuthorName: tt.authorName}
			if doc.Author() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, doc.Author())
			}
		})
	}
}

func TestSearchDocFirstISBN(t *testing.T) {
	tests := []struct {
		name     string
		isbns    []string
		expected *string
	}{
		{"with isbns", []string{"978-0135957059", "978-0201633610"}, strPtr("978-0135957059")},
		{"empty isbns", []string{}, nil},
		{"nil isbns", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := SearchDoc{ISBN: tt.isbns}
			result := doc.FirstISBN()
			if tt.expected == nil && result != nil {
				t.Errorf("expected nil, got %s", *result)
			}
			if tt.expected != nil && (result == nil || *result != *tt.expected) {
				t.Errorf("expected %s, got %v", *tt.expected, result)
			}
		})
	}
}

func TestSearchDocCoverURL(t *testing.T) {
	tests := []struct {
		name     string
		coverI   int
		hasURL   bool
	}{
		{"with cover", 12345, true},
		{"no cover", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := SearchDoc{CoverI: tt.coverI}
			result := doc.CoverURL()
			if tt.hasURL && result == nil {
				t.Error("expected URL, got nil")
			}
			if !tt.hasURL && result != nil {
				t.Errorf("expected nil, got %s", *result)
			}
			if tt.hasURL && result != nil {
				expected := "https://covers.openlibrary.org/b/id/12345-M.jpg"
				if *result != expected {
					t.Errorf("expected %s, got %s", expected, *result)
				}
			}
		})
	}
}

func TestSearchDocPages(t *testing.T) {
	tests := []struct {
		name     string
		pages    int
		expected *int
	}{
		{"with pages", 352, intPtr(352)},
		{"no pages", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := SearchDoc{NumberOfPages: tt.pages}
			result := doc.Pages()
			if tt.expected == nil && result != nil {
				t.Errorf("expected nil, got %d", *result)
			}
			if tt.expected != nil && (result == nil || *result != *tt.expected) {
				t.Errorf("expected %d, got %v", *tt.expected, result)
			}
		})
	}
}

func TestClientSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query().Get("q")
		// Query is URL-decoded by Go's URL parser
		if query != "test query" {
			t.Errorf("unexpected query: %s", query)
		}

		response := SearchResult{
			NumFound: 2,
			Docs: []SearchDoc{
				{Key: "/works/OL1", Title: "Book 1", AuthorName: []string{"Author 1"}},
				{Key: "/works/OL2", Title: "Book 2", AuthorName: []string{"Author 2"}},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override baseURL for testing
	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	docs, err := client.Search("test query", 5)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("expected 2 docs, got %d", len(docs))
	}

	if docs[0].Title != "Book 1" {
		t.Errorf("expected 'Book 1', got %s", docs[0].Title)
	}
}

func TestClientSearchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	_, err := client.Search("test", 5)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestWorkDetailsDescriptionText(t *testing.T) {
	tests := []struct {
		name        string
		description interface{}
		expected    *string
	}{
		{"string description", "A great book", strPtr("A great book")},
		{"map description", map[string]interface{}{"value": "A map description"}, strPtr("A map description")},
		{"nil description", nil, nil},
		{"other type", 123, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			work := WorkDetails{Description: tt.description}
			result := work.DescriptionText()
			if tt.expected == nil && result != nil {
				t.Errorf("expected nil, got %s", *result)
			}
			if tt.expected != nil && (result == nil || *result != *tt.expected) {
				t.Errorf("expected %s, got %v", *tt.expected, result)
			}
		})
	}
}

func TestClientGetWorkDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/works/OL123.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := WorkDetails{
			Key:         "/works/OL123",
			Title:       "Test Book",
			Description: "A test description",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()

	// Test with /works/ prefix
	work, err := client.GetWorkDetails("/works/OL123")
	if err != nil {
		t.Fatalf("failed to get work details: %v", err)
	}

	if work.Title != "Test Book" {
		t.Errorf("expected 'Test Book', got %s", work.Title)
	}
}

func TestClientGetWorkDetailsWithoutPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/works/OL456.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := WorkDetails{Key: "/works/OL456", Title: "Another Book"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()

	// Test without /works/ prefix
	work, err := client.GetWorkDetails("OL456")
	if err != nil {
		t.Fatalf("failed to get work details: %v", err)
	}

	if work.Title != "Another Book" {
		t.Errorf("expected 'Another Book', got %s", work.Title)
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
