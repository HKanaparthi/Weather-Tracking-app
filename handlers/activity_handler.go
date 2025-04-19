package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"weather-app/models"
)

// ActivityHandler handles requests for activity-related features
type ActivityHandler struct {
	apiKey string
}

// NewActivityHandler creates a new instance of ActivityHandler
func NewActivityHandler(apiKey string) *ActivityHandler {
	return &ActivityHandler{
		apiKey: apiKey,
	}
}

// GetActivitiesPageHandler handles rendering the activities page
func (h *ActivityHandler) GetActivitiesPageHandler(c *gin.Context) {
	// Get user from context if available
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		userStore := c.MustGet("user_store").(models.UserStore)
		var err error
		user, err = userStore.GetUserByID(userID.(int))
		if err != nil {
			log.Printf("Error getting user: %v", err)
		}
	}

	// Render the activities template
	c.HTML(http.StatusOK, "activities.html", gin.H{
		"title":     "Weather Activities",
		"User":      user,
		"isPremium": true,
	})
}

// GetActivitiesHandler handles API requests for activity recommendations
func (h *ActivityHandler) GetActivitiesHandler(c *gin.Context) {
	// Get weather parameters from query
	city := c.Query("city")
	tempStr := c.DefaultQuery("temp", "20")
	condition := c.DefaultQuery("condition", "clear")
	humidity := c.DefaultQuery("humidity", "50")
	windSpeed := c.DefaultQuery("wind", "5")

	log.Printf("Activity recommendation request for city: %s, condition: %s", city, condition)

	// Generate activity recommendations based on weather conditions
	// In a real implementation, you would have more sophisticated logic
	activities := generateActivityRecommendations(condition, tempStr, humidity, windSpeed)

	c.JSON(http.StatusOK, gin.H{
		"city":       city,
		"activities": activities,
		"success":    true,
	})
}

// generateActivityRecommendations returns activities appropriate for the given weather
func generateActivityRecommendations(condition, tempStr, humidity, windSpeed string) []map[string]interface{} {
	condition = strings.ToLower(condition)

	// Default activities for various weather conditions
	outdoorActivities := []map[string]interface{}{
		{
			"name":        "Walking",
			"description": "Perfect conditions for a walk in the park",
			"suitability": "Excellent",
			"icon":        "fas fa-walking",
		},
		{
			"name":        "Cycling",
			"description": "Good weather for cycling",
			"suitability": "Good",
			"icon":        "fas fa-bicycle",
		},
		{
			"name":        "Picnic",
			"description": "Consider a picnic in this weather",
			"suitability": "Good",
			"icon":        "fas fa-utensils",
		},
	}

	indoorActivities := []map[string]interface{}{
		{
			"name":        "Visit Museum",
			"description": "A great day to enjoy indoor cultural activities",
			"suitability": "Excellent",
			"icon":        "fas fa-landmark",
		},
		{
			"name":        "Shopping",
			"description": "Good time for shopping",
			"suitability": "Good",
			"icon":        "fas fa-shopping-bag",
		},
		{
			"name":        "Watch a Movie",
			"description": "Enjoy a movie at home or theater",
			"suitability": "Excellent",
			"icon":        "fas fa-film",
		},
	}

	// Return appropriate activities based on weather condition
	if strings.Contains(condition, "rain") ||
		strings.Contains(condition, "snow") ||
		strings.Contains(condition, "storm") ||
		strings.Contains(condition, "thunder") {
		return indoorActivities
	}

	return outdoorActivities
}

// RegisterActivityRoutes registers activity-related routes
func (h *ActivityHandler) RegisterActivityRoutes(router *gin.Engine) {
	router.GET("/activities", h.GetActivitiesPageHandler)
	router.GET("/weather-impact", h.GetActivitiesPageHandler)
	router.GET("/api/activities", h.GetActivitiesHandler)
}
