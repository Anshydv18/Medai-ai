package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	config "report/base"
	"report/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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


	summary, err := utils.SummarizeWithGemini(config.LoadConfig().APIKey, text, "english")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server is down"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

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

func HandlePrediction(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	uploadedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}
	defer uploadedFile.Close()

	imageData, err := io.ReadAll(uploadedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}

	prediction, err := callGeminiAPI(imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server is down"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prediction": prediction})
}

func callGeminiAPI(imageData []byte) (string, error) {
	apiKey := config.LoadConfig().APIKey
	if apiKey == "" {
		return "", fmt.Errorf("server went down")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	img := genai.ImageData("jpeg", imageData)

	prompt := `You are a medical AI assistant. Analyze this image and:
    1. List top 3 possible conditions
    2. For each, provide:
       - Confidence percentage (XX%)
       - Key visual findings
       - Urgency level (Low/Medium/High)
    3. Always add: "Consult a healthcare professional for accurate diagnosis."`

	resp, err := model.GenerateContent(ctx, img, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Extract and format response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No diagnosis could be generated", nil
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil

}
