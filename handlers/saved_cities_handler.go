package handlers

import (
	"database/sql" // Add this import
	"net/http"
	"weather-app/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SavedCitiesHandler handles saved cities functionality
type SavedCitiesHandler struct {
	db *sql.DB
}

// NewSavedCitiesHandler creates a new saved cities handler
func NewSavedCitiesHandler(db *sql.DB) *SavedCitiesHandler {
	return &SavedCitiesHandler{
		db: db,
	}
}

// RegisterRoutes registers saved cities routes
func (h *SavedCitiesHandler) RegisterRoutes(router *gin.Engine) {
	// API routes with authentication middleware
	citiesGroup := router.Group("/api/cities")
	citiesGroup.Use(middleware.AuthRequired())
	{
		citiesGroup.GET("/saved", h.getSavedCities)
		citiesGroup.POST("/save", h.saveCity)
		citiesGroup.DELETE("/remove", h.removeCity)
	}
}

// getSavedCities handles GET requests to retrieve a user's saved cities
func (h *SavedCitiesHandler) getSavedCities(c *gin.Context) {
	// Get user ID from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Query database for saved cities
	rows, err := h.db.Query(
		"SELECT city_name FROM saved_cities WHERE user_id = ?",
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get saved cities: " + err.Error()})
		return
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process saved cities: " + err.Error()})
			return
		}
		cities = append(cities, city)
	}

	// Return the saved cities
	c.JSON(http.StatusOK, gin.H{"cities": cities})
}

// saveCity handles POST requests to save a city for a user
func (h *SavedCitiesHandler) saveCity(c *gin.Context) {
	// Parse request body
	var req struct {
		CityName string `json:"city_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Insert into database, ignoring if already exists
	_, err := h.db.Exec(
		"INSERT IGNORE INTO saved_cities (user_id, city_name) VALUES (?, ?)",
		userID, req.CityName,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save city: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "City saved successfully"})
}

// removeCity handles DELETE requests to remove a saved city
func (h *SavedCitiesHandler) removeCity(c *gin.Context) {
	// Parse request body
	var req struct {
		CityName string `json:"city_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Delete from database
	_, err := h.db.Exec(
		"DELETE FROM saved_cities WHERE user_id = ? AND city_name = ?",
		userID, req.CityName,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove city: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "City removed successfully"})
}
