// save-city.js - Common functions for saving and managing cities

document.addEventListener('DOMContentLoaded', function() {
    // Initialize save city buttons
    initSaveCityButtons();
});

// Initialize save city buttons
function initSaveCityButtons() {
    const saveCityBtn = document.getElementById('save-city-btn');
    if (saveCityBtn) {
        // Get current city name from the page
        const cityName = saveCityBtn.dataset.city;

        // Check if this city is already saved
        checkIfCitySaved(cityName).then(isSaved => {
            updateSaveButtonState(saveCityBtn, isSaved);
        });

        // Add click event to toggle saved state
        saveCityBtn.addEventListener('click', function(e) {
            e.preventDefault();
            const isSaved = saveCityBtn.classList.contains('saved');

            if (isSaved) {
                removeCity(cityName).then(success => {
                    if (success) {
                        updateSaveButtonState(saveCityBtn, false);
                        showToast('City removed from your saved cities');
                    }
                });
            } else {
                saveCity(cityName).then(success => {
                    if (success) {
                        updateSaveButtonState(saveCityBtn, true);
                        showToast('City saved to your favorites');
                    }
                });
            }
        });
    }
}

// Check if a city is already saved for the current user
async function checkIfCitySaved(cityName) {
    try {
        const response = await fetch('/api/cities/saved');
        if (!response.ok) {
            console.error('Failed to fetch saved cities');
            return false;
        }

        const data = await response.json();
        return data.cities && data.cities.includes(cityName);
    } catch (error) {
        console.error('Error checking if city is saved:', error);
        return false;
    }
}

// Save a city for the current user
async function saveCity(cityName) {
    try {
        const response = await fetch('/api/cities/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                city_name: cityName
            })
        });

        if (!response.ok) {
            console.error('Failed to save city');
            return false;
        }

        return true;
    } catch (error) {
        console.error('Error saving city:', error);
        return false;
    }
}

// Remove a saved city for the current user
async function removeCity(cityName) {
    try {
        const response = await fetch('/api/cities/remove', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                city_name: cityName
            })
        });

        if (!response.ok) {
            console.error('Failed to remove city');
            return false;
        }

        return true;
    } catch (error) {
        console.error('Error removing city:', error);
        return false;
    }
}

// Update the save button appearance based on saved state
function updateSaveButtonState(button, isSaved) {
    if (isSaved) {
        button.classList.add('saved');
        button.innerHTML = '<i class="fas fa-star"></i> Saved';
    } else {
        button.classList.remove('saved');
        button.innerHTML = '<i class="far fa-star"></i> Save City';
    }
}

// Show a toast notification
function showToast(message) {
    // Check if toast container exists, create if not
    let toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toast-container';
        document.body.appendChild(toastContainer);
    }

    // Create toast element
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;

    // Add to container
    toastContainer.appendChild(toast);

    // Show animation
    setTimeout(() => {
        toast.classList.add('show');
    }, 10);

    // Hide and remove after 3 seconds
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => {
            toastContainer.removeChild(toast);
        }, 300);
    }, 3000);
}