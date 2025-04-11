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
	Location string         `json:"location"`
	Weather  CurrentWeather `json:"weather"`
	Timezone TimezoneInfo   `json:"timezone"`
	Forecast []ForecastDay  `json:"forecast"`
}

// CurrentWeather represents current weather conditions
type CurrentWeather struct {
	Temperature   float64 `json:"temperature"`
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
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
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

	// Sample data (in a real app, this would come from API)
	if location == "New York" {
		return LocationWeather{
			Location: location,
			Weather: CurrentWeather{
				Temperature:   72,
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
				{Day: "Mon", Condition: "Sunny", High: 75, Low: 62},
				{Day: "Tue", Condition: "Partly Cloudy", High: 73, Low: 60},
				{Day: "Wed", Condition: "Cloudy", High: 70, Low: 58},
				{Day: "Thu", Condition: "Chance of Rain", High: 68, Low: 55},
				{Day: "Fri", Condition: "Sunny", High: 72, Low: 58},
			},
		}, nil
	} else {
		// Default data for any other location
		return LocationWeather{
			Location: location,
			Weather: CurrentWeather{
				Temperature:   65,
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
				{Day: "Mon", Condition: "Cloudy", High: 65, Low: 55},
				{Day: "Tue", Condition: "Partly Cloudy", High: 67, Low: 54},
				{Day: "Wed", Condition: "Rain", High: 62, Low: 52},
				{Day: "Thu", Condition: "Cloudy", High: 64, Low: 53},
				{Day: "Fri", Condition: "Sunny", High: 70, Low: 56},
			},
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

	// Temperature-based suggestions
	if origin.Weather.Temperature > 75 || destination.Weather.Temperature > 75 {
		packingSuggestions = append(packingSuggestions, "Sunscreen", "Sunglasses", "Hat")
	}

	if origin.Weather.Temperature < 65 || destination.Weather.Temperature < 65 {
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
