package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brunomgama/go_rag/internal/common"
)

type Client struct {
	host  string
	model string
	http  *http.Client
}

type generatedRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

type generatedResponse struct {
	Response string `json:"response"`
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

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(generatedRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]any{
			"temperature": 0.2,
		},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", c.host+"/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)

	if !common.IsNilValue(err) {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("ollama generate status %d", res.StatusCode)
	}

	var out generatedResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Response, nil

}
