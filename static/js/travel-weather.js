document.addEventListener('DOMContentLoaded', function() {
    // Map-related variables
    let map;
    let routeLayer;
    let markersLayer;

    // Event listener for the include stops checkbox
    const includeStopsCheckbox = document.getElementById('include-stops');
    const stopsContainer = document.getElementById('stops-container');
    const stopsList = document.getElementById('stops-list');
    const addStopButton = document.getElementById('add-stop');
    const checkWeatherBtn = document.getElementById('check-weather');
    let stopCount = 0;

    // Toggle stops section when checkbox is clicked
    if (includeStopsCheckbox) {
        includeStopsCheckbox.addEventListener('change', function() {
            if (this.checked) {
                stopsContainer.classList.remove('hidden');
            } else {
                stopsContainer.classList.add('hidden');
            }
        });
    }

    // Add stop when add stop button is clicked
    if (addStopButton) {
        addStopButton.addEventListener('click', function() {
            addStop();
        });
    }

    // Function to add a new stop input field
    function addStop() {
        stopCount++;
        const stopDiv = document.createElement('div');
        stopDiv.className = 'stop-item';
        stopDiv.innerHTML = `
            <div class="form-group">
                <label for="stop-${stopCount}">Stop ${stopCount}:</label>
                <div class="stop-input-group">
                    <input type="text" id="stop-${stopCount}" class="stop-input" placeholder="Enter location">
                    <button type="button" class="remove-stop btn secondary small">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
        `;
        stopsList.appendChild(stopDiv);

        // Add event listener to remove button
        const removeButton = stopDiv.querySelector('.remove-stop');
        removeButton.addEventListener('click', function() {
            stopDiv.remove();
        });
    }

    // Check weather button click handler
    if (checkWeatherBtn) {
        checkWeatherBtn.addEventListener('click', function() {
            // Get form values
            const fromLocation = document.getElementById('from-location').value;
            const toLocation = document.getElementById('to-location').value;
            const travelDate = document.getElementById('travel-date').value;
            const includeStops = includeStopsCheckbox.checked;

            // Validate inputs
            if (!fromLocation || !toLocation) {
                alert('Please enter both origin and destination locations.');
                return;
            }

            if (!travelDate) {
                alert('Please select a travel date.');
                return;
            }

            // Collect stops if included
            const stops = [];
            if (includeStops) {
                const stopInputs = document.querySelectorAll('.stop-input');
                stopInputs.forEach(input => {
                    if (input.value.trim()) {
                        stops.push(input.value.trim());
                    }
                });
            }

            // Show loading state
            checkWeatherBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Loading...';
            checkWeatherBtn.disabled = true;

            // Create URL with query parameters
            let apiUrl = `/api/travel-weather?origin=${encodeURIComponent(fromLocation)}&destination=${encodeURIComponent(toLocation)}&date=${encodeURIComponent(travelDate)}`;

            // Add stops to URL if any
            if (stops.length > 0) {
                stops.forEach(stop => {
                    apiUrl += `&stops=${encodeURIComponent(stop)}`;
                });
            }

            // Call the API endpoint
            fetch(apiUrl)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(data => {
                    // Show results container
                    document.getElementById('results-container').classList.remove('hidden');

                    // Update the UI with the weather data
                    updateWeatherUI(data);

                    // Reset button
                    checkWeatherBtn.innerHTML = 'Check Weather';
                    checkWeatherBtn.disabled = false;

                    // Scroll to results
                    document.getElementById('results-container').scrollIntoView({ behavior: 'smooth' });
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Error fetching weather data. Please try again.');

                    // Reset button
                    checkWeatherBtn.innerHTML = 'Check Weather';
                    checkWeatherBtn.disabled = false;
                });
        });
    }

    // Function to update UI with weather data
    function updateWeatherUI(data) {
        // Update origin data
        document.getElementById('origin-name').textContent = data.origin.location;
        document.getElementById('origin-temp').textContent = `${Math.round(data.origin.weather.temperature)}°C`;
        document.getElementById('origin-condition').textContent = data.origin.weather.condition;
        document.getElementById('origin-wind').textContent = `${data.origin.weather.wind_speed} m/s`;
        document.getElementById('origin-humidity').textContent = `${data.origin.weather.humidity}%`;
        document.getElementById('origin-precipitation').textContent = `${data.origin.weather.precipitation}%`;

        // Update destination data
        document.getElementById('destination-name').textContent = data.destination.location;
        document.getElementById('destination-temp').textContent = `${Math.round(data.destination.weather.temperature)}°C`;
        document.getElementById('destination-condition').textContent = data.destination.weather.condition;
        document.getElementById('destination-wind').textContent = `${data.destination.weather.wind_speed} m/s`;
        document.getElementById('destination-humidity').textContent = `${data.destination.weather.humidity}%`;
        document.getElementById('destination-precipitation').textContent = `${data.destination.weather.precipitation}%`;

        // Update weather icons based on conditions
        updateWeatherIcon('origin', data.origin.weather.condition);
        updateWeatherIcon('destination', data.destination.weather.condition);

        // Update travel advice
        document.getElementById('weather-advice').textContent = data.travel_advice.weather_advisory;

        // Update packing list
        const packingList = document.getElementById('packing-list');
        packingList.innerHTML = '';
        data.travel_advice.packing_suggestions.forEach(item => {
            const li = document.createElement('li');
            li.textContent = item;
            packingList.appendChild(li);
        });

        // Update forecast sections
        updateForecast('origin-forecast', data.origin.forecast);
        updateForecast('destination-forecast', data.destination.forecast);

        // Update timezone information
        document.getElementById('origin-timezone-name').textContent = data.origin.location;
        document.getElementById('destination-timezone-name').textContent = data.destination.location;
        document.getElementById('origin-time').textContent = data.origin.timezone.current_time;
        document.getElementById('destination-time').textContent = data.destination.timezone.current_time;
        document.getElementById('origin-sunrise').textContent = data.origin.timezone.sunrise;
        document.getElementById('origin-sunset').textContent = data.origin.timezone.sunset;
        document.getElementById('destination-sunrise').textContent = data.destination.timezone.sunrise;
        document.getElementById('destination-sunset').textContent = data.destination.timezone.sunset;
        document.getElementById('time-diff').textContent = data.travel_advice.time_difference;

        // Check if data includes coordinates for mapping
        if (data.origin.coordinates && data.destination.coordinates) {
            updateMap(data);
        } else {
            // If no coordinates, fetch geocoding data and then update the map
            fetchGeocodingData(data).then(updatedData => {
                updateMap(updatedData);
            });
        }
    }

    // Function to update weather icon based on condition
    function updateWeatherIcon(location, condition) {
        const iconElement = document.querySelector(`.${location} .weather-icon i`);
        if (!iconElement) return;

        // Remove all existing classes except 'fas'
        iconElement.className = 'fas';

        // Add appropriate icon class based on weather condition
        const condition_lower = condition.toLowerCase();
        if (condition_lower.includes('clear') || condition_lower.includes('sunny')) {
            iconElement.classList.add('fa-sun');
        } else if (condition_lower.includes('partly cloudy') || condition_lower.includes('broken clouds')) {
            iconElement.classList.add('fa-cloud-sun');
        } else if (condition_lower.includes('cloud')) {
            iconElement.classList.add('fa-cloud');
        } else if (condition_lower.includes('rain') || condition_lower.includes('drizzle')) {
            iconElement.classList.add('fa-cloud-rain');
        } else if (condition_lower.includes('thunder') || condition_lower.includes('lightning')) {
            iconElement.classList.add('fa-bolt');
        } else if (condition_lower.includes('snow')) {
            iconElement.classList.add('fa-snowflake');
        } else if (condition_lower.includes('mist') || condition_lower.includes('fog')) {
            iconElement.classList.add('fa-smog');
        } else {
            iconElement.classList.add('fa-cloud');  // Default icon
        }
    }

    // Function to update forecast sections
    function updateForecast(forecastId, forecastData) {
        const forecastContainer = document.getElementById(forecastId);
        if (!forecastContainer) return;

        const forecastDaysContainer = forecastContainer.querySelector('.forecast-days');
        forecastDaysContainer.innerHTML = '';  // Clear existing forecast

        // Add forecast data
        forecastData.forEach(day => {
            const dayElement = document.createElement('div');
            dayElement.className = 'forecast-day';

            // Determine weather icon class
            let iconClass = 'fa-cloud';  // Default
            const condition = day.condition.toLowerCase();
            if (condition.includes('sunny') || condition.includes('clear')) {
                iconClass = 'fa-sun';
            } else if (condition.includes('partly') || condition.includes('broken')) {
                iconClass = 'fa-cloud-sun';
            } else if (condition.includes('rain')) {
                iconClass = 'fa-cloud-rain';
            } else if (condition.includes('thunder')) {
                iconClass = 'fa-bolt';
            } else if (condition.includes('snow')) {
                iconClass = 'fa-snowflake';
            }

            dayElement.innerHTML = `
                <div class="day-name">${day.day}</div>
                <div class="day-icon"><i class="fas ${iconClass}"></i></div>
                <div class="day-temp">${Math.round(day.high)}°C</div>
            `;
            forecastDaysContainer.appendChild(dayElement);
        });
    }

    // Handle forecast tab clicks
    const forecastTabs = document.querySelectorAll('.forecast-tabs .tab');
    forecastTabs.forEach(tab => {
        tab.addEventListener('click', function() {
            // Remove active class from all tabs and panels
            forecastTabs.forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.forecast-panel').forEach(p => p.classList.remove('active'));

            // Add active class to clicked tab
            this.classList.add('active');

            // Show corresponding panel
            const tabId = this.getAttribute('data-tab');
            document.getElementById(tabId).classList.add('active');
        });
    });

    // MAP FUNCTIONALITY

    // Function to initialize the map
    function initMap() {
        if (map) {
            map.remove();
        }

        map = L.map('travel-map').setView([20, 0], 2);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            maxZoom: 18
        }).addTo(map);

        routeLayer = L.layerGroup().addTo(map);
        markersLayer = L.layerGroup().addTo(map);
    }

    // Function to update the map with route data
    function updateMap(data) {
        // Make sure the map container is visible
        const mapContainer = document.getElementById('map-container');
        if (mapContainer) {
            mapContainer.classList.remove('hidden');
        }

        // Initialize map if it doesn't exist
        if (!map) {
            initMap();
        }

        // Clear existing markers and routes
        routeLayer.clearLayers();
        markersLayer.clearLayers();

        // Create markers array to store all locations for bounds calculation
        const markers = [];

        // Add origin marker
        const originMarker = createMarker(
            data.origin.location,
            [data.origin.coordinates.lat, data.origin.coordinates.lon],
            'blue',
            data.origin
        );
        markers.push(originMarker);

        // Add destination marker
        const destinationMarker = createMarker(
            data.destination.location,
            [data.destination.coordinates.lat, data.destination.coordinates.lon],
            'green',
            data.destination
        );
        markers.push(destinationMarker);

        // Add stop markers if any
        const waypoints = [];
        if (data.stops && data.stops.length > 0) {
            data.stops.forEach((stop, index) => {
                const stopMarker = createMarker(
                    stop.location,
                    [stop.coordinates.lat, stop.coordinates.lon],
                    'orange',
                    stop
                );
                markers.push(stopMarker);

                // Add to waypoints for route
                waypoints.push([stop.coordinates.lat, stop.coordinates.lon]);
            });
        }

        // Draw route from origin to destination, through any stops
        drawRoute([
            [data.origin.coordinates.lat, data.origin.coordinates.lon],
            ...waypoints,
            [data.destination.coordinates.lat, data.destination.coordinates.lon]
        ]);

        // Set map bounds to fit all markers
        if (markers.length > 0) {
            const group = L.featureGroup(markers);
            map.fitBounds(group.getBounds(), { padding: [50, 50] });
        }
    }

    // Create a marker with popup showing weather info
    function createMarker(title, position, color, locationData) {
        // Define icon colors
        const colors = {
            blue: { primary: '#1e88e5', border: '#0d47a1' },
            green: { primary: '#26a69a', border: '#00695c' },
            orange: { primary: '#ff9800', border: '#e65100' }
        };

        // Create custom icon
        const selectedColor = colors[color] || colors.blue;
        const icon = L.divIcon({
            className: 'custom-marker',
            html: `<div style="background-color: ${selectedColor.primary}; border: 2px solid ${selectedColor.border}; width: 16px; height: 16px; border-radius: 50%;"></div>`,
            iconSize: [16, 16],
            iconAnchor: [8, 8]
        });

        // Create marker
        const marker = L.marker(position, { icon }).addTo(markersLayer);

        // Create popup with weather info
        const popupContent = `
            <div class="weather-popup">
                <strong>${title}</strong>
                <div class="temp">${Math.round(locationData.weather.temperature)}°C</div>
                <div class="condition">${locationData.weather.condition}</div>
            </div>
        `;

        marker.bindPopup(popupContent);
        return marker;
    }

    // Draw route line between points
    function drawRoute(points) {
        if (points.length < 2) return;

        // Create a polyline with the points
        const polyline = L.polyline(points, {
            color: '#1e88e5',
            weight: 4,
            opacity: 0.7,
            dashArray: '10, 10',
            lineJoin: 'round'
        }).addTo(routeLayer);
    }

    // Function to fetch geocoding data if needed
    async function fetchGeocodingData(data) {
        // Clone the data object
        const updatedData = JSON.parse(JSON.stringify(data));

        // If origin doesn't have coordinates, fetch them
        if (!updatedData.origin.coordinates) {
            updatedData.origin.coordinates = await geocodeLocation(updatedData.origin.location);
        }

        // If destination doesn't have coordinates, fetch them
        if (!updatedData.destination.coordinates) {
            updatedData.destination.coordinates = await geocodeLocation(updatedData.destination.location);
        }

        // If there are stops, fetch coordinates for each
        if (updatedData.stops && updatedData.stops.length > 0) {
            for (let i = 0; i < updatedData.stops.length; i++) {
                if (!updatedData.stops[i].coordinates) {
                    updatedData.stops[i].coordinates = await geocodeLocation(updatedData.stops[i].location);
                }
            }
        }

        return updatedData;
    }

    // Function to geocode a location name to coordinates
    async function geocodeLocation(locationName) {
        // Note: In a production app, you'd use a geocoding service like Nominatim or Google Maps
        // For this demo, we'll use a free Nominatim endpoint
        try {
            const response = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(locationName)}`);
            const data = await response.json();

            if (data && data.length > 0) {
                return {
                    lat: parseFloat(data[0].lat),
                    lon: parseFloat(data[0].lon)
                };
            }

            // If geocoding fails, return default coordinates
            return { lat: 0, lon: 0 };
        } catch (error) {
            console.error('Geocoding error:', error);
            return { lat: 0, lon: 0 };
        }
    }
});