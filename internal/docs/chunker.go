package docs

import (
	"strconv"
	"strings"
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
			chunk_id = "page-" + itoa(page) + "#" + itoa(index)
		} else {
			chunk_id = "doc#" + itoa(index)
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

func itoa(i int) string {
	return strconv.Itoa(i)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
