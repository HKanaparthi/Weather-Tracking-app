/**
 * Weather Impact App - Complete Solution
 *
 * This file handles all functionality for the Weather Impact page.
 */

document.addEventListener('DOMContentLoaded', function() {
    console.log('Weather Impact App initialized');

    // Elements
    const searchForm = document.querySelector('.location-search form');
    const cityInput = document.querySelector('.location-search input[name="city"]');
    const loadingIndicator = document.getElementById('loading-indicator');
    const currentWeatherSummary = document.getElementById('current-weather-summary');
    const impactFilter = document.getElementById('impact-filter');
    const impactCards = document.getElementById('impact-cards');
    const noResults = document.getElementById('no-results');
    const weatherAlert = document.getElementById('weather-alert');
    const alertText = document.getElementById('alert-text');
    const modal = document.getElementById('impact-modal');
    const modalClose = document.getElementById('modal-close');
    const modalTitle = document.getElementById('modal-title');
    const modalBody = document.getElementById('modal-body');

    // Initialize
    init();

    function init() {
        // Set up event listeners
        setupEventListeners();

        // Check if city is available in the URL
        const urlParams = new URLSearchParams(window.location.search);
        const city = urlParams.get('city');

        if (city) {
            // City specified in URL, fetch real weather data
            fetchRealWeatherData(city);
        } else {
            // Show form to enter city
            if (loadingIndicator) loadingIndicator.style.display = 'none';
            if (currentWeatherSummary) currentWeatherSummary.style.display = 'none';
            if (impactFilter) impactFilter.style.display = 'none';
            if (impactCards) impactCards.style.display = 'none';
            if (noResults) noResults.style.display = 'block';
        }
    }

    function setupEventListeners() {
        // Form submission
        if (searchForm) {
            searchForm.addEventListener('submit', function(e) {
                const city = cityInput.value.trim();
                if (!city) {
                    e.preventDefault();
                    alert('Please enter a city name');
                }
            });
        }

        // Category filtering
        const categories = document.querySelectorAll('.impact-category');
        categories.forEach(category => {
            category.addEventListener('click', function() {
                // Remove active class from all categories
                categories.forEach(cat => cat.classList.remove('active'));

                // Add active class to clicked category
                this.classList.add('active');

                // Filter cards
                filterImpactCards(this.getAttribute('data-category'));
            });
        });

        // Modal close button
        if (modalClose) {
            modalClose.addEventListener('click', function() {
                modal.classList.remove('active');
            });
        }

        // Close modal when clicking outside
        window.addEventListener('click', function(event) {
            if (event.target === modal) {
                modal.classList.remove('active');
            }
        });
    }

    function fetchRealWeatherData(city) {
        // Show loading indicator
        if (loadingIndicator) loadingIndicator.style.display = 'block';
        if (currentWeatherSummary) currentWeatherSummary.style.display = 'none';
        if (impactFilter) impactFilter.style.display = 'none';
        if (impactCards) impactCards.style.display = 'none';
        if (noResults) noResults.style.display = 'none';

        // Fetch from your backend API endpoint
        fetch(`/api/weather?city=${encodeURIComponent(city)}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Weather data not found for this location');
                }
                return response.json();
            })
            .then(weatherData => {
                console.log('Weather data received:', weatherData);

                // Format the weather data for our frontend
                const formattedWeather = {
                    location: weatherData.city || city,
                    current: {
                        temp: weatherData.current.temperature ?
                            parseFloat(weatherData.current.temperature.replace('°C', '')) : 20,
                        weather: {
                            main: weatherData.current.condition.split(' ')[0] || 'Clear',
                            description: weatherData.current.condition || 'Clear sky',
                            icon: weatherData.current.icon || '01d'
                        },
                        humidity: weatherData.current.humidity ?
                            parseInt(weatherData.current.humidity.replace('%', '')) : 50,
                        wind_speed: weatherData.current.wind ?
                            parseFloat(weatherData.current.wind.replace(' m/s', '')) : 3,
                        visibility: 10, // Default visibility
                        uvi: weatherData.current.uvIndex || 5
                    }
                };

                // Update UI with weather data
                updateWeatherDisplay(formattedWeather);
                generateImpactAssessments(formattedWeather);
                checkWeatherAlerts(formattedWeather);

                // Show UI components
                if (loadingIndicator) loadingIndicator.style.display = 'none';
                if (currentWeatherSummary) currentWeatherSummary.style.display = 'flex';
                if (impactFilter) impactFilter.style.display = 'block';
                if (impactCards) impactCards.style.display = 'grid';
            })
            .catch(error => {
                console.error('Error fetching weather data:', error);

                if (loadingIndicator) loadingIndicator.style.display = 'none';

                if (noResults) {
                    noResults.style.display = 'block';
                    const noResultsTitle = noResults.querySelector('h3');
                    const noResultsDesc = noResults.querySelector('p');
                    if (noResultsTitle) {
                        noResultsTitle.textContent = 'Weather data not available';
                    }
                    if (noResultsDesc) {
                        noResultsDesc.textContent = 'Could not fetch weather data for this location. Please try another city or try again later.';
                    }
                }
            });
    }

    function updateWeatherDisplay(data) {
        if (!data || !data.current) return;

        // Update location name
        const currentLocation = document.getElementById('current-location');
        if (currentLocation) {
            currentLocation.textContent = data.location;
        }

        // Update weather condition
        const conditionText = document.getElementById('current-condition-text');
        if (conditionText) {
            conditionText.textContent = data.current.weather.description;
        }

        // Update weather icon
        const conditionIcon = document.getElementById('current-condition-icon');
        if (conditionIcon) {
            conditionIcon.src = `https://openweathermap.org/img/wn/${data.current.weather.icon}@2x.png`;
            conditionIcon.alt = data.current.weather.description;
        }

        // Update temperature
        const tempElement = document.getElementById('current-temp');
        if (tempElement) {
            tempElement.textContent = `${Math.round(data.current.temp)}°C`;
        }

        // Update humidity
        const humidityElement = document.getElementById('current-humidity');
        if (humidityElement) {
            humidityElement.textContent = `Humidity: ${data.current.humidity}%`;
        }

        // Update wind
        const windElement = document.getElementById('current-wind');
        if (windElement) {
            windElement.textContent = `Wind: ${data.current.wind_speed.toFixed(1)} m/s`;
        }

        // Update visibility
        const visibilityElement = document.getElementById('current-visibility');
        if (visibilityElement) {
            visibilityElement.textContent = `Visibility: ${data.current.visibility.toFixed(1)} km`;
        }
    }

    function checkWeatherAlerts(data) {
        // Hide alert by default
        if (weatherAlert) {
            weatherAlert.classList.remove('active');
        }

        if (!data || !data.current) return;

        // Check for conditions that warrant alerts
        const temp = data.current.temp;
        const weatherCondition = data.current.weather.main;
        const windSpeed = data.current.wind_speed;
        const visibility = data.current.visibility;

        let alertMessage = null;

        // Check for extreme weather conditions
        if (temp > 35) {
            alertMessage = "Extreme heat warning. Stay hydrated and avoid prolonged outdoor activities.";
        } else if (temp < 0) {
            alertMessage = "Freezing temperatures. Take precautions against cold exposure.";
        } else if (["Thunderstorm", "Tornado"].includes(weatherCondition)) {
            alertMessage = "Severe weather alert. Seek shelter and stay indoors.";
        } else if (weatherCondition === "Snow" && windSpeed > 10) {
            alertMessage = "Blizzard conditions possible. Avoid unnecessary travel.";
        } else if (weatherCondition === "Rain" && windSpeed > 15) {
            alertMessage = "Heavy rain and wind. Use caution if outdoors.";
        } else if (["Fog", "Mist"].includes(weatherCondition) && visibility < 2) {
            alertMessage = "Poor visibility conditions. Drive with caution.";
        }

        // Display alert if needed
        if (alertMessage && weatherAlert && alertText) {
            alertText.textContent = alertMessage;
            weatherAlert.classList.add('active');
        }
    }

    function generateImpactAssessments(data) {
        // Clear existing cards
        if (impactCards) {
            impactCards.innerHTML = '';
        } else {
            return;
        }

        if (!data || !data.current) {
            if (noResults) {
                noResults.style.display = 'block';
            }
            return;
        }

        // Generate activities based on weather conditions
        const activities = getActivitiesData(data);

        if (activities.length === 0) {
            if (noResults) {
                noResults.style.display = 'block';
            }
            return;
        }

        // Create card for each activity
        activities.forEach(activity => {
            const card = createImpactCard(activity);
            if (impactCards) {
                impactCards.appendChild(card);
            }

            // Add click event to show modal
            card.addEventListener('click', function() {
                showDetailedImpact(activity);
            });
        });
    }

    function getActivitiesData(weather) {
        const temp = weather.current.temp;
        const weatherCondition = weather.current.weather.main;
        const weatherDescription = weather.current.weather.description;
        const humidity = weather.current.humidity;
        const windSpeed = weather.current.wind_speed;
        const visibility = weather.current.visibility;
        const uvi = weather.current.uvi || 5; // Default UV index if not available

        // Generate activities with impact ratings
        const activities = [];

        // Outdoor Activities
        activities.push({
            title: "Running",
            category: "outdoor",
            icon: "fa-running",
            rating: getRatingForRunning(temp, weatherCondition, humidity, windSpeed),
            description: getDescriptionForRunning(temp, weatherCondition, humidity, windSpeed),
            factors: [
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "running") },
                { name: "Humidity", value: `${humidity}%`, rating: getRatingForHumidity(humidity) },
                { name: "Wind", value: `${windSpeed.toFixed(1)} m/s`, rating: getRatingForWind(windSpeed) },
                { name: "UV Index", value: uvi.toString(), rating: getRatingForUV(uvi) }
            ],
            recommendations: getRecommendationsForRunning(temp, weatherCondition, humidity, windSpeed, uvi)
        });

        activities.push({
            title: "Cycling",
            category: "outdoor",
            icon: "fa-biking",
            rating: getRatingForCycling(temp, weatherCondition, windSpeed, visibility),
            description: getDescriptionForCycling(temp, weatherCondition, windSpeed, visibility),
            factors: [
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "cycling") },
                { name: "Wind", value: `${windSpeed.toFixed(1)} m/s`, rating: getRatingForWind(windSpeed, "cycling") },
                { name: "Precipitation", value: getPrecipitationValue(weatherCondition), rating: getRatingForPrecipitation(weatherCondition) },
                { name: "Road Conditions", value: getRoadConditions(weatherCondition), rating: getRatingForRoadConditions(weatherCondition) }
            ],
            recommendations: getRecommendationsForCycling(temp, weatherCondition, windSpeed, uvi)
        });

        activities.push({
            title: "Hiking",
            category: "outdoor",
            icon: "fa-hiking",
            rating: getRatingForHiking(temp, weatherCondition, windSpeed, visibility),
            description: getDescriptionForHiking(temp, weatherCondition, visibility),
            factors: [
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "hiking") },
                { name: "Weather", value: weatherDescription, rating: getRatingForWeatherType(weatherCondition, "hiking") },
                { name: "Visibility", value: `${visibility.toFixed(1)} km`, rating: getRatingForVisibility(visibility) },
                { name: "UV Index", value: uvi.toString(), rating: getRatingForUV(uvi) }
            ],
            recommendations: getRecommendationsForHiking(temp, weatherCondition, visibility, uvi)
        });

        // Agriculture Activities
        activities.push({
            title: "Gardening",
            category: "agriculture",
            icon: "fa-seedling",
            rating: getRatingForGardening(temp, weatherCondition),
            description: getDescriptionForGardening(temp, weatherCondition),
            factors: [
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "gardening") },
                { name: "Precipitation", value: getPrecipitationValue(weatherCondition), rating: getRatingForPrecipitation(weatherCondition, "gardening") },
                { name: "Soil Condition", value: getSoilCondition(weatherCondition), rating: getRatingForSoilCondition(weatherCondition) }
            ],
            recommendations: getRecommendationsForGardening(temp, weatherCondition)
        });

        // Construction Activity
        activities.push({
            title: "Construction",
            category: "construction",
            icon: "fa-hard-hat",
            rating: getRatingForConstruction(temp, weatherCondition, windSpeed),
            description: getDescriptionForConstruction(temp, weatherCondition, windSpeed),
            factors: [
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "construction") },
                { name: "Weather", value: weatherDescription, rating: getRatingForWeatherType(weatherCondition, "construction") },
                { name: "Wind", value: `${windSpeed.toFixed(1)} m/s`, rating: getRatingForWind(windSpeed, "construction") },
                { name: "Working Hours", value: getWorkingHours(weatherCondition, temp), rating: getRatingForWorkingHours(weatherCondition, temp) }
            ],
            recommendations: getRecommendationsForConstruction(temp, weatherCondition, windSpeed)
        });

        // Health Activity
        activities.push({
            title: "Respiratory Health",
            category: "health",
            icon: "fa-lungs",
            rating: getRatingForRespiratoryHealth(temp, weatherCondition, humidity),
            description: getDescriptionForRespiratoryHealth(temp, weatherCondition, humidity),
            factors: [
                { name: "Air Quality", value: getAirQualityEstimate(weatherCondition), rating: getRatingForAirQuality(weatherCondition) },
                { name: "Humidity", value: `${humidity}%`, rating: getRatingForHumidity(humidity, "respiratory") },
                { name: "Temperature", value: `${Math.round(temp)}°C`, rating: getRatingForTemp(temp, "respiratory") }
            ],
            recommendations: getRecommendationsForRespiratoryHealth(temp, weatherCondition, humidity)
        });

        // Travel Activity
        activities.push({
            title: "Driving Conditions",
            category: "travel",
            icon: "fa-car",
            rating: getRatingForDriving(weatherCondition, visibility),
            description: getDescriptionForDriving(weatherCondition, visibility),
            factors: [
                { name: "Road Condition", value: getRoadConditions(weatherCondition), rating: getRatingForRoadConditions(weatherCondition) },
                { name: "Visibility", value: `${visibility.toFixed(1)} km`, rating: getRatingForVisibility(visibility, "driving") },
                { name: "Precipitation", value: getPrecipitationValue(weatherCondition), rating: getRatingForPrecipitation(weatherCondition, "driving") }
            ],
            recommendations: getRecommendationsForDriving(weatherCondition, visibility)
        });

        return activities;
    }

    function createImpactCard(activity) {
        const card = document.createElement('div');
        card.className = 'impact-card';
        card.setAttribute('data-category', activity.category);

        const ratingClass = activity.rating.toLowerCase(); // favorable, moderate, or unfavorable

        card.innerHTML = `
            <div class="impact-card-header">
                <div class="impact-card-title">
                    <i class="fas ${activity.icon} impact-card-icon"></i>
                    ${activity.title}
                </div>
                <div class="impact-card-rating">
                    <i class="fas fa-circle ${ratingClass}"></i>
                    <span class="impact-card-rating-text ${ratingClass}">${activity.rating}</span>
                </div>
            </div>
            <div class="impact-card-body">
                <div class="impact-card-description">
                    ${activity.description}
                </div>
                <div class="impact-card-factors">
                    ${activity.factors.map(factor => `
                        <div class="impact-card-factor">
                            <span class="impact-card-factor-name">${factor.name}</span>
                            <span class="impact-card-factor-value ${factor.rating.toLowerCase()}">${factor.value} (${factor.rating})</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;

        return card;
    }

    function filterImpactCards(category) {
        const cards = document.querySelectorAll('.impact-card');
        let visibleCount = 0;

        cards.forEach(card => {
            if (category === 'all' || card.getAttribute('data-category') === category) {
                card.style.display = 'block';
                visibleCount++;
            } else {
                card.style.display = 'none';
            }
        });

        // Show/hide no results message
        if (visibleCount === 0 && noResults) {
            noResults.style.display = 'block';
        } else if (noResults) {
            noResults.style.display = 'none';
        }
    }

    function showDetailedImpact(activity) {
        if (!modal || !modalTitle || !modalBody) return;

        modalTitle.textContent = activity.title;

        // Create modal content
        const ratingClass = activity.rating.toLowerCase();

        const modalContent = `
            <div class="modal-rating">
                <i class="fas fa-circle ${ratingClass}"></i>
                <span class="impact-card-rating-text ${ratingClass}">${activity.rating}</span>
            </div>
            <div class="modal-description">
                ${activity.description}
            </div>
            <div class="modal-factors">
                ${activity.factors.map(factor => `
                    <div class="modal-factor">
                        <span class="impact-card-factor-name">${factor.name}</span>
                        <span class="impact-card-factor-value ${factor.rating.toLowerCase()}">${factor.value} (${factor.rating})</span>
                    </div>
                `).join('')}
            </div>
            <div class="modal-recommendations">
                <h3>Recommendations</h3>
                ${activity.recommendations.map(rec => `
                    <div class="recommendation-item">
                        <span class="recommendation-icon"><i class="fas fa-check-circle"></i></span>
                        <span>${rec}</span>
                    </div>
                `).join('')}
            </div>
        `;

        modalBody.innerHTML = modalContent;
        modal.classList.add('active');
    }

    // Rating Calculation Functions
    function getRatingForRunning(temp, weather, humidity, wind) {
        if (temp < 0 || temp > 30 || humidity > 85 || ['Thunderstorm', 'Rain', 'Snow'].includes(weather)) {
            return "Unfavorable";
        } else if ((temp < 10 || temp > 25) || (humidity > 70) || wind > 5 || weather === 'Drizzle') {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForCycling(temp, weather, wind, visibility) {
        if (temp < -5 || temp > 35 || wind > 8 || visibility < 5 || ['Thunderstorm', 'Snow'].includes(weather)) {
            return "Unfavorable";
        } else if ((temp < 5 || temp > 30) || wind > 5 || weather === 'Rain') {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForHiking(temp, weather, wind, visibility) {
        if (temp < -5 || temp > 35 || visibility < 3 || ['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) {
            return "Unfavorable";
        } else if ((temp < 5 || temp > 30) || visibility < 5 || weather === 'Rain') {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForGardening(temp, weather) {
        if (temp < 0 || temp > 32 || ['Thunderstorm', 'Snow'].includes(weather)) {
            return "Unfavorable";
        } else if ((temp < 10 || temp > 28) || weather === 'Rain') {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForConstruction(temp, weather, wind) {
        if (temp < -5 || temp > 35 || wind > 10 || ['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) {
            return "Unfavorable";
        } else if ((temp < 5 || temp > 30) || wind > 7 || weather === 'Rain') {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForRespiratoryHealth(temp, weather, humidity) {
        if ((temp < 0 || temp > 35) || ['Dust', 'Smoke', 'Fog'].includes(weather) || humidity > 85 || humidity < 25) {
            return "Unfavorable";
        } else if ((temp < 5 || temp > 30) || ['Mist'].includes(weather) || humidity > 75 || humidity < 30) {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    function getRatingForDriving(weather, visibility) {
        if (['Thunderstorm', 'Heavy Rain', 'Snow', 'Fog'].includes(weather) || visibility < 2) {
            return "Unfavorable";
        } else if (['Rain', 'Drizzle', 'Mist'].includes(weather) || visibility < 5) {
            return "Moderate";
        } else {
            return "Favorable";
        }
    }

    // Factor Rating Functions
    function getRatingForTemp(temp, activity) {
        switch(activity) {
            case "running":
                if (temp < 5 || temp > 30) return "Unfavorable";
                if (temp < 10 || temp > 25) return "Moderate";
                return "Favorable";
            case "cycling":
                if (temp < 0 || temp > 35) return "Unfavorable";
                if (temp < 5 || temp > 30) return "Moderate";
                return "Favorable";
            case "hiking":
                if (temp < 0 || temp > 35) return "Unfavorable";
                if (temp < 5 || temp > 30) return "Moderate";
                return "Favorable";
            case "gardening":
                if (temp < 0 || temp > 32) return "Unfavorable";
                if (temp < 10 || temp > 28) return "Moderate";
                return "Favorable";
            case "construction":
                if (temp < -5 || temp > 35) return "Unfavorable";
                if (temp < 5 || temp > 30) return "Moderate";
                return "Favorable";
            case "respiratory":
                if (temp < 0 || temp > 35) return "Unfavorable";
                if (temp < 5 || temp > 30) return "Moderate";
                return "Favorable";
            default:
                if (temp < 5 || temp > 30) return "Unfavorable";
                if (temp < 10 || temp > 25) return "Moderate";
                return "Favorable";
        }
    }

    function getRatingForHumidity(humidity, activity) {
        if (activity === "respiratory") {
            if (humidity > 85 || humidity < 25) return "Unfavorable";
            if (humidity > 75 || humidity < 30) return "Moderate";
            return "Favorable";
        }

        // Default
        if (humidity > 85) return "Unfavorable";
        if (humidity > 70) return "Moderate";
        return "Favorable";
    }

    function getRatingForWind(wind, activity) {
        switch(activity) {
            case "cycling":
                if (wind > 8) return "Unfavorable";
                if (wind > 5) return "Moderate";
                return "Favorable";
            case "construction":
                if (wind > 10) return "Unfavorable";
                if (wind > 7) return "Moderate";
                return "Favorable";
            default:
                if (wind > 7) return "Unfavorable";
                if (wind > 4) return "Moderate";
                return "Favorable";
        }
    }

    function getRatingForUV(uvi) {
        if (uvi > 7) return "Unfavorable";
        if (uvi > 5) return "Moderate";
        return "Favorable";
    }

    function getRatingForPrecipitation(weather, activity) {
        switch(activity) {
            case "gardening":
                if (['Thunderstorm', 'Snow'].includes(weather)) return "Unfavorable";
                if (['Heavy Rain'].includes(weather)) return "Moderate";
                if (['Rain', 'Drizzle'].includes(weather)) return "Favorable";
                return "Moderate"; // No rain is moderate for gardening
            case "driving":
                if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) return "Unfavorable";
                if (['Rain', 'Drizzle'].includes(weather)) return "Moderate";
                return "Favorable";
            default:
                if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) return "Unfavorable";
                if (['Rain', 'Drizzle'].includes(weather)) return "Moderate";
                return "Favorable";
        }
    }

    function getRatingForRoadConditions(weather) {
        if (['Thunderstorm', 'Snow', 'Heavy Rain', 'Fog'].includes(weather)) return "Unfavorable";
        if (['Rain', 'Drizzle', 'Mist'].includes(weather)) return "Moderate";
        return "Favorable";
    }

    function getRatingForWeatherType(weather, activity) {
        switch(activity) {
            case "hiking":
                if (['Thunderstorm', 'Snow', 'Heavy Rain', 'Fog'].includes(weather)) return "Unfavorable";
                if (['Rain', 'Drizzle'].includes(weather)) return "Moderate";
                return "Favorable";
            case "construction":
                if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) return "Unfavorable";
                if (['Rain', 'Drizzle', 'Fog'].includes(weather)) return "Moderate";
                return "Favorable";
            default:
                if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) return "Unfavorable";
                if (['Rain', 'Drizzle', 'Fog', 'Mist'].includes(weather)) return "Moderate";
                return "Favorable";
        }
    }

    function getRatingForVisibility(visibility, activity) {
        if (activity === "driving") {
            if (visibility < 2) return "Unfavorable";
            if (visibility < 5) return "Moderate";
            return "Favorable";
        }

        // Default
        if (visibility < 3) return "Unfavorable";
        if (visibility < 8) return "Moderate";
        return "Favorable";
    }

    function getRatingForSoilCondition(weather) {
        if (['Thunderstorm', 'Heavy Rain'].includes(weather)) return "Unfavorable";
        if (['Snow'].includes(weather)) return "Unfavorable";
        if (['Rain'].includes(weather)) return "Moderate";
        if (['Drizzle', 'Mist', 'Fog'].includes(weather)) return "Favorable";
        return "Moderate"; // Dry soil is moderate
    }

    function getRatingForAirQuality(weather) {
        if (['Dust', 'Smoke', 'Fog'].includes(weather)) return "Unfavorable";
        if (['Haze', 'Mist'].includes(weather)) return "Moderate";
        return "Favorable";
    }

    function getRatingForWorkingHours(weather, temp) {
        if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather) || temp < -5 || temp > 35) return "Unfavorable";
        if (['Rain', 'Drizzle', 'Fog'].includes(weather) || temp < 5 || temp > 30) return "Moderate";
        return "Favorable";
    }

    // Value Calculation Functions
    function getPrecipitationValue(weather) {
        if (weather === 'Thunderstorm') return "Heavy";
        if (weather === 'Rain') return "Moderate";
        if (weather === 'Drizzle') return "Light";
        if (weather === 'Snow') return "Snow";
        return "None";
    }

    function getRoadConditions(weather) {
        if (weather === 'Thunderstorm' || weather === 'Heavy Rain') return "Very Wet";
        if (weather === 'Rain') return "Wet";
        if (weather === 'Drizzle' || weather === 'Mist') return "Damp";
        if (weather === 'Snow') return "Snow Covered";
        if (weather === 'Fog') return "Foggy";
        return "Dry";
    }

    function getSoilCondition(weather) {
        if (weather === 'Thunderstorm' || weather === 'Heavy Rain') return "Saturated";
        if (weather === 'Rain') return "Wet";
        if (weather === 'Drizzle' || weather === 'Mist') return "Moist";
        if (weather === 'Snow') return "Frozen";
        return "Dry";
    }

    function getWorkingHours(weather, temp) {
        if (['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather) || temp < -5 || temp > 35) {
            return "Limited";
        } else if (['Rain', 'Drizzle', 'Fog'].includes(weather) || temp < 5 || temp > 30) {
            return "Reduced";
        } else {
            return "Normal";
        }
    }

    function getAirQualityEstimate(weather) {
        if (['Dust', 'Smoke'].includes(weather)) return "Poor";
        if (['Fog', 'Haze'].includes(weather)) return "Moderate";
        if (['Rain', 'Drizzle'].includes(weather)) return "Good";
        return "Good";
    }

    // Description Functions
    function getDescriptionForRunning(temp, weather, humidity, wind) {
        if (temp < 0 || temp > 30 || humidity > 85 || ['Thunderstorm', 'Rain', 'Snow'].includes(weather)) {
            return "Current conditions are challenging for running. Consider indoor alternatives or adjusting your routine based on specific factors.";
        } else if ((temp < 10 || temp > 25) || (humidity > 70) || wind > 5 || weather === 'Drizzle') {
            return "Running is possible but you may need to modify your routine. Pay attention to the specific factors that could impact your comfort and performance.";
        } else {
            return "Excellent conditions for running with comfortable temperatures and good overall weather. Enjoy your run while maintaining normal precautions.";
        }
    }

    function getDescriptionForCycling(temp, weather, wind, visibility) {
        if (temp < -5 || temp > 35 || wind > 8 || visibility < 5 || ['Thunderstorm', 'Snow'].includes(weather)) {
            return "Current conditions present significant challenges for cycling. Consider postponing or choosing indoor cycling options instead.";
        } else if ((temp < 5 || temp > 30) || wind > 5 || weather === 'Rain') {
            return "Cycling is possible but you'll face some challenges that require appropriate gear and adjustments to your ride plans.";
        } else {
            return "Good conditions for cycling with manageable weather factors. With proper preparation, you should have an enjoyable ride.";
        }
    }

    function getDescriptionForHiking(temp, weather, visibility) {
        if (temp < -5 || temp > 35 || visibility < 3 || ['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) {
            return "Weather conditions are not suitable for hiking. Safety concerns include limited visibility, risk of hypothermia, or dangerous trail conditions.";
        } else if ((temp < 5 || temp > 30) || visibility < 5 || weather === 'Rain') {
            return "Hiking is possible but challenging. Trail conditions may be affected and you'll need appropriate gear for comfort and safety.";
        } else {
            return "Good conditions for hiking with clear trails and comfortable temperatures. Take standard precautions and enjoy your outdoor adventure.";
        }
    }

    function getDescriptionForGardening(temp, weather) {
        if (temp < 0 || temp > 32 || ['Thunderstorm', 'Snow'].includes(weather)) {
            return "Current conditions are not suitable for most gardening activities. Consider indoor gardening tasks or postpone to a more favorable day.";
        } else if ((temp < 10 || temp > 28) || weather === 'Heavy Rain') {
            return "Limited gardening activities are possible. Focus on tasks that can be done with weather protection or consider shorter work periods.";
        } else {
            return "Good conditions for most gardening activities. Soil moisture and temperature are generally favorable for plant growth and garden work.";
        }
    }

    function getDescriptionForConstruction(temp, weather, wind) {
        if (temp < -5 || temp > 35 || wind > 10 || ['Thunderstorm', 'Snow', 'Heavy Rain'].includes(weather)) {
            return "Construction activities should be limited due to safety concerns and reduced efficiency. Consider indoor work or postponing weather-sensitive tasks.";
        } else if ((temp < 5 || temp > 30) || wind > 7 || weather === 'Rain') {
            return "Construction can proceed with caution. Some activities may need to be modified or additional safety measures implemented due to weather conditions.";
        } else {
            return "Favorable conditions for construction work. Standard safety procedures apply, but weather should not significantly impact most construction activities.";
        }
    }

    function getDescriptionForRespiratoryHealth(temp, weather, humidity) {
        if ((temp < 0 || temp > 35) || ['Dust', 'Smoke', 'Fog'].includes(weather) || humidity > 85 || humidity < 25) {
            return "Current conditions may be challenging for those with respiratory conditions. Consider limiting outdoor exposure and using air purifiers indoors.";
        } else if ((temp < 5 || temp > 30) || ['Mist'].includes(weather) || humidity > 75 || humidity < 30) {
            return "Moderately suitable conditions for respiratory health. Those with sensitive airways should take normal precautions when outdoors.";
        } else {
            return "Generally favorable conditions for respiratory health. Air quality is good and temperature/humidity levels are in a comfortable range.";
        }
    }

    function getDescriptionForDriving(weather, visibility) {
        if (['Thunderstorm', 'Heavy Rain', 'Snow', 'Fog'].includes(weather) || visibility < 2) {
            return "Difficult driving conditions present. Reduced visibility and potentially hazardous road surfaces require extreme caution if travel is necessary.";
        } else if (['Rain', 'Drizzle', 'Mist'].includes(weather) || visibility < 5) {
            return "Drive with additional caution. Leave extra stopping distance and reduce speed as road conditions may be affected by the weather.";
        } else {
            return "Generally good driving conditions. Standard road safety practices apply, with normal visibility and dry road surfaces.";
        }
    }

    // Recommendation Functions
    function getRecommendationsForRunning(temp, weather, humidity, wind, uvi) {
        const recommendations = [];

        if (temp < 10) {
            recommendations.push("Wear layers to stay warm and consider gloves/hat");
            recommendations.push("Warm up thoroughly before running");
        } else if (temp > 25) {
            recommendations.push("Run during cooler morning/evening hours");
            recommendations.push("Stay hydrated and wear breathable clothing");
        }

        if (humidity > 70) {
            recommendations.push("Reduce intensity and duration of your run");
            recommendations.push("Wear moisture-wicking clothing");
        }

        if (wind > 5) {
            recommendations.push("Plan route with wind direction in mind (start into the wind)");
            recommendations.push("Wear wind-resistant outer layer if needed");
        }

        if (['Rain', 'Drizzle'].includes(weather)) {
            recommendations.push("Wear water-resistant jacket and reflective gear");
            recommendations.push("Avoid slippery surfaces and puddles");
        }

        if (uvi > 5) {
            recommendations.push("Apply sweat-resistant sunscreen (SPF 30+)");
            recommendations.push("Consider wearing a hat and sunglasses");
        }

        if (recommendations.length === 0) {
            recommendations.push("Standard running precautions apply");
            recommendations.push("Stay hydrated and listen to your body");
        }

        return recommendations;
    }

    function getRecommendationsForCycling(temp, weather, wind, uvi) {
        const recommendations = [];

        if (temp < 5) {
            recommendations.push("Wear thermal cycling gear and windproof layers");
            recommendations.push("Protect extremities with gloves, shoe covers, and balaclava");
        } else if (temp > 30) {
            recommendations.push("Ride early morning or evening to avoid heat");
            recommendations.push("Increase fluid intake and consider electrolyte replacement");
        }

        if (wind > 5) {
            recommendations.push("Plan route to have tailwind on return leg");
            recommendations.push("Use deeper section wheels with caution in crosswinds");
        }

        if (['Rain', 'Drizzle'].includes(weather)) {
            recommendations.push("Install fenders/mudguards if riding with others");
            recommendations.push("Brake earlier and avoid painted road markings");
            recommendations.push("Use lights even during daytime for visibility");
        }

        if (uvi > 5) {
            recommendations.push("Apply sunscreen to exposed skin");
            recommendations.push("Consider wearing arm sleeves for sun protection");
        }

        if (recommendations.length === 0) {
            recommendations.push("Standard cycling safety checks apply");
            recommendations.push("Carry basic repair kit and hydration");
        }

        return recommendations;
    }

    function getRecommendationsForHiking(temp, weather, visibility, uvi) {
        const recommendations = [];

        if (temp < 5) {
            recommendations.push("Wear insulated layers and protect against wind chill");
            recommendations.push("Pack emergency blanket and extra warm clothing");
        } else if (temp > 30) {
            recommendations.push("Hike early to avoid peak heat");
            recommendations.push("Carry more water than usual (minimum 3L)");
        }

        if (visibility < 5) {
            recommendations.push("Stay on marked trails and use GPS/map");
            recommendations.push("Hike with a partner and carry whistle");
        }

        if (['Rain', 'Drizzle'].includes(weather)) {
            recommendations.push("Wear waterproof boots and gaiters");
            recommendations.push("Pack extra socks and rain gear");
        }

        if (uvi > 5) {
            recommendations.push("Wear wide-brimmed hat and UV-protective clothing");
            recommendations.push("Reapply sunscreen every 2 hours");
        }

        if (recommendations.length === 0) {
            recommendations.push("Carry the 10 essentials for hiking");
            recommendations.push("Tell someone your planned route and return time");
        }

        return recommendations;
    }

    function getRecommendationsForGardening(temp, weather) {
        const recommendations = [];

        if (temp < 10) {
            recommendations.push("Focus on cold-weather tasks like pruning dormant plants");
            recommendations.push("Protect tender plants with frost cloth or mulch");
        } else if (temp > 28) {
            recommendations.push("Work in early morning or late afternoon");
            recommendations.push("Water plants deeply in the cooler hours");
        }

        if (weather === 'Rain') {
            recommendations.push("Avoid working with wet soil to prevent compaction");
            recommendations.push("Collect rainwater for future use");
        }

        if (weather === 'Snow') {
            recommendations.push("Postpone gardening until snow melts");
            recommendations.push("Gently brush snow from evergreen branches if heavy");
        }

        if (recommendations.length === 0) {
            recommendations.push("Ideal conditions for planting and soil work");
            recommendations.push("Apply mulch to retain moisture and suppress weeds");
        }

        return recommendations;
    }

    function getRecommendationsForConstruction(temp, weather, wind) {
        const recommendations = [];

        if (temp < 5) {
            recommendations.push("Use cold-weather concrete mixes if pouring");
            recommendations.push("Provide heated break areas for workers");
        } else if (temp > 30) {
            recommendations.push("Implement frequent hydration breaks (every 15-20 min)");
            recommendations.push("Schedule heavy work for cooler morning hours");
        }

        if (wind > 7) {
            recommendations.push("Secure loose materials and scaffolding");
            recommendations.push("Avoid crane operations in high winds");
        }

        if (['Rain', 'Drizzle'].includes(weather)) {
            recommendations.push("Use non-slip footwear and proper rain gear");
            recommendations.push("Cover electrical tools and connections");
        }

        if (recommendations.length === 0) {
            recommendations.push("Standard safety protocols apply");
            recommendations.push("Conduct regular toolbox talks on weather awareness");
        }

        return recommendations;
    }

    function getRecommendationsForRespiratoryHealth(temp, weather, humidity) {
        const recommendations = [];

        if (humidity < 30) {
            recommendations.push("Use saline nasal sprays to moisten nasal passages");
            recommendations.push("Consider a humidifier in living/sleeping areas");
        } else if (humidity > 75) {
            recommendations.push("Use dehumidifier to reduce mold/dust mite exposure");
            recommendations.push("Ensure proper ventilation in damp areas");
        }

        if (['Dust', 'Smoke'].includes(weather)) {
            recommendations.push("Limit outdoor activities when air quality is poor");
            recommendations.push("Use N95 mask if going outside is necessary");
        }

        if (temp < 5 || temp > 30) {
            recommendations.push("Breathe through nose to warm/humidify air");
            recommendations.push("Keep rescue medications easily accessible");
        }

        if (recommendations.length === 0) {
            recommendations.push("Good conditions for respiratory health");
            recommendations.push("Continue normal management of any conditions");
        }

        return recommendations;
    }

    function getRecommendationsForDriving(weather, visibility) {
        const recommendations = [];

        if (visibility < 5) {
            recommendations.push("Use low beam headlights (high beams reduce visibility in fog)");
            recommendations.push("Increase following distance to 4+ seconds");
        }

        if (['Rain', 'Drizzle'].includes(weather)) {
            recommendations.push("Check tire tread depth and inflation");
            recommendations.push("Avoid cruise control on wet roads");
        }

        if (weather === 'Snow') {
            recommendations.push("Carry winter emergency kit (blanket, shovel, etc.)");
            recommendations.push("Practice braking in empty parking lot to test traction");
        }

        if (weather === 'Fog') {
            recommendations.push("Use fog lights if available (not just hazard lights)");
            recommendations.push("Listen for traffic you can't see at intersections");
        }

        if (recommendations.length === 0) {
            recommendations.push("Standard safe driving practices apply");
            recommendations.push("Ensure all lights and wipers are functional");
        }

        return recommendations;
    }
});