// Location options handling script
document.addEventListener('DOMContentLoaded', function() {
    // Get URL parameters for lat and lon if they exist
    const urlParams = new URLSearchParams(window.location.search);
    const lat = urlParams.get('lat');
    const lon = urlParams.get('lon');

    // If we have coordinates in the URL but no location cards are displayed yet,
    // fetch nearby locations using the coordinates
    if (lat && lon && document.querySelectorAll('.location-card').length === 0) {
        fetchNearbyLocations(lat, lon);
    }

    // Function to fetch nearby locations using coordinates
    function fetchNearbyLocations(latitude, longitude) {
        // Show loading state
        const locationOptions = document.querySelector('.location-options');
        if (locationOptions) {
            locationOptions.innerHTML = '<div class="loading">Loading nearby locations...</div>';
        }

        // Fetch nearby locations from the server
        fetch(`/api/nearby-locations?lat=${latitude}&lon=${longitude}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to fetch nearby locations');
                }
                return response.json();
            })
            .then(data => {
                // If we have locations, display them
                if (data.locations && data.locations.length > 0) {
                    displayLocations(data.locations);
                } else {
                    // If no locations found, redirect directly to weather page
                    window.location.href = `/weather?lat=${latitude}&lon=${longitude}`;
                }
            })
            .catch(error => {
                console.error('Error fetching nearby locations:', error);
                // On error, redirect to the weather page with coordinates
                window.location.href = `/weather?lat=${latitude}&lon=${longitude}`;
            });
    }

    // Function to display the locations in the UI
    function displayLocations(locations) {
        const locationOptions = document.querySelector('.location-options');
        if (!locationOptions) return;

        // Clear any loading message
        locationOptions.innerHTML = '';

        // Add each location as a card
        locations.forEach(location => {
            const card = document.createElement('a');
            card.href = `/weather?lat=${location.lat}&lon=${location.lon}`;
            card.className = 'location-card';

            const distanceText = location.distance === 0 ?
                'Current' :
                (location.distance < 15 ? 'In City' : `${location.distance.toFixed(1)} km`);

            card.innerHTML = `
        <div class="location-info">
          <div class="location-name">${location.name}</div>
          <div class="location-detail">
            ${location.state ? location.state + ', ' : ''}${location.country}
          </div>
        </div>
        <div class="distance">
          ${distanceText}
        </div>
      `;

            // Add click event
            card.addEventListener('click', function(e) {
                e.preventDefault();
                window.location.href = this.href;
            });

            locationOptions.appendChild(card);
        });
    }
});