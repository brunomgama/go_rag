package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	host  string
	model string
	http  *http.Client
}

type reqInputArray struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type reqInputSingle struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type reqPromptSingle struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type respSingleObj struct {
	Embedding []float32 `json:"embedding"`
}

type respBatchObj struct {
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

func tryEmbeddingsFromMap(raw []byte) ([][]float32, bool) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, false
	}
	if v, ok := m["embeddings"]; ok {
		arr, ok := v.([]any)
		if !ok || len(arr) == 0 {
			return nil, false
		}
		out := make([][]float32, 0, len(arr))
		for _, e := range arr {
			af, ok := toFloat32Slice(e)
			if !ok {
				return nil, false
			}
			out = append(out, af)
		}
		return out, true
	}
	if v, ok := m["embedding"]; ok {
		af, ok := toFloat32Slice(v)
		if !ok {
			return nil, false
		}
		return [][]float32{af}, true
	}
	return nil, false
}

func toFloat32Slice(v any) ([]float32, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]float32, 0, len(list))
	for _, x := range list {
		switch t := x.(type) {
		case float64:
			out = append(out, float32(t))
		case float32:
			out = append(out, t)
		default:
			return nil, false
		}
	}
	return out, true
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	// --- 1) Try batch request with input: []string
	{
		bReq, _ := json.Marshal(reqInputArray{Model: c.model, Input: texts})
		vectors, ok, err := c.callAndParse(ctx, bReq)
		if err != nil {
			return nil, err
		}
		if ok {
			return vectors, nil
		}
	}

	// --- 2) Fallback: single request per text (input: string)
	if len(texts) == 1 {
		sReq, _ := json.Marshal(reqInputSingle{Model: c.model, Input: texts[0]})
		vectors, ok, err := c.callAndParse(ctx, sReq)
		if err != nil {
			return nil, err
		}
		if ok {
			return vectors, nil
		}
		// 3) Legacy fallback: "prompt"
		pReq, _ := json.Marshal(reqPromptSingle{Model: c.model, Prompt: texts[0]})
		vectors, ok, err = c.callAndParse(ctx, pReq)
		if err != nil {
			return nil, err
		}
		if ok {
			return vectors, nil
		}
	}

	return nil, fmt.Errorf("ollama embeddings: could not parse response in any known shape")
}

func (c *Client) callAndParse(ctx context.Context, body []byte) ([][]float32, bool, error) {
	req, _ := http.NewRequestWithContext(ctx, "POST", c.host+"/api/embeddings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("ollama embeddings status %d: %s", res.StatusCode, truncate(raw, 300))
	}

	var batch respBatchObj
	if err := json.Unmarshal(raw, &batch); err == nil && len(batch.Embeddings) > 0 {
		return batch.Embeddings, true, nil
	}
	var single respSingleObj
	if err := json.Unmarshal(raw, &single); err == nil && len(single.Embedding) > 0 {
		return [][]float32{single.Embedding}, true, nil
	}

	if out, ok := tryEmbeddingsFromMap(raw); ok {
		return out, true, nil
	}

	var bare []float32
	if err := json.Unmarshal(raw, &bare); err == nil && len(bare) > 0 {
		return [][]float32{bare}, true, nil
	}

	return nil, false, fmt.Errorf("unexpected embeddings response: %s", truncate(raw, 300))
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
