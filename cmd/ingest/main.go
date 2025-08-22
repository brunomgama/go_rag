package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/brunomgama/go_rag/internal/config"
	"github.com/brunomgama/go_rag/internal/docs"
	"github.com/brunomgama/go_rag/internal/embed"
	"github.com/brunomgama/go_rag/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	start := time.Now()
	ctx := context.Background()

	// COLECT FILES
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

	//CLIENTS
	emb := embed.NewOllama(cfg.OllamaHost, cfg.EmbeddingsModel)
	qd := store.NewQdrant(cfg.QdrantURL, cfg.QdrantCollection)

	totalChunks := 0
	totalVectors := 0

	var dim int

	for _, p := range paths {
		doc, err := docs.ParseFile(p)
		if err != nil {
			log.Printf("parse error %s: %v", p, err)
			continue
		}

		chunks := docs.ChunkByWord(doc, cfg.ChunkTarget, cfg.ChunkOverlap)
		if len(chunks) == 0 {
			continue
		}

		batch := 64
		var points []store.Point

		for i := 0; i < len(chunks); i += batch {
			j := i + batch
			if j > len(chunks) {
				j = len(chunks)
			}

			// allocate exactly for this batch
			texts := make([]string, j-i)
			for k := i; k < j; k++ {
				texts[k-i] = chunks[k].Text
			}

			vecs, err := emb.Embed(ctx, texts)
			if err != nil {
				log.Fatalf("embed error: %v", err) // use Fatalf so %v is formatted
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

			totalVectors += len(vecs)
		}

		if err := qd.Upsert(ctx, points); err != nil {
			log.Fatalf("upsert error: %v", err)
		}

		totalChunks += len(chunks)
		log.Printf("Upserted %d chunks for %s", len(chunks), doc.ID)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Ingest complete\nDocs: %d\nChunks: %d\nVectors: %d\nTime: %s\n",
		len(paths), totalChunks, totalVectors, elapsed)
}
