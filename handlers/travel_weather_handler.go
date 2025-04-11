package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// TravelLocation represents a location in a travel plan
type TravelLocation struct {
	Name        string    `json:"name"`
	Coordinates []float64 `json:"coordinates"`
}

// TravelWeatherResponse represents the complete travel weather data
type TravelWeatherResponse struct {
	Origin       TravelLocationWeather   `json:"origin"`
	Destination  TravelLocationWeather   `json:"destination"`
	Stops        []TravelLocationWeather `json:"stops,omitempty"`
	TravelAdvice TravelAdviceData        `json:"travel_advice"`
}

// TravelLocationWeather represents weather data for a specific location
type TravelLocationWeather struct {
	Location string            `json:"location"`
	Weather  TravelWeatherData `json:"weather"`
	Timezone TimezoneData      `json:"timezone"`
	Forecast []ForecastDay     `json:"forecast"`
}

// TravelWeatherData represents current weather conditions
type TravelWeatherData struct {
	Temperature   float64 `json:"temperature"`
	Condition     string  `json:"condition"`
	WindSpeed     float64 `json:"wind_speed"`
	Humidity      int     `json:"humidity"`
	Precipitation int     `json:"precipitation"`
}

// TimezoneData represents timezone data for a location
type TimezoneData struct {
	CurrentTime string `json:"current_time"`
	Sunrise     string `json:"sunrise"`
	Sunset      string `json:"sunset"`
}

// ForecastDay represents a single day's forecast
type ForecastDay struct {
	Day       string  `json:"day"`
	Condition string  `json:"condition"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
}

// TravelAdviceData contains travel suggestions based on weather
type TravelAdviceData struct {
	WeatherAdvisory    string   `json:"weather_advisory"`
	PackingSuggestions []string `json:"packing_suggestions"`
	TimeDifference     string   `json:"time_difference"`
}

// OpenWeatherResponse represents the response from OpenWeather API
type OpenWeatherResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Sys struct {
		Sunrise int64 `json:"sunrise"`
		Sunset  int64 `json:"sunset"`
	} `json:"sys"`
	Timezone int64 `json:"timezone"`
	Coord    struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
}

// ForecastResponse represents the forecast response from OpenWeather API
type ForecastResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp     float64 `json:"temp"`
			TempMin  float64 `json:"temp_min"`
			TempMax  float64 `json:"temp_max"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
		} `json:"weather"`
		Pop   float64 `json:"pop"`
		DtTxt string  `json:"dt_txt"`
	} `json:"list"`
}

// TravelWeatherHandler handles the travel weather planner page
func TravelWeatherHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/travel-weather.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// TravelWeatherAPIHandler handles API requests for travel weather data
func TravelWeatherAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Get query parameters
	origin := r.URL.Query().Get("origin")
	destination := r.URL.Query().Get("destination")
	travelDateStr := r.URL.Query().Get("date")
	stops := r.URL.Query()["stops"]

	// Validate inputs
	if origin == "" || destination == "" {
		http.Error(w, `{"error": "Both origin and destination are required"}`, http.StatusBadRequest)
		return
	}

	// Parse travel date (default to today if not provided or invalid)
	var travelDate time.Time
	var err error
	if travelDateStr != "" {
		travelDate, err = time.Parse("2006-01-02", travelDateStr)
		if err != nil {
			travelDate = time.Now()
		}
	} else {
		travelDate = time.Now()
	}

	// Get API key from environment variable
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		apiKey = "0c2e2084bdd01a671b1b450215191f89" // Fallback to the one from your code
	}

	// Get weather data for origin
	originWeather, err := getLocationWeather(origin, travelDate, apiKey)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to get weather data for origin: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Get weather data for destination
	destinationWeather, err := getLocationWeather(destination, travelDate, apiKey)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to get weather data for destination: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Process any stops
	var stopsWeather []TravelLocationWeather
	for _, stop := range stops {
		if stop != "" {
			stopWeather, err := getLocationWeather(stop, travelDate, apiKey)
			if err != nil {
				log.Printf("Error getting weather for stop %s: %v", stop, err)
				continue
			}
			stopsWeather = append(stopsWeather, stopWeather)
		}
	}

	// Generate travel advice
	travelAdvice := generateTravelAdvice(originWeather, destinationWeather, stopsWeather)

	// Create response
	response := TravelWeatherResponse{
		Origin:       originWeather,
		Destination:  destinationWeather,
		Stops:        stopsWeather,
		TravelAdvice: travelAdvice,
	}

	// Convert to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to encode response: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

// getLocationWeather gets weather data for a specific location
func getLocationWeather(location string, date time.Time, apiKey string) (TravelLocationWeather, error) {
	// Call OpenWeather API for current weather
	currentWeatherURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=imperial",
		location, apiKey)

	resp, err := http.Get(currentWeatherURL)
	if err != nil {
		return TravelLocationWeather{}, fmt.Errorf("failed to call weather API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TravelLocationWeather{}, fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}

	var weatherResp OpenWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return TravelLocationWeather{}, fmt.Errorf("failed to decode weather response: %v", err)
	}

	// Get forecast data
	forecastURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=imperial",
		location, apiKey)

	forecastResp, err := http.Get(forecastURL)
	if err != nil {
		return TravelLocationWeather{}, fmt.Errorf("failed to call forecast API: %v", err)
	}
	defer forecastResp.Body.Close()

	if forecastResp.StatusCode != http.StatusOK {
		return TravelLocationWeather{}, fmt.Errorf("forecast API returned status code %d", forecastResp.StatusCode)
	}

	var forecast ForecastResponse
	if err := json.NewDecoder(forecastResp.Body).Decode(&forecast); err != nil {
		return TravelLocationWeather{}, fmt.Errorf("failed to decode forecast response: %v", err)
	}

	// Process the data
	weatherCondition := "Unknown"
	if len(weatherResp.Weather) > 0 {
		weatherCondition = weatherResp.Weather[0].Description
	}

	// Calculate precipitation chance based on the first day's forecast
	precipitationChance := 0
	if len(forecast.List) > 0 {
		precipitationChance = int(forecast.List[0].Pop * 100)
	}

	// Convert sunrise/sunset times to location time zone
	locationOffset := time.Duration(weatherResp.Timezone) * time.Second
	sunriseTime := time.Unix(weatherResp.Sys.Sunrise, 0).UTC().Add(locationOffset)
	sunsetTime := time.Unix(weatherResp.Sys.Sunset, 0).UTC().Add(locationOffset)
	currentTime := time.Now().UTC().Add(locationOffset)

	// Process forecast days
	forecastDays := processForecastDays(forecast, date)

	return TravelLocationWeather{
		Location: weatherResp.Name,
		Weather: TravelWeatherData{
			Temperature:   weatherResp.Main.Temp,
			Condition:     weatherCondition,
			WindSpeed:     weatherResp.Wind.Speed,
			Humidity:      weatherResp.Main.Humidity,
			Precipitation: precipitationChance,
		},
		Timezone: TimezoneData{
			CurrentTime: currentTime.Format("3:04 PM"),
			Sunrise:     sunriseTime.Format("3:04 PM"),
			Sunset:      sunsetTime.Format("3:04 PM"),
		},
		Forecast: forecastDays,
	}, nil
}

// processForecastDays processes the forecast response into daily forecast data
func processForecastDays(forecast ForecastResponse, date time.Time) []ForecastDay {
	// Map to store forecasts by date
	dailyMap := make(map[string]*ForecastDay)

	// Process each forecast item
	for _, item := range forecast.List {
		forecastTime := time.Unix(item.Dt, 0)
		dateStr := forecastTime.Format("2006-01-02")

		// If this date doesn't exist in our map, create it
		if _, exists := dailyMap[dateStr]; !exists {
			dayName := forecastTime.Format("Mon")
			dailyMap[dateStr] = &ForecastDay{
				Day:       dayName,
				Condition: item.Weather[0].Description,
				High:      item.Main.TempMax,
				Low:       item.Main.TempMin,
			}
		} else {
			// Update min/max temperatures if needed
			if item.Main.TempMax > dailyMap[dateStr].High {
				dailyMap[dateStr].High = item.Main.TempMax
			}
			if item.Main.TempMin < dailyMap[dateStr].Low {
				dailyMap[dateStr].Low = item.Main.TempMin
			}

			// Use the most common weather condition for the day
			// A more complex implementation would count occurrences
			dailyMap[dateStr].Condition = item.Weather[0].Description
		}
	}

	// Convert map to slice
	var result []ForecastDay
	for _, forecast := range dailyMap {
		result = append(result, *forecast)
	}

	// If we have fewer than 5 days, add placeholder days
	// This is a simplified approach - in a real app, you'd want to
	// extrapolate based on seasonal patterns or historical data
	for len(result) < 5 {
		dayIndex := len(result)
		day := date.AddDate(0, 0, dayIndex).Format("Mon")

		// Clone weather from the last day with slight variations
		var condition string
		var high, low float64

		if len(result) > 0 {
			lastDay := result[len(result)-1]
			condition = lastDay.Condition
			high = lastDay.High + (float64(dayIndex%3) - 1)
			low = lastDay.Low + (float64(dayIndex%3) - 1)
		} else {
			condition = "Clear sky"
			high = 75.0
			low = 60.0
		}

		result = append(result, ForecastDay{
			Day:       day,
			Condition: condition,
			High:      high,
			Low:       low,
		})
	}

	return result
}

// generateTravelAdvice creates travel recommendations based on weather data
func generateTravelAdvice(origin, destination TravelLocationWeather, stops []TravelLocationWeather) TravelAdviceData {
	var packingSuggestions []string

	// Basic items everyone needs
	packingSuggestions = append(packingSuggestions, "Travel documents", "Phone charger", "Water bottle")

	// Temperature-based suggestions
	maxTemp := math.Max(origin.Weather.Temperature, destination.Weather.Temperature)
	minTemp := math.Min(origin.Weather.Temperature, destination.Weather.Temperature)

	if maxTemp > 80 {
		packingSuggestions = append(packingSuggestions, "Sunscreen", "Sunglasses", "Hat", "Light clothing")
	} else if maxTemp > 70 {
		packingSuggestions = append(packingSuggestions, "Light jacket", "Sunglasses")
	} else if maxTemp > 50 {
		packingSuggestions = append(packingSuggestions, "Medium jacket", "Layers of clothing")
	} else {
		packingSuggestions = append(packingSuggestions, "Heavy coat", "Gloves", "Warm hat", "Scarf")
	}

	// If significant temperature difference between locations
	if maxTemp-minTemp > 20 {
		packingSuggestions = append(packingSuggestions, "Versatile clothing for temperature changes")
	}

	// Rain-based suggestions
	if origin.Weather.Precipitation > 30 || destination.Weather.Precipitation > 30 {
		packingSuggestions = append(packingSuggestions, "Umbrella", "Waterproof jacket", "Waterproof shoes")
	} else if origin.Weather.Precipitation > 10 || destination.Weather.Precipitation > 10 {
		packingSuggestions = append(packingSuggestions, "Umbrella", "Light rain jacket")
	}

	// Check for rain in any stops
	for _, stop := range stops {
		if stop.Weather.Precipitation > 30 && !contains(packingSuggestions, "Umbrella") {
			packingSuggestions = append(packingSuggestions, "Umbrella", "Waterproof jacket")
			break
		}
	}

	// Generate weather advisory
	var weatherAdvisory string

	// Check for extreme conditions
	if strings.Contains(strings.ToLower(origin.Weather.Condition), "storm") ||
		strings.Contains(strings.ToLower(destination.Weather.Condition), "storm") {
		weatherAdvisory = "Storm conditions detected. Consider postponing travel or prepare for severe weather."
	} else if origin.Weather.Precipitation > 70 || destination.Weather.Precipitation > 70 {
		weatherAdvisory = "Heavy rain expected. Consider adjusting travel plans or prepare for wet conditions."
	} else if origin.Weather.Precipitation > 30 || destination.Weather.Precipitation > 30 {
		weatherAdvisory = "Moderate rain possible. Pack appropriate gear but no major weather concerns."
	} else if maxTemp > 95 {
		weatherAdvisory = "Extreme heat expected. Stay hydrated and avoid prolonged sun exposure."
	} else if minTemp < 32 {
		weatherAdvisory = "Freezing temperatures expected. Pack warm clothing and be prepared for potential ice."
	} else {
		weatherAdvisory = "Weather conditions look favorable for travel. Enjoy your trip!"
	}

	// Check stops for any warnings that might override the above
	for _, stop := range stops {
		if strings.Contains(strings.ToLower(stop.Weather.Condition), "storm") {
			weatherAdvisory = "Storm conditions detected at one of your stops. Review your route carefully."
			break
		}
	}

	// Calculate time difference (simplified)
	// In a real app, this would use the timezone offsets
	timeDifference := calculateTimeDifference(origin, destination)

	return TravelAdviceData{
		WeatherAdvisory:    weatherAdvisory,
		PackingSuggestions: packingSuggestions,
		TimeDifference:     timeDifference,
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// calculateTimeDifference calculates the time difference between two locations
func calculateTimeDifference(origin, destination TravelLocationWeather) string {
	// Parse the current times
	originTime, err1 := time.Parse("3:04 PM", origin.Timezone.CurrentTime)
	destTime, err2 := time.Parse("3:04 PM", destination.Timezone.CurrentTime)

	if err1 != nil || err2 != nil {
		return "Unknown"
	}

	// Calculate difference in hours
	diff := int(math.Abs(destTime.Sub(originTime).Hours()))

	if diff == 0 {
		return "No time difference"
	} else if diff == 1 {
		return "1 hour"
	} else {
		return fmt.Sprintf("%d hours", diff)
	}
}

// RegisterTravelWeatherRoutes registers all travel weather related routes
func RegisterTravelWeatherRoutes(mux *http.ServeMux) {
	// Register the page handler
	mux.HandleFunc("/travel-weather", TravelWeatherHandler)

	// Register the API handler
	mux.HandleFunc("/api/travel-weather", TravelWeatherAPIHandler)

	log.Println("Travel Weather routes registered successfully")
}
