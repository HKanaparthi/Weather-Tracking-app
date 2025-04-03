// services/weather_utils.go
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GetCurrentWeather fetches current weather data for a city
func GetCurrentWeather(city string) (*WeatherData, error) {
	// URL encode the city name to properly handle spaces and special characters
	encodedCity := url.QueryEscape(city)

	// Get API key
	apiKey := GetAPIKey()

	weatherUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		encodedCity, apiKey)

	resp, err := http.Get(weatherUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	var data WeatherData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// GetCurrentWeatherByCoords fetches current weather data for coordinates
func GetCurrentWeatherByCoords(lat, lon float64) (*WeatherData, error) {
	// Get API key
	apiKey := GetAPIKey()

	weatherUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
		lat, lon, apiKey)

	resp, err := http.Get(weatherUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	var data WeatherData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
