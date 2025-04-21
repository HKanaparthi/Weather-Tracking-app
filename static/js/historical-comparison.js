// historical-comparison.js - Handles the historical weather comparison functionality

document.addEventListener('DOMContentLoaded', function() {
    console.log("Historical comparison page loaded");

    // Add auth token check
    const authToken = localStorage.getItem('authToken');
    console.log("Auth token exists:", !!authToken);

    // Initialize the page
    initializePage();
});

// Global variables
let currentLocation = null;
let comparisonData = null;

// Helper function for safely parsing values
function safeParse(value, defaultValue = 0) {
    if (value === undefined || value === null || isNaN(parseFloat(value))) {
        return defaultValue;
    }
    return parseFloat(value);
}

function initializePage() {
    // Set up event listeners for buttons
    setupEventListeners();

    // Try to get saved location from localStorage
    const savedLocation = localStorage.getItem('weatherLocation');
    if (savedLocation) {
        try {
            currentLocation = JSON.parse(savedLocation);
            loadHistoricalComparison(currentLocation.lat, currentLocation.lon);
        } catch (e) {
            console.error("Error parsing saved location:", e);
            showLocationSelector();
        }
    } else {
        // If no saved location, show location selector
        showLocationSelector();
    }
}

function setupEventListeners() {
    // Location search button
    const searchBtn = document.getElementById('location-search-btn');
    if (searchBtn) {
        searchBtn.addEventListener('click', handleLocationSearch);
    }

    // Search form submission
    const searchForm = document.getElementById('searchForm');
    if (searchForm) {
        searchForm.addEventListener('submit', function(e) {
            e.preventDefault();
            handleLocationSearch();
        });
    }

    // Current location button
    const currentLocBtn = document.getElementById('current-location-btn');
    if (currentLocBtn) {
        currentLocBtn.addEventListener('click', handleCurrentLocation);
    }

    // Change location button
    const changeLocationBtn = document.getElementById('change-location-btn');
    if (changeLocationBtn) {
        changeLocationBtn.addEventListener('click', function() {
            showLocationSelector();
        });
    }
}

function handleLocationSearch() {
    const locationInput = document.getElementById('location-input');
    if (!locationInput || !locationInput.value.trim()) {
        showError('Please enter a location to search');
        return;
    }

    const query = locationInput.value.trim();

    // Show loading state
    showLoading('Searching for location...');

    // Call API to geocode the location
    fetch(`/api/geocode?location=${encodeURIComponent(query)}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to find location');
            }
            return response.json();
        })
        .then(data => {
            if (data && data.length > 0) {
                // Use the first result
                currentLocation = {
                    lat: data[0].lat,
                    lon: data[0].lon,
                    name: data[0].name,
                    country: data[0].country
                };

                // Save to localStorage
                localStorage.setItem('weatherLocation', JSON.stringify(currentLocation));

                // Load historical comparison for this location
                loadHistoricalComparison(currentLocation.lat, currentLocation.lon);
            } else {
                showError('No locations found. Please try a different search.');
            }
        })
        .catch(error => {
            showError('Error finding location: ' + error.message);
            console.error("Location search error:", error);
        })
        .finally(() => {
            hideLoading();
        });
}

function handleCurrentLocation() {
    // Show loading state
    showLoading('Getting your location...');

    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
            // Success callback
            position => {
                const lat = position.coords.latitude;
                const lon = position.coords.longitude;

                // Get location name from coordinates
                fetch(`/api/reverse-geocode?lat=${lat}&lon=${lon}`)
                    .then(response => {
                        if (!response.ok) {
                            throw new Error('Failed to get location name');
                        }
                        return response.json();
                    })
                    .then(data => {
                        currentLocation = {
                            lat: lat,
                            lon: lon,
                            name: data.name || 'Your Location',
                            country: data.country || ''
                        };

                        // Save to localStorage
                        localStorage.setItem('weatherLocation', JSON.stringify(currentLocation));

                        // Load historical comparison for this location
                        loadHistoricalComparison(lat, lon);
                    })
                    .catch(error => {
                        // If reverse geocoding fails, still use the coordinates
                        currentLocation = {
                            lat: lat,
                            lon: lon,
                            name: 'Your Location',
                            country: ''
                        };

                        // Save to localStorage
                        localStorage.setItem('weatherLocation', JSON.stringify(currentLocation));

                        // Load historical comparison for this location
                        loadHistoricalComparison(lat, lon);
                    })
                    .finally(() => {
                        hideLoading();
                    });
            },
            // Error callback
            error => {
                hideLoading();
                switch (error.code) {
                    case error.PERMISSION_DENIED:
                        showError('Location permission denied. Please allow location access or search for a location.');
                        break;
                    case error.POSITION_UNAVAILABLE:
                        showError('Location information is unavailable.');
                        break;
                    case error.TIMEOUT:
                        showError('Location request timed out.');
                        break;
                    default:
                        showError('An unknown error occurred while getting your location.');
                        break;
                }
            },
            // Options
            {
                enableHighAccuracy: true,
                timeout: 10000,
                maximumAge: 0
            }
        );
    } else {
        hideLoading();
        showError('Geolocation is not supported by your browser. Please search for a location instead.');
    }
}

function loadHistoricalComparison(lat, lon) {
    // Check auth before making request
    if (!isLoggedIn()) {
        console.log("User not logged in, redirecting to login page");
        return;
    }

    // Show loading state
    showLoading('Loading weather comparison data...');

    // Hide location selector if visible
    hideLocationSelector();

    // Calculate timestamps for different periods to include in the request
    const now = Math.floor(Date.now() / 1000); // Current time in seconds
    const oneMonthAgo = getTimestampMonthsAgo(1);
    const threeMonthsAgo = getTimestampMonthsAgo(3);
    const sixMonthsAgo = getTimestampMonthsAgo(6);

    // Log timestamps for debugging
    console.log("Timestamps for API request:");
    console.log("Current:", new Date(now * 1000).toLocaleString());
    console.log("1 Month Ago:", new Date(oneMonthAgo * 1000).toLocaleString());
    console.log("3 Months Ago:", new Date(threeMonthsAgo * 1000).toLocaleString());
    console.log("6 Months Ago:", new Date(sixMonthsAgo * 1000).toLocaleString());

    // Build the query string with parameters
    const queryParams = new URLSearchParams({
        lat: lat,
        lon: lon,
        city: currentLocation?.name || 'Unknown Location', // Fallback if no location name
        now: now,
        oneMonthAgo: oneMonthAgo,
        threeMonthsAgo: threeMonthsAgo,
        sixMonthsAgo: sixMonthsAgo
    }).toString();

    // Make API request for historical comparison
    const apiUrl = `/api/weather/historical-comparison?${queryParams}`;
    console.log("Making API request to:", apiUrl);

    fetch(apiUrl)
        .then(response => {
            console.log("API response status:", response.status);
            if (!response.ok) {
                return response.text().then(text => {
                    console.error("API error response:", text);
                    throw new Error(`Failed to load historical weather data: ${response.status} ${text}`);
                });
            }
            return response.json();
        })
        .then(data => {
            console.log("Received comparison data:", data);
            if (!data) {
                throw new Error("Empty response from server");
            }
            comparisonData = data;

            // If API returns data in different format than expected, transform it
            if (!comparisonData.historical) {
                comparisonData = transformApiData(comparisonData);
            }

            displayComparisonData();

            // Create graph visualization
            createWeatherGraph(comparisonData);
        })
        .catch(error => {
            console.error("API request error:", error);
            showError('Error loading weather comparison: ' + error.message);

            // Create test data to show UI even if API fails
            createTestData();
            displayComparisonData();

            // Create graph with test data
            createWeatherGraph(comparisonData);
        })
        .finally(() => {
            hideLoading();
        });
}

// Helper function to get timestamp for X months ago
function getTimestampMonthsAgo(months) {
    const now = new Date();
    const pastDate = new Date(now); // Make a copy
    pastDate.setMonth(now.getMonth() - months);
    return Math.floor(pastDate.getTime() / 1000);
}


// Transform API data if it comes in a different format
function transformApiData(data) {
    // Check if data needs transformation (missing expected structure)
    if (data && !data.historical) {
        console.log("Transforming API data to expected format");

        // Create the structure we expect
        return {
            city: data.city || currentLocation?.name || "Unknown Location",
            current: {
                temperature: safeParse(data.current?.temperature),
                feelsLike: safeParse(data.current?.feelsLike),
                humidity: safeParse(data.current?.humidity, 70),
                pressure: safeParse(data.current?.pressure, 1013),
                condition: data.current?.condition || "Unknown",
                icon: data.current?.icon || "01d",
                windSpeed: safeParse(data.current?.windSpeed),
                date: data.current?.date || new Date().toISOString()
            },
            historical: {
                "1month": {
                    temperature: safeParse(data.oneMonth?.temperature || data.historical?.["1month"]?.temperature),
                    feelsLike: safeParse(data.oneMonth?.feelsLike || data.historical?.["1month"]?.feelsLike),
                    humidity: safeParse(data.oneMonth?.humidity || data.historical?.["1month"]?.humidity, 65),
                    pressure: safeParse(data.oneMonth?.pressure || data.historical?.["1month"]?.pressure, 1010),
                    condition: data.oneMonth?.condition || data.historical?.["1month"]?.condition || "Unknown",
                    icon: data.oneMonth?.icon || data.historical?.["1month"]?.icon || "01d",
                    windSpeed: safeParse(data.oneMonth?.windSpeed || data.historical?.["1month"]?.windSpeed),
                    date: data.oneMonth?.date || data.historical?.["1month"]?.date || new Date(getTimestampMonthsAgo(1) * 1000).toISOString()
                },
                "3months": {
                    temperature: safeParse(data.threeMonths?.temperature || data.historical?.["3months"]?.temperature),
                    feelsLike: safeParse(data.threeMonths?.feelsLike || data.historical?.["3months"]?.feelsLike),
                    humidity: safeParse(data.threeMonths?.humidity || data.historical?.["3months"]?.humidity, 60),
                    pressure: safeParse(data.threeMonths?.pressure || data.historical?.["3months"]?.pressure, 1008),
                    condition: data.threeMonths?.condition || data.historical?.["3months"]?.condition || "Unknown",
                    icon: data.threeMonths?.icon || data.historical?.["3months"]?.icon || "01d",
                    windSpeed: safeParse(data.threeMonths?.windSpeed || data.historical?.["3months"]?.windSpeed),
                    date: data.threeMonths?.date || data.historical?.["3months"]?.date || new Date(getTimestampMonthsAgo(3) * 1000).toISOString()
                },
                "6months": {
                    temperature: safeParse(data.sixMonths?.temperature || data.historical?.["6months"]?.temperature),
                    feelsLike: safeParse(data.sixMonths?.feelsLike || data.historical?.["6months"]?.feelsLike),
                    humidity: safeParse(data.sixMonths?.humidity || data.historical?.["6months"]?.humidity, 55),
                    pressure: safeParse(data.sixMonths?.pressure || data.historical?.["6months"]?.pressure, 1005),
                    condition: data.sixMonths?.condition || data.historical?.["6months"]?.condition || "Unknown",
                    icon: data.sixMonths?.icon || data.historical?.["6months"]?.icon || "01d",
                    windSpeed: safeParse(data.sixMonths?.windSpeed || data.historical?.["6months"]?.windSpeed),
                    date: data.sixMonths?.date || data.historical?.["6months"]?.date || new Date(getTimestampMonthsAgo(6) * 1000).toISOString()
                }
            }
        };
    }

    return data;
}

// Create test data if API fails
function createTestData() {
    console.log("Creating test data with seasonal variations for Charlotte, NC");

    // Charlotte, NC monthly average temperatures (°C)
    const charlotteMonthlyTemps = {
        0: 6,    // January
        1: 8,    // February
        2: 12,   // March
        3: 17,   // April
        4: 21,   // May
        5: 25,   // June
        6: 27,   // July
        7: 26,   // August
        8: 23,   // September
        9: 17,   // October
        10: 12,  // November
        11: 7    // December
    };

    // Calculate dates for the historical periods
    const currentDate = new Date();
    const oneMonthAgo = new Date(currentDate);
    oneMonthAgo.setMonth(oneMonthAgo.getMonth() - 1);

    const threeMonthsAgo = new Date(currentDate);
    threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3);

    const sixMonthsAgo = new Date(currentDate);
    sixMonthsAgo.setMonth(sixMonthsAgo.getMonth() - 6);

    // Get temperatures based on month
    const currentMonth = currentDate.getMonth();
    const oneMonthMonth = oneMonthAgo.getMonth();
    const threeMonthMonth = threeMonthsAgo.getMonth();
    const sixMonthMonth = sixMonthsAgo.getMonth();

    // Add small random variations to make it more realistic
    const randomVariation = () => (Math.random() * 4) - 2; // Random number between -2 and +2

    const currentTemp = charlotteMonthlyTemps[currentMonth] + randomVariation();
    const oneMonthTemp = charlotteMonthlyTemps[oneMonthMonth] + randomVariation();
    const threeMonthTemp = charlotteMonthlyTemps[threeMonthMonth] + randomVariation();
    const sixMonthTemp = charlotteMonthlyTemps[sixMonthMonth] + randomVariation();

    comparisonData = {
        city: currentLocation?.name || "Charlotte",
        current: {
            temperature: currentTemp,
            feelsLike: currentTemp - 2,
            humidity: 70,
            pressure: 1015,
            condition: getSeasonalCondition(currentDate),
            icon: getSeasonalIcon(currentDate),
            windSpeed: 2 + (Math.random() * 3),
            date: currentDate.toISOString()
        },
        historical: {
            "1month": {
                temperature: oneMonthTemp,
                feelsLike: oneMonthTemp - 2,
                humidity: 75,
                pressure: 1010,
                condition: getSeasonalCondition(oneMonthAgo),
                icon: getSeasonalIcon(oneMonthAgo),
                windSpeed: 2 + (Math.random() * 3),
                date: oneMonthAgo.toISOString()
            },
            "3months": {
                temperature: threeMonthTemp,
                feelsLike: threeMonthTemp - 2,
                humidity: 80,
                pressure: 1005,
                condition: getSeasonalCondition(threeMonthsAgo),
                icon: getSeasonalIcon(threeMonthsAgo),
                windSpeed: 2 + (Math.random() * 3),
                date: threeMonthsAgo.toISOString()
            },
            "6months": {
                temperature: sixMonthTemp,
                feelsLike: sixMonthTemp - 2,
                humidity: 60,
                pressure: 1020,
                condition: getSeasonalCondition(sixMonthsAgo),
                icon: getSeasonalIcon(sixMonthsAgo),
                windSpeed: 2 + (Math.random() * 3),
                date: sixMonthsAgo.toISOString()
            }
        }
    };
}

// Helper function to get seasonal weather condition
function getSeasonalCondition(date) {
    const month = date.getMonth();

    // Typical weather conditions by month for Charlotte
    const conditions = [
        "partly cloudy",     // January
        "partly cloudy",     // February
        "partly cloudy",     // March
        "clear sky",         // April
        "clear sky",         // May
        "light rain",        // June
        "moderate rain",     // July
        "light rain",        // August
        "clear sky",         // September
        "clear sky",         // October
        "scattered clouds",  // November
        "overcast clouds"    // December
    ];

    return conditions[month];
}

// Helper function to get seasonal weather icon
function getSeasonalIcon(date) {
    const month = date.getMonth();

    // Weather icons by month for Charlotte
    const icons = [
        "03d",  // January - partly cloudy
        "03d",  // February - partly cloudy
        "03d",  // March - partly cloudy
        "01d",  // April - clear sky
        "01d",  // May - clear sky
        "10d",  // June - light rain
        "10d",  // July - moderate rain
        "10d",  // August - light rain
        "01d",  // September - clear sky
        "01d",  // October - clear sky
        "03d",  // November - scattered clouds
        "04d"   // December - overcast clouds
    ];

    return icons[month];
}

function displayComparisonData() {
    if (!comparisonData) {
        console.error("No comparison data available");
        showError('No comparison data available');
        createTestData(); // Create test data if none is available
    }

    // Update location title
    const locationTitle = document.getElementById('location-title');
    if (locationTitle) {
        const locationName = comparisonData.city ||
            (currentLocation ?
                `${currentLocation.name}${currentLocation.country ? ', ' + currentLocation.country : ''}` :
                "Charlotte, NC");
        locationTitle.textContent = `Weather Comparison for ${locationName}`;
    }

    // Display current weather
    displayWeatherCard('current-weather', 'Current Weather', comparisonData.current);

    // Display historical weather
    if (comparisonData.historical['1month']) {
        displayWeatherCard('month-1-weather', '1 Month Ago', comparisonData.historical['1month']);
    }

    if (comparisonData.historical['3months']) {
        displayWeatherCard('month-3-weather', '3 Months Ago', comparisonData.historical['3months']);
    }

    if (comparisonData.historical['6months']) {
        displayWeatherCard('month-6-weather', '6 Months Ago', comparisonData.historical['6months']);
    }

    // Show comparison section
    const comparisonSection = document.getElementById('comparison-section');
    if (comparisonSection) {
        comparisonSection.classList.remove('hidden');
    }

    // Calculate and display differences
    calculateDifferences();
}

function displayWeatherCard(containerId, title, weatherData) {
    const container = document.getElementById(containerId);
    if (!container) return;

    // Provide fallbacks for all values
    const icon = weatherData?.icon || weatherData?.weatherIcon || '01d'; // Default to clear sky
    const iconUrl = `https://openweathermap.org/img/wn/${icon}@2x.png`;

            // Format the date nicely
            let dateDisplay = "Unknown Date";
        try {
            if (weatherData?.date) {
                const dateObj = new Date(weatherData.date);
                dateDisplay = dateObj.toLocaleDateString(undefined, {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                });
            }
        } catch (e) {
            console.error("Error formatting date:", e);
        }

        const condition = weatherData?.condition || weatherData?.weatherDesc || weatherData?.weatherMain || 'Unknown';
        const temperature = safeParse(weatherData?.temperature, 0);
        const feelsLike = safeParse(weatherData?.feelsLike, 0);
        const humidity = safeParse(weatherData?.humidity, 0);

        // Round wind speed to 1 decimal place
        const windSpeed = Math.round(safeParse(weatherData?.windSpeed, 0) * 10) / 10;

        container.innerHTML = `
        <div class="weather-card">
            <h3>${title}</h3>
            <p class="date">${dateDisplay}</p>
            <div class="weather-icon">
                <img src="${iconUrl}" alt="${condition}">
            </div>
            <div class="weather-main">${condition}</div>
            <div class="weather-temp">
                <span class="temp-value">${Math.round(temperature)}</span>°C
            </div>
            <div class="weather-details">
                <div class="detail">
                    <span class="label">Feels Like:</span>
                    <span class="value">${Math.round(feelsLike)}°C</span>
                </div>
                <div class="detail">
                    <span class="label">Humidity:</span>
                    <span class="value">${Math.round(humidity)}%</span>
                </div>
                <div class="detail">
                    <span class="label">Wind:</span>
                    <span class="value">${windSpeed} m/s</span>
                </div>
            </div>
        </div>
    `;
    }

    function calculateDifferences() {
        const diffContainer = document.getElementById('differences-container');
        if (!diffContainer || !comparisonData) return;

        const current = comparisonData.current || {};
        const historical = comparisonData.historical || {};

        let diffHtml = '<h3>Temperature Differences</h3><div class="differences">';

        // Add difference for 1 month ago
        if (historical['1month']) {
            const tempDiff = safeParse(current.temperature) - safeParse(historical['1month'].temperature);
            const diffClass = tempDiff > 0 ? 'warmer' : (tempDiff < 0 ? 'cooler' : 'same');
            diffHtml += `
            <div class="diff-item">
                <span class="period">From 1 month ago:</span>
                <span class="diff ${diffClass}">${tempDiff > 0 ? '+' : ''}${tempDiff.toFixed(1)}°C</span>
                <span class="desc">${getComparisonText(tempDiff)}</span>
            </div>
        `;
        }

        // Add difference for 3 months ago
        if (historical['3months']) {
            const tempDiff = safeParse(current.temperature) - safeParse(historical['3months'].temperature);
            const diffClass = tempDiff > 0 ? 'warmer' : (tempDiff < 0 ? 'cooler' : 'same');
            diffHtml += `
            <div class="diff-item">
                <span class="period">From 3 months ago:</span>
                <span class="diff ${diffClass}">${tempDiff > 0 ? '+' : ''}${tempDiff.toFixed(1)}°C</span>
                <span class="desc">${getComparisonText(tempDiff)}</span>
            </div>
        `;
        }

        // Add difference for 6 months ago
        if (historical['6months']) {
            const tempDiff = safeParse(current.temperature) - safeParse(historical['6months'].temperature);
            const diffClass = tempDiff > 0 ? 'warmer' : (tempDiff < 0 ? 'cooler' : 'same');
            diffHtml += `
            <div class="diff-item">
                <span class="period">From 6 months ago:</span>
                <span class="diff ${diffClass}">${tempDiff > 0 ? '+' : ''}${tempDiff.toFixed(1)}°C</span>
                <span class="desc">${getComparisonText(tempDiff)}</span>
            </div>
        `;
        }

        diffHtml += '</div>';
        diffContainer.innerHTML = diffHtml;
        diffContainer.classList.remove('hidden');
    }

    function getComparisonText(tempDiff) {
        const absDiff = Math.abs(tempDiff);

        if (absDiff < 1) {
            return 'Almost the same temperature';
        } else if (absDiff < 3) {
            return tempDiff > 0 ? 'Slightly warmer now' : 'Slightly cooler now';
        } else if (absDiff < 7) {
            return tempDiff > 0 ? 'Noticeably warmer now' : 'Noticeably cooler now';
        } else if (absDiff < 15) {
            return tempDiff > 0 ? 'Significantly warmer now' : 'Significantly cooler now';
        } else {
            return tempDiff > 0 ? 'Drastically warmer now' : 'Drastically cooler now';
        }
    }

// Weather graph creation function
    function createWeatherGraph(data) {
        const graphContainer = document.getElementById('weather-graph');
        if (!graphContainer || !data) return;

        // Clear previous graph
        graphContainer.innerHTML = '<canvas id="temperatureChart"></canvas>';

        // Get the canvas element
        const canvas = document.getElementById('temperatureChart');
        if (!canvas) return;

        // Prepare data points
        const periods = ['6 Months Ago', '3 Months Ago', '1 Month Ago', 'Current'];

        // Get temperature values, ensuring they are numbers
        const temperatures = [
            safeParse(data.historical['6months']?.temperature),
            safeParse(data.historical['3months']?.temperature),
            safeParse(data.historical['1month']?.temperature),
            safeParse(data.current?.temperature)
        ];

        // Format dates for tooltips
        const dates = [
            formatShortDate(data.historical['6months']?.date),
            formatShortDate(data.historical['3months']?.date),
            formatShortDate(data.historical['1month']?.date),
            formatShortDate(data.current?.date)
        ];

        // Get colors based on temperatures (warmer = more red, cooler = more blue)
        const colors = temperatures.map(temp => getTemperatureColor(temp));

        // Create chart
        const ctx = canvas.getContext('2d');
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: periods,
                datasets: [{
                    label: 'Temperature (°C)',
                    data: temperatures,
                    backgroundColor: colors,
                    borderColor: 'rgba(75, 192, 192, 1)',
                    borderWidth: 2,
                    pointBackgroundColor: colors,
                    pointBorderColor: '#fff',
                    pointBorderWidth: 2,
                    pointRadius: 6,
                    pointHoverRadius: 8,
                    fill: false,
                    tension: 0.3
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: false,
                        title: {
                            display: true,
                            text: 'Temperature (°C)'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time Period'
                        }
                    }
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                const value = context.parsed.y;
                                const index = context.dataIndex;
                                return `Temperature: ${value.toFixed(1)}°C (${dates[index]})`;
                            }
                        }
                    },
                    legend: {
                        display: false
                    },
                    title: {
                        display: true,
                        text: 'Temperature Comparison Over Time',
                        font: {
                            size: 16
                        }
                    }
                }
            }
        });
    }

// Helper function to format date for tooltips
    function formatShortDate(dateString) {
        if (!dateString) return 'Unknown';
        try {
            const date = new Date(dateString);
            return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
        } catch (e) {
            console.error("Error formatting date:", e);
            return 'Unknown';
        }
    }

// Helper function to get color based on temperature
    function getTemperatureColor(temp) {
        // Convert Celsius to a color:
        // Cold (below 0°C) = Blue
        // Cool (0-15°C) = Light Blue to Green
        // Moderate (15-25°C) = Green to Yellow
        // Warm (25-35°C) = Yellow to Orange
        // Hot (above 35°C) = Orange to Red

        if (temp < 0) {
            return 'rgba(0, 0, 255, 0.7)'; // Blue for below freezing
        } else if (temp < 15) {
            // Transition from blue to green
            const ratio = temp / 15;
            return `rgba(${Math.floor(ratio * 50)}, ${Math.floor(100 + (ratio * 155))}, ${Math.floor(255 - (ratio * 200))}, 0.7)`;
        } else if (temp < 25) {
            // Transition from green to yellow
            const ratio = (temp - 15) / 10;
            return `rgba(${Math.floor(50 + (ratio * 205))}, ${Math.floor(255 - (ratio * 55))}, 0, 0.7)`;
        } else if (temp < 35) {
            // Transition from yellow to orange
            const ratio = (temp - 25) / 10;
            return `rgba(255, ${Math.floor(200 - (ratio * 120))}, 0, 0.7)`;
        } else {
            // Transition from orange to red for very hot
            const ratio = Math.min((temp - 35) / 15, 1);
            return `rgba(255, ${Math.floor(80 - (ratio * 80))}, 0, 0.7)`;
        }
    }

// Utility functions
    function showLoading(message) {
        const loadingEl = document.getElementById('loading');
        const loadingMsgEl = document.getElementById('loading-message');

        if (loadingMsgEl) {
            loadingMsgEl.textContent = message || 'Loading...';
        }

        if (loadingEl) {
            loadingEl.classList.remove('hidden');
        }
    }

    function hideLoading() {
        const loadingEl = document.getElementById('loading');
        if (loadingEl) {
            loadingEl.classList.add('hidden');
        }
    }

    function showError(message) {
        console.error(message);
        const errorEl = document.getElementById('error-message');
        if (errorEl) {
            errorEl.textContent = message;
            errorEl.classList.remove('hidden');

            // Auto-hide after 5 seconds
            setTimeout(() => {
                errorEl.classList.add('hidden');
            }, 5000);
        }
    }

    function showLocationSelector() {
        const locationSelectorEl = document.getElementById('location-selector');
        if (locationSelectorEl) {
            locationSelectorEl.classList.remove('hidden');
        }

        const comparisonSectionEl = document.getElementById('comparison-section');
        if (comparisonSectionEl) {
            comparisonSectionEl.classList.add('hidden');
        }
    }

    function hideLocationSelector() {
        const locationSelectorEl = document.getElementById('location-selector');
        if (locationSelectorEl) {
            locationSelectorEl.classList.add('hidden');
        }
    }

// Always return true for login check during testing
    function isLoggedIn() {
        return true;
    }