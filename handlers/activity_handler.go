package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"weather-app/models"
	"weather-app/services"
)

// ActivityHandler handles requests for activity and health recommendations
type ActivityHandler struct {
	activityService *services.ActivityService
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(apiKey string) *ActivityHandler {
	return &ActivityHandler{
		activityService: services.NewActivityService(apiKey),
	}
}

// GetActivitiesHandler handles API requests for activity recommendations
func (h *ActivityHandler) GetActivitiesHandler(c *gin.Context) {
	// Get lat and lon from query parameters
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	// Get city name if provided, otherwise use default
	city := c.DefaultQuery("city", "")

	// Get current weather data first
	var currentWeather *services.WeatherData
	var err error

	if city != "" {
		currentWeather, err = services.GetCurrentWeather(city)
	} else if latStr != "" && lonStr != "" {
		// Parse lat/lon
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid latitude format",
			})
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid longitude format",
			})
			return
		}

		// Use coordinates instead if city isn't provided
		currentWeather, err = services.GetCurrentWeatherByCoords(lat, lon)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either city name or latitude/longitude are required",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch weather data: " + err.Error(),
		})
		return
	}

	// Get activity and health recommendations
	activities, err := h.activityService.GetActivitiesAndHealth(currentWeather,
		currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		log.Printf("Warning: Error getting activity recommendations: %v", err)
		// Continue with partial data if possible
	}

	// Return the activity recommendations
	c.JSON(http.StatusOK, activities)
}

// GetActivitiesPageHandler handles web requests for the activities page
func (h *ActivityHandler) GetActivitiesPageHandler(c *gin.Context) {
	// Get city from query parameter
	city := c.DefaultQuery("city", "")

	// If no city provided but user is logged in, use their home city
	if city == "" {
		userID, exists := c.Get("user_id")
		if exists {
			userStore := c.MustGet("user_store").(models.UserStore)
			user, err := userStore.GetUserByID(userID.(int))
			if err == nil && user.HomeCity != "" {
				city = user.HomeCity
			}
		}
	}

	// Get weather data
	currentWeather, err := services.GetCurrentWeather(city)
	if err != nil {
		c.HTML(http.StatusOK, "weather-impact.html", gin.H{
			"error": "Failed to fetch weather data: " + err.Error(),
		})
		return
	}

	// Get activity recommendations
	activities, err := h.activityService.GetActivitiesAndHealth(currentWeather,
		currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		log.Printf("Error getting activity recommendations: %v", err)
	}

	// Render template with data
	c.HTML(http.StatusOK, "weather-impact.html", gin.H{
		"title":          "Weather Activities for " + currentWeather.Name,
		"city":           currentWeather.Name,
		"currentWeather": currentWeather,
		"activities":     activities,
		"currentYear":    time.Now().Year(),
	})
}

// GetActivitiesAndHealthForCity is a helper method to get recommendations for a city
func (h *ActivityHandler) GetActivitiesAndHealthForCity(city string) (*services.WeatherActivities, error) {
	// Get weather data
	weatherData, err := services.GetCurrentWeather(city)
	if err != nil {
		return nil, err
	}

	// Get activity recommendations
	return h.activityService.GetActivitiesAndHealth(weatherData, weatherData.Coord.Lat, weatherData.Coord.Lon)
}

// RedirectToWeatherImpact redirects users from activities.html to weather-impact.html
func (h *ActivityHandler) RedirectToWeatherImpact(c *gin.Context) {
	// Preserve any query parameters when redirecting
	query := c.Request.URL.RawQuery
	redirectURL := "/weather-impact"

	if query != "" {
		redirectURL += "?" + query
	}

	c.Redirect(http.StatusMovedPermanently, redirectURL)
}

// DirectActivitiesLink handles direct linking to the activities page
func (h *ActivityHandler) DirectActivitiesLink(c *gin.Context) {
	// This handler can be used for custom paths like /my-activities
	// that should still show the weather impact page
	h.GetActivitiesPageHandler(c)
}
