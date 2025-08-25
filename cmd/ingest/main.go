package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brunomgama/go_rag/internal/config"
	"github.com/brunomgama/go_rag/internal/docs"
	"github.com/brunomgama/go_rag/internal/embed"
	"github.com/brunomgama/go_rag/internal/store"
	"github.com/joho/godotenv"
)

type metrics struct {
	docs         int
	chunks       int
	vectors      int
	parseChunkMs time.Duration
	embedMs      time.Duration
	upsertMs     time.Duration
	approxTokens int
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	start := time.Now()
	ctx := context.Background()
	m := metrics{}

	// track duplicates within a run
	seen := make(map[string]bool)

	// collect files
	var paths []string
	root := "data"
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if len(paths) == 0 {
		log.Printf("No files found in %s", root)
		return
	}

	// clients
	emb := embed.NewOllama(cfg.OllamaHost, cfg.EmbeddingsModel)
	qd := store.NewQdrant(cfg.QdrantURL, cfg.QdrantCollection)

	var dim int

	for _, p := range paths {
		// dedup by checksum
		sum, err := fileSHA256(p)
		if err == nil {
			if seen[sum] {
				log.Printf("Skipping duplicate content: %s", p)
				continue
			}
			seen[sum] = true
		}

		// parse + chunk
		t0 := time.Now()
		doc, err := docs.ParseFile(p)
		if err != nil {
			log.Printf("parse error %s: %v", p, err)
			continue
		}
		chunks := docs.ChunkByWord(doc, cfg.ChunkTarget, cfg.ChunkOverlap)
		m.parseChunkMs += time.Since(t0)
		m.docs++
		m.chunks += len(chunks)
		for _, c := range chunks {
			m.approxTokens += countWords(c.Text)
		}
		if len(chunks) == 0 {
			continue
		}

		// embed + upsert
		batch := 64
		var points []store.Point
		vectorsThisDoc := 0

		for i := 0; i < len(chunks); i += batch {
			j := i + batch
			if j > len(chunks) {
				j = len(chunks)
			}

			texts := make([]string, j-i)
			for k := i; k < j; k++ {
				texts[k-i] = chunks[k].Text
			}

			t1 := time.Now()
			vecs, err := emb.Embed(ctx, texts)
			m.embedMs += time.Since(t1)
			if err != nil {
				log.Fatalf("embed error: %v", err)
			}

			if dim == 0 && len(vecs) > 0 {
				dim = len(vecs[0])
				if err := qd.EnsureCollection(ctx, dim); err != nil {
					log.Fatalf("ensure collection: %v", err)
				}
			}

			for k := range vecs {
				c := chunks[i+k]
				points = append(points, store.Point{
					ID:     fmt.Sprintf("%s::%s", c.DocID, c.ChunkID),
					Vector: vecs[k],
					Payload: map[string]any{
						"doc_id":   c.DocID,
						"page":     c.Page,
						"chunk_id": c.ChunkID,
						"text":     c.Text,
					},
				})
			}
			vectorsThisDoc += len(vecs)
			m.vectors += len(vecs)
		}

		if len(points) > 0 {
			t2 := time.Now()
			if err := qd.Upsert(ctx, points); err != nil {
				log.Fatalf("upsert error: %v", err)
			}
			m.upsertMs += time.Since(t2)
			log.Printf("Upserted %d vectors for %s", vectorsThisDoc, doc.ID)
		}
	}

	// summary
	elapsed := time.Since(start)
	chunksPerSec := float64(m.chunks) / elapsed.Seconds()
	vectorsPerSec := float64(m.vectors) / elapsed.Seconds()

	fmt.Printf(`
	✅ Ingest complete
		Docs:    %d
		Chunks:  %d
		Vectors: %d
		~Tokens: %d

	⏰ Timing:
		Total:        %s
		Parse+Chunk:  %s
		Embed:        %s
		Upsert:       %s

	🔄 Throughput:
		Chunks/sec:   %.2f
		Vectors/sec:  %.2f
`[1:],m.docs, m.chunks, m.vectors, m.approxTokens,elapsed, m.parseChunkMs, m.embedMs, m.upsertMs,chunksPerSec, vectorsPerSec,)

}

func fileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func countWords(s string) int { return len(strings.Fields(s)) }
