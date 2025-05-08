package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"report/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
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

func handlePrediction(c *gin.Context) {
	// Get uploaded image
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	// Open the file
	uploadedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}
	defer uploadedFile.Close()

	// Read file data
	imageData, err := io.ReadAll(uploadedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}

	// Call Gemini API
	prediction, err := callGeminiAPI(imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI prediction failed"})
		return
	}

	// Return prediction
	c.JSON(http.StatusOK, gin.H{"prediction": prediction})
}

func callGeminiAPI(imageData []byte) (string, error) {
	// Replace with your Gemini API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Gemini API key not set")
	}

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Initialize model (Gemini Pro Vision)
	model := client.GenerativeModel("gemini-pro-vision")

	// Create image part
	img := genai.ImageData("jpeg", imageData)

	// Define prompt
	prompt := "Analyze this medical image and suggest possible diseases. Be concise."

	// Generate response
	resp, err := model.GenerateContent(ctx, img, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	// Extract prediction
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "No prediction generated", nil
}
