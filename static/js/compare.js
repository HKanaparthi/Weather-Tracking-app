document.addEventListener('DOMContentLoaded', function() {
    const cityInput = document.getElementById('city-input');
    const addCityButton = document.getElementById('add-city');
    const clearAllButton = document.getElementById('clear-all');
    const citiesContainer = document.getElementById('cities-container');

    // Set background based on weather conditions
    updatePageBackground();

    // Add city button
    if (addCityButton) {
        addCityButton.addEventListener('click', function() {
            addCity();
        });
    }

    // Allow Enter key to submit the form
    if (cityInput) {
        cityInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                addCity();
            }
        });

        // Add validation message for empty input
        cityInput.addEventListener('focus', function() {
            this.classList.remove('input-error');
            const errorMsg = document.querySelector('.field-error');
            if (errorMsg) errorMsg.remove();
        });
    }

    // Clear all button
    if (clearAllButton) {
        clearAllButton.addEventListener('click', function() {
            window.location.href = '/compare';
        });
    }

    // Remove city buttons
    const removeCityButtons = document.querySelectorAll('.remove-city');
    removeCityButtons.forEach(button => {
        button.addEventListener('click', function() {
            const cityToRemove = this.getAttribute('data-city');
            removeCity(cityToRemove);
        });
    });

    // Function to add a city
    function addCity() {
        const city = cityInput.value.trim();

        if (city) {
            // Save original input value
            const originalText = cityInput.value;

            // Show loading state
            cityInput.disabled = true;
            cityInput.value = "Loading...";
            addCityButton.disabled = true;

            // Get current cities from URL params
            const urlParams = new URLSearchParams(window.location.search);
            const cities = urlParams.getAll('cities');

            // Check if city is already in the list
            if (!cities.includes(city)) {
                cities.push(city);

                // Build new URL with updated cities
                const newParams = new URLSearchParams();
                cities.forEach(c => newParams.append('cities', c));

                // Redirect to the updated URL
                window.location.href = `/compare?${newParams.toString()}`;
            } else {
                // City already exists, restore input state
                cityInput.disabled = false;
                cityInput.value = originalText;
                addCityButton.disabled = false;

                // Highlight existing city card
                const existingCard = document.querySelector(`.city-card[data-city="${city}"]`);
                if (existingCard) {
                    existingCard.classList.add('highlight');
                    setTimeout(() => {
                        existingCard.classList.remove('highlight');
                    }, 2000);

                    // Scroll to the card
                    existingCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }

                // Show alert
                alert(`${city} is already in your comparison`);
            }
        } else {
            // Show error for empty input
            cityInput.classList.add('input-error');

            // Check if error message already exists
            if (!document.querySelector('.field-error')) {
                const errorMsg = document.createElement('div');
                errorMsg.className = 'field-error';
                errorMsg.textContent = 'Please fill out this field.';
                errorMsg.style.color = 'white';
                errorMsg.style.backgroundColor = 'rgba(244, 67, 54, 0.8)';
                errorMsg.style.padding = '5px 10px';
                errorMsg.style.borderRadius = '4px';
                errorMsg.style.marginTop = '5px';
                errorMsg.style.fontSize = '14px';

                // Insert after the input
                cityInput.parentNode.insertBefore(errorMsg, cityInput.nextSibling);
            }
        }
    }

    // Function to remove a city
    function removeCity(cityToRemove) {
        // Get current cities from URL params
        const urlParams = new URLSearchParams(window.location.search);
        const cities = urlParams.getAll('cities');

        // Filter out the city to remove
        const updatedCities = cities.filter(city => city !== cityToRemove);

        // Build new URL with updated cities
        const newParams = new URLSearchParams();
        updatedCities.forEach(city => newParams.append('cities', city));

        // Redirect to the updated URL
        window.location.href = updatedCities.length ? `/compare?${newParams.toString()}` : '/compare';
    }

    // Function to update the page background based on weather conditions
    function updatePageBackground() {
        const weatherCards = document.querySelectorAll('.city-card');
        if (weatherCards.length === 0) return;

        // Count weather conditions
        const conditions = {};
        let dominantCondition = 'default';
        let maxCount = 0;

        weatherCards.forEach(card => {
            const conditionElement = card.querySelector('.condition');
            if (!conditionElement) return;

            const conditionText = conditionElement.textContent.toLowerCase();
            let weatherType = 'default';

            // Determine weather type from condition text
            if (conditionText.includes('clear') || conditionText.includes('sky')) {
                weatherType = 'clear';
            } else if (conditionText.includes('cloud')) {
                weatherType = 'cloud';
            } else if (conditionText.includes('rain') || conditionText.includes('drizzle')) {
                weatherType = 'rain';
            } else if (conditionText.includes('thunder') || conditionText.includes('storm')) {
                weatherType = 'thunder';
            } else if (conditionText.includes('snow') || conditionText.includes('ice')) {
                weatherType = 'snow';
            } else if (conditionText.includes('fog') || conditionText.includes('mist')) {
                weatherType = 'fog';
            }

            // Count occurrences
            conditions[weatherType] = (conditions[weatherType] || 0) + 1;

            // Update dominant condition
            if (conditions[weatherType] > maxCount) {
                maxCount = conditions[weatherType];
                dominantCondition = weatherType;
            }
        });

        // Apply background class
        document.body.className = document.body.className.replace(/weather-\w+/g, '');
        document.body.classList.add(`weather-${dominantCondition}`);
    }

    // Add styles
    const style = document.createElement('style');
    style.textContent = `
        .highlight {
            animation: pulse 1.5s;
        }
        
        @keyframes pulse {
            0% { box-shadow: 0 0 0 0 rgba(255, 255, 255, 0.7); }
            70% { box-shadow: 0 0 0 10px rgba(255, 255, 255, 0); }
            100% { box-shadow: 0 0 0 0 rgba(255, 255, 255, 0); }
        }
        
        .input-error {
            border-color: rgba(244, 67, 54, 0.8) !important;
            box-shadow: 0 0 0 2px rgba(244, 67, 54, 0.4);
        }
        
        .search-input::placeholder {
            color: rgba(255, 255, 255, 0.8);
        }
    `;
    document.head.appendChild(style);
});