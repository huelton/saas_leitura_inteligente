package service

import (
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractTextFromPDF extracts plain text from a PDF file path.
func ExtractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var b strings.Builder
	total := r.NumPage()
	for pageIndex := 1; pageIndex <= total; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		content, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		b.WriteString(content)
	}
	return b.String(), nil
}

// FileExists returns true if path exists and is a regular file.
func FileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
