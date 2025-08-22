package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	host  string
	model string
	http  *http.Client
}

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func NewOllama(host, model string) *Client {
	if host == "" {
		host = "http://localhost:11434"
	}

	return &Client{
		host:  host,
		model: model,
		http:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, _ := json.Marshal(ollamaEmbedRequest{Model: c.model, Input: texts})
	req, _ := http.NewRequestWithContext(ctx, "POST", c.host+"/api/embeddings", bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	// if res.StatusCode/100 != 2 {
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("ollama embeddings status %d", res.StatusCode)
	}

	var out ollamaEmbedResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Embeddings, nil

}
