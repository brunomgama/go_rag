package docs

import (
	"strings"

	"github.com/brunomgama/go_rag/internal/common"
)

type Chunk struct {
	DocID   string
	Page    int
	Index   int
	Text    string
	ChunkID string
}

func ChunkByWord(doc Document, target, overlap int) []Chunk {
	var out []Chunk

	if doc.MIME == "application/pdf" && len(doc.PageText) > 0 {
		for i, page := range doc.PageText {
			out = append(out, chunkOne(doc.ID, page, i+1, target, overlap)...)
		}
	} else {
		out = append(out, chunkOne(doc.ID, doc.Content, 0, target, overlap)...)
	}
	return out
}

func chunkOne(docID, text string, page, target, overlap int) []Chunk {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return nil
	}

	var chunks []Chunk
	step := max(1, target-overlap)
	for start := 0; start < len(words); start += step {
		end := min(len(words), start+target)
		segments := strings.Join(words[start:end], " ")
		index := len(chunks)
		chunk_id := ""

		if page > 0 {
			chunk_id = "page-" + common.Itoa(page) + "#" + common.Itoa(index)
		} else {
			chunk_id = "doc#" + common.Itoa(index)
		}

		chunks = append(chunks, Chunk{
			DocID:   docID,
			Page:    page,
			Index:   index,
			Text:    segments,
			ChunkID: chunk_id,
		})
		if end == len(words) {
			break
		}
	}
	return chunks
}
