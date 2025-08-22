package store

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

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
	path := fmt.Sprintf("collections/%s", q.Collection)

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
	path := fmt.Sprintf("collections/%s/points", q.Collection)

	_, err := q.http.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&res).
		Put(path)
	return err
}