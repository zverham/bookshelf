package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var baseURL = "https://openlibrary.org"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SearchResult struct {
	NumFound int          `json:"numFound"`
	Docs     []SearchDoc  `json:"docs"`
}

type SearchDoc struct {
	Key           string   `json:"key"`
	Title         string   `json:"title"`
	AuthorName    []string `json:"author_name"`
	FirstPublishYear int  `json:"first_publish_year"`
	ISBN          []string `json:"isbn"`
	NumberOfPages int      `json:"number_of_pages_median"`
	CoverI        int      `json:"cover_i"`
}

func (d *SearchDoc) Author() string {
	if len(d.AuthorName) > 0 {
		return d.AuthorName[0]
	}
	return "Unknown Author"
}

func (d *SearchDoc) FirstISBN() *string {
	if len(d.ISBN) > 0 {
		return &d.ISBN[0]
	}
	return nil
}

func (d *SearchDoc) CoverURL() *string {
	if d.CoverI > 0 {
		url := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", d.CoverI)
		return &url
	}
	return nil
}

func (d *SearchDoc) Pages() *int {
	if d.NumberOfPages > 0 {
		return &d.NumberOfPages
	}
	return nil
}

func (c *Client) Search(query string, limit int) ([]SearchDoc, error) {
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s/search.json?q=%s&limit=%d", baseURL, encodedQuery, limit)

	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Docs, nil
}

type WorkDetails struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description interface{} `json:"description"`
}

func (w *WorkDetails) DescriptionText() *string {
	switch v := w.Description.(type) {
	case string:
		return &v
	case map[string]interface{}:
		if val, ok := v["value"].(string); ok {
			return &val
		}
	}
	return nil
}

func (c *Client) GetWorkDetails(key string) (*WorkDetails, error) {
	if !strings.HasPrefix(key, "/works/") {
		key = "/works/" + key
	}
	workURL := fmt.Sprintf("%s%s.json", baseURL, key)

	resp, err := c.httpClient.Get(workURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get work details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("work details returned status %d", resp.StatusCode)
	}

	var work WorkDetails
	if err := json.NewDecoder(resp.Body).Decode(&work); err != nil {
		return nil, fmt.Errorf("failed to decode work details: %w", err)
	}

	return &work, nil
}
