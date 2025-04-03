package services

// Global variable to store the API key
var apiKey string

// SetAPIKey sets the API key for weather services
func SetAPIKey(key string) {
	apiKey = key
}

// GetAPIKey returns the API key used for weather services
func GetAPIKey() string {
	return apiKey
}
