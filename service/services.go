package service

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"report/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type DocumentRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func GenerateReportSummary(c *gin.Context) {
	req := &DocumentRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file upload"})
		return
	}

	filePath := filepath.Join("/temp", req.File.Filename)
	if err := c.SaveUploadedFile(req.File, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	text, err := extractTextFromFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text"})
		return
	}

	fmt.Println(text)

	summary, err := utils.SummarizeWithGemini("AIzaSyDcUWblqA1BeMdtJgYvjtpSXggG7D9PaNM", text, "english")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

// Helper: Extract text from PDF/DOCX/TXT
func extractTextFromFile(filePath string) (string, error) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".pdf":
		return utils.ExtractTextFromPDF(filePath)
	case ".docx":
		return utils.ExtractTextFromDOCX(filePath)
	case ".txt":
		return utils.ExtractTextFromTXT(filePath)
	default:
		return "", fmt.Errorf("unsupported file type")
	}
}
