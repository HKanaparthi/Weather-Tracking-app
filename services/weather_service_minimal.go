package services

import "fmt"

// services/weather_service_minimal.go

// WeatherService is a minimal interface for weather operations
type WeatherService struct {
	ApiKey string
}

// NewWeatherService creates a new weather service
func NewWeatherService(apiKey string) *WeatherService {
	return &WeatherService{
		ApiKey: apiKey,
	}
}

// Placeholder for the method expected by the handler
func (s *WeatherService) GetHistoricalComparison(lat, lon float64) (interface{}, error) {
	// This will intentionally always return an error to trigger our fallback
	return nil, fmt.Errorf("using fallback implementation")
}
