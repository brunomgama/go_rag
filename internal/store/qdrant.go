package store

import (
	"context"
	"fmt"

	"github.com/brunomgama/go_rag/internal/common"
	"github.com/go-resty/resty/v2"
)

type SearchRequest struct {
	Vector         []float32      `json:"vector"`
	TopK           int            `json:"top"`
	WithPayload    bool           `json:"with_payload"`
	WithVector     bool           `json:"with_vector"`
	ScoreThreshold *float32       `json:"score_threshold,omitempty"`
	Filter         map[string]any `json:"filter,omitempty"`
}

type SearchResult struct {
	ID      any            `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

type searchResp struct {
	Result []SearchResult `json:"result"`
	Status string         `json:"status"`
	Time   float64        `json:"time"`
}

type Qdrant struct {
	http       *resty.Client
	BaseURL    string
	Collection string
}

type Point struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

func NewQdrant(baseUrl, collection string) *Qdrant {
	client := resty.New().SetBaseURL(baseUrl)
	return &Qdrant{http: client, BaseURL: baseUrl, Collection: collection}
}

func (q *Qdrant) EnsureCollection(ctx context.Context, dim int) error {
	body := map[string]any{
		"vectors": map[string]any{
			"size":     dim,
			"distance": "Cosine",
		},
	}

	var res map[string]any
	path := fmt.Sprintf("/collections/%s", q.Collection)

	_, err := q.http.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&res).
		Put(path)
	return err
}

func (q *Qdrant) Upsert(ctx context.Context, points []Point) error {
	body := map[string]any{
		"points": points,
	}

	var res map[string]any
	path := fmt.Sprintf("/collections/%s/points", q.Collection)

	_, err := q.http.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&res).
		Put(path)
	return err
}

func (q *Qdrant) Search(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	var out searchResp

	path := fmt.Sprintf("/collection/%s/points/search", q.Collection)
	_, err := q.http.R().SetContext(ctx).SetBody(req).SetResult(&out).Post(path)

	if !common.IsNilValue(err) {
		return nil, err
	}

	return out.Result, nil
}
