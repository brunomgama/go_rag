package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brunomgama/go_rag/internal/common"
	"github.com/brunomgama/go_rag/internal/embed"
	"github.com/brunomgama/go_rag/internal/llm"
	"github.com/brunomgama/go_rag/internal/store"
)

type Service struct {
	Embed    *embed.Client
	LLM      *llm.Client
	Store    *store.Qdrant
	TopK     int
	MinScore *float32
}

type Citation struct {
	DocID   string  `json:"doc_id"`
	Page    int     `json:"page"`
	ChunkID string  `json:"chunk_id"`
	Score   float32 `json:"score"`
	Snippet string  `json:"snippet"`
}

type Answer struct {
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
	LatencyMS int64      `json:"latency_ms"`
}

func (s *Service) embedQuery(ctx context.Context, question string) ([]float32, error) {
	vecs, err := s.Embed.Embed(ctx, []string{question})

	if !common.IsNilValue(err) {
		return nil, err
	}

	if len(vecs) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return vecs[0], nil
}

func (s *Service) Query(ctx context.Context, question string, topK int) (Answer, error) {
	start := time.Now()
	if topK <= 0 {
		topK = s.TopK
	}

	vec, err := s.embedQuery(ctx, question)
	if !common.IsNilValue(err) {
		return Answer{}, nil
	}

	req := store.SearchRequest{
		Vector:      vec,
		TopK:        topK,
		WithPayload: true,
		WithVector:  false,
	}

	if !common.IsNilValue(s.MinScore) {
		req.ScoreThreshold = s.MinScore
	}

	results, err := s.Store.Search(ctx, req)

	if !common.IsNilValue(err) {
		return Answer{}, nil
	}

	var src strings.Builder
	citations := make([]Citation, 0, len(results))

	for i, r := range results {
		if common.IsNilValue(r.Payload) {
			continue
		}

		text, _ := r.Payload["text"].(string)
		doc, _ := r.Payload["doc_id"].(string)
		page := common.AsInt(r.Payload["page"])
		chunk, _ := r.Payload["chunk_id"].(string)

		fmt.Fprintf(&src, "\n[Source %d] (%s p.%d, %s)\n%s\n", i+1, doc, page, chunk, common.Clamp(text, 900))
		citations = append(citations, Citation{
			DocID: doc, Page: page, ChunkID: chunk, Score: r.Score, Snippet: common.Snippet(text, 280),
		})
	}

	prompt := strings.TrimSpace(fmt.Sprintf(`
You are a precise assistant. Answer the user USING ONLY the Sources below.
If the answer is not in the sources, say "I don't know".
Cite like [Source 1], [Source 2] referencing the source blocks.

Question:
%s

Sources:
%s

Answer with citations:
`, question, src.String()))

	out, err := s.LLM.Generate(ctx, prompt)

	if !common.IsNilValue(err) {
		return Answer{}, err
	}

	return Answer{
		Answer:    strings.TrimSpace(out),
		Citations: citations,
		LatencyMS: time.Since(start).Milliseconds(),
	}, nil
}
