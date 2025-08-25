package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/brunomgama/go_rag/internal/common"
	"github.com/brunomgama/go_rag/internal/config"
	"github.com/brunomgama/go_rag/internal/embed"
	"github.com/brunomgama/go_rag/internal/llm"
	"github.com/brunomgama/go_rag/internal/rag"
	"github.com/brunomgama/go_rag/internal/store"
)

type queryRequest struct {
	Query string `json:query"`
	TopK  int    `json:top_k"`
}

type queryResponse = rag.Answer

func main() {
	cfg := config.Load()

	emb := embed.NewOllama(cfg.OllamaHost, cfg.EmbeddingsModel)
	llmClient := llm.NewOllama(cfg.OllamaHost, envDefault("LLM_MODEL", "llama3.1:8b"))
	st := store.NewQdrant(cfg.QdrantURL, cfg.QdrantCollection)

	minScroe := float32(0.15)
	svc := &rag.Service{
		Embed:    emb,
		LLM:      llmClient,
		Store:    st,
		TopK:     6,
		MinScore: &minScroe,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /query", func(w http.ResponseWriter, r *http.Request) {
		var req queryRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if !common.IsNilValue(err) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()

		ans, err := svc.Query(ctx, req.Query, req.TopK)
		if !common.IsNilValue(err) {
			log.Println("query error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ans)
	})

	port := envDefault("PORT", "8080")
	log.Printf("API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(mux)))
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func envDefault(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}
