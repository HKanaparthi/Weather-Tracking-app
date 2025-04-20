package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"weather-app/middleware"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadDir string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir string) *UploadHandler {
	// Create upload directory if it doesn't exist
	os.MkdirAll(filepath.Join(uploadDir, "chat_images"), 0755)

	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

// RegisterRoutes registers upload-related routes
func (h *UploadHandler) RegisterRoutes(router *gin.Engine) {
	uploadGroup := router.Group("/api/upload")
	uploadGroup.Use(middleware.AuthRequired())
	{
		uploadGroup.POST("/image", h.handleImageUpload)
	}
}

// handleImageUpload processes image uploads
func (h *UploadHandler) handleImageUpload(c *gin.Context) {
	// Get the file from the request
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file: " + err.Error()})
		return
	}
	defer file.Close()

	// Check if the file is an image
	if !isImageFile(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files are allowed"})
		return
	}

	// Check file size (limit to 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size should be less than 5MB"})
		return
	}

	// Generate a unique filename
	filename := generateUniqueFilename(header.Filename)

	// Get upload purpose (chat, profile, etc.)
	purpose := c.PostForm("purpose")
	if purpose == "" {
		purpose = "misc"
	}

	// Determine the directory based on purpose
	var uploadPath string
	switch purpose {
	case "chat":
		uploadPath = filepath.Join(h.uploadDir, "chat_images", filename)
	case "profile":
		uploadPath = filepath.Join(h.uploadDir, "profile_photos", filename)
	default:
		uploadPath = filepath.Join(h.uploadDir, "misc", filename)
	}

	// Create a new file
	dst, err := os.Create(uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file: " + err.Error()})
		return
	}
	defer dst.Close()

	// Copy the file
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	// Determine the URL path
	var urlPath string
	switch purpose {
	case "chat":
		urlPath = "/static/chat_images/" + filename
	case "profile":
		urlPath = "/static/profile_photos/" + filename
	default:
		urlPath = "/static/misc/" + filename
	}

	// Return the file URL
	c.JSON(http.StatusOK, gin.H{"url": urlPath})
}

// isImageFile checks if a filename has an image extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// generateUniqueFilename creates a unique filename
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + ext
}
