package services

import (
	"fmt"
	"time"
)

// TravelWeatherData represents the complete travel weather data
type TravelWeatherData struct {
	Origin       LocationWeather   `json:"origin"`
	Destination  LocationWeather   `json:"destination"`
	Stops        []LocationWeather `json:"stops,omitempty"`
	TravelAdvice TravelAdvice      `json:"travel_advice"`
}

// LocationWeather represents weather data for a specific location
type LocationWeather struct {
	Location    string         `json:"location"`
	Weather     CurrentWeather `json:"weather"`
	Timezone    TimezoneInfo   `json:"timezone"`
	Forecast    []ForecastDay  `json:"forecast"`
	Coordinates Coordinates    `json:"coordinates"` // Added for map functionality
}

// Coordinates represents geographical coordinates
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	Temperature   float64 `json:"temperature"` // Now in Celsius
	Condition     string  `json:"condition"`
	WindSpeed     float64 `json:"wind_speed"`
	Humidity      int     `json:"humidity"`
	Precipitation int     `json:"precipitation"`
}

// TimezoneInfo represents timezone data for a location
type TimezoneInfo struct {
	CurrentTime string `json:"current_time"`
	Sunrise     string `json:"sunrise"`
	Sunset      string `json:"sunset"`
}

// ForecastDay represents a single day's forecast
type ForecastDay struct {
	Day       string  `json:"day"`
	Condition string  `json:"condition"`
	High      float64 `json:"high"` // Now in Celsius
	Low       float64 `json:"low"`  // Now in Celsius
}

// TravelAdvice contains travel suggestions based on weather
type TravelAdvice struct {
	WeatherAdvisory    string   `json:"weather_advisory"`
	PackingSuggestions []string `json:"packing_suggestions"`
	TimeDifference     string   `json:"time_difference"`
}

// TravelWeatherService provides methods for travel weather planning
type TravelWeatherService struct {
	weatherService *WeatherService
}

// NewTravelWeatherService creates a new travel weather service instance
func NewTravelWeatherService(weatherService *WeatherService) *TravelWeatherService {
	return &TravelWeatherService{
		weatherService: weatherService,
	}
}

// Convert Fahrenheit to Celsius
func fahrenheitToCelsius(fahrenheit float64) float64 {
	return (fahrenheit - 32) * 5 / 9
}

// GetTravelWeatherData fetches and processes weather data for travel planning
func (s *TravelWeatherService) GetTravelWeatherData(origin, destination string, travelDate time.Time, stops []string) (*TravelWeatherData, error) {
	// This would normally call the weather service to get real data
	// For now, we'll return sample data

	// In a real implementation, you would:
	// 1. Get weather data for the origin
	// 2. Get weather data for the destination
	// 3. If stops provided, get weather for those
	// 4. Generate travel advice based on the weather data
	// 5. Return the combined data

	originWeather, err := s.getLocationWeather(origin, travelDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching origin weather: %w", err)
	}

	destWeather, err := s.getLocationWeather(destination, travelDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching destination weather: %w", err)
	}

	// Process any stops
	var stopsWeather []LocationWeather
	for _, stop := range stops {
		stopWeather, err := s.getLocationWeather(stop, travelDate)
		if err != nil {
			return nil, fmt.Errorf("error fetching weather for stop %s: %w", stop, err)
		}
		stopsWeather = append(stopsWeather, stopWeather)
	}

	// Generate travel advice
	travelAdvice := s.generateTravelAdvice(originWeather, destWeather, stopsWeather)

	return &TravelWeatherData{
		Origin:       originWeather,
		Destination:  destWeather,
		Stops:        stopsWeather,
		TravelAdvice: travelAdvice,
	}, nil
}

// getLocationWeather gets weather data for a specific location
// This is a placeholder that would normally call your weather service
func (s *TravelWeatherService) getLocationWeather(location string, date time.Time) (LocationWeather, error) {
	// In a real implementation, this would call your weather service API
	// For now, return sample data based on the location name

	// Map of locations to coordinates (simplified for demo)
	locationCoordinates := map[string]Coordinates{
		"new york":     {Lat: 40.7128, Lon: -74.0060},
		"los angeles":  {Lat: 34.0522, Lon: -118.2437},
		"chicago":      {Lat: 41.8781, Lon: -87.6298},
		"houston":      {Lat: 29.7604, Lon: -95.3698},
		"phoenix":      {Lat: 33.4484, Lon: -112.0740},
		"philadelphia": {Lat: 39.9526, Lon: -75.1652},
		"san antonio":  {Lat: 29.4241, Lon: -98.4936},
		"san diego":    {Lat: 32.7157, Lon: -117.1611},
		"dallas":       {Lat: 32.7767, Lon: -96.7970},
		"san jose":     {Lat: 37.3382, Lon: -121.8863},
		"austin":       {Lat: 30.2672, Lon: -97.7431},
		"boston":       {Lat: 42.3601, Lon: -71.0589},
		"las vegas":    {Lat: 36.1699, Lon: -115.1398},
		"miami":        {Lat: 25.7617, Lon: -80.1918},
		"denver":       {Lat: 39.7392, Lon: -104.9903},
		"seattle":      {Lat: 47.6062, Lon: -122.3321},
		"atlanta":      {Lat: 33.7490, Lon: -84.3880},
		"london":       {Lat: 51.5074, Lon: -0.1278},
		"paris":        {Lat: 48.8566, Lon: 2.3522},
		"tokyo":        {Lat: 35.6762, Lon: 139.6503},
		"sydney":       {Lat: -33.8688, Lon: 151.2093},
	}

	// Get coordinates for the location (case-insensitive)
	locationLower := location
	if locationLower == "" {
		locationLower = "new york" // Default to New York if empty
	}

	coords, exists := locationCoordinates[locationLower]
	if !exists {
		// Default coordinates if location not found
		coords = Coordinates{Lat: 0, Lon: 0}
	}

	// Sample data (in a real app, this would come from API)
	if location == "New York" {
		// Convert Fahrenheit values to Celsius
		tempF := 72.0
		highF := 75.0
		lowF := 62.0

		return LocationWeather{
			Location: location,
			Weather: CurrentWeather{
				Temperature:   fahrenheitToCelsius(tempF), // Now in Celsius
				Condition:     "Sunny",
				WindSpeed:     8,
				Humidity:      45,
				Precipitation: 0,
			},
			Timezone: TimezoneInfo{
				CurrentTime: "10:30 AM",
				Sunrise:     "6:45 AM",
				Sunset:      "7:30 PM",
			},
			Forecast: []ForecastDay{
				{Day: "Mon", Condition: "Sunny", High: fahrenheitToCelsius(highF), Low: fahrenheitToCelsius(lowF)},
				{Day: "Tue", Condition: "Partly Cloudy", High: fahrenheitToCelsius(73), Low: fahrenheitToCelsius(60)},
				{Day: "Wed", Condition: "Cloudy", High: fahrenheitToCelsius(70), Low: fahrenheitToCelsius(58)},
				{Day: "Thu", Condition: "Chance of Rain", High: fahrenheitToCelsius(68), Low: fahrenheitToCelsius(55)},
				{Day: "Fri", Condition: "Sunny", High: fahrenheitToCelsius(72), Low: fahrenheitToCelsius(58)},
			},
			Coordinates: coords,
		}, nil
	} else {
		// Default data for any other location
		// Convert Fahrenheit values to Celsius
		tempF := 65.0

		return LocationWeather{
			Location: location,
			Weather: CurrentWeather{
				Temperature:   fahrenheitToCelsius(tempF), // Now in Celsius
				Condition:     "Partly Cloudy",
				WindSpeed:     12,
				Humidity:      60,
				Precipitation: 10,
			},
			Timezone: TimezoneInfo{
				CurrentTime: "1:30 PM",
				Sunrise:     "6:15 AM",
				Sunset:      "8:00 PM",
			},
			Forecast: []ForecastDay{
				{Day: "Mon", Condition: "Cloudy", High: fahrenheitToCelsius(65), Low: fahrenheitToCelsius(55)},
				{Day: "Tue", Condition: "Partly Cloudy", High: fahrenheitToCelsius(67), Low: fahrenheitToCelsius(54)},
				{Day: "Wed", Condition: "Rain", High: fahrenheitToCelsius(62), Low: fahrenheitToCelsius(52)},
				{Day: "Thu", Condition: "Cloudy", High: fahrenheitToCelsius(64), Low: fahrenheitToCelsius(53)},
				{Day: "Fri", Condition: "Sunny", High: fahrenheitToCelsius(70), Low: fahrenheitToCelsius(56)},
			},
			Coordinates: coords,
		}, nil
	}
}

// generateTravelAdvice creates travel recommendations based on weather data
func (s *TravelWeatherService) generateTravelAdvice(origin, destination LocationWeather, stops []LocationWeather) TravelAdvice {
	// This would have more sophisticated logic in a real app
	// For now, return sample advice

	var packingSuggestions []string

	// Basic items everyone needs
	packingSuggestions = append(packingSuggestions, "Travel documents", "Phone charger", "Water bottle")

	// Temperature-based suggestions (adjusted for Celsius)
	if origin.Weather.Temperature > 24 || destination.Weather.Temperature > 24 { // ~75°F in Celsius
		packingSuggestions = append(packingSuggestions, "Sunscreen", "Sunglasses", "Hat")
	}

	if origin.Weather.Temperature < 18 || destination.Weather.Temperature < 18 { // ~65°F in Celsius
		packingSuggestions = append(packingSuggestions, "Light jacket or sweater")
	}

	// Rain-based suggestions
	if origin.Weather.Precipitation > 0 || destination.Weather.Precipitation > 0 {
		packingSuggestions = append(packingSuggestions, "Umbrella", "Waterproof jacket")
	}

	// Weather advisory
	var weatherAdvisory string
	if origin.Weather.Precipitation > 30 || destination.Weather.Precipitation > 30 {
		weatherAdvisory = "Heavy rain expected. Consider adjusting travel plans or prepare for wet conditions."
	} else if origin.Weather.Precipitation > 0 || destination.Weather.Precipitation > 0 {
		weatherAdvisory = "Light rain possible. Pack appropriate gear but no major weather concerns."
	} else {
		weatherAdvisory = "Weather conditions look favorable for travel. Enjoy your trip!"
	}

	// Time difference (simplified)
	timeDifference := "3 hours" // This would be calculated based on actual timezone data

	return TravelAdvice{
		WeatherAdvisory:    weatherAdvisory,
		PackingSuggestions: packingSuggestions,
		TimeDifference:     timeDifference,
	}
}
