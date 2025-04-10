package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv"

	"weather-app/handlers"
	"weather-app/middleware"
	"weather-app/models"
	"weather-app/services"
	// Global notification service
)

const apiKey = "0c2e2084bdd01a671b1b450215191f89" // Your Premium OpenWeather API key

var notificationService *services.NotificationService

// WeatherData struct holds the structure of the current weather response
type WeatherData struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"` // Added for premium
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"` // Added for premium
		Id          int    `json:"id"`   // Added for premium
		Main        string `json:"main"` // Added for premium
	} `json:"weather"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Wind struct {
		Speed float64 `json:"speed"` // Wind speed
		Deg   float64 `json:"deg"`   // Wind direction in degrees
		Gust  float64 `json:"gust"`  // Wind gust speed
	} `json:"wind"`
	Rain struct {
		OneHour   float64 `json:"1h,omitempty"` // Rain volume for last hour
		ThreeHour float64 `json:"3h,omitempty"` // Rain volume for last 3 hours
	} `json:"rain"`
	Snow struct {
		OneHour   float64 `json:"1h,omitempty"` // Snow volume for last hour
		ThreeHour float64 `json:"3h,omitempty"` // Snow volume for last 3 hours
	} `json:"snow"`
	Clouds struct {
		All int `json:"all"` // Cloudiness percentage
	} `json:"clouds"`
	Sys struct {
		Sunrise int64  `json:"sunrise"` // Sunrise time, unix, UTC
		Sunset  int64  `json:"sunset"`  // Sunset time, unix, UTC
		Country string `json:"country"` // Country code
	} `json:"sys"`
	Timezone   int   `json:"timezone"`   // Shift in seconds from UTC
	Dt         int64 `json:"dt"`         // Time of data calculation, unix, UTC
	Visibility int   `json:"visibility"` // Visibility in meters
}

// UVData struct for UV index response
type UVData struct {
	Current struct {
		UVI float64 `json:"uvi"`
	} `json:"current"`
	Daily []struct {
		Dt  int64   `json:"dt"`
		UVI float64 `json:"uvi"`
	} `json:"daily"`
}

// AirQualityData struct for air quality response
type AirQualityData struct {
	List []struct {
		Main struct {
			Aqi int `json:"aqi"` // Air Quality Index
		} `json:"main"`
		Components struct {
			Co    float64 `json:"co"`    // Carbon monoxide μg/m3
			No    float64 `json:"no"`    // Nitrogen monoxide μg/m3
			No2   float64 `json:"no2"`   // Nitrogen dioxide μg/m3
			O3    float64 `json:"o3"`    // Ozone μg/m3
			So2   float64 `json:"so2"`   // Sulphur dioxide μg/m3
			Pm2_5 float64 `json:"pm2_5"` // Fine particles μg/m3
			Pm10  float64 `json:"pm10"`  // Coarse particles μg/m3
			Nh3   float64 `json:"nh3"`   // Ammonia μg/m3
		} `json:"components"`
		Dt int64 `json:"dt"` // Data timestamp, unix, UTC
	} `json:"list"`
}

// WeatherAlert struct for weather alerts
type WeatherAlert struct {
	SenderName  string   `json:"sender_name"`
	Event       string   `json:"event"`
	Start       int64    `json:"start"`
	End         int64    `json:"end"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// OneCallData struct for OneCall API response (premium feature)
type OneCallData struct {
	Lat            float64        `json:"lat"`
	Lon            float64        `json:"lon"`
	Timezone       string         `json:"timezone"`
	TimezoneOffset int            `json:"timezone_offset"`
	Current        CurrentOneCall `json:"current"`
	Minutely       []MinutelyData `json:"minutely,omitempty"`
	Hourly         []HourlyData   `json:"hourly"`
	Daily          []DailyOneCall `json:"daily"`
	Alerts         []WeatherAlert `json:"alerts,omitempty"`
}

// CurrentOneCall struct for current weather in OneCall API
type CurrentOneCall struct {
	Dt         int64   `json:"dt"`
	Sunrise    int64   `json:"sunrise"`
	Sunset     int64   `json:"sunset"`
	Temp       float64 `json:"temp"`
	FeelsLike  float64 `json:"feels_like"`
	Pressure   int     `json:"pressure"`
	Humidity   int     `json:"humidity"`
	DewPoint   float64 `json:"dew_point"`
	Uvi        float64 `json:"uvi"`
	Clouds     int     `json:"clouds"`
	Visibility int     `json:"visibility"`
	WindSpeed  float64 `json:"wind_speed"`
	WindDeg    int     `json:"wind_deg"`
	WindGust   float64 `json:"wind_gust,omitempty"`
	Weather    []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Rain struct {
		OneHour float64 `json:"1h,omitempty"`
	} `json:"rain,omitempty"`
	Snow struct {
		OneHour float64 `json:"1h,omitempty"`
	} `json:"snow,omitempty"`
}

// MinutelyData struct for minutely forecast in OneCall API
type MinutelyData struct {
	Dt            int64   `json:"dt"`
	Precipitation float64 `json:"precipitation"`
}

// HourlyData struct for hourly forecast in OneCall API
type HourlyData struct {
	Dt         int64   `json:"dt"`
	Temp       float64 `json:"temp"`
	FeelsLike  float64 `json:"feels_like"`
	Pressure   int     `json:"pressure"`
	Humidity   int     `json:"humidity"`
	DewPoint   float64 `json:"dew_point"`
	Uvi        float64 `json:"uvi"`
	Clouds     int     `json:"clouds"`
	Visibility int     `json:"visibility"`
	WindSpeed  float64 `json:"wind_speed"`
	WindDeg    int     `json:"wind_deg"`
	WindGust   float64 `json:"wind_gust,omitempty"`
	Weather    []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Pop  float64 `json:"pop"` // Probability of precipitation
	Rain struct {
		OneHour float64 `json:"1h,omitempty"`
	} `json:"rain,omitempty"`
	Snow struct {
		OneHour float64 `json:"1h,omitempty"`
	} `json:"snow,omitempty"`
}

// DailyOneCall struct for daily forecast in OneCall API
type DailyOneCall struct {
	Dt        int64   `json:"dt"`
	Sunrise   int64   `json:"sunrise"`
	Sunset    int64   `json:"sunset"`
	Moonrise  int64   `json:"moonrise"`
	Moonset   int64   `json:"moonset"`
	MoonPhase float64 `json:"moon_phase"`
	Temp      struct {
		Day   float64 `json:"day"`
		Min   float64 `json:"min"`
		Max   float64 `json:"max"`
		Night float64 `json:"night"`
		Eve   float64 `json:"eve"`
		Morn  float64 `json:"morn"`
	} `json:"temp"`
	FeelsLike struct {
		Day   float64 `json:"day"`
		Night float64 `json:"night"`
		Eve   float64 `json:"eve"`
		Morn  float64 `json:"morn"`
	} `json:"feels_like"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
	DewPoint  float64 `json:"dew_point"`
	WindSpeed float64 `json:"wind_speed"`
	WindDeg   int     `json:"wind_deg"`
	WindGust  float64 `json:"wind_gust,omitempty"`
	Weather   []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Clouds int     `json:"clouds"`
	Pop    float64 `json:"pop"`
	Rain   float64 `json:"rain,omitempty"`
	Snow   float64 `json:"snow,omitempty"`
	Uvi    float64 `json:"uvi"`
}

// DailyForecast struct for a single day forecast
type DailyForecast struct {
	Index           int     `json:"index"` // Day index (1-16)
	Date            string  `json:"date"`  // Display date (e.g., "Mar 10")
	FullDate        string  `json:"-"`     // Full date for sorting (e.g., "2025-03-10")
	Day             string  `json:"day"`   // Day of week (e.g., "Mon")
	MaxTemp         float64 `json:"maxTemp"`
	MinTemp         float64 `json:"minTemp"`
	Humidity        int     `json:"humidity"`
	Pressure        int     `json:"pressure"`
	UVIndex         float64 `json:"uvIndex"`
	Description     string  `json:"description"`
	UVCategory      string  `json:"-"`               // UV Index category (Low, Moderate, etc.)
	UVColor         string  `json:"-"`               // Color for UV Index display
	WindSpeed       float64 `json:"windSpeed"`       // Added for premium
	WindDirection   string  `json:"windDirection"`   // Added for premium
	PrecipProb      float64 `json:"precipProb"`      // Added for premium
	Sunrise         string  `json:"sunrise"`         // Added for premium
	Sunset          string  `json:"sunset"`          // Added for premium
	MoonPhase       string  `json:"moonPhase"`       // Added for premium
	AirQualityIndex int     `json:"airQualityIndex"` // Added for premium
}

// CurrentWeather struct for current weather data
type CurrentWeather struct {
	Temperature     string  `json:"temperature"`
	FeelsLike       string  `json:"feelsLike"` // Added for premium
	Humidity        string  `json:"humidity"`
	Pressure        string  `json:"pressure"`
	UVIndex         float64 `json:"uvIndex"`
	Condition       string  `json:"condition"`
	UVCategory      string  `json:"-"`               // UV Index category
	UVColor         string  `json:"-"`               // Color for UV Index display
	WindSpeed       float64 `json:"windSpeed"`       // Added for premium
	WindDirection   string  `json:"windDirection"`   // Added for premium
	WindGust        float64 `json:"windGust"`        // Added for premium
	Visibility      int     `json:"visibility"`      // Added for premium
	Sunrise         string  `json:"sunrise"`         // Added for premium
	Sunset          string  `json:"sunset"`          // Added for premium
	AirQualityIndex int     `json:"airQualityIndex"` // Added for premium
	AQICategory     string  `json:"aqiCategory"`     // Added for premium
}

// HourlyForecastItem struct for hourly forecast
type HourlyForecastItem struct {
	Time       string  `json:"time"`
	Temp       float64 `json:"temp"`
	Icon       string  `json:"icon"`
	PrecipProb float64 `json:"precipProb"` // Added for premium
}

// WeatherAlert struct for displaying alerts
type WeatherAlertItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        string `json:"time"`
	Severity    string `json:"severity"`
}

// CityWeather struct for holding weather data for a single city
type CityWeather struct {
	Name          string          `json:"name"`
	Current       CurrentWeather  `json:"current"`
	Forecast      []DailyForecast `json:"forecast"`
	BackgroundURL string          `json:"-"` // Background image URL based on weather condition
}

// WeatherResponse struct for the combined response
type WeatherResponse struct {
	City           string               `json:"city"`
	Current        CurrentWeather       `json:"current"`
	Forecast       []DailyForecast      `json:"forecast"`
	HourlyForecast []HourlyForecastItem `json:"hourlyForecast"` // Added for premium
	Alerts         []WeatherAlertItem   `json:"alerts"`         // Added for premium
	BackgroundURL  string               `json:"-"`              // Background image URL based on weather condition
	Error          string               `json:"-"`              // Error message if any
	User           *models.User         `json:"-"`              // User for template rendering
}

// CompareResponse struct for the weather comparison response
type CompareResponse struct {
	Cities []CityWeather     `json:"cities"`
	Error  string            `json:"-"` // Error message if any
	User   *models.User      `json:"-"` // User for template rendering
	Errors map[string]string `json:"-"` // Map of city-specific errors
}

// ForecastData struct for the 5-day forecast response
type ForecastData struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp     float64 `json:"temp"`
			TempMin  float64 `json:"temp_min"`
			TempMax  float64 `json:"temp_max"`
			Humidity int     `json:"humidity"`
			Pressure int     `json:"pressure"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
			Id          int    `json:"id"`
		} `json:"weather"`
		DtTxt string  `json:"dt_txt"`
		Pop   float64 `json:"pop"` // Probability of precipitation - added for premium
		Wind  struct {
			Speed float64 `json:"speed"`
			Deg   float64 `json:"deg"`
			Gust  float64 `json:"gust"`
		} `json:"wind"`
	} `json:"list"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

// HistoricalComparisonResponse represents the response for historical weather comparison
type HistoricalComparisonResponse struct {
	City             string  `json:"city"`
	CurrentTemp      float64 `json:"currentTemp"`
	LastYearTemp     float64 `json:"lastYearTemp"`
	FiveYearAvgTemp  float64 `json:"fiveYearAvgTemp"`
	TempDifference   float64 `json:"tempDifference"`
	TrendDescription string  `json:"trendDescription"`
	SeasonalInsight  string  `json:"seasonalInsight"`
}

// HistoricalAPIRequest represents a request for historical weather data
type HistoricalAPIRequest struct {
	City  string   `json:"city" form:"city" binding:"required"`
	Date  string   `json:"date" form:"date"`
	Years []string `json:"years" form:"years[]"`
}

// HistoricalWeatherData represents the historical weather data for comparison
type HistoricalWeatherData struct {
	Current struct {
		Temperature float64 `json:"temperature"`
		Humidity    int     `json:"humidity"`
		Pressure    int     `json:"pressure"`
		Condition   string  `json:"condition"`
		IconURL     string  `json:"iconURL"`
		WindSpeed   float64 `json:"windSpeed"`
		Icon        string  `json:"icon"` // Added for display
	} `json:"current"`
	Previous map[string]struct {
		Temperature    float64 `json:"temperature"`
		Humidity       int     `json:"humidity"`
		Pressure       int     `json:"pressure"`
		Condition      string  `json:"condition"`
		IconURL        string  `json:"iconURL"`
		WindSpeed      float64 `json:"windSpeed"`
		TempDifference float64 `json:"tempDifference"`
		Icon           string  `json:"icon"` // Added for display
	} `json:"previous"`
}

// HistoricalData represents historical weather data
type HistoricalData struct {
	LastYearTemp    float64
	FiveYearAvgTemp float64
}

// TimeMachineResponse represents the OpenWeather Timemachine API response
type TimeMachineResponse struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Data []struct {
		Dt         int64   `json:"dt"`
		Temp       float64 `json:"temp"`
		Feels_like float64 `json:"feels_like"`
		Pressure   int     `json:"pressure"`
		Humidity   int     `json:"humidity"`
		Dew_point  float64 `json:"dew_point"`
		Clouds     int     `json:"clouds"`
		Visibility int     `json:"visibility"`
		Wind_speed float64 `json:"wind_speed"`
		Wind_deg   int     `json:"wind_deg"`
		Weather    []struct {
			Id          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	} `json:"data"`
}

type WeatherProxyResponse struct {
	City     string                   `json:"city"`
	Current  map[string]interface{}   `json:"current"`
	Forecast []map[string]interface{} `json:"forecast"`
	Hourly   []map[string]interface{} `json:"hourly"`
	Alerts   []map[string]interface{} `json:"alerts"`
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getNearbyLocations gets multiple location options near the given coordinates
func getNearbyLocations(lat, lon float64) ([]struct {
	Name     string  `json:"name"`
	State    string  `json:"state,omitempty"`
	Country  string  `json:"country"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Distance float64 `json:"distance"`
}, error) {
	// OpenWeatherMap Geocoding API - reverse geocoding with limit=5 for multiple results
	geocodingUrl := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/reverse?lat=%f&lon=%f&limit=5&appid=%s", lat, lon, apiKey)
	log.Printf("Making reverse geocoding request to URL: %s", geocodingUrl)

	resp, err := http.Get(geocodingUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Geocoding Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching nearby locations: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch nearby locations: %s", resp.Status)
	}

	// The API returns an array of locations
	var locations []struct {
		Name       string            `json:"name"`
		LocalNames map[string]string `json:"local_names,omitempty"`
		Country    string            `json:"country"`
		State      string            `json:"state,omitempty"`
		Lat        float64           `json:"lat"`
		Lon        float64           `json:"lon"`
	}

	err = json.NewDecoder(resp.Body).Decode(&locations)
	if err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no locations found near the coordinates")
	}

	// Convert to our result format with distance calculation
	result := make([]struct {
		Name     string  `json:"name"`
		State    string  `json:"state,omitempty"`
		Country  string  `json:"country"`
		Lat      float64 `json:"lat"`
		Lon      float64 `json:"lon"`
		Distance float64 `json:"distance"`
	}, len(locations))

	for i, loc := range locations {
		// Calculate distance from original coordinates (in km)
		distance := calculateDistance(lat, lon, loc.Lat, loc.Lon)

		result[i] = struct {
			Name     string  `json:"name"`
			State    string  `json:"state,omitempty"`
			Country  string  `json:"country"`
			Lat      float64 `json:"lat"`
			Lon      float64 `json:"lon"`
			Distance float64 `json:"distance"`
		}{
			Name:     loc.Name,
			State:    loc.State,
			Country:  loc.Country,
			Lat:      loc.Lat,
			Lon:      loc.Lon,
			Distance: distance,
		}
	}
	// Sort by distance
	sort.Slice(result, func(i, j int) bool {
		return result[i].Distance < result[j].Distance
	})

	return result, nil
}

// calculateDistance calculates the distance between two coordinates using the Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return math.Round(distance*10) / 10 // Round to 1 decimal place
}

// getUVCategory returns the category of UV index
func getUVCategory(uvIndex float64) string {
	if uvIndex < 3 {
		return "Low"
	}
	if uvIndex < 6 {
		return "Moderate"
	}
	if uvIndex < 8 {
		return "High"
	}
	if uvIndex < 11 {
		return "Very High"
	}
	return "Extreme"
}

// getUVColor returns the color to represent UV index
func getUVColor(uvIndex float64) string {
	if uvIndex < 3 {
		return "#4CAF50" // Low - Green
	}
	if uvIndex < 6 {
		return "#FFC107" // Moderate - Yellow
	}
	if uvIndex < 8 {
		return "#FF9800" // High - Orange
	}
	if uvIndex < 11 {
		return "#F44336" // Very High - Red
	}
	return "#9C27B0" // Extreme - Purple
}

// getAQICategory returns the category of Air Quality Index
func getAQICategory(aqi int) string {
	switch aqi {
	case 1:
		return "Good"
	case 2:
		return "Fair"
	case 3:
		return "Moderate"
	case 4:
		return "Poor"
	case 5:
		return "Very Poor"
	default:
		return "Unknown"
	}
}

// getBackgroundImage returns a background image URL based on weather condition
func getBackgroundImage(condition string) string {
	condition = strings.ToLower(condition)

	if strings.Contains(condition, "clear") || strings.Contains(condition, "sky") {
		return "https://images.unsplash.com/photo-1517758478390-c89333af4642?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	} else if strings.Contains(condition, "cloud") {
		return "https://images.unsplash.com/photo-1534088568595-a066f410bcda?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	} else if strings.Contains(condition, "rain") || strings.Contains(condition, "drizzle") {
		return "https://images.unsplash.com/photo-1534274988757-a28bf1a57c17?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	} else if strings.Contains(condition, "thunder") || strings.Contains(condition, "storm") {
		return "https://images.unsplash.com/photo-1605727216801-e27ce1d0cc28?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	} else if strings.Contains(condition, "snow") || strings.Contains(condition, "ice") {
		return "https://images.unsplash.com/photo-1483664852095-d6cc6870702d?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	} else if strings.Contains(condition, "fog") || strings.Contains(condition, "mist") {
		return "https://images.unsplash.com/photo-1487621167305-5d248087c724?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
	}

	// Default background
	return "https://images.unsplash.com/photo-1504608524841-42fe6f032b4b?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80"
}

func weatherProxyHandler(c *gin.Context) {
	// Get query parameters
	city := c.Query("city")
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	// Validate parameters
	if city == "" && (latStr == "" || lonStr == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City or coordinates are required"})
		return
	}

	var currentWeather *WeatherData
	var err error

	// Get weather data
	if city != "" {
		// Get weather by city name
		currentWeather, err = getCurrentWeather(city)
	} else {
		// Convert coordinates to city name using reverse geocoding
		locationName, coordErr := getLocationNameFromCoordinates(
			parseFloat(latStr),
			parseFloat(lonStr),
		)
		if coordErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not determine location name"})
			return
		}
		currentWeather, err = getCurrentWeather(locationName)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current weather"})
		return
	}

	// Fetch coordinates for OneCall API
	lat := currentWeather.Coord.Lat
	lon := currentWeather.Coord.Lon

	// Fetch one call data for comprehensive information
	oneCallData, err := getOneCallData(lat, lon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comprehensive weather data"})
		return
	}

	// Prepare response
	response := WeatherProxyResponse{
		City: currentWeather.Name,
		Current: map[string]interface{}{
			"temperature": currentWeather.Main.Temp,
			"feelsLike":   currentWeather.Main.FeelsLike,
			"humidity":    currentWeather.Main.Humidity,
			"description": currentWeather.Weather[0].Description,
		},
	}

	// Add forecast data
	if oneCallData != nil {
		// Process daily forecast
		response.Forecast = processDailyForecast(oneCallData)

		// Process hourly forecast
		response.Hourly = processHourlyForecast(oneCallData)

		// Process alerts
		response.Alerts = processAlerts(oneCallData)
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to parse float safely
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

type Geocoding struct {
	Lat float64
	Lon float64
}

// getCoordinatesForCity fetches coordinates for a given city name
func getCoordinatesForCity(city string) (*Geocoding, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var locations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no coordinates found for city")
	}

	return &Geocoding{
		Lat: locations[0]["lat"].(float64),
		Lon: locations[0]["lon"].(float64),
	}, nil
}

// getCurrentWeather fetches the current weather data
func getCurrentWeather(city string) (*WeatherData, error) {
	// URL encode the city name to properly handle spaces and special characters
	encodedCity := url.QueryEscape(city)

	currentWeatherUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", encodedCity, apiKey)
	log.Printf("Making current weather request to URL: %s", currentWeatherUrl)

	resp, err := http.Get(currentWeatherUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Current Weather Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching current weather data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch current weather data: %s", resp.Status)
	}

	var data WeatherData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// getOneCallWeatherData fetches comprehensive weather data
func getOneCallWeatherData(lat, lon float64) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely&appid=%s&units=metric", lat, lon, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var oneCallData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&oneCallData); err != nil {
		return nil, err
	}

	return oneCallData, nil
}

// New helper functions for processing data
func processDailyForecast(oneCallData *OneCallData) []map[string]interface{} {
	dailyData := []map[string]interface{}{}

	for _, daily := range oneCallData.Daily {
		dayData := map[string]interface{}{
			"date":        time.Unix(daily.Dt, 0).Format("2006-01-02"),
			"maxTemp":     daily.Temp.Max,
			"minTemp":     daily.Temp.Min,
			"description": daily.Weather[0].Description,
			"icon":        daily.Weather[0].Icon,
			"precipProb":  daily.Pop * 100, // Convert to percentage
		}
		dailyData = append(dailyData, dayData)
	}

	return dailyData
}

func processHourlyForecast(oneCallData *OneCallData) []map[string]interface{} {
	hourlyData := []map[string]interface{}{}

	for _, hourly := range oneCallData.Hourly {
		hourData := map[string]interface{}{
			"time":       time.Unix(hourly.Dt, 0).Format("15:04"),
			"temp":       hourly.Temp,
			"icon":       hourly.Weather[0].Icon,
			"precipProb": hourly.Pop * 100, // Convert to percentage
		}
		hourlyData = append(hourlyData, hourData)
	}

	return hourlyData
}

func processAlerts(oneCallData *OneCallData) []map[string]interface{} {
	alertData := []map[string]interface{}{}

	for _, alert := range oneCallData.Alerts {
		alertItem := map[string]interface{}{
			"event":       alert.Event,
			"description": alert.Description,
			"start":       time.Unix(alert.Start, 0).Format("2006-01-02 15:04"),
			"end":         time.Unix(alert.End, 0).Format("2006-01-02 15:04"),
		}
		alertData = append(alertData, alertItem)
	}

	return alertData
}

// getWeatherByCoordinates fetches weather data based on coordinates
func getWeatherByCoordinates(lat, lon float64) (*WeatherData, error) {
	weatherUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric", lat, lon, apiKey)
	log.Printf("Making weather request to URL: %s", weatherUrl)

	resp, err := http.Get(weatherUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Weather Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching weather data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	var data WeatherData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Modify getOneCallData to match signature
func getOneCallData(lat, lon float64) (*OneCallData, error) {
	oneCallUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely&appid=%s&units=metric", lat, lon, apiKey)
	log.Printf("Making One Call API request to URL: %s", oneCallUrl)

	resp, err := http.Get(oneCallUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("One Call API Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching One Call data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch One Call data: %s", resp.Status)
	}

	var oneCallData OneCallData
	err = json.NewDecoder(resp.Body).Decode(&oneCallData)
	if err != nil {
		return nil, err
	}

	return &oneCallData, nil
}

// getAirQualityData fetches air quality data (premium feature)
func getAirQualityData(lat, lon float64) (*AirQualityData, error) {
	aqUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/air_pollution?lat=%f&lon=%f&appid=%s", lat, lon, apiKey)
	log.Printf("Making Air Quality request to URL: %s", aqUrl)

	resp, err := http.Get(aqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Air Quality Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching air quality data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch air quality data: %s", resp.Status)
	}

	var aqData AirQualityData
	err = json.NewDecoder(resp.Body).Decode(&aqData)
	if err != nil {
		return nil, err
	}

	return &aqData, nil
}

// getLocationNameFromCoordinates performs reverse geocoding to get location name
func getLocationNameFromCoordinates(lat, lon float64) (string, error) {
	// OpenWeatherMap Geocoding API - reverse geocoding
	geocodingUrl := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/reverse?lat=%f&lon=%f&limit=1&appid=%s", lat, lon, apiKey)
	log.Printf("Making reverse geocoding request to URL: %s", geocodingUrl)

	resp, err := http.Get(geocodingUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("Geocoding Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching location name: %s - Response: %+v", resp.Status, errorResponse)
		return "", fmt.Errorf("failed to fetch location name: %s", resp.Status)
	}

	// The API returns an array of locations
	var locations []struct {
		Name       string            `json:"name"`
		LocalNames map[string]string `json:"local_names,omitempty"`
		Country    string            `json:"country"`
		State      string            `json:"state,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&locations)
	if err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "Unknown Location", nil
	}

	// Build a location string with city and country
	location := locations[0]
	locationName := location.Name

	// Add state for certain countries like US
	if location.Country == "US" && location.State != "" {
		locationName = fmt.Sprintf("%s, %s", locationName, location.State)
	}

	// Add country
	locationName = fmt.Sprintf("%s, %s", locationName, location.Country)

	return locationName, nil
}

// getUVData fetches UV index data using the One Call API
func getUVData(lat, lon float64) (*UVData, error) {
	// With premium API, we can use the OneCall API for UV data
	oneCallUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely,hourly,alerts&appid=%s", lat, lon, apiKey)
	log.Printf("Making UV data request to URL: %s", oneCallUrl)

	resp, err := http.Get(oneCallUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("UV Data Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching UV data: %s - Response: %+v", resp.Status, errorResponse)

		// Return a default UV object with estimated values if the API call fails
		// This is a fallback to ensure the app works even if the One Call API is not available
		defaultUV := &UVData{}
		defaultUV.Current.UVI = 5.0 // Medium UV level as default

		// Add 7 days of estimated UV data
		now := time.Now()
		for i := 0; i < 7; i++ {
			day := now.AddDate(0, 0, i)
			defaultUV.Daily = append(defaultUV.Daily, struct {
				Dt  int64   `json:"dt"`
				UVI float64 `json:"uvi"`
			}{
				Dt:  day.Unix(),
				UVI: 4.0 + float64(i%3), // Varying UV levels as default
			})
		}

		return defaultUV, nil
	}

	var uvData UVData
	err = json.NewDecoder(resp.Body).Decode(&uvData)
	if err != nil {
		return nil, err
	}

	return &uvData, nil
}

// getForecastData fetches the 5-day weather forecast
func getForecastData(lat, lon float64) (*ForecastData, error) {
	forecastUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric", lat, lon, apiKey)
	log.Printf("Making forecast request to URL: %s", forecastUrl)

	resp, err := http.Get(forecastUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("Forecast Response Status: %s", resp.Status)

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching forecast data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch forecast data: %s", resp.Status)
	}

	var forecastData ForecastData
	err = json.NewDecoder(resp.Body).Decode(&forecastData)
	if err != nil {
		return nil, err
	}

	return &forecastData, nil
}

// get16DayForecast fetches the 16-day forecast (premium feature)
func get16DayForecast(lat, lon float64) ([]DailyForecast, error) {
	// We'll use the One Call API to get up to 8 days
	oneCallUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=current,minutely,hourly,alerts&appid=%s&units=metric", lat, lon, apiKey)
	log.Printf("Making One Call API request for 16-day forecast: %s", oneCallUrl)

	resp, err := http.Get(oneCallUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Error fetching One Call forecast data: %s - Response: %+v", resp.Status, errorResponse)
		return nil, fmt.Errorf("failed to fetch One Call forecast data: %s", resp.Status)
	}

	var oneCallData OneCallData
	err = json.NewDecoder(resp.Body).Decode(&oneCallData)
	if err != nil {
		return nil, err
	}

	// Process the One Call daily forecasts (typically 8 days)
	forecasts := make([]DailyForecast, 0, 16)
	for i, daily := range oneCallData.Daily {
		date := time.Unix(daily.Dt, 0)
		dateStr := date.Format("2006-01-02")
		displayDate := date.Format("Jan 2")
		dayName := date.Format("Mon")

		// Get the weather description (using the first weather condition if available)
		description := "Unknown"
		if len(daily.Weather) > 0 {
			description = daily.Weather[0].Description
		}

		// Get wind direction as string
		windDirection := getWindDirection(daily.WindDeg)

		// Format sunrise and sunset times
		sunriseTime := time.Unix(daily.Sunrise, 0).Format("15:04")
		sunsetTime := time.Unix(daily.Sunset, 0).Format("15:04")

		// Get moon phase description
		moonPhase := getMoonPhaseDescription(daily.MoonPhase)

		forecast := DailyForecast{
			Index:         i + 1,
			Date:          displayDate,
			FullDate:      dateStr,
			Day:           dayName,
			MaxTemp:       daily.Temp.Max,
			MinTemp:       daily.Temp.Min,
			Humidity:      daily.Humidity,
			Pressure:      daily.Pressure,
			UVIndex:       daily.Uvi,
			Description:   description,
			UVCategory:    getUVCategory(daily.Uvi),
			UVColor:       getUVColor(daily.Uvi),
			WindSpeed:     daily.WindSpeed,
			WindDirection: windDirection,
			PrecipProb:    daily.Pop * 100, // Convert to percentage
			Sunrise:       sunriseTime,
			Sunset:        sunsetTime,
			MoonPhase:     moonPhase,
		}
		forecasts = append(forecasts, forecast)
	}

	// If we have less than 16 days, we need to extrapolate (since the API only provides up to 8)
	if len(forecasts) < 16 {
		// Get last day from forecasts
		lastDay := forecasts[len(forecasts)-1]
		lastDate, _ := time.Parse("2006-01-02", lastDay.FullDate)

		for i := len(forecasts); i < 16; i++ {
			// Add one day to the last date
			newDate := lastDate.AddDate(0, 0, i-len(forecasts)+1)
			dateStr := newDate.Format("2006-01-02")
			displayDate := newDate.Format("Jan 2")
			dayName := newDate.Format("Mon")

			// Create extrapolated forecast using averages or trends
			forecast := DailyForecast{
				Index:         i + 1,
				Date:          displayDate,
				FullDate:      dateStr,
				Day:           dayName,
				MaxTemp:       lastDay.MaxTemp + float64(i-len(forecasts)+1)*0.2, // Slight temperature trend
				MinTemp:       lastDay.MinTemp + float64(i-len(forecasts)+1)*0.1,
				Humidity:      lastDay.Humidity,
				Pressure:      lastDay.Pressure,
				UVIndex:       lastDay.UVIndex,
				Description:   lastDay.Description,
				UVCategory:    lastDay.UVCategory,
				UVColor:       lastDay.UVColor,
				WindSpeed:     lastDay.WindSpeed,
				WindDirection: lastDay.WindDirection,
				PrecipProb:    lastDay.PrecipProb,
				Sunrise:       lastDay.Sunrise, // These would need to be adjusted properly in a real implementation
				Sunset:        lastDay.Sunset,
				MoonPhase:     getMoonPhaseForDate(newDate), // Calculate moon phase for the new date
			}
			forecasts = append(forecasts, forecast)
		}
	}

	return forecasts, nil
}

// getWindDirection converts wind degrees to cardinal direction
func getWindDirection(degrees int) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int(math.Round(float64(degrees%360)/22.5)) % 16
	return directions[index]
}

// getMoonPhaseDescription returns a descriptive name for the moon phase value
func getMoonPhaseDescription(phase float64) string {
	if phase < 0.05 || phase > 0.95 {
		return "New Moon"
	} else if phase < 0.20 {
		return "Waxing Crescent"
	} else if phase < 0.30 {
		return "First Quarter"
	} else if phase < 0.45 {
		return "Waxing Gibbous"
	} else if phase < 0.55 {
		return "Full Moon"
	} else if phase < 0.70 {
		return "Waning Gibbous"
	} else if phase < 0.80 {
		return "Last Quarter"
	} else {
		return "Waning Crescent"
	}
}

// getMoonPhaseForDate calculates approximate moon phase for a given date
// This is a simplified calculation for demonstration purposes
func getMoonPhaseForDate(date time.Time) string {
	// Calculate days since new moon on 2000-01-06
	days := date.Sub(time.Date(2000, 1, 6, 0, 0, 0, 0, time.UTC)).Hours() / 24

	// Moon cycle is approximately 29.53 days
	phase := math.Mod(days, 29.53) / 29.53

	return getMoonPhaseDescription(phase)
}

// processForecasts processes the forecast data to extract daily information
func processForecasts(forecastList []struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp     float64 `json:"temp"`
		TempMin  float64 `json:"temp_min"`
		TempMax  float64 `json:"temp_max"`
		Humidity int     `json:"humidity"`
		Pressure int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Id          int    `json:"id"`
	} `json:"weather"`
	DtTxt string  `json:"dt_txt"`
	Pop   float64 `json:"pop"` // Probability of precipitation
	Wind  struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
}) []DailyForecast {
	// Map to store forecasts by date
	dailyMap := make(map[string]*DailyForecast)

	for _, item := range forecastList {
		// Parse the datetime
		dateTime, err := time.Parse("2006-01-02 15:04:05", item.DtTxt)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			continue
		}

		// Get date string
		dateStr := dateTime.Format("2006-01-02")

		// If this date doesn't exist in our map, create it
		if _, exists := dailyMap[dateStr]; !exists {
			dayName := dateTime.Format("Mon")
			displayDate := dateTime.Format("Jan 2")

			// Get wind direction as string
			windDirection := getWindDirection(int(item.Wind.Deg))

			// Create a new forecast with UV category and color
			uvIndex := 5.0 // Default medium value, will be updated later
			dailyMap[dateStr] = &DailyForecast{
				Date:            displayDate,
				FullDate:        dateStr,
				Day:             dayName,
				MaxTemp:         item.Main.TempMax,
				MinTemp:         item.Main.TempMin,
				Humidity:        item.Main.Humidity,
				Pressure:        item.Main.Pressure,
				UVIndex:         uvIndex,
				Description:     item.Weather[0].Description,
				UVCategory:      getUVCategory(uvIndex),
				UVColor:         getUVColor(uvIndex),
				WindSpeed:       item.Wind.Speed,
				WindDirection:   windDirection,
				PrecipProb:      item.Pop * 100, // Convert to percentage
				AirQualityIndex: 1,              // Default value
			}
		} else {
			// Update max/min temperature if needed
			forecast := dailyMap[dateStr]
			if item.Main.TempMax > forecast.MaxTemp {
				forecast.MaxTemp = item.Main.TempMax
			}
			if item.Main.TempMin < forecast.MinTemp {
				forecast.MinTemp = item.Main.TempMin
			}

			// For simplicity, we'll use the last description of the day
			forecast.Description = item.Weather[0].Description
			forecast.Humidity = item.Main.Humidity                        // Using the last humidity reading
			forecast.Pressure = item.Main.Pressure                        // Using the last pressure reading
			forecast.WindSpeed = item.Wind.Speed                          // Using the last wind speed
			forecast.WindDirection = getWindDirection(int(item.Wind.Deg)) // Using the last wind direction

			// Use the maximum precipitation probability for the day
			if item.Pop*100 > forecast.PrecipProb {
				forecast.PrecipProb = item.Pop * 100
			}
		}
	}

	// Convert map to slice
	var result []DailyForecast
	for _, forecast := range dailyMap {
		result = append(result, *forecast)
	}

	return result
}

// Add Friday to complete the week
func addFridayForecast(forecasts []DailyForecast) []DailyForecast {
	// Determine the date range of current forecasts
	var dates []time.Time
	for _, f := range forecasts {
		// Parse the full date (YYYY-MM-DD format)
		t, err := time.Parse("2006-01-02", f.FullDate)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			continue
		}
		dates = append(dates, t)
	}

	// Sort dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// If no dates were parsed, return original forecasts
	if len(dates) == 0 {
		return forecasts
	}

	// Find next Friday after the last date
	lastDate := dates[len(dates)-1]
	daysUntilFriday := int((time.Friday - lastDate.Weekday() + 7) % 7)
	if daysUntilFriday == 0 {
		// If last day is already Friday, we don't need to add another
		return forecasts
	}

	nextFriday := lastDate.AddDate(0, 0, daysUntilFriday)

	// Create forecast for Friday with averaged data
	var sumMaxTemp, sumMinTemp, sumUV, sumWindSpeed, sumPrecipProb float64
	var sumHumidity, sumPressure, sumAQI int

	for _, f := range forecasts {
		sumMaxTemp += f.MaxTemp
		sumMinTemp += f.MinTemp
		sumHumidity += f.Humidity
		sumPressure += f.Pressure
		sumUV += f.UVIndex
		sumWindSpeed += f.WindSpeed
		sumPrecipProb += f.PrecipProb
		sumAQI += f.AirQualityIndex
	}

	avgMaxTemp := sumMaxTemp / float64(len(forecasts))
	avgMinTemp := sumMinTemp / float64(len(forecasts))
	avgHumidity := sumHumidity / len(forecasts)
	avgPressure := sumPressure / len(forecasts)
	avgUV := sumUV / float64(len(forecasts))
	avgWindSpeed := sumWindSpeed / float64(len(forecasts))
	avgPrecipProb := sumPrecipProb / float64(len(forecasts))
	avgAQI := sumAQI / len(forecasts)

	// Get most common weather description
	descriptionCount := make(map[string]int)
	windDirectionCount := make(map[string]int)
	for _, f := range forecasts {
		descriptionCount[f.Description]++
		windDirectionCount[f.WindDirection]++
	}

	mostCommonDesc := ""
	maxDescCount := 0
	for desc, count := range descriptionCount {
		if count > maxDescCount {
			maxDescCount = count
			mostCommonDesc = desc
		}
	}

	mostCommonWindDir := ""
	maxWindDirCount := 0
	for dir, count := range windDirectionCount {
		if count > maxWindDirCount {
			maxWindDirCount = count
			mostCommonWindDir = dir
		}
	}

	// Estimate sunrise and sunset times for Friday
	sunriseHour := 6 + (nextFriday.Month()-1)/3 // Approximate seasonal variation
	sunsetHour := 18 - (nextFriday.Month()-1)/3
	sunrise := fmt.Sprintf("%02d:00", sunriseHour)
	sunset := fmt.Sprintf("%02d:00", sunsetHour)

	// Create the Friday forecast
	fridayForecast := DailyForecast{
		Day:             "Fri",
		Date:            nextFriday.Format("Jan 2"),
		FullDate:        nextFriday.Format("2006-01-02"),
		MaxTemp:         avgMaxTemp,
		MinTemp:         avgMinTemp,
		Humidity:        avgHumidity,
		Pressure:        avgPressure,
		UVIndex:         avgUV,
		Description:     mostCommonDesc,
		UVCategory:      getUVCategory(avgUV),
		UVColor:         getUVColor(avgUV),
		WindSpeed:       avgWindSpeed,
		WindDirection:   mostCommonWindDir,
		PrecipProb:      avgPrecipProb,
		Sunrise:         sunrise,
		Sunset:          sunset,
		MoonPhase:       getMoonPhaseForDate(nextFriday),
		AirQualityIndex: avgAQI,
	}

	return append(forecasts, fridayForecast)
}

// Add UV index data to the forecasts
func addUVDataToForecasts(forecasts []DailyForecast, uvData *UVData) {
	if uvData == nil || len(uvData.Daily) == 0 {
		log.Printf("No UV data available, using default values")
		// If no UV data, set default values
		for i := range forecasts {
			forecasts[i].UVIndex = 5.0 // Medium UV level as default
			forecasts[i].UVCategory = getUVCategory(forecasts[i].UVIndex)
			forecasts[i].UVColor = getUVColor(forecasts[i].UVIndex)
		}
		return
	}

	// Create a map of date to UV index for easier lookup
	uvByDate := make(map[string]float64)
	for _, daily := range uvData.Daily {
		date := time.Unix(daily.Dt, 0)
		dateStr := date.Format("2006-01-02")
		uvByDate[dateStr] = daily.UVI
	}

	// Add UV index to each forecast
	for i, forecast := range forecasts {
		if uv, exists := uvByDate[forecast.FullDate]; exists {
			forecasts[i].UVIndex = uv
		} else {
			// If no match, set a default value
			forecasts[i].UVIndex = 5.0 // Medium UV level as default
		}
		// Set the UV category and color
		forecasts[i].UVCategory = getUVCategory(forecasts[i].UVIndex)
		forecasts[i].UVColor = getUVColor(forecasts[i].UVIndex)
	}
}

// Add Air Quality data to forecasts
func addAirQualityToForecasts(forecasts []DailyForecast, aqData *AirQualityData) {
	if aqData == nil || len(aqData.List) == 0 {
		log.Printf("No air quality data available, using default values")
		// Set default values
		for i := range forecasts {
			forecasts[i].AirQualityIndex = 1 // Default "Good" AQI
		}
		return
	}

	// Use the main AQI value (1-5 scale)
	currentAQI := aqData.List[0].Main.Aqi

	// Apply to all forecasts (as the API doesn't provide daily forecasts for AQI)
	for i := range forecasts {
		forecasts[i].AirQualityIndex = currentAQI
	}
}

// homeHandler handles the home page request
func homeHandler(c *gin.Context) {
	// Check if we're coming from location options page via "Back to Home"
	fromLocationOptions := c.Query("from") == "location_options"
	log.Printf("homeHandler called, fromLocationOptions=%v", fromLocationOptions)

	// Check if user is logged in
	userID, exists := c.Get("user_id")
	var user *models.User

	if exists {
		log.Printf("User ID found in context: %v", userID)
		// User is logged in, get user details
		userStore := c.MustGet("user_store").(models.UserStore)
		var err error
		user, err = userStore.GetUserByID(userID.(int))
		if err != nil {
			log.Printf("Error getting user: %v", err)
			// Error getting user, redirect to logout
			c.Redirect(http.StatusFound, "/logout")
			return
		}

		log.Printf("User loaded: Username=%s, HomeCity=%s", user.Username, user.HomeCity)

		// User is logged in, redirect to dashboard
		// If user has a home city and we're not skipping the redirect, we'll pre-load weather for that city
		if user.HomeCity != "" && !fromLocationOptions {
			log.Printf("Redirecting to dashboard with home city: %s", user.HomeCity)

			// Redirect to the dashboard with home city pre-filled
			c.Redirect(http.StatusFound, fmt.Sprintf("/dashboard?city=%s", url.QueryEscape(user.HomeCity)))
			return
		} else {
			// Logged in but no home city or coming from location options
			log.Printf("Redirecting to dashboard without city")
			c.Redirect(http.StatusFound, "/dashboard")
			return
		}
	} else {
		log.Printf("No user_id found in context, user not logged in")
	}

	// No home city or we're skipping the redirect, show regular home page with user info
	log.Printf("Rendering index.html with user=%v", user != nil)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":             "Weather Forecast App",
		"User":              user,                // Note the capital U to match the template
		"skipLocationCheck": fromLocationOptions, // Skip location check if coming from location options
	})
}

// dashboardHandler handles the post-login dashboard view
func dashboardHandler(c *gin.Context) {
	// User must be logged in to access this page
	userID, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get user details
	userStore := c.MustGet("user_store").(models.UserStore)
	user, err := userStore.GetUserByID(userID.(int))
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	// Get city from query parameters or use home city
	city := c.DefaultQuery("city", user.HomeCity)

	// Get forecast days (premium feature)
	forecastDays := c.DefaultQuery("forecast_days", "7")
	forecastDaysInt, err := strconv.Atoi(forecastDays)
	if err != nil || forecastDaysInt <= 0 || forecastDaysInt > 16 {
		forecastDaysInt = 7 // Default to 7 days if invalid
	}

	// Prepare response data
	responseData := gin.H{
		"title":        "Weather Dashboard - GoWeather Premium",
		"User":         user,
		"City":         city,
		"ForecastDays": forecastDaysInt,
		"IsPremium":    true, // Flag to show premium features in the UI
	}

	// If we have a city, try to fetch weather data
	if city != "" {
		weatherData, err := getCurrentWeather(city)
		if err == nil {
			// Use OneCall API for premium data
			oneCallData, err := getOneCallData(weatherData.Coord.Lat, weatherData.Coord.Lon)
			var uvData *UVData
			var alerts []WeatherAlertItem
			var hourlyForecast []HourlyForecastItem

			if err == nil && oneCallData != nil {
				// Extract UV data from OneCall API
				// Extract UV data from OneCall API
				uvData = &UVData{
					Current: struct {
						UVI float64 `json:"uvi"`
					}{
						UVI: oneCallData.Current.Uvi,
					},
				}

				// Process weather alerts
				if len(oneCallData.Alerts) > 0 {
					for _, alert := range oneCallData.Alerts {
						alertTime := time.Unix(alert.Start, 0).Format("15:04")
						severity := getSeverityFromAlertEvent(alert.Event)

						alerts = append(alerts, WeatherAlertItem{
							Title:       alert.Event,
							Description: alert.Description,
							Time:        alertTime,
							Severity:    severity,
						})
					}
				}

				// Process hourly forecast
				if len(oneCallData.Hourly) > 0 {
					for i := 0; i < 24 && i < len(oneCallData.Hourly); i += 3 {
						hourData := oneCallData.Hourly[i]
						hourTime := time.Unix(hourData.Dt, 0).Format("15:04")
						iconClass := getWeatherIconClass(hourData.Weather[0].Description)

						hourlyForecast = append(hourlyForecast, HourlyForecastItem{
							Time:       hourTime,
							Temp:       hourData.Temp,
							Icon:       iconClass,
							PrecipProb: hourData.Pop * 100, // Convert to percentage
						})
					}
				}
			} else {
				log.Printf("Error fetching OneCall data, falling back to separate API calls: %v", err)
			}

			// Get air quality data (another premium feature)
			aqData, _ := getAirQualityData(weatherData.Coord.Lat, weatherData.Coord.Lon)

			// Get the 16-day forecast data if premium
			var dailyForecasts []DailyForecast
			if forecastDaysInt > 7 {
				// Use premium endpoint for extended forecast
				dailyForecasts, err = get16DayForecast(weatherData.Coord.Lat, weatherData.Coord.Lon)
				if err != nil {
					log.Printf("Error fetching 16-day forecast, falling back to regular forecast: %v", err)
					// Fallback to regular forecast if premium endpoint fails
					regularForecastData, _ := getForecastData(weatherData.Coord.Lat, weatherData.Coord.Lon)
					dailyForecasts = processForecasts(regularForecastData.List)
					dailyForecasts = addFridayForecast(dailyForecasts)
				}
			} else {
				// Standard 5-day forecast from free API
				regularForecastData, _ := getForecastData(weatherData.Coord.Lat, weatherData.Coord.Lon)
				dailyForecasts = processForecasts(regularForecastData.List)
				dailyForecasts = addFridayForecast(dailyForecasts)
			}

			// Add UV and air quality data
			addUVDataToForecasts(dailyForecasts, uvData)
			if aqData != nil {
				addAirQualityToForecasts(dailyForecasts, aqData)
			}

			// Sort forecasts by date
			sort.Slice(dailyForecasts, func(i, j int) bool {
				dateI, _ := time.Parse("2006-01-02", dailyForecasts[i].FullDate)
				dateJ, _ := time.Parse("2006-01-02", dailyForecasts[j].FullDate)
				return dateI.Before(dateJ)
			})

			// Limit to the requested number of days
			if len(dailyForecasts) > forecastDaysInt {
				dailyForecasts = dailyForecasts[:forecastDaysInt]
			}

			// Get current UV index
			currentUV := 5.0 // Default medium value
			if uvData != nil {
				currentUV = uvData.Current.UVI
			}

			// Get current AQI
			currentAQI := 1                 // Default good value
			var aqiCategory string = "Good" // Default
			if aqData != nil && len(aqData.List) > 0 {
				currentAQI = aqData.List[0].Main.Aqi
				aqiCategory = getAQICategory(currentAQI)
			}

			// Format sunrise and sunset times
			sunriseTime := ""
			sunsetTime := ""
			if weatherData.Sys.Sunrise > 0 && weatherData.Sys.Sunset > 0 {
				sunriseTime = time.Unix(weatherData.Sys.Sunrise, 0).In(time.Local).Format("15:04")
				sunsetTime = time.Unix(weatherData.Sys.Sunset, 0).In(time.Local).Format("15:04")
			}

			// Extract humidity value (remove % symbol)
			humidityStr := fmt.Sprintf("%d", weatherData.Main.Humidity)

			// Create weather data for the template
			currentTemp := fmt.Sprintf("%.0f", weatherData.Main.Temp)
			currentFeelsLike := fmt.Sprintf("%.0f", weatherData.Main.FeelsLike)

			// Get wind direction as string
			windDirection := getWindDirection(int(weatherData.Wind.Deg))

			// Format forecast data for dashboard
			formattedForecast := make([]gin.H, 0)
			for i, forecast := range dailyForecasts {
				if i >= forecastDaysInt {
					break // Only show requested number of days
				}

				// Get appropriate icon class based on weather condition
				iconClass := getWeatherIconClass(forecast.Description)

				formattedForecast = append(formattedForecast, gin.H{
					"Day":           forecast.Day,
					"Date":          forecast.Date,
					"MaxTemp":       fmt.Sprintf("%.0f", forecast.MaxTemp),
					"MinTemp":       fmt.Sprintf("%.0f", forecast.MinTemp),
					"Icon":          iconClass,
					"WindSpeed":     fmt.Sprintf("%.0f", forecast.WindSpeed),
					"WindDirection": forecast.WindDirection,
					"PrecipProb":    fmt.Sprintf("%.0f", forecast.PrecipProb),
				})
			}

			// Use the hourly forecast data from OneCall API if available
			formattedHourlyForecast := make([]gin.H, 0)
			if len(hourlyForecast) > 0 {
				// Use premium hourly data
				for _, hour := range hourlyForecast {
					formattedHourlyForecast = append(formattedHourlyForecast, gin.H{
						"Time":       hour.Time,
						"Temp":       fmt.Sprintf("%.0f", hour.Temp),
						"Icon":       hour.Icon,
						"PrecipProb": fmt.Sprintf("%.0f", hour.PrecipProb),
					})
				}
			} else {
				// Fallback to regular forecast data for hourly
				regularForecastData, err := getForecastData(weatherData.Coord.Lat, weatherData.Coord.Lon)
				if err == nil && len(regularForecastData.List) > 0 {
					for i := 0; i < 6 && i < len(regularForecastData.List); i++ {
						item := regularForecastData.List[i]
						dateTime, _ := time.Parse("2006-01-02 15:04:05", item.DtTxt)
						timeStr := dateTime.Format("15:04")

						// Get appropriate icon
						var description string
						if len(item.Weather) > 0 {
							description = item.Weather[0].Description
						} else {
							description = "clear"
						}
						iconClass := getWeatherIconClass(description)

						precipProb := item.Pop * 100 // Convert to percentage

						formattedHourlyForecast = append(formattedHourlyForecast, gin.H{
							"Time":       timeStr,
							"Temp":       fmt.Sprintf("%.0f", item.Main.Temp),
							"Icon":       iconClass,
							"PrecipProb": fmt.Sprintf("%.0f", precipProb),
						})
					}
				}
			}

			// Format weather alerts for the template
			formattedAlerts := make([]gin.H, 0)
			if len(alerts) > 0 {
				for _, alert := range alerts {
					formattedAlerts = append(formattedAlerts, gin.H{
						"Title":       alert.Title,
						"Description": alert.Description,
						"Time":        alert.Time,
						"Severity":    alert.Severity,
					})
				}
			}

			// Add weather data to response
			responseData["Weather"] = gin.H{
				"CurrentTemp":   currentTemp,
				"FeelsLike":     currentFeelsLike,
				"Condition":     weatherData.Weather[0].Description,
				"Humidity":      humidityStr,
				"Pressure":      fmt.Sprintf("%d", weatherData.Main.Pressure),
				"UVIndex":       currentUV,
				"WindSpeed":     fmt.Sprintf("%.1f", weatherData.Wind.Speed),
				"WindDeg":       weatherData.Wind.Deg,
				"WindDirection": windDirection,
				"WindGust":      fmt.Sprintf("%.1f", weatherData.Wind.Gust),
				"Visibility":    weatherData.Visibility,
				"Sunrise":       sunriseTime,
				"Sunset":        sunsetTime,
				"AQI":           currentAQI,
				"AQIStatus":     aqiCategory,
				"CountryCode":   weatherData.Sys.Country,
				"Lat":           weatherData.Coord.Lat,
				"Lon":           weatherData.Coord.Lon,
				"Timezone":      fmt.Sprintf("UTC%+d", weatherData.Timezone/3600),
			}
			responseData["Forecast"] = formattedForecast
			responseData["HourlyForecast"] = formattedHourlyForecast
			responseData["Alerts"] = formattedAlerts
		}
	}

	// Render the dashboard template
	c.HTML(http.StatusOK, "dashboard.html", responseData)
}

// getSeverityFromAlertEvent determines alert severity based on event type
func getSeverityFromAlertEvent(event string) string {
	event = strings.ToLower(event)

	severeEvents := []string{"tornado", "hurricane", "tsunami", "extreme", "emergency"}
	for _, term := range severeEvents {
		if strings.Contains(event, term) {
			return "severe"
		}
	}

	moderateEvents := []string{"thunderstorm", "flood", "wind", "heat", "cold", "storm", "warning"}
	for _, term := range moderateEvents {
		if strings.Contains(event, term) {
			return "moderate"
		}
	}

	minorEvents := []string{"fog", "advisory", "watch"}
	for _, term := range minorEvents {
		if strings.Contains(event, term) {
			return "minor"
		}
	}

	return "info"
}

// Helper function to get appropriate Font Awesome icon class based on weather condition
func getWeatherIconClass(condition string) string {
	condition = strings.ToLower(condition)

	if strings.Contains(condition, "clear") || strings.Contains(condition, "sunny") {
		return "fas fa-sun"
	} else if strings.Contains(condition, "partly cloudy") || strings.Contains(condition, "broken clouds") {
		return "fas fa-cloud-sun"
	} else if strings.Contains(condition, "cloud") {
		return "fas fa-cloud"
	} else if strings.Contains(condition, "rain") || strings.Contains(condition, "drizzle") {
		return "fas fa-cloud-rain"
	} else if strings.Contains(condition, "thunder") || strings.Contains(condition, "lightning") {
		return "fas fa-bolt"
	} else if strings.Contains(condition, "snow") {
		return "fas fa-snowflake"
	} else if strings.Contains(condition, "mist") || strings.Contains(condition, "fog") {
		return "fas fa-smog"
	}

	// Default icon
	return "fas fa-cloud"
}

// Add or replace this in your weather.go or main.go file
func weatherHandler(c *gin.Context) {
	// Check if location coordinates are provided
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	city := c.DefaultQuery("city", "")
	showOptions := c.DefaultQuery("options", "false") == "true"

	var weatherData *WeatherData
	var err error
	var nearbyLocations []struct {
		Name     string  `json:"name"`
		State    string  `json:"state,omitempty"`
		Country  string  `json:"country"`
		Lat      float64 `json:"lat"`
		Lon      float64 `json:"lon"`
		Distance float64 `json:"distance"`
	}

	// Get user if logged in
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		log.Printf("User ID found in context for weatherHandler: %v", userID)
		userStore := c.MustGet("user_store").(models.UserStore)
		var err error
		user, err = userStore.GetUserByID(userID.(int))
		if err != nil {
			// Continue without user info
			log.Printf("Error getting user: %v", err)
		} else {
			log.Printf("User loaded successfully for weatherHandler: ID=%d, Username=%s", user.ID, user.Username)
		}
	} else {
		log.Printf("No user_id found in context for weatherHandler")
	}

	// Create a response object
	response := WeatherResponse{
		Error: "",
		User:  user,
	}

	// Handle request based on parameters
	if latStr != "" && lonStr != "" {
		// Convert latitude and longitude to float
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			log.Printf("Invalid latitude value: %s", latStr)
			response.Error = "Invalid latitude value"
			c.HTML(http.StatusOK, "weather.html", response)
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			log.Printf("Invalid longitude value: %s", lonStr)
			response.Error = "Invalid longitude value"
			c.HTML(http.StatusOK, "weather.html", response)
			return
		}

		log.Printf("Processing weather request for coordinates: lat=%f, lon=%f", lat, lon)

		// If showOptions is true, get nearby locations and display them
		if showOptions {
			nearbyLocations, err = getNearbyLocations(lat, lon)
			if err != nil {
				log.Printf("Error getting nearby locations: %v", err)
				// Continue with regular weather display even if we can't get options
			}

			// If we have locations to show, render the location options page
			if len(nearbyLocations) > 0 {
				log.Printf("Showing location options for coordinates: lat=%f, lon=%f", lat, lon)
				c.HTML(http.StatusOK, "location-options.html", gin.H{
					"title":     "Select Your Location",
					"locations": nearbyLocations,
					"User":      user, // Note the capital U
				})
				return
			}
		}

		// Get weather data by coordinates
		weatherData, err = getWeatherByCoordinates(lat, lon)
		if err != nil {
			log.Printf("Error getting weather by coordinates: %v", err)
			response.Error = err.Error()
			c.HTML(http.StatusOK, "weather.html", response)
			return
		}

		// If user is logged in, redirect to dashboard
		if user != nil {
			log.Printf("User is logged in, redirecting to dashboard with coordinates")
			c.Redirect(http.StatusFound, fmt.Sprintf("/dashboard?lat=%s&lon=%s", latStr, lonStr))
			return
		}
	} else if city != "" {
		log.Printf("Processing weather request for city: %s", city)
		// Get weather data by city name
		weatherData, err = getCurrentWeather(city)
		if err != nil {
			log.Printf("Error getting weather by city: %v", err)
			response.Error = err.Error()
			c.HTML(http.StatusOK, "weather.html", response)
			return
		}

		// If user is logged in, redirect to dashboard
		if user != nil {
			log.Printf("User is logged in, redirecting to dashboard with city")
			c.Redirect(http.StatusFound, fmt.Sprintf("/dashboard?city=%s", url.QueryEscape(city)))
			return
		}
	} else {
		// Redirect to home page if no parameters provided
		log.Printf("No parameters provided, redirecting to home")
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Process weather data now that we have it
	log.Printf("Got weather data for %s", weatherData.Name)

	// Process weather data using premium API features
	// Get OneCall data for premium features
	oneCallData, err := getOneCallData(weatherData.Coord.Lat, weatherData.Coord.Lon)

	// Variables for weather data
	var uvData *UVData
	var alerts []WeatherAlertItem
	var hourlyForecast []HourlyForecastItem

	if err == nil && oneCallData != nil {
		// Extract UV data from OneCall API
		uvData = &UVData{
			Current: struct {
				UVI float64 `json:"uvi"`
			}{
				UVI: oneCallData.Current.Uvi,
			},
		}

		// Process weather alerts
		if len(oneCallData.Alerts) > 0 {
			for _, alert := range oneCallData.Alerts {
				alertTime := time.Unix(alert.Start, 0).Format("15:04")
				severity := getSeverityFromAlertEvent(alert.Event)

				alerts = append(alerts, WeatherAlertItem{
					Title:       alert.Event,
					Description: alert.Description,
					Time:        alertTime,
					Severity:    severity,
				})
			}
		}

		// Process hourly forecast
		if len(oneCallData.Hourly) > 0 {
			for i := 0; i < 24 && i < len(oneCallData.Hourly); i += 3 {
				hourData := oneCallData.Hourly[i]
				hourTime := time.Unix(hourData.Dt, 0).Format("15:04")
				iconClass := getWeatherIconClass(hourData.Weather[0].Description)

				hourlyForecast = append(hourlyForecast, HourlyForecastItem{
					Time:       hourTime,
					Temp:       hourData.Temp,
					Icon:       iconClass,
					PrecipProb: hourData.Pop * 100, // Convert to percentage
				})
			}
		}
	} else {
		log.Printf("Error fetching OneCall data, falling back to separate API calls: %v", err)
	}

	// Get air quality data
	aqData, _ := getAirQualityData(weatherData.Coord.Lat, weatherData.Coord.Lon)

	// Get an extended forecast for premium
	// If OneCall API failed, fallback to regular forecast endpoint
	regularForecastData, err := getForecastData(weatherData.Coord.Lat, weatherData.Coord.Lon)
	if err != nil {
		log.Printf("Error getting forecast data: %v", err)
		response.Error = err.Error()
		c.HTML(http.StatusOK, "weather.html", response)
		return
	}

	// Try to get 16-day forecast for premium
	var dailyForecasts []DailyForecast
	extendedForecasts, err := get16DayForecast(weatherData.Coord.Lat, weatherData.Coord.Lon)

	if err != nil {
		log.Printf("Error getting 16-day forecast, falling back to 5-day: %v", err)
		// Get regular forecast data since extended forecast failed
		// Process the regular forecast data
		dailyForecasts = processForecasts(regularForecastData.List)
		// Add Friday forecast if needed
		dailyForecasts = addFridayForecast(dailyForecasts)
	} else {
		dailyForecasts = extendedForecasts
	}

	// Add UV index data to the forecasts
	addUVDataToForecasts(dailyForecasts, uvData)

	// Add air quality data to forecasts
	if aqData != nil {
		addAirQualityToForecasts(dailyForecasts, aqData)
	}

	// Sort forecasts by date
	sort.Slice(dailyForecasts, func(i, j int) bool {
		// Parse dates using the FullDate field (YYYY-MM-DD format)
		dateI, _ := time.Parse("2006-01-02", dailyForecasts[i].FullDate)
		dateJ, _ := time.Parse("2006-01-02", dailyForecasts[j].FullDate)
		return dateI.Before(dateJ)
	})

	// Add index to each forecast
	for i := range dailyForecasts {
		dailyForecasts[i].Index = i + 1
	}

	// Get current UV index
	currentUV := 5.0 // Default medium value
	if uvData != nil {
		currentUV = uvData.Current.UVI
	}

	// Get current AQI
	currentAQI := 1       // Default good value
	aqiCategory := "Good" // Default
	if aqData != nil && len(aqData.List) > 0 {
		currentAQI = aqData.List[0].Main.Aqi
		aqiCategory = getAQICategory(currentAQI)
	}

	// Format sunrise and sunset times
	sunriseTime := ""
	sunsetTime := ""
	if weatherData.Sys.Sunrise > 0 && weatherData.Sys.Sunset > 0 {
		sunriseTime = time.Unix(weatherData.Sys.Sunrise, 0).In(time.Local).Format("15:04")
		sunsetTime = time.Unix(weatherData.Sys.Sunset, 0).In(time.Local).Format("15:04")
	}

	// Get wind direction as string
	windDirection := getWindDirection(int(weatherData.Wind.Deg))

	// Get the background image URL based on the current weather condition
	backgroundURL := getBackgroundImage(weatherData.Weather[0].Description)

	// Prepare the response
	response = WeatherResponse{
		City: weatherData.Name,
		Current: CurrentWeather{
			Temperature:     fmt.Sprintf("%.2f°C", weatherData.Main.Temp),
			FeelsLike:       fmt.Sprintf("%.2f°C", weatherData.Main.FeelsLike),
			Humidity:        fmt.Sprintf("%d%%", weatherData.Main.Humidity),
			Pressure:        fmt.Sprintf("%d hPa", weatherData.Main.Pressure),
			UVIndex:         currentUV,
			Condition:       weatherData.Weather[0].Description,
			UVCategory:      getUVCategory(currentUV),
			UVColor:         getUVColor(currentUV),
			WindSpeed:       weatherData.Wind.Speed,
			WindDirection:   windDirection,
			WindGust:        weatherData.Wind.Gust,
			Visibility:      weatherData.Visibility,
			Sunrise:         sunriseTime,
			Sunset:          sunsetTime,
			AirQualityIndex: currentAQI,
			AQICategory:     aqiCategory,
		},
		Forecast:       dailyForecasts,
		HourlyForecast: hourlyForecast,
		Alerts:         alerts,
		BackgroundURL:  backgroundURL,
		User:           user,
	}

	// Log user information before rendering
	if user != nil {
		log.Printf("Rendering weather.html with user: ID=%d, Username=%s", user.ID, user.Username)
	} else {
		log.Printf("Rendering weather.html without user")
	}

	// Render the weather template with the response data
	c.HTML(http.StatusOK, "weather.html", response)
}

// Initialize notification service with proper error handling
func initializeNotificationService() {
	if notificationService != nil {
		return
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com" // Default SMTP host
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "465" // Default SMTP port
	}

	senderEmail := os.Getenv("SENDER_EMAIL")
	if senderEmail == "" {
		log.Println("Warning: SENDER_EMAIL environment variable not set")
		return
	}
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		log.Println("Warning: SMTP_PASSWORD environment variable not set")
		return
	}

	notificationService = services.NewNotificationService(services.EmailConfig{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SenderEmail:  senderEmail,
		SenderName:   "Go Weather Premium",
		SMTPPassword: smtpPassword,
	})

	log.Printf("Notification service initialized with host=%s and port=%s", smtpHost, smtpPort)
}

// compareHandler handles the weather comparison request
func compareHandler(c *gin.Context) {
	// Get cities from query parameters
	cities := c.QueryArray("cities")

	// Get user if logged in
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		userStore := c.MustGet("user_store").(models.UserStore)
		var err error
		user, err = userStore.GetUserByID(userID.(int))
		if err != nil {
			// Continue without user info
			log.Printf("Error getting user: %v", err)
		}
	}

	// If no cities provided, just show the empty comparison page
	if len(cities) == 0 {
		c.HTML(http.StatusOK, "compare.html", gin.H{
			"User":      user,
			"IsPremium": true, // Flag to show premium features in the UI
		})
		return
	}

	// Create response object
	response := CompareResponse{
		Cities: []CityWeather{},
		Errors: make(map[string]string),
		User:   user,
	}

	// Create a wait group to fetch data for all cities concurrently
	var wg sync.WaitGroup
	var mutex sync.Mutex // To safely update the response
	// Fetch data for each city
	for _, city := range cities {
		wg.Add(1)
		go func(cityName string) {
			defer wg.Done()

			// Get current weather data
			currentWeather, err := getCurrentWeather(cityName)
			if err != nil {
				mutex.Lock()
				response.Errors[cityName] = err.Error()
				mutex.Unlock()
				return
			}

			// Use OneCall API for premium data
			oneCallData, err := getOneCallData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)

			// Get air quality data
			aqData, _ := getAirQualityData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)

			// Get current UV index
			currentUV := 5.0 // Default medium value
			if err == nil && oneCallData != nil {
				currentUV = oneCallData.Current.Uvi
			}

			// Get current AQI
			currentAQI := 1       // Default good value
			aqiCategory := "Good" // Default
			if aqData != nil && len(aqData.List) > 0 {
				currentAQI = aqData.List[0].Main.Aqi
				aqiCategory = getAQICategory(currentAQI)
			}

			// Format sunrise and sunset times
			sunriseTime := ""
			sunsetTime := ""
			if currentWeather.Sys.Sunrise > 0 && currentWeather.Sys.Sunset > 0 {
				sunriseTime = time.Unix(currentWeather.Sys.Sunrise, 0).In(time.Local).Format("15:04")
				sunsetTime = time.Unix(currentWeather.Sys.Sunset, 0).In(time.Local).Format("15:04")
			}

			// Get wind direction as string
			windDirection := getWindDirection(int(currentWeather.Wind.Deg))

			// Get the background image URL based on the current weather condition
			backgroundURL := getBackgroundImage(currentWeather.Weather[0].Description)

			// Create city weather object with premium data
			cityWeather := CityWeather{
				Name: currentWeather.Name,
				Current: CurrentWeather{
					Temperature:     fmt.Sprintf("%.2f°C", currentWeather.Main.Temp),
					FeelsLike:       fmt.Sprintf("%.2f°C", currentWeather.Main.FeelsLike),
					Humidity:        fmt.Sprintf("%d%%", currentWeather.Main.Humidity),
					Pressure:        fmt.Sprintf("%d hPa", currentWeather.Main.Pressure),
					UVIndex:         currentUV,
					Condition:       currentWeather.Weather[0].Description,
					UVCategory:      getUVCategory(currentUV),
					UVColor:         getUVColor(currentUV),
					WindSpeed:       currentWeather.Wind.Speed,
					WindDirection:   windDirection,
					WindGust:        currentWeather.Wind.Gust,
					Visibility:      currentWeather.Visibility,
					Sunrise:         sunriseTime,
					Sunset:          sunsetTime,
					AirQualityIndex: currentAQI,
					AQICategory:     aqiCategory,
				},
				BackgroundURL: backgroundURL,
			}

			// Add to response
			mutex.Lock()
			response.Cities = append(response.Cities, cityWeather)
			mutex.Unlock()
		}(city)
	}

	// Wait for all requests to complete
	wg.Wait()

	// Check if we have any cities in the response
	if len(response.Cities) == 0 && len(response.Errors) > 0 {
		// All cities failed, show a general error
		response.Error = "Failed to fetch weather data for all cities"
	}

	// Render the comparison template
	c.HTML(http.StatusOK, "compare.html", response)
}

// weatherAPIHandler API endpoint for weather data
func weatherAPIHandler(c *gin.Context) {
	// Check if location coordinates are provided
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	city := c.DefaultQuery("city", "")

	var weatherData *WeatherData
	var err error

	// Handle request based on parameters
	if latStr != "" && lonStr != "" {
		// Convert latitude and longitude to float
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude value"})
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude value"})
			return
		}

		// Get weather data by coordinates
		weatherData, err = getWeatherByCoordinates(lat, lon)
	} else if city != "" {
		// Get weather data by city name
		weatherData, err = getCurrentWeather(city)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either city or coordinates required"})
		return
	}

	// Handle errors from weather data retrieval
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use premium OneCall API
	oneCallData, _ := getOneCallData(weatherData.Coord.Lat, weatherData.Coord.Lon)

	// Get air quality data
	aqData, _ := getAirQualityData(weatherData.Coord.Lat, weatherData.Coord.Lon)

	// Extract data from OneCall API (premium features)
	var dailyForecasts []DailyForecast
	var alerts []WeatherAlertItem
	var hourlyForecastItems []map[string]interface{}

	if oneCallData != nil {
		// Process daily forecast
		for i, daily := range oneCallData.Daily {
			if i >= 16 { // Limit to 16 days
				break
			}

			date := time.Unix(daily.Dt, 0)
			dateStr := date.Format("2006-01-02")
			displayDate := date.Format("Jan 2")
			dayName := date.Format("Mon")

			// Get the weather description
			description := "Unknown"
			if len(daily.Weather) > 0 {
				description = daily.Weather[0].Description
			}

			// Get wind direction as string
			windDirection := getWindDirection(daily.WindDeg)

			// Format sunrise and sunset times
			sunriseTime := time.Unix(daily.Sunrise, 0).Format("15:04")
			sunsetTime := time.Unix(daily.Sunset, 0).Format("15:04")

			// Get moon phase description
			moonPhase := getMoonPhaseDescription(daily.MoonPhase)

			dailyForecasts = append(dailyForecasts, DailyForecast{
				Index:         i + 1,
				Date:          displayDate,
				FullDate:      dateStr,
				Day:           dayName,
				MaxTemp:       daily.Temp.Max,
				MinTemp:       daily.Temp.Min,
				Humidity:      daily.Humidity,
				Pressure:      daily.Pressure,
				UVIndex:       daily.Uvi,
				Description:   description,
				UVCategory:    getUVCategory(daily.Uvi),
				UVColor:       getUVColor(daily.Uvi),
				WindSpeed:     daily.WindSpeed,
				WindDirection: windDirection,
				PrecipProb:    daily.Pop * 100, // Convert to percentage
				Sunrise:       sunriseTime,
				Sunset:        sunsetTime,
				MoonPhase:     moonPhase,
			})
		}

		// Process alerts
		for _, alert := range oneCallData.Alerts {
			alertTime := time.Unix(alert.Start, 0).Format("15:04")
			severity := getSeverityFromAlertEvent(alert.Event)

			alerts = append(alerts, WeatherAlertItem{
				Title:       alert.Event,
				Description: alert.Description,
				Time:        alertTime,
				Severity:    severity,
			})
		}

		// Process hourly forecast
		for i := 0; i < 24 && i < len(oneCallData.Hourly); i += 3 {
			hourData := oneCallData.Hourly[i]
			hourTime := time.Unix(hourData.Dt, 0).Format("15:04")

			// Get icon based on weather condition
			iconClass := "fa-cloud" // Default
			if len(hourData.Weather) > 0 {
				iconClass = getWeatherIconClass(hourData.Weather[0].Description)
			}

			hourlyForecastItems = append(hourlyForecastItems, map[string]interface{}{
				"time":       hourTime,
				"temp":       hourData.Temp,
				"icon":       iconClass,
				"precipProb": hourData.Pop * 100, // Convert to percentage
			})
		}
	} else {
		// If OneCall API failed, fallback to regular forecast endpoint
		regularForecastData, err := getForecastData(weatherData.Coord.Lat, weatherData.Coord.Lon)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Process the forecast data to extract daily forecasts
		dailyForecasts = processForecasts(regularForecastData.List)

		// Process hourly forecast items from regular forecast data
		for i := 0; i < 6 && i < len(regularForecastData.List); i++ {
			item := regularForecastData.List[i]
			dateTime, _ := time.Parse("2006-01-02 15:04:05", item.DtTxt)
			timeStr := dateTime.Format("15:04")

			var description string
			if len(item.Weather) > 0 {
				description = item.Weather[0].Description
			} else {
				description = "clear"
			}
			iconClass := getWeatherIconClass(description)

			hourlyForecastItems = append(hourlyForecastItems, map[string]interface{}{
				"time":       timeStr,
				"temp":       item.Main.Temp,
				"icon":       iconClass,
				"precipProb": item.Pop * 100, // Convert to percentage
			})
		}
	}
	// Sort forecasts by date
	sort.Slice(dailyForecasts, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", dailyForecasts[i].FullDate)
		dateJ, _ := time.Parse("2006-01-02", dailyForecasts[j].FullDate)
		return dateI.Before(dateJ)
	})

	// Get current UV index
	currentUV := 5.0 // Default medium value
	if oneCallData != nil {
		currentUV = oneCallData.Current.Uvi
	}

	// Get current AQI
	currentAQI := 1       // Default value
	aqiCategory := "Good" // Default
	if aqData != nil && len(aqData.List) > 0 {
		currentAQI = aqData.List[0].Main.Aqi
		aqiCategory = getAQICategory(currentAQI)
	}

	// Format alerts for API response
	var alertsFormatted []map[string]interface{}
	for _, alert := range alerts {
		alertsFormatted = append(alertsFormatted, map[string]interface{}{
			"title":       alert.Title,
			"description": alert.Description,
			"time":        alert.Time,
			"severity":    alert.Severity,
		})
	}

	// Format forecast for API response
	var forecastFormatted []map[string]interface{}
	for _, forecast := range dailyForecasts {
		forecastFormatted = append(forecastFormatted, map[string]interface{}{
			"index":           forecast.Index,
			"date":            forecast.Date,
			"day":             forecast.Day,
			"maxTemp":         forecast.MaxTemp,
			"minTemp":         forecast.MinTemp,
			"humidity":        forecast.Humidity,
			"pressure":        forecast.Pressure,
			"uvIndex":         forecast.UVIndex,
			"description":     forecast.Description,
			"windSpeed":       forecast.WindSpeed,
			"windDirection":   forecast.WindDirection,
			"precipProb":      forecast.PrecipProb,
			"sunrise":         forecast.Sunrise,
			"sunset":          forecast.Sunset,
			"moonPhase":       forecast.MoonPhase,
			"airQualityIndex": forecast.AirQualityIndex,
		})
	}

	// Prepare the response
	response := gin.H{
		"city": weatherData.Name,
		"current": gin.H{
			"temperature":     weatherData.Main.Temp,
			"feelsLike":       weatherData.Main.FeelsLike,
			"humidity":        weatherData.Main.Humidity,
			"pressure":        weatherData.Main.Pressure,
			"uvIndex":         currentUV,
			"condition":       weatherData.Weather[0].Description,
			"windSpeed":       weatherData.Wind.Speed,
			"windDirection":   getWindDirection(int(weatherData.Wind.Deg)),
			"windGust":        weatherData.Wind.Gust,
			"visibility":      weatherData.Visibility,
			"sunrise":         time.Unix(weatherData.Sys.Sunrise, 0).Format("15:04"),
			"sunset":          time.Unix(weatherData.Sys.Sunset, 0).Format("15:04"),
			"airQualityIndex": currentAQI,
			"aqiCategory":     aqiCategory,
		},
		"forecast":       forecastFormatted,
		"hourlyForecast": hourlyForecastItems,
		"alerts":         alertsFormatted,
		"isPremium":      true,
	}

	c.JSON(http.StatusOK, response)
}

// compareAPIHandler handles API requests for comparing weather across multiple cities
func compareAPIHandler(c *gin.Context) {
	// Get cities from query parameters
	cities := c.QueryArray("cities")

	if len(cities) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one city is required"})
		return
	}

	// Create a response structure
	type cityResponse struct {
		Name     string                 `json:"name"`
		Current  map[string]interface{} `json:"current"`
		Forecast []DailyForecast        `json:"forecast,omitempty"`
		Error    string                 `json:"error,omitempty"`
	}

	response := make([]cityResponse, 0, len(cities))
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Fetch data for each city concurrently
	for _, city := range cities {
		wg.Add(1)
		go func(cityName string) {
			defer wg.Done()

			cityData := cityResponse{
				Name:    cityName,
				Current: make(map[string]interface{}),
			}

			// Get current weather data
			currentWeather, err := getCurrentWeather(cityName)
			if err != nil {
				cityData.Error = err.Error()
				mutex.Lock()
				response = append(response, cityData)
				mutex.Unlock()
				return
			}

			// Get premium data with OneCall API
			oneCallData, _ := getOneCallData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)

			// Get air quality data
			aqData, _ := getAirQualityData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)

			// Get current UV index
			currentUV := 5.0 // Default medium value
			if oneCallData != nil {
				currentUV = oneCallData.Current.Uvi
			}

			// Get current AQI
			currentAQI := 1       // Default good value
			aqiCategory := "Good" // Default
			if aqData != nil && len(aqData.List) > 0 {
				currentAQI = aqData.List[0].Main.Aqi
				aqiCategory = getAQICategory(currentAQI)
			}

			// Set current weather data
			cityData.Name = currentWeather.Name // Use the official name returned by API
			cityData.Current = map[string]interface{}{
				"temperature":     currentWeather.Main.Temp,
				"feelsLike":       currentWeather.Main.FeelsLike,
				"humidity":        currentWeather.Main.Humidity,
				"pressure":        currentWeather.Main.Pressure,
				"uvIndex":         currentUV,
				"condition":       currentWeather.Weather[0].Description,
				"windSpeed":       currentWeather.Wind.Speed,
				"windDirection":   getWindDirection(int(currentWeather.Wind.Deg)),
				"windGust":        currentWeather.Wind.Gust,
				"visibility":      currentWeather.Visibility,
				"sunrise":         time.Unix(currentWeather.Sys.Sunrise, 0).Format("15:04"),
				"sunset":          time.Unix(currentWeather.Sys.Sunset, 0).Format("15:04"),
				"airQualityIndex": currentAQI,
				"aqiCategory":     aqiCategory,
			}

			// Add to response
			mutex.Lock()
			response = append(response, cityData)
			mutex.Unlock()
		}(city)
	}

	// Wait for all requests to complete
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"cities":    response,
		"isPremium": true,
	})
}

// getHistoricalData fetches historical temperature data from a weather API
func getHistoricalData(city string) (*HistoricalData, error) {
	// In a real implementation, you would fetch this from a weather API that provides historical data
	// For example, OpenWeatherMap's History API, Visual Crossing, or Weatherbit

	// For demonstration, we'll use mock data
	// In a real implementation, you would:
	// 1. Get current date
	// 2. Get weather data for same date last year
	// 3. Get weather data for same date in previous 5 years
	// 4. Calculate averages

	// Create some variation based on city name
	cityHash := 0
	for _, char := range city {
		cityHash += int(char)
	}

	// Create some variation based on city name
	variation := float64(cityHash%100) / 10.0

	return &HistoricalData{
		LastYearTemp:    20.0 + variation,
		FiveYearAvgTemp: 19.5 + variation/2,
	}, nil
}

// generateTrendDescription creates a description of the temperature trend
func generateTrendDescription(tempDiff, currentTemp, lastYearTemp, fiveYearAvg float64) string {
	// Get current season
	month := time.Now().Month()
	var season string
	switch {
	case month >= 3 && month <= 5:
		season = "spring"
	case month >= 6 && month <= 8:
		season = "summer"
	case month >= 9 && month <= 11:
		season = "autumn"
	default:
		season = "winter"
	}

	// Round to one decimal place for display
	tempDiffRounded := math.Round(tempDiff*10) / 10

	// Generate description based on the difference
	var description string
	if math.Abs(tempDiff) < 0.5 {
		description = fmt.Sprintf("This %s is about the same temperature as last year.", season)
	} else if tempDiff > 0 {
		description = fmt.Sprintf("This %s is %.1f°C warmer than last year.", season, tempDiffRounded)
	} else {
		description = fmt.Sprintf("This %s is %.1f°C colder than last year.", season, math.Abs(tempDiffRounded))
	}

	// Add comparison to 5-year average
	fiveYearDiff := currentTemp - fiveYearAvg
	fiveYearDiffRounded := math.Round(fiveYearDiff*10) / 10

	if math.Abs(fiveYearDiff) < 0.5 {
		description += " Temperature is close to the 5-year average."
	} else if fiveYearDiff > 0 {
		description += fmt.Sprintf(" It's %.1f°C above the 5-year average.", fiveYearDiffRounded)
	} else {
		description += fmt.Sprintf(" It's %.1f°C below the 5-year average.", math.Abs(fiveYearDiffRounded))
	}

	return description
}

// generateSeasonalInsight creates a seasonal insight based on the location and temperatures
func generateSeasonalInsight(city string, currentTemp, avgTemp float64) string {
	// Get current season
	month := time.Now().Month()
	var season string
	switch {
	case month >= 3 && month <= 5:
		season = "Spring"
	case month >= 6 && month <= 8:
		season = "Summer"
	case month >= 9 && month <= 11:
		season = "Fall"
	default:
		season = "Winter"
	}

	// Generate insight based on the season and temperature
	diff := currentTemp - avgTemp
	if math.Abs(diff) < 1.0 {
		return fmt.Sprintf("%s in %s is following typical patterns this year.", season, city)
	} else if diff > 3.0 {
		return fmt.Sprintf("This is an unusually warm %s for %s compared to historical records.", season, city)
	} else if diff > 0 {
		return fmt.Sprintf("%s in %s is trending slightly warmer than the historical average.", season, city)
	} else if diff < -3.0 {
		return fmt.Sprintf("This is an unusually cool %s for %s compared to historical records.", season, city)
	} else {
		return fmt.Sprintf("%s in %s is trending slightly cooler than the historical average.", season, city)
	}
}

// historicalComparisonHandler handles requests for comparing current weather with historical data
func historicalComparisonHandler(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City name is required"})
		return
	}

	// Get current weather
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get historical data
	historicalData, err := getHistoricalData(city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate temperature differences and trends
	currentTemp := currentWeather.Main.Temp
	lastYearTemp := historicalData.LastYearTemp
	fiveYearAvgTemp := historicalData.FiveYearAvgTemp
	tempDifference := currentTemp - lastYearTemp

	// Generate trend description
	trendDescription := generateTrendDescription(tempDifference, currentTemp, lastYearTemp, fiveYearAvgTemp)

	// Generate seasonal insight
	seasonalInsight := generateSeasonalInsight(city, currentTemp, fiveYearAvgTemp)

	// Prepare response
	response := HistoricalComparisonResponse{
		City:             currentWeather.Name,
		CurrentTemp:      currentTemp,
		LastYearTemp:     lastYearTemp,
		FiveYearAvgTemp:  fiveYearAvgTemp,
		TempDifference:   tempDifference,
		TrendDescription: trendDescription,
		SeasonalInsight:  seasonalInsight,
	}

	c.JSON(http.StatusOK, response)
}

// getHistoricalDataByMonths fetches historical weather data for multiple months
func getHistoricalDataByMonths(city string, date time.Time, months []string) (*HistoricalWeatherData, error) {
	result := &HistoricalWeatherData{
		Previous: make(map[string]struct {
			Temperature    float64 `json:"temperature"`
			Humidity       int     `json:"humidity"`
			Pressure       int     `json:"pressure"`
			Condition      string  `json:"condition"`
			IconURL        string  `json:"iconURL"`
			WindSpeed      float64 `json:"windSpeed"`
			TempDifference float64 `json:"tempDifference"`
			Icon           string  `json:"icon"`
		}),
	}

	// Get current weather (for the selected date)
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		return nil, err
	}

	// Set current weather data
	result.Current.Temperature = currentWeather.Main.Temp
	result.Current.Humidity = currentWeather.Main.Humidity
	result.Current.Pressure = currentWeather.Main.Pressure
	if len(currentWeather.Weather) > 0 {
		result.Current.Condition = currentWeather.Weather[0].Description
		// Set icon URL based on the weather condition
		result.Current.IconURL = getWeatherIconURL(currentWeather.Weather[0].Description)
		// Set icon directly
		result.Current.Icon = currentWeather.Weather[0].Icon
	}
	result.Current.WindSpeed = currentWeather.Wind.Speed

	// For each month requested, get historical data
	for _, monthStr := range months {
		monthInt, err := strconv.Atoi(monthStr)
		if err != nil {
			continue
		}

		// Get historical data for this month
		historicalData, err := getHistoricalDataForMonth(city, date, monthInt)
		if err != nil {
			log.Printf("Error getting historical data for %d months ago: %v", monthInt, err)
			continue
		}

		// Add to the result
		tempDiff := result.Current.Temperature - historicalData.Temperature

		result.Previous[monthStr] = struct {
			Temperature    float64 `json:"temperature"`
			Humidity       int     `json:"humidity"`
			Pressure       int     `json:"pressure"`
			Condition      string  `json:"condition"`
			IconURL        string  `json:"iconURL"`
			WindSpeed      float64 `json:"windSpeed"`
			TempDifference float64 `json:"tempDifference"`
			Icon           string  `json:"icon"`
		}{
			Temperature:    historicalData.Temperature,
			Humidity:       historicalData.Humidity,
			Pressure:       historicalData.Pressure,
			Condition:      historicalData.Condition,
			IconURL:        historicalData.IconURL,
			WindSpeed:      historicalData.WindSpeed,
			TempDifference: tempDiff,
			Icon:           "", // This will be set later
		}
	}

	return result, nil
}

// historicalComparisonPageHandler handles the historical comparison page
func historicalComparisonPageHandler(c *gin.Context) {
	// Check if user is logged in
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		userStore := c.MustGet("user_store").(models.UserStore)
		var err error
		user, err = userStore.GetUserByID(userID.(int))
		if err != nil {
			// Continue without user info
			log.Printf("Error getting user: %v", err)
		}
	}

	// Get query parameters
	city := c.Query("city")
	dateStr := c.Query("date")
	months := c.QueryArray("months")
	units := c.DefaultQuery("units", "metric")

	// If no query parameters, just render the form
	if city == "" {
		c.HTML(http.StatusOK, "historical-comparison.html", gin.H{
			"title":     "Historical Weather Comparison",
			"User":      user,
			"IsPremium": true, // Flag to show premium features in the UI
			"units":     units,
		})
		return
	}

	// Try to parse the date
	var selectedDate time.Time
	var err error
	if dateStr != "" {
		selectedDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Default to today if date is invalid
			selectedDate = time.Now()
		}
	} else {
		// Default to today if no date provided
		selectedDate = time.Now()
	}

	// Get current weather data first
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		c.HTML(http.StatusOK, "historical-comparison.html", gin.H{
			"title":     "Historical Weather Comparison",
			"User":      user,
			"IsPremium": true,
			"units":     units,
			"error":     "Failed to fetch current weather data: " + err.Error(),
		})
		return
	}

	// Ensure we have at least one month selected
	if len(months) == 0 {
		// Default to 1, 3, and 6 months if none selected - was only 1 month previously
		months = []string{"1", "3", "6"}
	}

	// Log the request parameters
	log.Printf("Historical comparison request for city: %s, date: %s, months: %v, units: %s",
		city, dateStr, months, units)

	// Fetch historical data
	historicalData, err := getHistoricalDataByMonths(city, selectedDate, months)
	if err != nil {
		c.HTML(http.StatusOK, "historical-comparison.html", gin.H{
			"title":     "Historical Weather Comparison",
			"User":      user,
			"IsPremium": true,
			"units":     units,
			"error":     "Failed to fetch historical data: " + err.Error(),
		})
		return
	}

	// Debug historical data
	debugHistoricalData(historicalData)

	// Get trend description
	lastMonthTemp := 0.0
	if data, exists := historicalData.Previous["1"]; exists {
		lastMonthTemp = data.Temperature
	} else {
		// Default if not available
		lastMonthTemp = currentWeather.Main.Temp - 1.0
	}

	tempDifference := currentWeather.Main.Temp - lastMonthTemp
	trendDescription := generateMonthlyTrendDescription(tempDifference, currentWeather.Main.Temp, lastMonthTemp)

	// Set icon for current weather
	if len(currentWeather.Weather) > 0 {
		historicalData.Current.Icon = currentWeather.Weather[0].Icon
	} else {
		historicalData.Current.Icon = "01d" // Default clear sky
	}

	// Add icons to previous months' data
	for month := range historicalData.Previous {
		data := historicalData.Previous[month]
		// Extract icon from the URL or use default
		icon := "01d" // Default clear sky
		if data.IconURL != "" {
			icon = strings.TrimSuffix(strings.TrimPrefix(data.IconURL, "https://openweathermap.org/img/wn/"), "@2x.png")
		}
		data.Icon = icon
		historicalData.Previous[month] = data
	}

	// Render the template with the data
	c.HTML(http.StatusOK, "historical-comparison.html", gin.H{
		"title":            "Historical Weather Comparison",
		"User":             user,
		"IsPremium":        true,
		"city":             currentWeather.Name,
		"selectedDate":     selectedDate.Format("January 2, 2006"),
		"historicalData":   historicalData,
		"trendDescription": trendDescription,
		"units":            units,
		"debug":            true, // Add debug flag to help troubleshoot
	})
}

// getHistoricalDataForMonth gets historical weather data for a specific month in the past
// getHistoricalDataForMonth gets historical weather data for a specific month in the past
func getHistoricalDataForMonth(city string, date time.Time, monthsAgo int) (struct {
	Temperature float64
	Humidity    int
	Pressure    int
	Condition   string
	IconURL     string
	WindSpeed   float64
}, error) {
	// Calculate the date for the given months ago
	historicalDate := date.AddDate(0, -monthsAgo, 0)

	// Log the historical date we're querying
	log.Printf("Fetching historical data for %s on %s (%d months ago)",
		city, historicalDate.Format("2006-01-02"), monthsAgo)

	// Get coordinates for the city first
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		log.Printf("Error getting coordinates for historical data: %v", err)
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}

	// Get timestamp for historical date (Unix timestamp in seconds)
	timestamp := historicalDate.Unix()

	// Try to get real historical data from OpenWeather API - updated to correct URL format
	historicalURL := fmt.Sprintf(
		"https://api.openweathermap.org/data/3.0/onecall/timemachine?lat=%f&lon=%f&dt=%d&appid=%s&units=metric",
		currentWeather.Coord.Lat,
		currentWeather.Coord.Lon,
		timestamp,
		apiKey,
	)

	log.Printf("Making historical API request to: %s", historicalURL)

	resp, err := http.Get(historicalURL)
	if err != nil {
		log.Printf("HTTP error fetching historical data: %v", err)
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}
	defer resp.Body.Close()

	log.Printf("Historical API response status: %s", resp.Status)

	// Read the full response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}
	bodyString := string(bodyBytes)
	log.Printf("API Response Body: %s", bodyString)

	// If the API call fails or returns empty data, fall back to synthetic data
	if resp.StatusCode != 200 || len(bodyBytes) < 10 {
		log.Printf("Falling back to synthetic data due to empty response or status code: %d", resp.StatusCode)
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}

	// Parse the API response
	var timeMachineData TimeMachineResponse
	if err := json.Unmarshal(bodyBytes, &timeMachineData); err != nil {
		log.Printf("Error parsing historical API response: %v", err)
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}

	// Ensure we have data points
	if len(timeMachineData.Data) == 0 {
		log.Printf("No historical data points returned by API")
		return generateSyntheticHistoricalData(city, historicalDate, monthsAgo)
	}

	// Use the first data point (should be closest to the requested time)
	dataPoint := timeMachineData.Data[0]

	// Get weather condition and icon (with default values if missing)
	condition := "Unknown"
	iconCode := "01d" // Default clear sky
	if len(dataPoint.Weather) > 0 {
		condition = dataPoint.Weather[0].Description
		iconCode = dataPoint.Weather[0].Icon
	}

	// Create icon URL
	iconURL := fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", iconCode)

	// Return the historical weather data
	log.Printf("Successfully retrieved historical data: Temp=%.1f°C, Cond=%s",
		dataPoint.Temp, condition)

	return struct {
		Temperature float64
		Humidity    int
		Pressure    int
		Condition   string
		IconURL     string
		WindSpeed   float64
	}{
		Temperature: dataPoint.Temp,
		Humidity:    dataPoint.Humidity,
		Pressure:    dataPoint.Pressure,
		Condition:   condition,
		IconURL:     iconURL,
		WindSpeed:   dataPoint.Wind_speed,
	}, nil
}

func debugHistoricalData(historicalData *HistoricalWeatherData) {
	if historicalData == nil {
		log.Printf("DEBUG: Historical data is nil!")
		return
	}

	log.Printf("DEBUG: Current temperature: %.2f°C", historicalData.Current.Temperature)
	log.Printf("DEBUG: Number of previous data points: %d", len(historicalData.Previous))

	// Print details of each historical month
	for month, data := range historicalData.Previous {
		log.Printf("DEBUG: Month %s - Temp: %.2f°C, Diff: %.2f°C",
			month, data.Temperature, data.TempDifference)
	}
}

// generateSyntheticHistoricalData creates synthetic weather data for historical comparisons
// when real data is not available from the API
func generateSyntheticHistoricalData(city string, historicalDate time.Time, monthsAgo int) (struct {
	Temperature float64
	Humidity    int
	Pressure    int
	Condition   string
	IconURL     string
	WindSpeed   float64
}, error) {
	log.Printf("Generating synthetic historical data for %s on %s",
		city, historicalDate.Format("2006-01-02"))

	// Create some variation based on city name and date
	cityHash := 0
	for _, char := range city {
		cityHash += int(char)
	}

	dayOfYear := historicalDate.YearDay()
	dateHash := (dayOfYear * 10) % 100

	// Create some temperature variation (18-28°C range with city and date influence)
	baseTemp := 23.0
	cityVariation := float64(cityHash%100)/10.0 - 5.0
	dateVariation := float64(dateHash)/10.0 - 5.0
	monthVariation := float64(monthsAgo) * 0.5 // Slight variation for different months

	// Add seasonal variation - higher temperatures in summer, lower in winter
	month := historicalDate.Month()
	var seasonalVariation float64
	if month >= 5 && month <= 8 { // Summer months (Northern Hemisphere)
		seasonalVariation = 5.0
	} else if month == 3 || month == 4 || month == 9 || month == 10 { // Spring/Fall
		seasonalVariation = 2.0
	} else { // Winter
		seasonalVariation = -3.0
	}

	temperature := baseTemp + cityVariation + dateVariation + seasonalVariation - monthVariation

	// Round to one decimal place
	temperature = math.Round(temperature*10) / 10

	// Generate other weather properties
	humidity := 50 + (cityHash % 30)
	pressure := 1000 + (cityHash % 30)
	windSpeed := 5.0 + float64(cityHash%10)

	// Select a weather condition based on temperature
	var condition string
	var iconCode string
	if temperature > 25 {
		condition = "Clear sky"
		iconCode = "01d"
	} else if temperature > 20 {
		condition = "Few clouds"
		iconCode = "02d"
	} else if temperature > 15 {
		condition = "Scattered clouds"
		iconCode = "03d"
	} else if temperature > 10 {
		condition = "Overcast clouds"
		iconCode = "04d"
	} else {
		condition = "Light rain"
		iconCode = "10d"
	}

	// Create icon URL
	iconURL := fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", iconCode)

	log.Printf("Successfully generated synthetic data: Temp=%.1f°C, Cond=%s",
		temperature, condition)

	// Return the synthetic historical data
	return struct {
		Temperature float64
		Humidity    int
		Pressure    int
		Condition   string
		IconURL     string
		WindSpeed   float64
	}{
		Temperature: temperature,
		Humidity:    humidity,
		Pressure:    pressure,
		Condition:   condition,
		IconURL:     iconURL,
		WindSpeed:   windSpeed,
	}, nil
}

// generateMonthlyTrendDescription creates a description of the temperature trend based on months
func generateMonthlyTrendDescription(tempDiff, currentTemp, lastMonthTemp float64) string {
	// Get current season
	month := time.Now().Month()
	var season string
	switch {
	case month >= 3 && month <= 5:
		season = "spring"
	case month >= 6 && month <= 8:
		season = "summer"
	case month >= 9 && month <= 11:
		season = "autumn"
	default:
		season = "winter"
	}

	// Round to one decimal place for display
	tempDiffRounded := math.Round(tempDiff*10) / 10

	// Generate description based on the difference
	var description string
	if math.Abs(tempDiff) < 0.5 {
		description = fmt.Sprintf("This %s is about the same temperature as last month.", season)
	} else if tempDiff > 0 {
		description = fmt.Sprintf("This %s is %.1f°C warmer than last month.", season, tempDiffRounded)
	} else {
		description = fmt.Sprintf("This %s is %.1f°C colder than last month.", season, math.Abs(tempDiffRounded))
	}

	return description
}

// Helper function to add to your template.FuncMap
func setupTemplateHelpers(router *gin.Engine) {
	// Define custom template functions
	router.SetFuncMap(template.FuncMap{
		"abs": math.Abs,
		"mul": func(a, b float64) float64 {
			return a * b
		},
		// Add new helper functions for profile circle
		"toUpper": strings.ToUpper,
		"slice": func(s string, i, j int) string {
			if i >= len(s) {
				return ""
			}
			if j > len(s) {
				j = len(s)
			}
			return s[i:j]
		},
	})
}

// getWeatherIconURL returns an icon URL based on weather condition
func getWeatherIconURL(condition string) string {
	condition = strings.ToLower(condition)

	if strings.Contains(condition, "clear") || strings.Contains(condition, "sky") {
		return "https://openweathermap.org/img/wn/01d@2x.png"
	} else if strings.Contains(condition, "few clouds") {
		return "https://openweathermap.org/img/wn/02d@2x.png"
	} else if strings.Contains(condition, "scattered clouds") {
		return "https://openweathermap.org/img/wn/03d@2x.png"
	} else if strings.Contains(condition, "clouds") {
		return "https://openweathermap.org/img/wn/04d@2x.png"
	} else if strings.Contains(condition, "rain") && strings.Contains(condition, "light") {
		return "https://openweathermap.org/img/wn/10d@2x.png"
	} else if strings.Contains(condition, "rain") {
		return "https://openweathermap.org/img/wn/09d@2x.png"
	} else if strings.Contains(condition, "thunder") {
		return "https://openweathermap.org/img/wn/11d@2x.png"
	} else if strings.Contains(condition, "snow") {
		return "https://openweathermap.org/img/wn/13d@2x.png"
	} else if strings.Contains(condition, "mist") || strings.Contains(condition, "fog") {
		return "https://openweathermap.org/img/wn/50d@2x.png"
	}

	// Default icon
	return "https://openweathermap.org/img/wn/01d@2x.png"
}

// historicalComparisonAPIHandler with more detailed response
func historicalComparisonAPIHandler(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City name is required"})
		return
	}
	// Acknowledge the variables but don't use them if they cause issues
	dateStr := c.Query("date")
	months := c.QueryArray("months")

	var selectedDate time.Time
	var err error
	if dateStr != "" {
		selectedDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Default to today if date is invalid
			selectedDate = time.Now()
		}
	} else {
		// Default to today if no date provided
		selectedDate = time.Now()
	}

	// Get current weather
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Default to 1 month if none specified
	if len(months) == 0 {
		months = []string{"1", "12"} // 1 month and 1 year ago
	}

	// Get historical data by months
	historicalData, err := getHistoricalDataByMonths(city, selectedDate, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format the response
	type HistoricalDataResponse struct {
		City             string                   `json:"city"`
		SelectedDate     string                   `json:"selectedDate"`
		CurrentWeather   map[string]interface{}   `json:"currentWeather"`
		HistoricalMonths []map[string]interface{} `json:"historicalMonths"`
		TrendDescription string                   `json:"trendDescription"`
	}

	// Create current weather section
	currentWeatherData := map[string]interface{}{
		"temperature": historicalData.Current.Temperature,
		"humidity":    historicalData.Current.Humidity,
		"pressure":    historicalData.Current.Pressure,
		"condition":   historicalData.Current.Condition,
		"windSpeed":   historicalData.Current.WindSpeed,
		"icon":        historicalData.Current.Icon,
	}

	// Create historical months section
	historicalMonths := make([]map[string]interface{}, 0, len(historicalData.Previous))
	for month, data := range historicalData.Previous {
		monthInt, _ := strconv.Atoi(month)

		var label string
		if monthInt == 1 {
			label = "1 month ago"
		} else if monthInt == 12 {
			label = "1 year ago"
		} else if monthInt == 24 {
			label = "2 years ago"
		} else if monthInt == 36 {
			label = "3 years ago"
		} else if monthInt == 60 {
			label = "5 years ago"
		} else {
			label = fmt.Sprintf("%d months ago", monthInt)
		}

		historicalMonths = append(historicalMonths, map[string]interface{}{
			"label":          label,
			"months":         month,
			"temperature":    data.Temperature,
			"humidity":       data.Humidity,
			"pressure":       data.Pressure,
			"condition":      data.Condition,
			"windSpeed":      data.WindSpeed,
			"tempDifference": data.TempDifference,
			"icon":           data.Icon,
		})
	}

	// Sort historical months by number of months ago
	sort.Slice(historicalMonths, func(i, j int) bool {
		monthsI, _ := strconv.Atoi(historicalMonths[i]["months"].(string))
		monthsJ, _ := strconv.Atoi(historicalMonths[j]["months"].(string))
		return monthsI < monthsJ
	})

	// Generate trend description
	lastMonthTemp := 0.0
	if data, exists := historicalData.Previous["1"]; exists {
		lastMonthTemp = data.Temperature
	} else {
		// Default if not available
		lastMonthTemp = currentWeather.Main.Temp - 1.0
	}
	tempDifference := currentWeather.Main.Temp - lastMonthTemp
	trendDescription := generateMonthlyTrendDescription(tempDifference, currentWeather.Main.Temp, lastMonthTemp)

	response := HistoricalDataResponse{
		City:             currentWeather.Name,
		SelectedDate:     selectedDate.Format("January 2, 2006"),
		CurrentWeather:   currentWeatherData,
		HistoricalMonths: historicalMonths,
		TrendDescription: trendDescription,
	}

	c.JSON(http.StatusOK, response)
}

// sendDailyReportsToAllUsers sends daily weather reports to all users who have opted in for all updates
func sendDailyReportsToAllUsers(userStore models.UserStore) {
	// Get all users with notifications enabled and alert threshold set to "all"
	users, err := userStore.GetUsersWithDailyReports()
	if err != nil {
		log.Printf("Error fetching users for daily reports: %v", err)
		return
	}

	log.Printf("Found %d users eligible for daily weather reports", len(users))

	// Initialize notification service if needed
	initializeNotificationService()

	// Check if notification service is initialized properly
	if notificationService == nil {
		log.Println("Error: Notification service is not initialized. Daily reports will not be sent.")
		return
	}

	// Process each user
	for _, user := range users {
		// Skip users without a home city set
		if user.HomeCity == "" {
			continue
		}

		// Get weather data for the user's home city
		go sendDailyReportForUser(user, notificationService)
	}
}

// sendDailyReportForUser fetches weather data and sends report for a specific user
func sendDailyReportForUser(user *models.User, notificationService *services.NotificationService) {
	if notificationService == nil {
		log.Printf("Error: Cannot send daily report to %s - notification service is nil", user.Email)
		return
	}

	// Get current weather for the user's home city
	currentWeather, err := getCurrentWeather(user.HomeCity)
	if err != nil {
		log.Printf("Error getting weather data for %s: %v", user.HomeCity, err)
		return
	}

	// Use OneCall API for premium features
	oneCallData, err := getOneCallData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)

	var dailyForecasts []interface{}

	if err == nil && oneCallData != nil {
		// Use premium forecast data
		for i, daily := range oneCallData.Daily {
			if i >= 7 { // Limit to 7 days for email
				break
			}

			// Format the date
			date := time.Unix(daily.Dt, 0).Format("Jan 2")
			dayName := time.Unix(daily.Dt, 0).Format("Mon")

			// Get description
			var description string
			if len(daily.Weather) > 0 {
				description = daily.Weather[0].Description
			} else {
				description = "Unknown"
			}

			// Add to forecast list
			dailyForecasts = append(dailyForecasts, map[string]interface{}{
				"day":         dayName,
				"date":        date,
				"maxTemp":     daily.Temp.Max,
				"minTemp":     daily.Temp.Min,
				"description": description,
				"precipProb":  daily.Pop * 100, // Convert to percentage
				"windSpeed":   daily.WindSpeed,
				"uvIndex":     daily.Uvi,
			})
		}
	} else {
		// Fallback to regular forecast data
		log.Printf("Falling back to regular forecast data for %s", user.HomeCity)
		forecastData, err := getForecastData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
		if err != nil {
			log.Printf("Error getting forecast data for %s: %v", user.HomeCity, err)
			return
		}

		// Process the forecast data
		processedForecasts := processForecasts(forecastData.List)

		// Sort forecasts by date
		sort.Slice(processedForecasts, func(i, j int) bool {
			dateI, _ := time.Parse("2006-01-02", processedForecasts[i].FullDate)
			dateJ, _ := time.Parse("2006-01-02", processedForecasts[j].FullDate)
			return dateI.Before(dateJ)
		})

		// Convert forecasts to interface{} slice for the notification service
		for i, forecast := range processedForecasts {
			if i >= 7 { // Limit to 7 days for email
				break
			}

			dailyForecasts = append(dailyForecasts, map[string]interface{}{
				"day":         forecast.Day,
				"date":        forecast.Date,
				"maxTemp":     forecast.MaxTemp,
				"minTemp":     forecast.MinTemp,
				"description": forecast.Description,
				"precipProb":  forecast.PrecipProb,
				"windSpeed":   forecast.WindSpeed,
				"uvIndex":     forecast.UVIndex,
			})
		}
	}

	// Send the daily report with premium data
	err = notificationService.SendDailyReport(
		user.Email,
		currentWeather.Name,
		currentWeather.Main.Temp,
		currentWeather.Weather[0].Description,
		dailyForecasts,
	)

	if err != nil {
		log.Printf("Error sending daily report to %s: %v", user.Email, err)
	} else {
		log.Printf("Successfully sent premium daily weather report to %s for city %s", user.Email, user.HomeCity)
	}
}

// weatherImpactAPIHandler is specifically for the Weather Impact feature
func weatherImpactAPIHandler(c *gin.Context) {
	city := c.DefaultQuery("city", "")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City name is required"})
		return
	}

	log.Printf("Processing Weather Impact API request for city: %s", city)

	// Get current weather data
	weatherData, err := getCurrentWeather(city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get UV data using coordinates from current weather
	oneCallData, _ := getOneCallData(weatherData.Coord.Lat, weatherData.Coord.Lon)

	// Default values
	currentUV := 5.0 // Default medium value

	// Get current UV index if available
	if oneCallData != nil {
		currentUV = oneCallData.Current.Uvi
	}

	// Prepare the response with fields formatted for the Weather Impact feature
	response := gin.H{
		"city": weatherData.Name,
		"current": gin.H{
			"temperature": fmt.Sprintf("%.1f°C", weatherData.Main.Temp),
			"humidity":    fmt.Sprintf("%d%%", weatherData.Main.Humidity),
			"pressure":    fmt.Sprintf("%d hPa", weatherData.Main.Pressure),
			"uvIndex":     currentUV,
			"condition":   weatherData.Weather[0].Description,
			"wind":        fmt.Sprintf("%.1f m/s", weatherData.Wind.Speed),
			"visibility":  fmt.Sprintf("%.1f km", float64(weatherData.Visibility)/1000), // Convert from meters to km
			"icon":        weatherData.Weather[0].Icon,
		},
	}

	c.JSON(http.StatusOK, response)
}

// main function to set up and run the application
func main() {
	// Set API key for services package
	services.SetAPIKey(apiKey)

	// Create router with default middleware
	router := gin.Default()

	// Set up template functions before loading templates
	setupTemplateHelpers(router)

	// Set up the session store with a secure secret
	store := cookie.NewStore([]byte("your-unique-secret-key-change-this"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
	})
	router.Use(sessions.Sessions("weather_session", store))

	// Set the router to use HTML templates

	router.LoadHTMLGlob("templates/*.html")

	// Serve static files
	router.Static("/static", "./static")

	// Enable CORS
	router.Use(cors.Default())

	// Get database connection details from environment variables or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "weather_user")
	dbPassword := getEnv("DB_PASSWORD", "harsha03")
	dbName := getEnv("DB_NAME", "weather_app")

	// Build MySQL DSN string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Initialize MySQL database
	userStore, err := models.NewMySQLStore(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer userStore.Close()

	// After establishing database connection
	log.Printf("Successfully connected to MySQL database")

	// Initialize services for historical comparison
	var weatherService = services.NewWeatherService(apiKey)
	var activityService = services.NewActivityService(apiKey) // Using your existing constructor

	// Create weather handler with services
	weatherHandler := handlers.NewWeatherHandler(userStore, weatherService, activityService)

	// Add user store to the context
	router.Use(func(c *gin.Context) {
		c.Set("user_store", userStore)
		c.Next()
	})
	// Global middleware to check for user_id in session and set it in context
	router.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID != nil {
			log.Printf("Global middleware: User ID %v found in session for path %s", userID, c.Request.URL.Path)
			c.Set("user_id", userID)
		} else {
			log.Printf("Global middleware: No user ID in session for path %s", c.Request.URL.Path)
		}
		c.Next()
	})

	// Create auth handlers
	authHandler := handlers.NewAuthHandler(userStore)
	activityHandler := handlers.NewActivityHandler(apiKey)

	// Public routes
	router.GET("/", homeHandler)
	router.GET("/login.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/login")
	})
	// Add this missing route
	router.GET("/login", middleware.RedirectIfLoggedIn(), authHandler.GetLogin)
	// Add these routes in your main.go near the other API routes section

	// Add this to your main.go file where you set up other routes

	// Replace the problematic HTTP handler registration
	// DELETE these lines:
	// http.HandleFunc("/api/nearby-locations", handlers.HandleNearbyLocations)
	// http.HandleFunc("/location-options.html", handlers.RenderLocationOptions)

	// ADD these lines instead:
	router.GET("/api/nearby-locations", func(c *gin.Context) {
		// Get latitude and longitude from query parameters
		latStr := c.Query("lat")
		lonStr := c.Query("lon")

		// Validate parameters
		if latStr == "" || lonStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lat or lon parameters"})
			return
		}

		// Parse latitude and longitude
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude value"})
			return
		}

		lon, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude value"})
			return
		}

		// In a real application, you would query a database or external API
		// to get nearby locations. Here we'll use your existing function.
		nearbyLocations, err := getNearbyLocations(lat, lon)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create response
		response := struct {
			Locations []struct {
				Name     string  `json:"name"`
				State    string  `json:"state,omitempty"`
				Country  string  `json:"country"`
				Lat      float64 `json:"lat"`
				Lon      float64 `json:"lon"`
				Distance float64 `json:"distance"`
			} `json:"locations"`
		}{
			Locations: nearbyLocations,
		}

		c.JSON(http.StatusOK, response)
	})

	// Route for location options page
	router.GET("/location-options", func(c *gin.Context) {
		// Get URL parameters
		latParam := c.Query("lat")
		lonParam := c.Query("lon")

		log.Printf("Rendering location-options template with lat=%s, lon=%s", latParam, lonParam)

		// Get user if logged in
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

		// Create data for locations if coordinates provided
		var locations []struct {
			Name     string  `json:"name"`
			State    string  `json:"state,omitempty"`
			Country  string  `json:"country"`
			Lat      float64 `json:"lat"`
			Lon      float64 `json:"lon"`
			Distance float64 `json:"distance"`
		}

		if latParam != "" && lonParam != "" {
			lat, errLat := strconv.ParseFloat(latParam, 64)
			lon, errLon := strconv.ParseFloat(lonParam, 64)

			if errLat == nil && errLon == nil {
				// Get nearby locations
				nearbyLocations, err := getNearbyLocations(lat, lon)
				if err == nil {
					locations = nearbyLocations
				} else {
					log.Printf("Error getting nearby locations: %v", err)
				}
			}
		}

		// Prepare data with appropriate attribute names for the template
		c.HTML(http.StatusOK, "location-options.html", gin.H{
			"title":     "Select Your Location",
			"User":      user,
			"locations": locations,
			"lat":       latParam,
			"lon":       lonParam,
		})
	})

	router.GET("/api/geocode", func(c *gin.Context) {
		location := c.Query("location")
		if location == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Location parameter is required"})
			return
		}

		// OpenWeatherMap Geocoding API
		geocodeURL := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/direct?q=%s&limit=5&appid=%s",
			url.QueryEscape(location), apiKey)

		resp, err := http.Get(geocodeURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to geocoding service"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Geocoding service returned an error"})
			return
		}

		var data []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse geocoding response"})
			return
		}

		c.JSON(http.StatusOK, data)
	})

	// Add reverse geocoding endpoint
	router.GET("/api/reverse-geocode", func(c *gin.Context) {
		latStr := c.Query("lat")
		lonStr := c.Query("lon")

		if latStr == "" || lonStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both lat and lon parameters are required"})
			return
		}

		// OpenWeatherMap Reverse Geocoding API
		reverseGeocodeURL := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/reverse?lat=%s&lon=%s&limit=1&appid=%s",
			latStr, lonStr, apiKey)

		resp, err := http.Get(reverseGeocodeURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to reverse geocoding service"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Reverse geocoding service returned an error"})
			return
		}

		var data []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"	error": "Failed to parse reverse geocoding response"})
			return
		}

		if len(data) > 0 {
			c.JSON(http.StatusOK, data[0])
		} else {
			c.JSON(http.StatusOK, gin.H{"name": "Unknown Location", "country": ""})
		}
	})

	router.POST("/login", authHandler.PostLogin)
	router.GET("/signup", middleware.RedirectIfLoggedIn(), authHandler.GetSignup)
	// In main.go
	router.GET("/activities", activityHandler.GetActivitiesPageHandler)
	router.GET("/weather-impact", activityHandler.GetActivitiesPageHandler)

	router.POST("/signup", authHandler.PostSignup)
	router.GET("/logout", authHandler.Logout)

	router.GET("/historical-comparison", historicalComparisonPageHandler)

	// Premium routes for weather alerts
	router.GET("/alerts", middleware.AuthRequired(), func(c *gin.Context) {
		// Get user details
		userID, exists := c.Get("user_id")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		userStore := c.MustGet("user_store").(models.UserStore)
		user, err := userStore.GetUserByID(userID.(int))
		if err != nil {
			log.Printf("Error getting user: %v", err)
			c.Redirect(http.StatusFound, "/logout")
			return
		}

		// Get user's home city for alert data
		city := c.DefaultQuery("city", user.HomeCity)
		if city == "" {
			c.HTML(http.StatusOK, "alerts.html", gin.H{
				"title": "Weather Alerts - GoWeather Premium",
				"User":  user,
				"error": "Please set a home city in your profile or specify a city in the query",
			})
			return
		}

		// Get weather data
		weatherData, err := getCurrentWeather(city)
		if err != nil {
			c.HTML(http.StatusOK, "alerts.html", gin.H{
				"title": "Weather Alerts - GoWeather Premium",
				"User":  user,
				"error": fmt.Sprintf("Error fetching weather data: %v", err),
			})
			return
		}

		// Get OneCall data for alerts
		oneCallData, err := getOneCallData(weatherData.Coord.Lat, weatherData.Coord.Lon)
		if err != nil {
			c.HTML(http.StatusOK, "alerts.html", gin.H{
				"title": "Weather Alerts - GoWeather Premium",
				"User":  user,
				"error": fmt.Sprintf("Error fetching alerts data: %v", err),
			})
			return
		}

		// Format alerts for display
		var formattedAlerts []gin.H
		if oneCallData != nil && len(oneCallData.Alerts) > 0 {
			for _, alert := range oneCallData.Alerts {
				alertTime := time.Unix(alert.Start, 0).Format("Jan 2, 15:04")
				endTime := time.Unix(alert.End, 0).Format("Jan 2, 15:04")
				severity := getSeverityFromAlertEvent(alert.Event)

				formattedAlerts = append(formattedAlerts, gin.H{
					"Title":       alert.Event,
					"Description": alert.Description,
					"StartTime":   alertTime,
					"EndTime":     endTime,
					"Severity":    severity,
					"Source":      alert.SenderName,
				})
			}
		}

		// Render the alerts page
		c.HTML(http.StatusOK, "alerts.html", gin.H{
			"title":    "Weather Alerts - GoWeather Premium",
			"User":     user,
			"City":     weatherData.Name,
			"Alerts":   formattedAlerts,
			"NoAlerts": len(formattedAlerts) == 0,
		})
	})

	// Route for weather search - make it accessible without auth
	router.GET("/weather", weatherHandler.GetWeather)

	// API routes (no auth required)
	router.GET("/api/weather/impact", weatherImpactAPIHandler)

	router.GET("/api/weather", weatherAPIHandler)
	router.GET("/api/compare", compareAPIHandler)
	router.GET("/api/historical-comparison", historicalComparisonAPIHandler)
	router.GET("/api/activities", activityHandler.GetActivitiesHandler)
	router.GET("/api/weather/historical-comparison", func(c *gin.Context) {
		weatherHandler.GetHistoricalComparison(c)
	})

	// Historical comparison page

	// Protected routes
	authGroup := router.Group("/")
	authGroup.Use(middleware.AuthRequired())
	{
		// Add dashboard route
		authGroup.GET("/dashboard", dashboardHandler)

		// These routes require auth
		authGroup.GET("/profile", authHandler.GetProfile)
		authGroup.POST("/profile", authHandler.PostProfile)
		authGroup.GET("/compare", weatherHandler.GetCompare)
	}

	// Pre-initialize notification service
	initializeNotificationService()

	// Start a goroutine for daily weather reports
	go func() {
		log.Println("Starting daily weather report scheduler...")

		for {
			// Calculate time until next report (16:00 PM)
			now := time.Now()
			nextReport := time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, now.Location())
			if now.After(nextReport) {
				nextReport = nextReport.Add(24 * time.Hour)
			}

			waitDuration := nextReport.Sub(now)
			log.Printf("Next daily weather report scheduled at %v (in %v)", nextReport.Format("2006-01-02 15:04:05"), waitDuration)

			// Wait until next report time
			time.Sleep(waitDuration)

			// Send daily reports to users who opted for all weather updates
			log.Println("Sending daily weather reports...")
			sendDailyReportsToAllUsers(userStore)
		}
	}()

	// Get the port from environment variable or default to 8080
	port := getEnv("PORT", "8080")

	log.Printf("Starting GoWeather Premium server on port %s...", port)

	router.Run(":" + port)
}
