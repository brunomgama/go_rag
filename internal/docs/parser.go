package docs

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/brunomgama/go_rag/internal/common"
	pdf "github.com/ledongthuc/pdf"
)

type Document struct {
	ID       string
	Path     string
	MIME     string
	Content  string
	PageText []string
}

func ParseFile(path string) (Document, error) {
	extension := strings.ToLower(filepath.Ext(path))

	switch extension {
	case ".pdf":
		return ParsePDF(path)
	default:
		b, err := os.ReadFile(path)

		if !common.IsNilValue(err) {
			log.Println("Error reading plain text document..")
			return Document{}, err
		}

		return Document{
			ID:       filepath.Base(path),
			Path:     path,
			MIME:     "text/plain",
			Content:  string(b),
			PageText: []string{string(b)},
		}, nil
	}
}

func ParsePDF(path string) (Document, error) {
	file, reader, err := pdf.Open(path)

	if !common.IsNilValue(err) {
		log.Println("Error reading document based text document..")
		return Document{}, err
	}
	defer file.Close()

	numbPages := reader.NumPage()
	pages := make([]string, 0, numbPages)
	var full bytes.Buffer

	for i := 1; i <= numbPages; i++ {
		page := reader.Page(i)

		if page.V.IsNull() {
			continue
		}

		content, _ := page.GetPlainText(nil)
		s := strings.TrimSpace(content)

		pages = append(pages, s)
		full.WriteString(s + "\n")
	}

	return Document{
		ID:       filepath.Base(path),
		Path:     path,
		MIME:     "application/pdf",
		Content:  full.String(),
		PageText: pages,
	}, nil
}
