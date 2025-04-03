package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"weather-app/models"
)

// WeatherData struct to match what's returned by getCurrentWeather
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
	Wind *struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

// ActivityRecommendation represents the recommendation for a specific activity
type ActivityRecommendation struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

// AllergenInfo represents information about a specific allergen
type AllergenInfo struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Value int    `json:"value"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// HealthImpact represents information about health impacts
type HealthImpact struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// WeatherActivities holds all recommendations and information
type WeatherActivities struct {
	OutdoorActivities []ActivityRecommendation `json:"outdoorActivities"`
	TravelCommute     []ActivityRecommendation `json:"travelCommute"`
	HomeGarden        []ActivityRecommendation `json:"homeGarden"`
	Allergens         []AllergenInfo           `json:"allergens"`
	HealthImpacts     []HealthImpact           `json:"healthImpacts"`
}

// PollutionData represents air quality and pollution data
type PollutionData struct {
	Main struct {
		Aqi int `json:"aqi"` // Air Quality Index
	} `json:"main"`
	Components struct {
		Co   float64 `json:"co"`    // Carbon monoxide
		No   float64 `json:"no"`    // Nitrogen monoxide
		No2  float64 `json:"no2"`   // Nitrogen dioxide
		O3   float64 `json:"o3"`    // Ozone
		So2  float64 `json:"so2"`   // Sulphur dioxide
		Pm25 float64 `json:"pm2_5"` // Fine particles
		Pm10 float64 `json:"pm10"`  // Coarse particles
		Nh3  float64 `json:"nh3"`   // Ammonia
	} `json:"components"`
}

// PollenData represents pollen counts and allergen data
// Note: This is a simplified model, as actual pollen data depends on your specific API
type PollenData struct {
	Tree    int `json:"tree"`
	Grass   int `json:"grass"`
	Weed    int `json:"weed"`
	Mold    int `json:"mold"`
	Ragweed int `json:"ragweed"`
	Dust    int `json:"dust"`
}

// ActivityService is responsible for providing activity recommendations
type ActivityService struct {
	apiKey string
}

// NewActivityService creates a new activity service with the given API key
func NewActivityService(apiKey string) *ActivityService {
	return &ActivityService{
		apiKey: apiKey,
	}
}

// GetAirPollutionData fetches air pollution data from OpenWeatherMap
func (s *ActivityService) GetAirPollutionData(lat, lon float64) (*PollutionData, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/air_pollution?lat=%f&lon=%f&appid=%s",
		lat, lon, s.apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch air pollution data: %s", resp.Status)
	}

	var result struct {
		List []PollutionData `json:"list"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if len(result.List) == 0 {
		return nil, fmt.Errorf("no pollution data available")
	}

	return &result.List[0], nil
}

// GetPollenData would fetch pollen data from a third-party API
// Note: You'll need to integrate with a specific provider for this data
func (s *ActivityService) GetPollenData(lat, lon float64) (*PollenData, error) {
	// This is a mock implementation
	// In a real implementation, you would call an actual pollen data API

	// For demonstration purposes, we're generating synthetic data
	// based on the current month (since pollen levels are seasonal)
	month := time.Now().Month()

	// Generate synthetic pollen data based on typical seasonal patterns
	pollenData := &PollenData{
		Tree:    10,
		Grass:   10,
		Weed:    10,
		Mold:    10,
		Ragweed: 10,
		Dust:    15,
	}

	// Adjust based on season (Northern Hemisphere patterns)
	switch {
	case month >= time.March && month <= time.May: // Spring
		pollenData.Tree = 60 + (time.Now().Day() % 20)  // High tree pollen in spring
		pollenData.Grass = 30 + (time.Now().Day() % 30) // Medium grass pollen
		pollenData.Weed = 20 + (time.Now().Day() % 10)  // Low-medium weed pollen
		pollenData.Mold = 15 + (time.Now().Day() % 15)  // Low-medium mold
	case month >= time.June && month <= time.August: // Summer
		pollenData.Tree = 20 + (time.Now().Day() % 20)    // Low-medium tree pollen
		pollenData.Grass = 70 + (time.Now().Day() % 20)   // High grass pollen
		pollenData.Weed = 40 + (time.Now().Day() % 30)    // Medium weed pollen
		pollenData.Ragweed = 60 + (time.Now().Day() % 30) // Higher ragweed in late summer
		pollenData.Mold = 40 + (time.Now().Day() % 20)    // Medium-high mold in humid summer
	case month >= time.September && month <= time.November: // Fall
		pollenData.Tree = 10 + (time.Now().Day() % 10)    // Low tree pollen
		pollenData.Grass = 20 + (time.Now().Day() % 20)   // Low-medium grass pollen
		pollenData.Weed = 60 + (time.Now().Day() % 30)    // High weed pollen
		pollenData.Ragweed = 70 + (time.Now().Day() % 20) // High ragweed pollen
		pollenData.Mold = 60 + (time.Now().Day() % 30)    // High mold with falling leaves
	default: // Winter
		pollenData.Tree = 5 + (time.Now().Day() % 10)    // Very low tree pollen
		pollenData.Grass = 5 + (time.Now().Day() % 10)   // Very low grass pollen
		pollenData.Weed = 5 + (time.Now().Day() % 10)    // Very low weed pollen
		pollenData.Ragweed = 5 + (time.Now().Day() % 10) // Very low ragweed pollen
		pollenData.Mold = 20 + (time.Now().Day() % 20)   // Medium indoor mold possible
		pollenData.Dust = 30 + (time.Now().Day() % 30)   // Higher dust due to heating systems
	}

	// Add some randomness to make it seem more realistic
	// In a real implementation, this would come from the API

	return pollenData, nil
}

// GetActivitiesAndHealth generates comprehensive activity recommendations and health information
func (s *ActivityService) GetActivitiesAndHealth(weatherData *WeatherData, lat, lon float64) (*WeatherActivities, error) {
	// Get additional data we need for recommendations
	pollutionData, err := s.GetAirPollutionData(lat, lon)
	if err != nil {
		log.Printf("Warning: Couldn't fetch pollution data: %v. Using estimates.", err)
		// Continue with default values
	}

	pollenData, err := s.GetPollenData(lat, lon)
	if err != nil {
		log.Printf("Warning: Couldn't fetch pollen data: %v. Using estimates.", err)
		// Continue with default values
	}

	// Extract weather conditions
	temperature := weatherData.Main.Temp
	humidity := weatherData.Main.Humidity
	windSpeed := 10.0 // Set a default windspeed if not available in your weatherData
	if weatherData.Wind != nil {
		windSpeed = weatherData.Wind.Speed
	}

	weatherCondition := ""
	if len(weatherData.Weather) > 0 {
		weatherCondition = strings.ToLower(weatherData.Weather[0].Description)
	}

	// Get time of day
	hour := time.Now().Hour()
	isDaytime := hour >= 6 && hour <= 18

	// Get AQI (Air Quality Index)
	aqi := 2 // Default to moderate AQI
	if pollutionData != nil {
		aqi = pollutionData.Main.Aqi
	}

	// Now generate all our recommendations
	activities := &WeatherActivities{
		OutdoorActivities: s.getOutdoorActivities(temperature, weatherCondition, windSpeed, isDaytime, aqi),
		TravelCommute:     s.getTravelRecommendations(weatherCondition, windSpeed, temperature, aqi),
		HomeGarden:        s.getHomeGardenActivities(weatherCondition, temperature, humidity),
		Allergens:         s.getAllergenInfo(pollenData, humidity),
		HealthImpacts:     s.getHealthImpacts(temperature, humidity, weatherCondition, aqi, pollenData),
	}

	return activities, nil
}

// getOutdoorActivities generates recommendations for outdoor activities
func (s *ActivityService) getOutdoorActivities(temp float64, condition string, windSpeed float64, isDaytime bool, aqi int) []ActivityRecommendation {
	activities := []ActivityRecommendation{
		{Name: "Fishing", Icon: "🎣"},
		{Name: "Running", Icon: "🏃"},
		{Name: "Golf", Icon: "⛳"},
		{Name: "Biking & Cycling", Icon: "🚲"},
		{Name: "Beach & Pool", Icon: "🏖️"},
		{Name: "Stargazing", Icon: "🔭"},
	}

	// Set default status
	for i := range activities {
		activities[i].Status = "Good"
		activities[i].Color = "#4CAF50" // Green
	}

	// Check if it's raining/snowing/etc.
	isPrecipitation := strings.Contains(condition, "rain") ||
		strings.Contains(condition, "snow") ||
		strings.Contains(condition, "drizzle") ||
		strings.Contains(condition, "thunderstorm")

	// Check if it's cloudy
	isCloudy := strings.Contains(condition, "cloud") || strings.Contains(condition, "overcast")

	// Check if it's windy
	isWindy := windSpeed > 15.0

	// Check if it's extremely hot or cold
	isExtremeCold := temp < 0
	isExtremeHot := temp > 35

	// Adjust recommendations based on conditions

	// Fishing
	if isPrecipitation || isWindy {
		activities[0].Status = "Fair"
		activities[0].Color = "#FFC107" // Yellow
	}
	if isExtremeCold || isExtremeHot {
		activities[0].Status = "Poor"
		activities[0].Color = "#F44336" // Red
	}

	// Running
	if isPrecipitation {
		activities[1].Status = "Fair"
		activities[1].Color = "#FFC107"
	}
	if isExtremeCold || isExtremeHot || aqi > 3 {
		activities[1].Status = "Poor"
		activities[1].Color = "#F44336"
	}

	// Golf
	if isPrecipitation || isWindy {
		activities[2].Status = "Poor"
		activities[2].Color = "#F44336"
	}
	if isCloudy && !isPrecipitation && !isWindy {
		activities[2].Status = "Fair"
		activities[2].Color = "#FFC107"
	}

	// Biking & Cycling
	if isPrecipitation || isWindy {
		activities[3].Status = "Poor"
		activities[3].Color = "#F44336"
	}
	if isExtremeCold || isExtremeHot || aqi > 3 {
		activities[3].Status = "Fair"
		activities[3].Color = "#FFC107"
	}

	// Beach & Pool
	if isPrecipitation || isWindy || temp < 20 {
		activities[4].Status = "Poor"
		activities[4].Color = "#F44336"
	}
	if isCloudy && temp >= 20 && temp < 25 {
		activities[4].Status = "Fair"
		activities[4].Color = "#FFC107"
	}

	// Stargazing
	if !isDaytime {
		if isPrecipitation || isCloudy {
			activities[5].Status = "Poor"
			activities[5].Color = "#F44336"
		} else {
			activities[5].Status = "Ideal"
			activities[5].Color = "#4CAF50"
		}
	} else {
		activities[5].Status = "Poor"
		activities[5].Color = "#F44336"
	}

	return activities
}

// getTravelRecommendations generates recommendations for travel and commute
func (s *ActivityService) getTravelRecommendations(condition string, windSpeed float64, temp float64, aqi int) []ActivityRecommendation {
	activities := []ActivityRecommendation{
		{Name: "Air Travel", Icon: "✈️"},
		{Name: "Driving", Icon: "🚗"},
	}

	// Set default status
	for i := range activities {
		activities[i].Status = "Good"
		activities[i].Color = "#4CAF50" // Green
	}

	// Check conditions
	hasThunderstorm := strings.Contains(condition, "thunderstorm")
	hasFog := strings.Contains(condition, "fog") || strings.Contains(condition, "mist")
	hasSnow := strings.Contains(condition, "snow")
	hasHeavyRain := strings.Contains(condition, "heavy rain")
	isVeryWindy := windSpeed > 25.0

	// Air Travel
	if hasThunderstorm || isVeryWindy {
		activities[0].Status = "Poor"
		activities[0].Color = "#F44336"
	} else if hasFog {
		activities[0].Status = "Fair"
		activities[0].Color = "#FFC107"
	} else {
		activities[0].Status = "Ideal"
		activities[0].Color = "#4CAF50"
	}

	// Driving
	if hasThunderstorm || hasHeavyRain || hasSnow || hasFog {
		activities[1].Status = "Poor"
		activities[1].Color = "#F44336"
	} else if strings.Contains(condition, "rain") || strings.Contains(condition, "drizzle") {
		activities[1].Status = "Fair"
		activities[1].Color = "#FFC107"
	} else {
		activities[1].Status = "Good"
		activities[1].Color = "#4CAF50"
	}

	return activities
}

// getHomeGardenActivities generates recommendations for home and garden activities
func (s *ActivityService) getHomeGardenActivities(condition string, temp float64, humidity int) []ActivityRecommendation {
	activities := []ActivityRecommendation{
		{Name: "Lawn Mowing", Icon: "🚜"},
		{Name: "Composting", Icon: "🌱"},
		{Name: "Outdoor Entertaining", Icon: "🍹"},
	}

	// Set default status
	for i := range activities {
		activities[i].Status = "Good"
		activities[i].Color = "#4CAF50" // Green
	}

	// Check conditions
	isPrecipitation := strings.Contains(condition, "rain") ||
		strings.Contains(condition, "snow") ||
		strings.Contains(condition, "drizzle") ||
		strings.Contains(condition, "thunderstorm")

	// Lawn Mowing
	if isPrecipitation {
		activities[0].Status = "Poor"
		activities[0].Color = "#F44336"
	} else if strings.Contains(condition, "cloudy") && !isPrecipitation {
		activities[0].Status = "Good"
		activities[0].Color = "#4CAF50"
	}

	// Composting
	if temp < 10 {
		activities[1].Status = "Fair"
		activities[1].Color = "#FFC107"
	} else if temp > 10 && humidity > 50 {
		activities[1].Status = "Ideal"
		activities[1].Color = "#4CAF50"
	}

	// Outdoor Entertaining
	if isPrecipitation {
		activities[2].Status = "Poor"
		activities[2].Color = "#F44336"
	} else if temp < 15 || temp > 30 {
		activities[2].Status = "Fair"
		activities[2].Color = "#FFC107"
	} else {
		activities[2].Status = "Good"
		activities[2].Color = "#4CAF50"
	}

	return activities
}

// getAllergenInfo generates information about allergen levels
func (s *ActivityService) getAllergenInfo(pollenData *PollenData, humidity int) []AllergenInfo {
	allergens := []AllergenInfo{
		{Name: "Tree Pollen", Icon: "🌳"},
		{Name: "Ragweed Pollen", Icon: "🌿"},
		{Name: "Mold", Icon: "🍄"},
		{Name: "Grass Pollen", Icon: "🌱"},
		{Name: "Dust & Dander", Icon: "🏠"},
	}

	// Default values if we don't have real data
	if pollenData == nil {
		for i := range allergens {
			allergens[i].Level = "Moderate"
			allergens[i].Value = 3
			allergens[i].Color = "#FFC107" // Yellow
		}
		return allergens
	}

	// Tree Pollen
	allergens[0].Value = determineLevel(pollenData.Tree)
	allergens[0].Level = getLevelName(allergens[0].Value)
	allergens[0].Color = getAllergenColor(allergens[0].Value)

	// Ragweed Pollen
	allergens[1].Value = determineLevel(pollenData.Ragweed)
	allergens[1].Level = getLevelName(allergens[1].Value)
	allergens[1].Color = getAllergenColor(allergens[1].Value)

	// Mold
	moldLevel := pollenData.Mold
	if humidity > 70 {
		moldLevel += 20 // Increase mold count in high humidity
	}
	allergens[2].Value = determineLevel(moldLevel)
	allergens[2].Level = getLevelName(allergens[2].Value)
	allergens[2].Color = getAllergenColor(allergens[2].Value)

	// Grass Pollen
	allergens[3].Value = determineLevel(pollenData.Grass)
	allergens[3].Level = getLevelName(allergens[3].Value)
	allergens[3].Color = getAllergenColor(allergens[3].Value)

	// Dust & Dander
	allergens[4].Value = determineLevel(pollenData.Dust)
	allergens[4].Level = getLevelName(allergens[4].Value)
	allergens[4].Color = getAllergenColor(allergens[4].Value)

	return allergens
}

// Helper function to determine allergen level based on count
func determineLevel(count int) int {
	if count < 20 {
		return 1 // Low
	} else if count < 50 {
		return 2 // Moderate
	} else if count < 80 {
		return 3 // High
	}
	return 4 // Very High
}

// Helper function to get level name
func getLevelName(level int) string {
	switch level {
	case 1:
		return "Low"
	case 2:
		return "Moderate"
	case 3:
		return "High"
	case 4:
		return "Very High"
	default:
		return "Unknown"
	}
}

// Helper function to get color for allergen level
func getAllergenColor(level int) string {
	switch level {
	case 1:
		return "#4CAF50" // Green
	case 2:
		return "#FFC107" // Yellow
	case 3:
		return "#FF9800" // Orange
	case 4:
		return "#F44336" // Red
	default:
		return "#9E9E9E" // Gray
	}
}

// RecordActivity records a user's weather-related activity
func (s *ActivityService) RecordActivity(activity models.Activity) error {
	// In a production implementation, you would store this in your database
	// For now, we'll just log it
	log.Printf("Recording activity: User ID=%d, Type=%s, Location=(%.4f, %.4f), Time=%v",
		activity.UserID, activity.Type, activity.Latitude, activity.Longitude, activity.Timestamp)

	return nil
}

// getHealthImpacts generates information about potential health impacts
func (s *ActivityService) getHealthImpacts(temp float64, humidity int, condition string, aqi int, pollenData *PollenData) []HealthImpact {
	impacts := []HealthImpact{
		{Name: "Arthritis", Icon: "🦴"},
		{Name: "Sinus Pressure", Icon: "👃"},
		{Name: "Common Cold", Icon: "🤧"},
		{Name: "Flu", Icon: "🤒"},
		{Name: "Migraine", Icon: "🤕"},
		{Name: "Asthma", Icon: "🫁"},
	}

	// Set defaults
	for i := range impacts {
		impacts[i].Level = "Low"
		impacts[i].Color = "#4CAF50" // Green
	}

	// Arthritis
	barometricPressure := 1013.25 // Default standard pressure (hPa)
	// In a real implementation, you'd get this from the weather data

	if temp < 10 || humidity > 80 || barometricPressure < 1000 {
		impacts[0].Level = "Moderate"
		impacts[0].Color = "#FFC107" // Yellow
	}
	if temp < 5 && humidity > 70 {
		impacts[0].Level = "High"
		impacts[0].Color = "#F44336" // Red
	}

	// Sinus Pressure
	if strings.Contains(condition, "change") || barometricPressure < 1000 || barometricPressure > 1020 {
		impacts[1].Level = "Moderate"
		impacts[1].Color = "#FFC107"
	}
	if pollenData != nil && (pollenData.Tree > 50 || pollenData.Ragweed > 50) {
		impacts[1].Level = "High"
		impacts[1].Color = "#F44336"
	}

	// Common Cold
	if temp < 10 && humidity < 40 {
		impacts[2].Level = "Moderate"
		impacts[2].Color = "#FFC107"
	}

	// Flu
	// Flu is generally a seasonal concern, not directly tied to daily weather
	// We'll keep it as "Low" for most conditions

	// Migraine
	if barometricPressure < 1000 || barometricPressure > 1025 {
		impacts[4].Level = "Moderate"
		impacts[4].Color = "#FFC107"
	}
	if strings.Contains(condition, "thunderstorm") {
		impacts[4].Level = "High"
		impacts[4].Color = "#F44336"
	}

	// Asthma
	if aqi > 2 {
		impacts[5].Level = "Moderate"
		impacts[5].Color = "#FFC107"
	}
	if aqi > 3 || (pollenData != nil && (pollenData.Tree > 70 || pollenData.Grass > 70)) {
		impacts[5].Level = "High"
		impacts[5].Color = "#F44336"
	}

	return impacts
}
