/**
 * dashboard.js - JavaScript for the weather dashboard
 * Handles dynamic weather data loading and UI updates
 */

document.addEventListener('DOMContentLoaded', function() {
    // Elements
    const searchForm = document.getElementById('searchForm');
    const cityInput = document.getElementById('cityInput');
    const getLocationBtn = document.getElementById('getLocation');
    const locationAlert = document.getElementById('locationAlert');
    const allowLocationBtn = document.getElementById('allowLocation');
    const denyLocationBtn = document.getElementById('denyLocation');
    const loadingIndicator = document.getElementById('loadingIndicator');

    // Weather elements
    const currentTempElement = document.getElementById('currentTemp');
    const weatherTypeElement = document.getElementById('weatherType');
    const locationNameElement = document.getElementById('locationName');
    const currentDegreeElement = document.getElementById('currentDegree');
    const cloudyDegreeElement = document.getElementById('cloudyDegree');
    const currentWeatherLabelElement = document.getElementById('currentWeatherLabel');
    const humidityValueElement = document.getElementById('humidityValue');
    const currentDateElement = document.getElementById('currentDate');

    // Initialize dashboard with default data
    initializeDashboard();

    // Check URL parameters to see if we're loading data
    const urlParams = new URLSearchParams(window.location.search);
    const city = urlParams.get('city');
    const lat = urlParams.get('lat');
    const lon = urlParams.get('lon');
    const skipLocationCheck = urlParams.get('from') === 'location_options';

    // If we have a city or coordinates, load weather data
    if (city) {
        loadWeatherData(city);
    } else if (lat && lon) {
        loadWeatherDataByCoordinates(lat, lon);
    } else {
        // If user has a home city set, use that
        const homeCity = cityInput.value;
        if (homeCity && !skipLocationCheck) {
            loadWeatherData(homeCity);
        } else if (!skipLocationCheck) {
            // Otherwise check for geolocation permission
            setTimeout(requestLocationPermission, 1000);
        }
    }

    // Event Listeners
    if (searchForm) {
        searchForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const city = cityInput.value.trim();
            if (city) {
                loadWeatherData(city);
            }
        });
    }

    if (getLocationBtn) {
        getLocationBtn.addEventListener('click', function() {
            requestLocationPermission();
        });
    }

    if (allowLocationBtn) {
        allowLocationBtn.addEventListener('click', function() {
            locationAlert.classList.remove('show');
            getLocationWeather();
        });
    }

    if (denyLocationBtn) {
        denyLocationBtn.addEventListener('click', function() {
            locationAlert.classList.remove('show');
            // User denied location, do nothing
        });
    }
    // Function to initialize dashboard with date and animations
    function initializeDashboard() {
        updateCurrentDate();
        setupHourlyForecastInteraction();
        setupDailyForecastInteraction();

        // Set humidity indicator (default value)
        const humidityValue = parseInt(humidityValueElement.textContent);
        const indicatorFill = document.querySelector('.indicator-fill');
        if (indicatorFill) {
            indicatorFill.style.width = `${humidityValue}%`;
        }
    }

    // Function to update the current date display
    function updateCurrentDate() {
        const now = new Date();
        const formattedDate = `${now.getDate().toString().padStart(2, '0')}.${(now.getMonth() + 1).toString().padStart(2, '0')}.${now.getFullYear()}`;

        if (currentDateElement) {
            currentDateElement.textContent = formattedDate;
        }
    }

    // Setup hourly forecast interaction
    function setupHourlyForecastInteraction() {
        const timeSlots = document.querySelectorAll('.time-slot');
        timeSlots.forEach(slot => {
            slot.addEventListener('click', function() {
                // Remove active class from all slots
                timeSlots.forEach(s => s.classList.remove('active'));
                // Add active class to clicked slot
                this.classList.add('active');
            });
        });
    }

    // Setup daily forecast interaction
    function setupDailyForecastInteraction() {
        const dayRows = document.querySelectorAll('.day-row');
        dayRows.forEach(row => {
            row.addEventListener('click', function() {
                // Remove active class from all rows
                dayRows.forEach(r => r.classList.remove('active'));
                // Add active class to clicked row
                this.classList.add('active');
            });
        });
    }

    // Function to request location permission
    function requestLocationPermission() {
        if (!navigator.geolocation) {
            console.log("Geolocation is not supported by this browser.");
            return;
        }

        if (navigator.permissions && navigator.permissions.query) {
            navigator.permissions.query({ name: 'geolocation' })
                .then(function(permissionStatus) {
                    if (permissionStatus.state === 'granted') {
                        // Permission already granted, get location weather
                        getLocationWeather();
                    } else if (permissionStatus.state === 'prompt') {
                        // Show our custom permission dialog
                        locationAlert.classList.add('show');
                    } else {
                        // Permission denied previously
                        console.log("Geolocation permission denied.");
                    }
                });
        } else {
            // Older browsers - just try to get location
            getLocationWeather();
        }
    }

    // Function to get weather for current location
    function getLocationWeather() {
        loadingIndicator.classList.add('show');

        navigator.geolocation.getCurrentPosition(
            // Success callback
            function(position) {
                const latitude = position.coords.latitude;
                const longitude = position.coords.longitude;

                // Redirect to location options page with coordinates
                window.location.href = `/weather?lat=${latitude}&lon=${longitude}&options=true`;
            },
            // Error callback
            function(error) {
                loadingIndicator.classList.remove('show');
                console.error("Error getting location:", error.message);

                // Show an error message based on the error code
                let errorMessage = "";
                switch(error.code) {
                    case error.PERMISSION_DENIED:
                        errorMessage = "Location access was denied by the user.";
                        break;
                    case error.POSITION_UNAVAILABLE:
                        errorMessage = "Location information is unavailable.";
                        break;
                    case error.TIMEOUT:
                        errorMessage = "The request to get location timed out.";
                        break;
                    case error.UNKNOWN_ERROR:
                        errorMessage = "An unknown error occurred.";
                        break;
                }

                alert("Could not get your location: " + errorMessage);
            },
            // Options
            {
                enableHighAccuracy: true,
                timeout: 10000,
                maximumAge: 0
            }
        );
    }

    // Function to load weather data for a city
    function loadWeatherData(city) {
        loadingIndicator.classList.add('show');

        // Make API request to your weather endpoint
        fetch(`/api/weather?city=${encodeURIComponent(city)}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Failed to fetch weather data: ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                updateDashboardWithWeatherData(data);
                loadingIndicator.classList.remove('show');
            })
            .catch(error => {
                console.error('Error fetching weather data:', error);
                loadingIndicator.classList.remove('show');
                alert(`Error loading weather data: ${error.message}`);
            });
    }

    // Function to load weather data using coordinates
    function loadWeatherDataByCoordinates(lat, lon) {
        loadingIndicator.classList.add('show');

        // Make API request to your weather endpoint with coordinates
        fetch(`/api/weather?lat=${lat}&lon=${lon}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Failed to fetch weather data: ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                updateDashboardWithWeatherData(data);
                loadingIndicator.classList.remove('show');
            })
            .catch(error => {
                console.error('Error fetching weather data:', error);
                loadingIndicator.classList.remove('show');
                alert(`Error loading weather data: ${error.message}`);
            });
    }

    // Function to update dashboard with weather data
    function updateDashboardWithWeatherData(data) {
        try {
            // Update city input if empty
            if (!cityInput.value) {
                cityInput.value = data.city;
            }

            // Extract temperature value from string (remove °C)
            const tempString = data.current.temperature;
            const tempValue = parseFloat(tempString);

            // Update main temperature
            if (currentTempElement) {
                currentTempElement.textContent = Math.round(tempValue);
            }

            // Update weather type
            if (weatherTypeElement) {
                weatherTypeElement.textContent = capitalizeFirstLetter(data.current.condition);
            }

            // Update location name
            if (locationNameElement) {
                locationNameElement.textContent = data.city;
            }

            // Update current degree
            if (currentDegreeElement) {
                currentDegreeElement.textContent = Math.round(tempValue);
            }

            // Update cloudy degree (for demo, using same temp)
            if (cloudyDegreeElement) {
                cloudyDegreeElement.textContent = Math.round(tempValue);
            }

            // Update current weather label
            if (currentWeatherLabelElement) {
                currentWeatherLabelElement.textContent = capitalizeFirstLetter(data.current.condition);
            }

            // Update humidity
            if (humidityValueElement) {
                // Extract humidity value from string (remove %)
                const humidityString = data.current.humidity;
                const humidityValue = parseInt(humidityString);
                humidityValueElement.textContent = humidityValue;

                // Update humidity indicator
                const indicatorFill = document.querySelector('.indicator-fill');
                if (indicatorFill) {
                    indicatorFill.style.width = `${humidityValue}%`;
                }
            }

            // Update UV index indicator if available
            if (data.current.uvIndex) {
                updateUVIndicator(data.current.uvIndex);
            }

            // Update forecast
            if (data.forecast && data.forecast.length > 0) {
                updateForecast(data.forecast);
            }

            // Update weather icons based on condition
            updateWeatherIcons(data.current.condition);

            // Update background image based on condition
            updateBackgroundImage(data.current.condition);

        } catch (error) {
            console.error('Error updating dashboard:', error);
        }
    }

    // Function to update hourly and daily forecast
    function updateForecast(forecastData) {
        // Update hourly slots
        const timeSlots = document.querySelectorAll('.time-slot');

        // Only update if forecast data exists and we have time slots
        if (forecastData.length > 0 && timeSlots.length > 0) {
            // Start from index 1 (skip "Now" slot)
            for (let i = 1; i < timeSlots.length && i <= forecastData.length; i++) {
                const forecast = forecastData[i - 1];
                const slot = timeSlots[i];

                const tempElement = slot.querySelector('.time-temp');
                const iconElement = slot.querySelector('.time-icon i');

                if (tempElement) {
                    tempElement.textContent = `${Math.round(forecast.maxTemp)}°`;
                }

                if (iconElement) {
                    // Update icon based on weather description
                    updateWeatherIcon(iconElement, forecast.description);
                }
            }
        }

        // Update daily forecast
        const dayRows = document.querySelectorAll('.day-row');

        if (forecastData.length > 0 && dayRows.length > 0) {
            for (let i = 0; i < dayRows.length && i < forecastData.length; i++) {
                const forecast = forecastData[i];
                const row = dayRows[i];

                const dayName = row.querySelector('.day-name');
                const iconElement = row.querySelector('.day-icon i');
                const tempElement = row.querySelector('.day-temp');

                if (dayName) {
                    dayName.textContent = forecast.day;
                }

                if (iconElement) {
                    // Update icon based on weather description
                    updateWeatherIcon(iconElement, forecast.description);
                }

                if (tempElement) {
                    // Update with maxTemp and minTemp
                    tempElement.innerHTML = `${Math.round(forecast.maxTemp)}° <span class="low-temp">${Math.round(forecast.minTemp)}</span>`;
                }
            }
        }
    }

    // Function to update weather icons based on condition
    function updateWeatherIcon(iconElement, condition) {
        condition = condition.toLowerCase();

        // Remove existing classes
        iconElement.className = '';

        // Add appropriate class based on condition
        if (condition.includes('clear') || condition.includes('sunny')) {
            iconElement.className = 'fas fa-sun';
        } else if (condition.includes('partly cloudy') || condition.includes('broken clouds')) {
            iconElement.className = 'fas fa-cloud-sun';
        } else if (condition.includes('cloud')) {
            iconElement.className = 'fas fa-cloud';
        } else if (condition.includes('rain') || condition.includes('drizzle')) {
            iconElement.className = 'fas fa-cloud-rain';
        } else if (condition.includes('thunder') || condition.includes('lightning')) {
            iconElement.className = 'fas fa-bolt';
        } else if (condition.includes('snow')) {
            iconElement.className = 'fas fa-snowflake';
        } else if (condition.includes('mist') || condition.includes('fog')) {
            iconElement.className = 'fas fa-smog';
        } else {
            // Default
            iconElement.className = 'fas fa-cloud';
        }
    }

    // Function to update weather icons throughout the dashboard
    function updateWeatherIcons(condition) {
        // Update main weather icon (assuming there's one)
        const mainWeatherIcon = document.querySelector('.card-content i');
        if (mainWeatherIcon) {
            updateWeatherIcon(mainWeatherIcon, condition);
        }
    }

    // Function to update UV index indicator
    function updateUVIndicator(uvIndex) {
        const uvIndicator = document.querySelector('.uv-indicator');
        if (!uvIndicator) return;

        // Remove existing classes
        uvIndicator.className = 'uv-indicator';

        // Add appropriate class based on UV index
        if (uvIndex < 3) {
            uvIndicator.classList.add('uv-low');
            uvIndicator.textContent = 'Low';
        } else if (uvIndex < 6) {
            uvIndicator.classList.add('uv-moderate');
            uvIndicator.textContent = 'Moderate';
        } else if (uvIndex < 8) {
            uvIndicator.classList.add('uv-high');
            uvIndicator.textContent = 'High';
        } else if (uvIndex < 11) {
            uvIndicator.classList.add('uv-very-high');
            uvIndicator.textContent = 'Very High';
        } else {
            uvIndicator.classList.add('uv-extreme');
            uvIndicator.textContent = 'Extreme';
        }
    }

    /**
     * Enhanced function to update the background image based on weather condition
     * Also adjusts for time of day and card visibility
     * @param {string} condition - The weather condition description
     */
    function updateBackgroundImage(condition) {
        condition = condition.toLowerCase();
        let backgroundUrl = '';
        let isDarkBg = false;

        // Check if it's day or night
        const hour = new Date().getHours();
        const isNighttime = hour < 6 || hour > 18;

        // Select background based on weather condition
        if (condition.includes('clear') || condition.includes('sunny')) {
            if (isNighttime) {
                backgroundUrl = "https://images.unsplash.com/photo-1475274047050-1d0c0975c63e?ixid=MnwxMjA3fDB8MHxzZWFyY2h8M3x8bmlnaHQlMjBza3l8ZW58MHx8MHx8&ixlib=rb-1.2.1&w=1000&q=80";
                isDarkBg = true;
            } else {
                backgroundUrl = "https://images.unsplash.com/photo-1517758478390-c89333af4642?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
                isDarkBg = false;
            }
        } else if (condition.includes('partly cloudy') || condition.includes('few clouds')) {
            if (isNighttime) {
                backgroundUrl = "https://images.unsplash.com/photo-1505322715123-95d7a4986e50?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
                isDarkBg = true;
            } else {
                backgroundUrl = "https://images.unsplash.com/photo-1611928482473-7b27d24eab80?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
                isDarkBg = false;
            }
        } else if (condition.includes('cloud')) {
            backgroundUrl = "https://images.unsplash.com/photo-1534088568595-a066f410bcda?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = true;
        } else if (condition.includes('rain') || condition.includes('drizzle')) {
            backgroundUrl = "https://images.unsplash.com/photo-1534274988757-a28bf1a57c17?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = true;
        } else if (condition.includes('thunder') || condition.includes('storm')) {
            backgroundUrl = "https://images.unsplash.com/photo-1605727216801-e27ce1d0cc28?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = true;
        } else if (condition.includes('snow')) {
            backgroundUrl = "https://images.unsplash.com/photo-1483664852095-d6cc6870702d?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = false;
        } else if (condition.includes('fog') || condition.includes('mist')) {
            backgroundUrl = "https://images.unsplash.com/photo-1487621167305-5d248087c724?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = true;
        } else {
            // Default background
            backgroundUrl = "https://images.unsplash.com/photo-1504608524841-42fe6f032b4b?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=1200&q=80";
            isDarkBg = false;
        }

        // Apply the background image
        document.body.style.backgroundImage = `url('${backgroundUrl}')`;

        // Add time of day class
        if (isNighttime) {
            document.body.classList.add('night');
            document.body.classList.remove('day');
        } else {
            document.body.classList.add('day');
            document.body.classList.remove('night');
        }

        // Add background type class
        if (isDarkBg) {
            document.body.classList.add('dark-bg');
            document.body.classList.remove('light-bg');
        } else {
            document.body.classList.add('light-bg');
            document.body.classList.remove('dark-bg');
        }

        console.log(`Background updated for condition: ${condition}, night: ${isNighttime}, dark: ${isDarkBg}`);
    }

    // Utility function to capitalize first letter
    function capitalizeFirstLetter(string) {
        return string.charAt(0).toUpperCase() + string.slice(1);
    }
});