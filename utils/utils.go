package utils

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf" // Free PDF extractor
	"github.com/nguyenthenguyen/docx"
)

// Extract text from PDF (using free library)
func ExtractTextFromPDF(filePath string) (string, error) {
    file, reader, err := pdf.Open(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to open PDF: %v", err)
    }
    defer file.Close()

    var textBuilder strings.Builder
    for i := 1; i <= reader.NumPage(); i++ {
        page := reader.Page(i)
        if page.V.IsNull() {
            continue
        }
        
        // Pass nil for fonts
        text, err := page.GetPlainText(nil)
        if err != nil {
            return "", fmt.Errorf("failed to extract page %d: %v", i, err)
        }
        textBuilder.WriteString(text)
    }
    return textBuilder.String(), nil
}

// Extract text from DOCX (unchanged)
func ExtractTextFromDOCX(filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	return r.Editable().GetContent(), nil
}

// Extract text from TXT (unchanged)
func ExtractTextFromTXT(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
