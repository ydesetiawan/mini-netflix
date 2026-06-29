package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ydesetiawan/mini-netflix/internal/config"
)

type Client struct {
	baseURL string
	index   string
	http    *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.ESUrl,
		index:   cfg.ESIndexContent,
		http:    &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("elasticsearch error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// CreateIndex creates the content index with autocomplete mapping.
func (c *Client) CreateIndex(ctx context.Context) error {
	mapping := map[string]any{
		"settings": map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"autocomplete_analyzer": map[string]any{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "autocomplete_filter"},
					},
				},
				"filter": map[string]any{
					"autocomplete_filter": map[string]any{
						"type":     "edge_ngram",
						"min_gram": 1,
						"max_gram": 20,
					},
				},
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":           map[string]any{"type": "keyword"},
				"title":        map[string]any{"type": "text", "analyzer": "autocomplete_analyzer", "search_analyzer": "standard"},
				"description":  map[string]any{"type": "text"},
				"type":         map[string]any{"type": "keyword"},
				"release_year": map[string]any{"type": "integer"},
				"genres":       map[string]any{"type": "keyword"},
				"view_count":   map[string]any{"type": "long"},
				"rating_avg":   map[string]any{"type": "float"},
				"thumbnail_url": map[string]any{"type": "keyword", "index": false},
				"title_suggest": map[string]any{
					"type":            "completion",
					"analyzer":        "simple",
					"search_analyzer": "simple",
				},
			},
		},
	}

	path := fmt.Sprintf("/%s", c.index)
	_, err := c.do(ctx, http.MethodPut, path, mapping)
	return err
}

// IndexContent upserts a content document into Elasticsearch.
func (c *Client) IndexContent(ctx context.Context, doc map[string]any) error {
	id, _ := doc["id"].(string)
	path := fmt.Sprintf("/%s/_doc/%s", c.index, id)
	_, err := c.do(ctx, http.MethodPut, path, doc)
	return err
}

// Autocomplete returns title suggestions using the completion suggester.
func (c *Client) Autocomplete(ctx context.Context, prefix string) ([]string, error) {
	query := map[string]any{
		"suggest": map[string]any{
			"title_suggest": map[string]any{
				"prefix": prefix,
				"completion": map[string]any{
					"field": "title_suggest",
					"size":  10,
				},
			},
		},
	}

	path := fmt.Sprintf("/%s/_search", c.index)
	data, err := c.do(ctx, http.MethodPost, path, query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Suggest map[string][]struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"suggest"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var suggestions []string
	for _, opts := range result.Suggest["title_suggest"] {
		for _, opt := range opts.Options {
			suggestions = append(suggestions, opt.Text)
		}
	}
	return suggestions, nil
}

// Search performs a full-text + fuzzy search on content.
func (c *Client) Search(ctx context.Context, q string, from, size int) ([]map[string]any, int, error) {
	query := map[string]any{
		"from": from,
		"size": size,
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     q,
				"fields":    []string{"title^3", "description", "genres"},
				"fuzziness": "AUTO",
			},
		},
	}

	path := fmt.Sprintf("/%s/_search", c.index)
	data, err := c.do(ctx, http.MethodPost, path, query)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, err
	}

	docs := make([]map[string]any, 0, len(result.Hits.Hits))
	for _, h := range result.Hits.Hits {
		docs = append(docs, h.Source)
	}
	return docs, result.Hits.Total.Value, nil
}
