/**
 * improved-dashboard.js - JavaScript for the GoWeather Premium dashboard
 * Handles dynamic weather data loading and UI updates with enhanced functionality
 */

// Capture server-provided username before anything else
(function() {
    // Check if the page is using server-side rendered username
    const usernameElement = document.getElementById('username');
    const serverProvidedUsername = usernameElement ? usernameElement.textContent.trim() : null;

    // If server provided a non-default username, use it as the source of truth
    if (serverProvidedUsername && serverProvidedUsername !== 'User') {
        console.log('Using server-provided username:', serverProvidedUsername);
        localStorage.setItem('username', serverProvidedUsername);
    }
})();

// API Key - for OpenWeatherMap API
const API_KEY = "0c2e2084bdd01a671b1b450215191f89";

// Global state for weather data
let currentWeatherData = null;
let hourlyForecastData = [];
let dailyForecastData = [];
let weatherMap = null;
let currentMapLayer = null;
let weatherTileLayer = null;
let locationMarker = null;
let locationPopup = null;

// Hourly forecast variables
let currentHourlyPage = 0; // 0 = current 24h, -1 = previous 24h, 1 = next 24h
const hoursPerPage = 24;
let allHourlyData = []; // Will store 72 hours of data (past 24h, current 24h, future 24h)

document.addEventListener('DOMContentLoaded', function() {
    // Elements
    const searchForm = document.getElementById('searchForm');
    const cityInput = document.getElementById('cityInput');
    const getLocationBtn = document.getElementById('getLocation');
    const loadingIndicator = document.getElementById('loadingIndicator');
    const forecastDaysSelect = document.getElementById('forecastDays');
    const forecastDaysDisplay = document.getElementById('forecastDaysDisplay');

    // Initialize dashboard
    initializeDashboard();

    // Check URL parameters for forecast days
    const urlParams = new URLSearchParams(window.location.search);
    const forecastDays = urlParams.get('forecast_days');

    // Set forecast days select value if provided in URL
    if (forecastDays && forecastDaysSelect) {
        forecastDaysSelect.value = forecastDays;
        if (forecastDaysDisplay) {
            if (forecastDays === "30") {
                forecastDaysDisplay.textContent = "1 month";
            } else {
                forecastDaysDisplay.textContent = forecastDays;
            }
        }
    }

    // Load weather data based on URL parameters
    const city = urlParams.get('city');
    const lat = urlParams.get('lat');
    const lon = urlParams.get('lon');

    if (city) {
        loadWeatherData(city);
    } else if (lat && lon) {
        loadWeatherDataByCoordinates(lat, lon);
    } else {
        const homeCity = cityInput && cityInput.value ? cityInput.value : '';
        if (homeCity) {
            loadWeatherData(homeCity);
        } else {
            // Try to get user location after a short delay
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

    if (forecastDaysSelect) {
        forecastDaysSelect.addEventListener('change', function() {
            const value = this.value;

            if (forecastDaysDisplay) {
                if (value === "30") {
                    forecastDaysDisplay.textContent = "1 month";
                } else {
                    forecastDaysDisplay.textContent = value;
                }
            }

            // If we have a city, reload with new forecast days
            const city = cityInput.value.trim();
            if (city) {
                // Update URL with forecast days
                const url = new URL(window.location);
                url.searchParams.set('forecast_days', value);
                window.history.pushState({}, '', url);

                loadWeatherData(city);
            }
        });
    }

    // Hourly forecast navigation event listeners
    const prevHoursBtn = document.getElementById('prevHours');
    const nextHoursBtn = document.getElementById('nextHours');
    const hourlyRangeDisplay = document.getElementById('hourlyRangeDisplay');

    if (prevHoursBtn) {
        prevHoursBtn.addEventListener('click', function() {
            if (currentHourlyPage > -1) {
                currentHourlyPage--;
                displayCurrentHourlyPage();
                updateHourlyNavigationStatus();
            }
        });
    }

    if (nextHoursBtn) {
        nextHoursBtn.addEventListener('click', function() {
            if (currentHourlyPage < 1) {
                currentHourlyPage++;
                displayCurrentHourlyPage();
                updateHourlyNavigationStatus();
            }
        });
    }

    // Initialize map layer selection
    const mapLayerSelect = document.getElementById('mapLayer');
    if (mapLayerSelect) {
        mapLayerSelect.addEventListener('change', function() {
            addWeatherLayer(this.value);
            updateLegendVisibility(this.value);
        });
    }

    // Setup profile dropdown
    setupProfileDropdown();

    // Set up dark mode toggle
    setupDarkModeToggle();
});

// Function to initialize dashboard
function initializeDashboard() {
    updateCurrentDate();
    setupHourlyForecasting();
    setupDailyForecastInteraction();
    setupCityAutocomplete();

    // Set default humidity indicator
    const humidityValueElement = document.getElementById('humidityValue');
    const humidityValue = parseInt(humidityValueElement?.textContent || "50");
    const indicatorFill = document.querySelector('.indicator-fill');
    if (indicatorFill) {
        indicatorFill.style.width = `${humidityValue}%`;
    }

    // Initialize premium features
    initPremiumFeatures();

    // Initialize username display
    initializeUsername();

    // Handle device-specific behavior
    handleDeviceSpecificBehavior();

    // Update sunrise-sunset indicator
    updateSunPositionIndicator();
}

// Function to initialize premium features
function initPremiumFeatures() {
    // Add premium badges and highlight premium features
    const premiumFeatures = document.querySelectorAll('.premium-feature');
    premiumFeatures.forEach(feature => {
        feature.classList.add('premium-highlight');
        setTimeout(() => {
            feature.classList.remove('premium-highlight');
        }, 2000);
    });

    // Update metrics section with current weather data
    updateWeatherMetrics();
}

// Function to update weather metrics section
function updateWeatherMetrics() {
    if (!currentWeatherData) return;

    // Set metrics values
    const metricUv = document.getElementById('metricUv');
    const metricFeelsLike = document.getElementById('metricFeelsLike');
    const metricHumidity = document.getElementById('metricHumidity');
    const metricWind = document.getElementById('metricWind');
    const metricPressure = document.getElementById('metricPressure');
    const metricVisibility = document.getElementById('metricVisibility');

    if (metricUv) metricUv.textContent = currentWeatherData.current.uvIndex.toFixed(1);
    if (metricFeelsLike) metricFeelsLike.textContent = Math.round(currentWeatherData.current.feelsLike) + '°';
    if (metricHumidity) metricHumidity.textContent = currentWeatherData.current.humidity + '%';
    if (metricWind) metricWind.textContent = Math.round(currentWeatherData.current.windSpeed) + ' mph';
    if (metricPressure) metricPressure.textContent = currentWeatherData.current.pressure;
    if (metricVisibility) {
        const visibilityMiles = (currentWeatherData.current.visibility / 1609).toFixed(1);
        metricVisibility.textContent = visibilityMiles;
    }
}

// Updated initializeUsername function that respects server-provided username
// Updated initializeUsername function that handles profile photo
function initializeUsername() {
    console.log('initializeUsername called');
    const usernameElement = document.getElementById('username');
    const profilePhotoElement = document.getElementById('profile-photo');

    if (!usernameElement) return;

    // Check if there's a server-provided username that's not the default
    const serverUsername = usernameElement.textContent.trim();
    const isServerCustomUsername = serverUsername !== 'User';

    // Get username from localStorage as backup
    let savedUsername = localStorage.getItem('username');

    // Determine the username to display
    let displayUsername;

    if (isServerCustomUsername) {
        // If server provided a custom username, use it and update localStorage
        displayUsername = serverUsername;
        localStorage.setItem('username', displayUsername);
    } else if (savedUsername && savedUsername !== 'User' && savedUsername !== 'GoWeather User') {
        // Otherwise use localStorage if it has a custom value
        displayUsername = savedUsername;
    } else {
        // Default fallback
        displayUsername = 'User';
        localStorage.setItem('username', displayUsername);
    }

    // Set the username in the UI
    usernameElement.textContent = displayUsername;

    console.log('Username initialized to:', displayUsername);
}

function updateUsername(newUsername) {
    if (!newUsername || newUsername.trim() === '') return;

    newUsername = newUsername.trim();

    // Update in localStorage
    localStorage.setItem('username', newUsername);

    // Update UI elements
    const usernameElement = document.getElementById('username');
    const profileInitialElement = document.getElementById('profile-initial');

    if (usernameElement) {
        usernameElement.textContent = newUsername;
    }

    if (profileInitialElement) {
        profileInitialElement.textContent = newUsername.charAt(0).toUpperCase();
    }
}

function setupUsernameChange() {
    const changeUsernameButton = document.getElementById('changeUsernameBtn');

    if (changeUsernameButton) {
        changeUsernameButton.addEventListener('click', function() {
            const newUsername = prompt('Enter your username:', localStorage.getItem('username') || '');
            if (newUsername) {
                updateUsername(newUsername);
            }
        });
    }
}

// Function to update sun position on the progress bar
function updateSunPositionIndicator(sunriseTime = '06:53', sunsetTime = '19:53') {
    // Get current time
    const now = new Date();
    const hours = now.getHours();
    const minutes = now.getMinutes();

    // Update the display times
    const sunriseLargeTime = document.getElementById('sunriseLargeTime');
    const sunsetLargeTime = document.getElementById('sunsetLargeTime');

    if (sunriseLargeTime) sunriseLargeTime.textContent = convertTo12HourFormat(sunriseTime);
    if (sunsetLargeTime) sunsetLargeTime.textContent = convertTo12HourFormat(sunsetTime);

    // Parse the sunrise and sunset times
    const [sunriseHour, sunriseMinute] = (typeof sunriseTime === 'string' ? sunriseTime : '06:53').split(':').map(Number);
    const [sunsetHour, sunsetMinute] = (typeof sunsetTime === 'string' ? sunsetTime : '19:53').split(':').map(Number);

    // Calculate day start, end, and length
    const dayStart = sunriseHour + (sunriseMinute / 60);
    const dayEnd = sunsetHour + (sunsetMinute / 60);
    const dayLength = dayEnd - dayStart;
    const currentTime = hours + (minutes / 60);

    // Calculate the position percentage
    let positionPercent = 0;
    if (currentTime < dayStart) {
        positionPercent = 0;
    } else if (currentTime > dayEnd) {
        positionPercent = 100;
    } else {
        positionPercent = ((currentTime - dayStart) / dayLength) * 100;
    }

    // Update the position of the indicator
    const dayProgressIndicator = document.querySelector('.day-progress-indicator');
    if (dayProgressIndicator) {
        dayProgressIndicator.style.left = `${positionPercent}%`;
    }
}

// Helper function to convert 24h time to 12h time
function convertTo12HourFormat(time24h) {
    if (!time24h || time24h === '--:--') return '--:--';

    const [hours, minutes] = time24h.split(':').map(Number);
    const period = hours >= 12 ? 'PM' : 'AM';
    const hours12 = hours % 12 || 12;

    return `${hours12}:${minutes.toString().padStart(2, '0')} ${period}`;
}

// Function to update current date display
function updateCurrentDate() {
    const currentDateElement = document.getElementById('currentDate');
    if (!currentDateElement) return;

    const now = new Date();
    const formattedDate = `${now.getDate().toString().padStart(2, '0')}.${(now.getMonth() + 1).toString().padStart(2, '0')}.${now.getFullYear()}`;
    currentDateElement.textContent = formattedDate;
}

// Setup hourly forecast display
function setupHourlyForecasting() {
    const hourlyForecastContainer = document.getElementById('hourlyForecast');
    if (!hourlyForecastContainer) return;

    // Clear existing content
    hourlyForecastContainer.innerHTML = '';

    // Get current hour
    const now = new Date();
    const currentHour = now.getHours();

    // Generate placeholder data for 24 hours
    for (let i = 0; i < 24; i++) {
        // Calculate hour for this slot
        const slotHour = (currentHour + i) % 24;
        const slotTime = `${slotHour.toString().padStart(2, '0')}:00`;

        // Create time slot
        const timeSlot = document.createElement('div');
        timeSlot.className = `time-slot ${i === 0 ? 'active' : ''}`;
        if (i === 0) timeSlot.classList.add('current-hour');

        // Set placeholder content
        timeSlot.innerHTML = `
            <div class="time-label">${i === 0 ? 'Now' : slotTime}</div>
            <div class="time-temp">--°</div>
            <div class="time-icon"><i class="fas fa-cloud"></i></div>
            <div class="precipitation">--%</div>
        `;

        // Add to container
        hourlyForecastContainer.appendChild(timeSlot);

        // Add click event
        timeSlot.addEventListener('click', function() {
            // Remove active class from all slots
            document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('active'));
            // Add active class to clicked slot
            this.classList.add('active');
        });
    }
}

// Setup daily forecast interaction
function setupDailyForecastInteraction() {
    const dailyForecastContainer = document.getElementById('dailyForecast');
    if (!dailyForecastContainer) return;

    // Clear existing content
    dailyForecastContainer.innerHTML = '';

    // Create 7 placeholder days (default)
    const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    const today = new Date().getDay();

    for (let i = 0; i < 7; i++) {
        const dayIndex = (today + i) % 7;
        const dayName = days[dayIndex];

        // Create day row
        const dayRow = document.createElement('div');
        dayRow.className = `day-row ${i === 0 ? 'active' : ''}`;

        // Set content
        dayRow.innerHTML = `
            <div class="day-name">${i === 0 ? 'Today' : dayName}</div>
            <div class="day-icon"><i class="fas fa-cloud"></i></div>
            <div class="day-temp">--° <span class="low-temp">--°</span></div>
            <div class="day-details premium-feature">
                <div class="precipitation-chance"><i class="fas fa-tint"></i> --%</div>
                <div class="wind-speed"><i class="fas fa-wind"></i> -- km/h</div>
            </div>
        `;

        // Add to container
        dailyForecastContainer.appendChild(dayRow);

        // Add click event
        dayRow.addEventListener('click', function() {
            // Remove active class from all rows
            document.querySelectorAll('.day-row').forEach(r => r.classList.remove('active'));
            // Add active class to clicked row
            this.classList.add('active');
        });
    }
}

// Function to set up city autocomplete
function setupCityAutocomplete() {
    const cityInput = document.getElementById('cityInput');
    if (!cityInput) return;

    // Create datalist if it doesn't exist
    let cityOptions = document.getElementById('cityOptions');
    if (!cityOptions) {
        cityOptions = document.createElement('datalist');
        cityOptions.id = 'cityOptions';
        document.body.appendChild(cityOptions);
        cityInput.setAttribute('list', 'cityOptions');
    }

    // Add common cities
    const commonCities = ['London', 'New York', 'Tokyo', 'Paris', 'Berlin',
        'Sydney', 'Beijing', 'Moscow', 'Rome', 'Madrid',
        'Dubai', 'Mumbai', 'São Paulo', 'Toronto', 'Cairo'];

    commonCities.forEach(city => {
        const option = document.createElement('option');
        option.value = city;
        cityOptions.appendChild(option);
    });

    // Add input event to fetch suggestions
    cityInput.addEventListener('input', function() {
        const query = this.value.trim();
        if (query.length > 2) {
            fetchCitySuggestions(query);
        }
    });
}

// Function to request location permission
function requestLocationPermission() {
    if (!navigator.geolocation) {
        console.log("Geolocation is not supported by this browser.");
        return;
    }

    getLocationWeather();
}

// Function to get weather for current location
function getLocationWeather() {
    const loadingIndicator = document.getElementById('loadingIndicator');
    if (loadingIndicator) {
        loadingIndicator.style.display = 'flex';
        loadingIndicator.classList.add('show');
    }

    navigator.geolocation.getCurrentPosition(
        // Success callback
        function(position) {
            const latitude = position.coords.latitude;
            const longitude = position.coords.longitude;

            // Update URL with coordinates
            const url = new URL(window.location);
            url.searchParams.set('lat', latitude);
            url.searchParams.set('lon', longitude);
            window.history.pushState({}, '', url);

            // Load weather data directly instead of redirecting
            loadWeatherDataByCoordinates(latitude, longitude);
        },
        // Error callback
        function(error) {
            if (loadingIndicator) {
                loadingIndicator.style.display = 'none';
                loadingIndicator.classList.remove('show');
            }
            console.error("Error getting location:", error.message);

            let errorMessage = "Could not get your location.";
            alert(errorMessage);
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
    const loadingIndicator = document.getElementById('loadingIndicator');
    if (loadingIndicator) {
        loadingIndicator.style.display = 'flex';
        loadingIndicator.classList.add('show');
    }

    // Get number of forecast days
    const forecastDaysSelect = document.getElementById('forecastDays');
    const forecastDays = forecastDaysSelect?.value || 7;

    // Update URL with the city parameter without reloading the page
    const url = new URL(window.location);
    url.searchParams.set('city', city);
    url.searchParams.set('forecast_days', forecastDays);
    window.history.pushState({}, '', url);

    // Make API request to OpenWeatherMap directly
    fetchWeatherData(city)
        .then(data => {
            updateDashboardWithWeatherData(data);
            if (loadingIndicator) {
                loadingIndicator.style.display = 'none';
                loadingIndicator.classList.remove('show');
            }
        })
        .catch(error => {
            console.error('Error fetching weather data:', error);
            if (loadingIndicator) {
                loadingIndicator.style.display = 'none';
                loadingIndicator.classList.remove('show');
            }
            alert(`Error loading weather data: ${error.message}`);
        });
}

// Function to load weather data using coordinates
function loadWeatherDataByCoordinates(lat, lon) {
    const loadingIndicator = document.getElementById('loadingIndicator');
    if (loadingIndicator) {
        loadingIndicator.style.display = 'flex';
        loadingIndicator.classList.add('show');
    }

    // Get number of forecast days
    const forecastDaysSelect = document.getElementById('forecastDays');
    const forecastDays = forecastDaysSelect?.value || 7;

    // Update URL with forecast days
    const url = new URL(window.location);
    url.searchParams.set('forecast_days', forecastDays);
    window.history.pushState({}, '', url);

    // Make API request to OpenWeatherMap directly with coordinates
    fetchWeatherDataByCoordinates(lat, lon)
        .then(data => {
            // Update city input with the fetched city name
            const cityInput = document.getElementById('cityInput');
            if (cityInput && data.city) {
                cityInput.value = data.city;
            }

            updateDashboardWithWeatherData(data);
            if (loadingIndicator) {
                loadingIndicator.style.display = 'none';
                loadingIndicator.classList.remove('show');
            }
        })
        .catch(error => {
            console.error('Error fetching weather data:', error);
            if (loadingIndicator) {
                loadingIndicator.style.display = 'none';
                loadingIndicator.classList.remove('show');
            }
            alert(`Error loading weather data: ${error.message}`);
        });
}

// Function to fetch weather data from OpenWeatherMap API for a city
async function fetchWeatherData(city) {
    try {
        // Step 1: Get coordinates from city name
        const geoUrl = `https://api.openweathermap.org/geo/1.0/direct?q=${encodeURIComponent(city)}&limit=1&appid=${API_KEY}`;
        const geoResponse = await fetch(geoUrl);

        if (!geoResponse.ok) {
            throw new Error(`Failed to fetch location data: ${geoResponse.status}`);
        }

        const geoData = await geoResponse.json();

        if (!geoData || geoData.length === 0) {
            throw new Error('Location not found');
        }

        const { lat, lon, name, country } = geoData[0];

        // Continue with coordinates
        return fetchWeatherDataByCoordinates(lat, lon, name, country);
    } catch (error) {
        console.error('Error in fetchWeatherData:', error);
        throw error;
    }
}

// Function to fetch weather data from coordinates
async function fetchWeatherDataByCoordinates(lat, lon, cityNameOverride = null, countryCode = null) {
    try {
        // Get number of forecast days
        const forecastDaysSelect = document.getElementById('forecastDays');
        const forecastDays = forecastDaysSelect?.value || 7;

        // Step 1: Get current weather
        const currentWeatherUrl = `https://api.openweathermap.org/data/2.5/weather?lat=${lat}&lon=${lon}&units=metric&appid=${API_KEY}`;
        const currentWeatherResponse = await fetch(currentWeatherUrl);

        if (!currentWeatherResponse.ok) {
            throw new Error(`Failed to fetch current weather: ${currentWeatherResponse.status}`);
        }

        const currentWeather = await currentWeatherResponse.json();

        // Step 2: Get One Call API data (hourly and daily forecasts)
        const oneCallUrl = `https://api.openweathermap.org/data/3.0/onecall?lat=${lat}&lon=${lon}&units=metric&exclude=minutely&appid=${API_KEY}`;
        const oneCallResponse = await fetch(oneCallUrl);

        if (!oneCallResponse.ok) {
            throw new Error(`Failed to fetch forecast data: ${oneCallResponse.status}`);
        }

        const oneCallData = await oneCallResponse.json();

        // Step 3: Get Air Quality data
        const airQualityUrl = `https://api.openweathermap.org/data/2.5/air_pollution?lat=${lat}&lon=${lon}&appid=${API_KEY}`;
        const airQualityResponse = await fetch(airQualityUrl);

        let airQualityData = null;
        if (airQualityResponse.ok) {
            airQualityData = await airQualityResponse.json();
        }

        // If we need extended forecast (beyond what OneCall provides)
        let extendedForecast = null;
        if (parseInt(forecastDays) > 8) {
            try {
                // For forecasts > 8 days, we'll use the 16-day forecast endpoint
                // Note: This is a fallback for illustration purposes
                const extendedForecastUrl = `https://api.openweathermap.org/data/2.5/forecast/daily?lat=${lat}&lon=${lon}&cnt=${forecastDays}&units=metric&appid=${API_KEY}`;
                const extendedResponse = await fetch(extendedForecastUrl);

                if (extendedResponse.ok) {
                    extendedForecast = await extendedResponse.json();
                }
            } catch (extendedError) {
                console.warn('Could not fetch extended forecast, falling back to available data:', extendedError);
                // We'll fall back to the data we have from OneCall
            }
        }

        // Step 4: Process all data into a unified format
        return processWeatherData(currentWeather, oneCallData, airQualityData, cityNameOverride, parseInt(forecastDays), extendedForecast);
    } catch (error) {
        console.error('Error in fetchWeatherDataByCoordinates:', error);
        throw error;
    }
}

// Function to process all weather data into a unified format
function processWeatherData(currentWeather, oneCallData, airQualityData, cityNameOverride = null, forecastDays = 7, extendedForecast = null) {
    // Use city name from currentWeather if not overridden
    const cityName = cityNameOverride || currentWeather.name;

    // Process current weather
    const currentData = {
        temperature: currentWeather.main.temp,
        feelsLike: currentWeather.main.feels_like,
        humidity: currentWeather.main.humidity,
        pressure: currentWeather.main.pressure,
        condition: currentWeather.weather[0]?.description || 'Unknown',
        icon: currentWeather.weather[0]?.icon || '01d',
        windSpeed: currentWeather.wind?.speed || 0,
        windDeg: currentWeather.wind?.deg || 0,
        windDirection: getWindDirection(currentWeather.wind?.deg || 0),
        windGust: currentWeather.wind?.gust || 0,
        countryCode: currentWeather.sys?.country || '',
        sunrise: formatTimeWithTimezone(currentWeather.sys?.sunrise || 0, currentWeather.timezone || 0),
        sunset: formatTimeWithTimezone(currentWeather.sys?.sunset || 0, currentWeather.timezone || 0),
        visibility: currentWeather.visibility || 0,
        timezone: formatTimezone(currentWeather.timezone || 0),
        lat: currentWeather.coord?.lat || 0,
        lon: currentWeather.coord?.lon || 0
    };

    // Calculate day length
    if (currentWeather.sys?.sunrise && currentWeather.sys?.sunset) {
        const sunriseTime = currentWeather.sys.sunrise * 1000;
        const sunsetTime = currentWeather.sys.sunset * 1000;
        const dayLengthMs = sunsetTime - sunriseTime;
        const dayLengthHours = Math.floor(dayLengthMs / (1000 * 60 * 60));
        const dayLengthMinutes = Math.floor((dayLengthMs % (1000 * 60 * 60)) / (1000 * 60));
        currentData.dayLength = `${dayLengthHours}h ${dayLengthMinutes}m`;
    } else {
        currentData.dayLength = '--h --m';
    }

    // Process UV index from OneCall data
    if (oneCallData && oneCallData.current) {
        currentData.uvIndex = oneCallData.current.uvi || 0;
        const { status, message } = getUVStatus(currentData.uvIndex);
        currentData.uvStatus = status;
        currentData.uvMessage = message;
    } else {
        currentData.uvIndex = 0;
        currentData.uvStatus = 'Low';
        currentData.uvMessage = 'No protection needed';
    }

    // Process Air Quality data
    if (airQualityData && airQualityData.list && airQualityData.list.length > 0) {
        const aqi = airQualityData.list[0].main.aqi;
        currentData.aqi = aqi;
        currentData.aqiStatus = getAQIStatus(aqi);
        currentData.pm25 = airQualityData.list[0].components.pm2_5?.toFixed(1) || '--';
        currentData.pm10 = airQualityData.list[0].components.pm10?.toFixed(1) || '--';
    } else {
        currentData.aqi = 1;
        currentData.aqiStatus = 'Good';
        currentData.pm25 = '--';
        currentData.pm10 = '--';
    }

    // Process moon phase
    if (oneCallData && oneCallData.daily && oneCallData.daily.length > 0) {
        currentData.moonPhase = getMoonPhaseDescription(oneCallData.daily[0].moon_phase || 0);
    } else {
        currentData.moonPhase = '--';
    }

    // Process precipitation data
    currentData.precipAmount = 0;
    currentData.rainAmount = 0;
    currentData.snowAmount = 0;
    currentData.precipProb = 0;

    if (oneCallData && oneCallData.daily && oneCallData.daily.length > 0) {
        // Get precipitation probability from the first day
        currentData.precipProb = Math.round((oneCallData.daily[0].pop || 0) * 100);

        // Get rain and snow amounts if available
        if (oneCallData.daily[0].rain) {
            currentData.rainAmount = oneCallData.daily[0].rain;
            currentData.precipAmount += oneCallData.daily[0].rain;
        }

        if (oneCallData.daily[0].snow) {
            currentData.snowAmount = oneCallData.daily[0].snow;
            currentData.precipAmount += oneCallData.daily[0].snow;
        }
    }

    // Process hourly forecast
    let hourlyForecast = [];
    if (oneCallData && oneCallData.hourly) {
        hourlyForecast = oneCallData.hourly.map(hour => {
            return {
                time: formatTime(hour.dt * 1000),
                temp: hour.temp,
                icon: getWeatherIconClass(hour.weather[0]?.description || 'clear'),
                precipProb: Math.round((hour.pop || 0) * 100),
                weatherId: hour.weather[0]?.id || 800,
                weatherIcon: hour.weather[0]?.icon || '01d'
            };
        });
    }

    // Process daily forecast
    let dailyForecast = [];

    // Get today's date for proper day sequencing
    const today = new Date();

    if (oneCallData && oneCallData.daily) {
        // First process the data we have from OneCall (usually 7-8 days)
        dailyForecast = oneCallData.daily.map((day, index) => {
            // Create a date based on today plus index days
            const forecastDate = new Date(today);
            forecastDate.setDate(today.getDate() + index);

            // Get the proper day name
            const dayName = index === 0 ? 'Today' : getDayName(forecastDate);

            return {
                day: dayName,
                date: formatDate(forecastDate),
                fullDate: formatFullDate(forecastDate),
                maxTemp: day.temp.max,
                minTemp: day.temp.min,
                humidity: day.humidity,
                pressure: day.pressure,
                description: day.weather[0]?.description || 'Unknown',
                uvIndex: day.uvi || 0,
                windSpeed: day.wind_speed || 0,
                windDirection: getWindDirection(day.wind_deg || 0),
                precipProb: Math.round((day.pop || 0) * 100),
                sunrise: formatTimeWithTimezone(day.sunrise, currentWeather.timezone || 0),
                sunset: formatTimeWithTimezone(day.sunset, currentWeather.timezone || 0),
                moonPhase: getMoonPhaseDescription(day.moon_phase || 0),
                weatherId: day.weather[0]?.id || 800,
                weatherIcon: day.weather[0]?.icon || '01d'
            };
        });
    }

    // If we need more days than provided by OneCall, add them
    if (forecastDays > dailyForecast.length) {
        // If we have extended forecast data, use it
        if (extendedForecast && extendedForecast.list) {
            // Process extended forecast data (if available)
            const existingDates = new Set(dailyForecast.map(d => d.fullDate));

            extendedForecast.list.forEach((day, index) => {
                // Calculate the correct date based on today + offset
                const forecastDate = new Date(today);
                forecastDate.setDate(today.getDate() + dailyForecast.length + index);

                const fullDate = formatFullDate(forecastDate);

                // Only add if not already in our forecast
                if (!existingDates.has(fullDate)) {
                    dailyForecast.push({
                        day: getDayName(forecastDate),
                        date: formatDate(forecastDate),
                        fullDate: fullDate,
                        maxTemp: day.temp.max,
                        minTemp: day.temp.min,
                        humidity: day.humidity,
                        pressure: day.pressure,
                        description: day.weather[0]?.description || 'Unknown',
                        uvIndex: 5, // Default UV index if not available
                        windSpeed: day.speed || 0,
                        windDirection: getWindDirection(day.deg || 0),
                        precipProb: Math.round((day.pop || 0) * 100),
                        sunrise: '--:--', // Not available in extended forecast
                        sunset: '--:--',  // Not available in extended forecast
                        moonPhase: '--',  // Not available in extended forecast
                        weatherId: day.weather[0]?.id || 800,
                        weatherIcon: day.weather[0]?.icon || '01d'
                    });
                }
            });
        } else {
            // Generate estimated data for missing days
            const lastAvailableDay = dailyForecast[dailyForecast.length - 1];

            for (let i = dailyForecast.length; i < forecastDays; i++) {
                // Add proper day based on sequence from today
                const forecastDate = new Date(today);
                forecastDate.setDate(today.getDate() + i);

                // Create extrapolated forecast with minor variations
                dailyForecast.push({
                    day: getDayName(forecastDate),
                    date: formatDate(forecastDate),
                    fullDate: formatFullDate(forecastDate),
                    maxTemp: lastAvailableDay.maxTemp + ((Math.random() * 4) - 2), // +/- 2°C variation
                    minTemp: lastAvailableDay.minTemp + ((Math.random() * 3) - 1.5), // +/- 1.5°C variation
                    humidity: lastAvailableDay.humidity + Math.floor((Math.random() * 10) - 5), // +/- 5% variation
                    pressure: lastAvailableDay.pressure,
                    description: lastAvailableDay.description,
                    uvIndex: lastAvailableDay.uvIndex,
                    windSpeed: lastAvailableDay.windSpeed + ((Math.random() * 2) - 1), // +/- 1 m/s variation
                    windDirection: lastAvailableDay.windDirection,
                    precipProb: Math.min(Math.max(lastAvailableDay.precipProb + Math.floor((Math.random() * 20) - 10), 0), 100), // +/- 10% variation
                    sunrise: lastAvailableDay.sunrise,
                    sunset: lastAvailableDay.sunset,
                    moonPhase: getMoonPhaseForDate(forecastDate),
                    weatherId: lastAvailableDay.weatherId,
                    weatherIcon: lastAvailableDay.weatherIcon
                });
            }
        }
    }

    // Process weather alerts
    let alerts = [];
    if (oneCallData && oneCallData.alerts) {
        alerts = oneCallData.alerts.map(alert => {
            return {
                title: alert.event,
                description: alert.description,
                time: formatTime(alert.start * 1000),
                severity: getSeverityClass(alert.event)
            };
        });
    }
    // Return complete weather data
    return {
        city: cityName,
        current: currentData,
        hourlyForecast,
        forecast: dailyForecast.slice(0, forecastDays), // Limit to requested number of days
        alerts
    };
}

// Function to update dashboard with weather data
function updateDashboardWithWeatherData(data) {
    try {
        // Store the data globally
        currentWeatherData = data;

        // Update city input
        const cityInput = document.getElementById('cityInput');
        if (cityInput) {
            cityInput.value = data.city;
        }

        // Update city name and location details
        const locationNameElement = document.getElementById('locationName');
        if (locationNameElement) {
            locationNameElement.textContent = data.city;
        }

        const countryCodeElement = document.getElementById('countryCode');
        if (countryCodeElement && data.current.countryCode) {
            countryCodeElement.textContent = data.current.countryCode;
        }

        const coordinatesElement = document.getElementById('coordinates');
        if (coordinatesElement) {
            coordinatesElement.textContent = `${data.current.lat.toFixed(4)}, ${data.current.lon.toFixed(4)}`;
        }

        const timezoneElement = document.getElementById('timezone');
        if (timezoneElement) {
            timezoneElement.textContent = data.current.timezone;
        }

        // Update temperature and weather condition
        const currentTempElement = document.getElementById('currentTemp');
        if (currentTempElement) {
            currentTempElement.textContent = Math.round(data.current.temperature);
        }

        const weatherTypeElement = document.getElementById('weatherType');
        if (weatherTypeElement) {
            weatherTypeElement.textContent = data.current.condition;
        }

        const feelsLikeElement = document.getElementById('feelsLike');
        if (feelsLikeElement) {
            feelsLikeElement.textContent = Math.round(data.current.feelsLike);
        }

        // Update humidity
        const humidityValueElement = document.getElementById('humidityValue');
        if (humidityValueElement) {
            humidityValueElement.textContent = data.current.humidity;

            // Update humidity indicator
            const indicatorFill = document.querySelector('.indicator-fill');
            if (indicatorFill) {
                indicatorFill.style.width = `${data.current.humidity}%`;
            }
        }

        // Update sun and moon information
        const sunriseTimeElement = document.getElementById('sunriseTime');
        if (sunriseTimeElement) {
            sunriseTimeElement.textContent = data.current.sunrise;
        }

        const sunsetTimeElement = document.getElementById('sunsetTime');
        if (sunsetTimeElement) {
            sunsetTimeElement.textContent = data.current.sunset;
        }

        const dayLengthElement = document.getElementById('dayLength');
        if (dayLengthElement) {
            dayLengthElement.textContent = data.current.dayLength;
        }

        const moonPhaseElement = document.getElementById('moonPhase');
        if (moonPhaseElement) {
            moonPhaseElement.textContent = data.current.moonPhase;
        }

        // Update sunrise/sunset times in the large display
        const sunriseLargeTimeElement = document.getElementById('sunriseLargeTime');
        const sunsetLargeTimeElement = document.getElementById('sunsetLargeTime');

        if (sunriseLargeTimeElement) {
            // Convert 24h to 12h format for display
            const sunriseParts = data.current.sunrise.split(':');
            let sunriseHour = parseInt(sunriseParts[0]);
            const sunriseMinute = sunriseParts[1];
            const sunriseAmPm = sunriseHour >= 12 ? 'PM' : 'AM';
            sunriseHour = sunriseHour % 12 || 12;
            sunriseLargeTimeElement.textContent = `${sunriseHour}:${sunriseMinute} ${sunriseAmPm}`;
        }

        if (sunsetLargeTimeElement) {
            // Convert 24h to 12h format for display
            const sunsetParts = data.current.sunset.split(':');
            let sunsetHour = parseInt(sunsetParts[0]);
            const sunsetMinute = sunsetParts[1];
            const sunsetAmPm = sunsetHour >= 12 ? 'PM' : 'AM';
            sunsetHour = sunsetHour % 12 || 12;
            sunsetLargeTimeElement.textContent = `${sunsetHour}:${sunsetMinute} ${sunsetAmPm}`;
        }

        // Update sun position indicator with actual sunrise/sunset times
        updateSunPositionIndicator(data.current.sunrise, data.current.sunset);

        // Update UV index
        const uvValueElement = document.getElementById('uvValue');
        if (uvValueElement) {
            uvValueElement.textContent = data.current.uvIndex.toFixed(1);
        }

        const uvStatusElement = document.getElementById('uvStatus');
        if (uvStatusElement) {
            uvStatusElement.textContent = data.current.uvStatus;
            uvStatusElement.dataset.status = data.current.uvStatus;
        }

        const uvMessageElement = document.getElementById('uvMessage');
        if (uvMessageElement) {
            uvMessageElement.textContent = data.current.uvMessage;
        }

        // Update air quality
        const aqiValueElement = document.getElementById('aqiValue');
        if (aqiValueElement) {
            aqiValueElement.textContent = data.current.aqi;
        }

        const aqiStatusElement = document.getElementById('aqiStatus');
        if (aqiStatusElement) {
            aqiStatusElement.textContent = data.current.aqiStatus;
            aqiStatusElement.dataset.status = data.current.aqiStatus;
        }

        const pm25Element = document.getElementById('pm25');
        if (pm25Element) {
            pm25Element.textContent = data.current.pm25;
        }

        const pm10Element = document.getElementById('pm10');
        if (pm10Element) {
            pm10Element.textContent = data.current.pm10;
        }

        // Update wind information
        const windSpeedElement = document.getElementById('windSpeed');
        if (windSpeedElement) {
            windSpeedElement.textContent = data.current.windSpeed.toFixed(1);
        }

        const windDirectionElement = document.getElementById('windDirection');
        if (windDirectionElement) {
            windDirectionElement.textContent = data.current.windDirection;
        }

        const windDirectionIconElement = document.getElementById('windDirectionIcon');
        if (windDirectionIconElement) {
            windDirectionIconElement.style.transform = `rotate(${data.current.windDeg}deg)`;
        }

        const windGustElement = document.getElementById('windGust');
        if (windGustElement) {
            windGustElement.textContent = data.current.windGust.toFixed(1);
        }

        // Update precipitation information
        const precipAmountElement = document.getElementById('precipAmount');
        if (precipAmountElement) {
            precipAmountElement.textContent = data.current.precipAmount.toFixed(1);
        }

        const precipProbElement = document.getElementById('precipProb');
        if (precipProbElement) {
            precipProbElement.textContent = data.current.precipProb;
        }

        const rainAmountElement = document.getElementById('rainAmount');
        if (rainAmountElement) {
            rainAmountElement.textContent = data.current.rainAmount.toFixed(1);
        }

        const snowAmountElement = document.getElementById('snowAmount');
        if (snowAmountElement) {
            snowAmountElement.textContent = data.current.snowAmount.toFixed(1);
        }

        // Update hourly forecast
        updateHourlyForecast(data.hourlyForecast);

        // Update daily forecast
        updateDailyForecast(data.forecast);

        // Update temperature chart
        updateTemperatureChart(data.hourlyForecast);

        // Update daily icons
        updateDailyIcons(data.forecast);

        // Update weather alerts
        updateWeatherAlerts(data.alerts);

        // Update weather metrics section
        updateWeatherMetrics();

        // Initialize or update weather map
        if (!weatherMap) {
            // If map not initialized, initialize it
            initializeWeatherMap();
        } else {
            // If map already initialized, just update location
            updateWeatherMapLocation(data.current.lat, data.current.lon);
        }

        // Update page title with current city and temperature
        document.title = `${Math.round(data.current.temperature)}°C | ${data.city} - GoWeather Premium`;

    } catch (error) {
        console.error('Error updating dashboard:', error);
    }
}

// Function to update hourly forecast
function updateHourlyForecast(hourlyData) {
    const hourlyForecastContainer = document.getElementById('hourlyForecast');
    if (!hourlyForecastContainer || !hourlyData || hourlyData.length === 0) return;

    // Clear existing content
    hourlyForecastContainer.innerHTML = '';

    // Get current hour
    const now = new Date();
    const currentHour = now.getHours();

    // Create 24 hour forecast display
    hourlyData.slice(0, 24).forEach((hour, index) => {
        // Create time slot
        const timeSlot = document.createElement('div');
        timeSlot.className = `time-slot ${index === 0 ? 'active' : ''}`;

        // Add current hour indicator
        const hourTime = parseInt(hour.time.split(':')[0]);
        if (hourTime === currentHour) {
            timeSlot.classList.add('current-hour');
        }

        // Set content with correct weather icon
        timeSlot.innerHTML = `
            <div class="time-label">${index === 0 ? 'Now' : hour.time}</div>
            <div class="time-temp">${Math.round(hour.temp)}°</div>
            <div class="time-icon"><i class="${hour.icon}"></i></div>
            <div class="precipitation">${hour.precipProb}%</div>
        `;

        // Add to container
        hourlyForecastContainer.appendChild(timeSlot);

        // Add click event
        timeSlot.addEventListener('click', function() {
            // Remove active class from all slots
            document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('active'));
            // Add active class to clicked slot
            this.classList.add('active');
        });
    });

    // Store the hourly data for later use
    allHourlyData = generateExtendedHourlyData(hourlyData);

    // Reset hourly page to current 24h
    currentHourlyPage = 0;
    updateHourlyNavigationStatus();
}