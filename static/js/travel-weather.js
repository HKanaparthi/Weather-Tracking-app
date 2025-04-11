document.addEventListener('DOMContentLoaded', function() {
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
        document.getElementById('origin-temp').textContent = `${Math.round(data.origin.weather.temperature)}°F`;
        document.getElementById('origin-condition').textContent = data.origin.weather.condition;
        document.getElementById('origin-wind').textContent = `${data.origin.weather.wind_speed} mph`;
        document.getElementById('origin-humidity').textContent = `${data.origin.weather.humidity}%`;
        document.getElementById('origin-precipitation').textContent = `${data.origin.weather.precipitation}%`;

        // Update destination data
        document.getElementById('destination-name').textContent = data.destination.location;
        document.getElementById('destination-temp').textContent = `${Math.round(data.destination.weather.temperature)}°F`;
        document.getElementById('destination-condition').textContent = data.destination.weather.condition;
        document.getElementById('destination-wind').textContent = `${data.destination.weather.wind_speed} mph`;
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
                <div class="day-temp">${Math.round(day.high)}°F</div>
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
});