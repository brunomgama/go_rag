package config

import (
	"os"
	"strconv"

	"github.com/brunomgama/go_rag/internal/common"
	"github.com/joho/godotenv"
)

type Config struct {
	EmbeddingsModel  string
	OllamaHost       string
	QdrantURL        string
	QdrantCollection string
	ChunkTarget      int
	ChunkOverlap     int
	llm_Model        string
	llm_Port         int
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		EmbeddingsModel:  envDefault("EMBEDDINGS_MODEL", "nomic-embed-text"),
		OllamaHost:       envDefault("OLLAMA_HOST", "http://localhost:11434"),
		QdrantURL:        envDefault("QDRANT_URL", "http://localhost:6333"),
		QdrantCollection: envDefault("QDRANT_COLLECTION", "docs"),
		ChunkTarget:      mustInt(os.Getenv("CHUNK_TOKEN_TARGET"), 800),
		ChunkOverlap:     mustInt(os.Getenv("CHUNK_OVERLAP"), 120),
		llm_Model:        envDefault("LLM_MODEL", "llama3.1:8b"),
		llm_Port:         mustInt(os.Getenv("LLM_PORT"), 8080),
	}
}

func envDefault(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}

func mustInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if !common.IsNilValue(err) {
		return def
	}
	return v
}
