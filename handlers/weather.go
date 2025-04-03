package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"weather-app/models"
	"weather-app/services"
)

const apiKey = "0c2e2084bdd01a671b1b450215191f89" // Your OpenWeather API key

// WeatherData struct holds the structure of the current weather response
type WeatherData struct {
	Name string `json:"name"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
		Pressure int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
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

// DailyForecast struct for a single day forecast
type DailyForecast struct {
	Index       int     `json:"index"` // Day index (1-7)
	Date        string  `json:"date"`  // Display date (e.g., "Mar 10")
	FullDate    string  `json:"-"`     // Full date for sorting (e.g., "2025-03-10")
	Day         string  `json:"day"`   // Day of week (e.g., "Mon")
	MaxTemp     float64 `json:"maxTemp"`
	MinTemp     float64 `json:"minTemp"`
	Humidity    int     `json:"humidity"`
	Pressure    int     `json:"pressure"`
	UVIndex     float64 `json:"uvIndex"` // Added UV Index
	Description string  `json:"description"`
	UVCategory  string  `json:"-"` // UV Index category (Low, Moderate, etc.)
	UVColor     string  `json:"-"` // Color for UV Index display
}

// CurrentWeather struct for current weather data
type CurrentWeather struct {
	Temperature string  `json:"temperature"`
	Humidity    string  `json:"humidity"`
	Pressure    string  `json:"pressure"`
	UVIndex     float64 `json:"uvIndex"` // Added UV Index
	Condition   string  `json:"condition"`
	UVCategory  string  `json:"-"` // UV Index category
	UVColor     string  `json:"-"` // Color for UV Index display
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
	City          string          `json:"city"`
	Current       CurrentWeather  `json:"current"`
	Forecast      []DailyForecast `json:"forecast"`
	BackgroundURL string          `json:"-"` // Background image URL based on weather condition
	Error         string          `json:"-"` // Error message if any
	User          *models.User    `json:"-"` // User for template rendering
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
		} `json:"weather"`
		DtTxt string `json:"dt_txt"`
	} `json:"list"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

// WeatherHandler handles weather-related routes
type WeatherHandler struct {
	userStore       models.UserStore
	weatherService  *services.WeatherService
	activityService *services.ActivityService
	logger          *log.Logger
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(userStore models.UserStore, weatherService *services.WeatherService, activityService *services.ActivityService) *WeatherHandler {
	return &WeatherHandler{
		userStore:       userStore,
		weatherService:  weatherService,
		activityService: activityService,
		logger:          log.New(os.Stdout, "[WeatherHandler] ", log.LstdFlags),
	}
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

// getUVData fetches UV index data using the One Call API
func getUVData(lat, lon float64) (*UVData, error) {
	uvUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely,hourly,alerts&appid=%s", lat, lon, apiKey)
	log.Printf("Making UV data request to URL: %s", uvUrl)

	resp, err := http.Get(uvUrl)
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
	} `json:"weather"`
	DtTxt string `json:"dt_txt"`
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

			// Create a new forecast with UV category and color
			uvIndex := 5.0 // Default medium value, will be updated later
			dailyMap[dateStr] = &DailyForecast{
				Date:        displayDate,
				FullDate:    dateStr,
				Day:         dayName,
				MaxTemp:     item.Main.TempMax,
				MinTemp:     item.Main.TempMin,
				Humidity:    item.Main.Humidity,
				Pressure:    item.Main.Pressure,
				UVIndex:     uvIndex,
				Description: item.Weather[0].Description,
				UVCategory:  getUVCategory(uvIndex),
				UVColor:     getUVColor(uvIndex),
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
			forecast.Humidity = item.Main.Humidity // Using the last humidity reading
			forecast.Pressure = item.Main.Pressure // Using the last pressure reading
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
	var sumMaxTemp, sumMinTemp, sumUV float64
	var sumHumidity, sumPressure int

	for _, f := range forecasts {
		sumMaxTemp += f.MaxTemp
		sumMinTemp += f.MinTemp
		sumHumidity += f.Humidity
		sumPressure += f.Pressure
		sumUV += f.UVIndex
	}

	avgMaxTemp := sumMaxTemp / float64(len(forecasts))
	avgMinTemp := sumMinTemp / float64(len(forecasts))
	avgHumidity := sumHumidity / len(forecasts)
	avgPressure := sumPressure / len(forecasts)
	avgUV := sumUV / float64(len(forecasts))

	// Get most common weather description
	descriptionCount := make(map[string]int)
	for _, f := range forecasts {
		descriptionCount[f.Description]++
	}

	mostCommonDesc := ""
	maxCount := 0
	for desc, count := range descriptionCount {
		if count > maxCount {
			maxCount = count
			mostCommonDesc = desc
		}
	}

	// Create the Friday forecast
	fridayForecast := DailyForecast{
		Day:         "Fri",
		Date:        nextFriday.Format("Jan 2"),
		FullDate:    nextFriday.Format("2006-01-02"),
		MaxTemp:     avgMaxTemp,
		MinTemp:     avgMinTemp,
		Humidity:    avgHumidity,
		Pressure:    avgPressure,
		UVIndex:     avgUV,
		Description: mostCommonDesc,
		UVCategory:  getUVCategory(avgUV),
		UVColor:     getUVColor(avgUV),
	}

	return append(forecasts, fridayForecast)
}

// Add UV index data to the forecasts
func addUVDataToForecasts(forecasts []DailyForecast, uvData *UVData) {
	if uvData == nil || len(uvData.Daily) == 0 {
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

// GetWeather handles the weather data request and rendering
func (h *WeatherHandler) GetWeather(c *gin.Context) {
	city := c.DefaultQuery("city", "")
	if city == "" {
		// Redirect to home page if no city provided
		c.Redirect(http.StatusFound, "/")
		return
	}

	log.Printf("Processing weather request for city: %s", city)

	// Get user if logged in
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		var err error
		user, err = h.userStore.GetUserByID(userID.(int))
		if err != nil {
			// Continue without user info
			log.Printf("Error getting user: %v", err)
		}
	}

	// Create a response object
	response := WeatherResponse{
		Error: "",
		User:  user,
	}

	// Get current weather data
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		response.Error = err.Error()
		c.HTML(http.StatusOK, "weather.html", response)
		return
	}

	// Get UV data using coordinates from current weather
	uvData, err := getUVData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		log.Printf("Warning: Failed to get UV data: %v. Using estimates.", err)
		// Continue with the process, we'll use default UV values if needed
	}

	// Get forecast data using coordinates from current weather
	forecastData, err := getForecastData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		response.Error = err.Error()
		c.HTML(http.StatusOK, "weather.html", response)
		return
	}

	// Process the forecast data to extract daily forecasts
	dailyForecasts := processForecasts(forecastData.List)

	// Add Friday forecast if needed
	dailyForecasts = addFridayForecast(dailyForecasts)

	// Add UV index data to the forecasts
	addUVDataToForecasts(dailyForecasts, uvData)

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

	// Get the background image URL based on the current weather condition
	backgroundURL := getBackgroundImage(currentWeather.Weather[0].Description)

	// Prepare the response
	response = WeatherResponse{
		City: currentWeather.Name,
		Current: CurrentWeather{
			Temperature: fmt.Sprintf("%.2f°C", currentWeather.Main.Temp),
			Humidity:    fmt.Sprintf("%d%%", currentWeather.Main.Humidity),
			Pressure:    fmt.Sprintf("%d hPa", currentWeather.Main.Pressure),
			UVIndex:     currentUV,
			Condition:   currentWeather.Weather[0].Description,
			UVCategory:  getUVCategory(currentUV),
			UVColor:     getUVColor(currentUV),
		},
		Forecast:      dailyForecasts,
		BackgroundURL: backgroundURL,
		User:          user,
	}

	// Render the weather template with the response data
	c.HTML(http.StatusOK, "weather.html", response)
}

// GetCompare handles the weather comparison request
func (h *WeatherHandler) GetCompare(c *gin.Context) {
	// Get cities from query parameters
	cityParams := c.QueryArray("cities")

	// Get user if logged in
	var user *models.User
	userID, exists := c.Get("user_id")
	if exists {
		var err error
		user, err = h.userStore.GetUserByID(userID.(int))
		if err != nil {
			// Continue without user info
			log.Printf("Error getting user: %v", err)
		}
	}

	// If no cities provided, just show the empty comparison page
	if len(cityParams) == 0 {
		c.HTML(http.StatusOK, "compare.html", gin.H{
			"User": user,
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
	for _, city := range cityParams {
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

			// Get UV data using coordinates from current weather
			uvData, err := getUVData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
			if err != nil {
				log.Printf("Warning: Failed to get UV data for %s: %v. Using estimates.", cityName, err)
				// Continue with the process, we'll use default UV values if needed
			}

			// Get current UV index
			currentUV := 5.0 // Default medium value
			if uvData != nil {
				currentUV = uvData.Current.UVI
			}

			// Get the background image URL based on the current weather condition
			backgroundURL := getBackgroundImage(currentWeather.Weather[0].Description)

			// Create city weather object
			cityWeather := CityWeather{
				Name: currentWeather.Name,
				Current: CurrentWeather{
					Temperature: fmt.Sprintf("%.2f°C", currentWeather.Main.Temp),
					Humidity:    fmt.Sprintf("%d%%", currentWeather.Main.Humidity),
					Pressure:    fmt.Sprintf("%d hPa", currentWeather.Main.Pressure),
					UVIndex:     currentUV,
					Condition:   currentWeather.Weather[0].Description,
					UVCategory:  getUVCategory(currentUV),
					UVColor:     getUVColor(currentUV),
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

// WeatherAPIHandler API endpoint for weather data (for AJAX requests or API consumers)
func (h *WeatherHandler) WeatherAPIHandler(c *gin.Context) {
	city := c.DefaultQuery("city", "")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City name is required"})
		return
	}

	log.Printf("Processing API weather request for city: %s", city)

	// Get current weather data
	currentWeather, err := getCurrentWeather(city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get UV data using coordinates from current weather
	uvData, err := getUVData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		log.Printf("Warning: Failed to get UV data: %v. Using estimates.", err)
		// Continue with the process, we'll use default UV values if needed
	}

	// Get forecast data using coordinates from current weather
	forecastData, err := getForecastData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Process the forecast data to extract daily forecasts
	dailyForecasts := processForecasts(forecastData.List)

	// Add Friday forecast if needed
	dailyForecasts = addFridayForecast(dailyForecasts)

	// Add UV index data to the forecasts
	addUVDataToForecasts(dailyForecasts, uvData)

	// Sort forecasts by date
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

	// Prepare the response
	response := gin.H{
		"city": currentWeather.Name,
		"current": gin.H{
			"temperature": fmt.Sprintf("%.2f°C", currentWeather.Main.Temp),
			"humidity":    fmt.Sprintf("%d%%", currentWeather.Main.Humidity),
			"pressure":    fmt.Sprintf("%d hPa", currentWeather.Main.Pressure),
			"uvIndex":     currentUV,
			"condition":   currentWeather.Weather[0].Description,
		},
		"forecast": dailyForecasts,
	}

	c.JSON(http.StatusOK, response)
}

// CompareAPIHandler API endpoint for weather comparison data
func (h *WeatherHandler) CompareAPIHandler(c *gin.Context) {
	// Get cities from query parameters
	cityParams := c.QueryArray("cities")

	if len(cityParams) == 0 {
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

	response := make([]cityResponse, 0, len(cityParams))
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Fetch data for each city concurrently
	for _, city := range cityParams {
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

			// Get UV data
			uvData, err := getUVData(currentWeather.Coord.Lat, currentWeather.Coord.Lon)
			currentUV := 5.0 // Default
			if err == nil && uvData != nil {
				currentUV = uvData.Current.UVI
			}

			// Set current weather data
			cityData.Name = currentWeather.Name // Use the official name returned by API
			cityData.Current = map[string]interface{}{
				"temperature": fmt.Sprintf("%.2f°C", currentWeather.Main.Temp),
				"humidity":    fmt.Sprintf("%d%%", currentWeather.Main.Humidity),
				"pressure":    fmt.Sprintf("%d hPa", currentWeather.Main.Pressure),
				"uvIndex":     currentUV,
				"condition":   currentWeather.Weather[0].Description,
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
		"cities": response,
	})
}

// SaveComparisonHandler saves a set of cities for comparison
func (h *WeatherHandler) SaveComparisonHandler(c *gin.Context) {
	// Get user if logged in
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You must be logged in to save comparisons"})
		return
	}

	// Get cities and name from the request
	var req struct {
		Name   string   `json:"name" binding:"required"`
		Cities []string `json:"cities" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Using userID to log or for future implementation
	log.Printf("User %v is saving comparison '%s' with cities: %v", userID, req.Name, req.Cities)

	// TODO: Implement saving to the database
	// This is a placeholder for the actual database operation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comparison saved successfully",
		"data": gin.H{
			"name":   req.Name,
			"cities": req.Cities,
		},
	})
}

// GetSavedComparisonsHandler retrieves saved comparisons for the user
func (h *WeatherHandler) GetSavedComparisonsHandler(c *gin.Context) {
	// Get user if logged in
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You must be logged in to view saved comparisons"})
		return
	}

	// Using userID to log or for future implementation
	log.Printf("Retrieving saved comparisons for user %v", userID)

	// TODO: Implement retrieval from the database
	// This is a placeholder for the actual database operation
	// Just returning sample data for now
	c.JSON(http.StatusOK, gin.H{
		"comparisons": []gin.H{
			{
				"id":     1,
				"name":   "My Favorite Cities",
				"cities": []string{"London", "New York", "Tokyo"},
			},
			{
				"id":     2,
				"name":   "Vacation Spots",
				"cities": []string{"Miami", "Barcelona", "Sydney"},
			},
		},
	})
}

// GetHistoricalComparison handles requests for historical weather comparison
func (h *WeatherHandler) GetHistoricalComparison(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		// Redirect to login page instead of returning 401
		c.Redirect(http.StatusFound, "/login.html")
		return
	}

	// Get user from ID
	user, err := h.userStore.GetUserByID(userID.(int))
	if err != nil {
		h.logger.Printf("Failed to get user: %v", err)
		c.Redirect(http.StatusFound, "/login.html")
		return
	}

	// Parse location coordinates
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude parameter"})
		return
	}

	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude parameter"})
		return
	}

	// Try to get historical comparison data first
	historicalData, err := h.weatherService.GetHistoricalComparison(lat, lon)

	// If the normal method fails, use our fallback
	if err != nil {
		h.logger.Printf("Regular historical data fetch failed: %v. Using fallback data.", err)

		// Use the fallback function instead
		fallbackData, fallbackErr := services.GenerateHistoricalFallbackData(lat, lon)
		if fallbackErr != nil {
			h.logger.Printf("Fallback also failed: %v", fallbackErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch historical weather data"})
			return
		}

		// Record this activity
		activity := models.Activity{
			UserID:    user.ID,
			Type:      "HISTORICAL_COMPARISON",
			Latitude:  lat,
			Longitude: lon,
			Timestamp: time.Now(),
		}

		if err := h.activityService.RecordActivity(activity); err != nil {
			h.logger.Printf("Failed to record historical comparison activity: %v", err)
		}

		c.JSON(http.StatusOK, fallbackData)
		return
	}

	// Record this activity
	activity := models.Activity{
		UserID:    user.ID,
		Type:      "HISTORICAL_COMPARISON",
		Latitude:  lat,
		Longitude: lon,
		Timestamp: time.Now(),
	}

	if err := h.activityService.RecordActivity(activity); err != nil {
		h.logger.Printf("Failed to record historical comparison activity: %v", err)
	}

	c.JSON(http.StatusOK, historicalData)
}
