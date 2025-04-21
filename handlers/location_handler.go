package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Location represents a geographical location
type Location struct {
	Name     string  `json:"name"`
	Country  string  `json:"country"`
	State    string  `json:"state,omitempty"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Distance float64 `json:"distance"`
}

// NearbyLocationsResponse is the structure returned by the API
type NearbyLocationsResponse struct {
	Locations []Location `json:"locations"`
}

// HandleReverseGeocode converts coordinates to a location name
func HandleReverseGeocode(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Get latitude and longitude from query parameters
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	// Validate parameters
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing lat or lon parameters", http.StatusBadRequest)
		return
	}

	// Parse latitude and longitude
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
		return
	}

	// In a real application, you would call an external geocoding API
	// For now, we'll determine the location based on the coordinates

	// For the Charlotte coordinates (approximately 35.3, -80.7)
	// You would integrate with a real geocoding service here
	cityName := "Charlotte"
	stateName := "North Carolina"
	countryName := "United States"

	// Create location response
	location := Location{
		Name:    cityName,
		State:   stateName,
		Country: countryName,
		Lat:     lat,
		Lon:     lon,
	}

	// Encode response to JSON
	err = json.NewEncoder(w).Encode(location)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleNearbyLocations handles the API request for nearby locations
func HandleNearbyLocations(w http.ResponseWriter, r *http.Request) {
	// Set content type for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Get latitude and longitude from query parameters
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	// Validate parameters
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing lat or lon parameters", http.StatusBadRequest)
		return
	}

	// Parse latitude and longitude
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
		return
	}

	// In a real application, you would query a database or external API
	// to get nearby locations. Here we'll simulate this with sample data.
	nearbyLocations := findNearbyLocations(lat, lon)

	// Create response
	response := NearbyLocationsResponse{
		Locations: nearbyLocations,
	}

	// Encode response to JSON
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// findNearbyLocations simulates finding nearby locations
// In a real application, this would query a database or external API
func findNearbyLocations(lat, lon float64) []Location {
	// This is a simulation - in a real app you would:
	// 1. Query a geocoding API like OpenStreetMap, Google Maps, etc.
	// 2. Or query your own database with geospatial capabilities

	// For demonstration, we'll create some sample locations
	// around the provided coordinates
	locations := []Location{
		{
			Name:     fmt.Sprintf("Current Location (%.4f, %.4f)", lat, lon),
			Country:  "Current Country",
			Lat:      lat,
			Lon:      lon,
			Distance: 0,
		},
	}

	// Add some nearby places (this is just for demonstration)
	// In reality, these would come from a geocoding service
	nearbyOffsets := []struct {
		latOffset  float64
		lonOffset  float64
		name       string
		state      string
		country    string
		distanceKm float64
	}{
		{0.01, 0.01, "Northwest Area", "State A", "Country A", 1.5},
		{-0.01, 0.01, "Northeast Area", "State A", "Country A", 1.7},
		{0.01, -0.01, "Southwest Area", "State B", "Country A", 1.6},
		{-0.01, -0.01, "Southeast Area", "State B", "Country A", 1.8},
		{0.05, 0.05, "Far North City", "State C", "Country B", 7.2},
	}

	for _, offset := range nearbyOffsets {
		// Calculate approximate distance
		// In a real app, use a proper distance formula or library
		// Removed unused variables dlat and dlon

		// Using the pre-calculated distance instead of calculating it
		// distance := math.Sqrt(dlat*dlat+dlon*dlon) * 111.0 // rough approximation

		locations = append(locations, Location{
			Name:     offset.name,
			State:    offset.state,
			Country:  offset.country,
			Lat:      lat + offset.latOffset,
			Lon:      lon + offset.lonOffset,
			Distance: offset.distanceKm,
		})
	}

	return locations
}

// RenderLocationOptions renders the location options page
func RenderLocationOptions(w http.ResponseWriter, r *http.Request) {
	// Get URL parameters - we'll pass these to the page for client-side processing
	// but not use them directly here
	latParam := r.URL.Query().Get("lat")
	lonParam := r.URL.Query().Get("lon")

	// In a real application, you would:
	// 1. Get the user session if they're logged in
	// 2. Fetch nearby locations based on lat/lon if provided
	// 3. Render the template with the locations data

	// This is a simplified example that doesn't actually fetch locations
	// Those would be fetched client-side via the API endpoint

	// We'll use these parameters in the rendered HTML
	_ = latParam
	_ = lonParam

	// Render the template
	// This assumes you have a template engine set up in your app
	// Replace this with your actual template rendering code
	// Example if using Go's html/template:
	// tmpl, err := template.ParseFiles("templates/location-options.html")
	// if err != nil {
	//     http.Error(w, "Template error", http.StatusInternalServerError)
	//     return
	// }
	// err = tmpl.Execute(w, data)
	// if err != nil {
	//     http.Error(w, "Template render error", http.StatusInternalServerError)
	// }

	// For now, just output a simple response
	// Replace this with your actual template rendering
	w.Header().Set("Content-Type", "text/html")

	// Create HTML string with data attributes if available
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Select Your Location - GoWeather</title>
			<link rel="stylesheet" href="/static/css/style.css">
		</head>
		<body>
			<h1>Select Your Location</h1>
			<div class="location-options"`

	// Add data attributes directly if available
	if latParam != "" && lonParam != "" {
		html += fmt.Sprintf(` data-lat="%s" data-lon="%s"`, latParam, lonParam)
	}

	html += `></div>
			<script src="/static/js/location-options.js"></script>
		</body>
		</html>
	`

	w.Write([]byte(html))
}
