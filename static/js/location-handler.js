// Enhanced location handling with debugging
document.addEventListener('DOMContentLoaded', function() {
    console.log("Location handler loaded");

    // Check if geolocation popup elements exist
    const allowBtn = document.querySelector('.allow-btn');
    const denyBtn = document.querySelector('.deny-btn');

    if (!allowBtn || !denyBtn) {
        console.error("Geolocation buttons not found in the DOM");
        return;
    }

    // Show location popup after a delay
    setTimeout(function() {
        const popup = document.querySelector('.location-popup');
        if (popup) {
            popup.classList.add('show');
            console.log("Location popup displayed");
        } else {
            console.error("Location popup element not found");
        }
    }, 2000);

    // Handle "Allow" button click
    allowBtn.addEventListener('click', function() {
        console.log("Allow button clicked");

        const popup = document.querySelector('.location-popup');
        if (popup) {
            popup.classList.remove('show');
        }

        // Request geolocation from browser
        if (navigator.geolocation) {
            console.log("Browser supports geolocation, requesting position...");
            navigator.geolocation.getCurrentPosition(
                // Success callback
                function(position) {
                    const latitude = position.coords.latitude;
                    const longitude = position.coords.longitude;

                    console.log("Geolocation successful:", latitude, longitude);
                    console.log("Making API request to:", `/api/nearby-locations?lat=${latitude}&lon=${longitude}`);

                    // Check if multiple locations are available nearby
                    fetch(`/api/nearby-locations?lat=${latitude}&lon=${longitude}`)
                        .then(response => {
                            console.log("API response status:", response.status);
                            if (!response.ok) {
                                throw new Error(`Failed to fetch nearby locations: ${response.status} ${response.statusText}`);
                            }
                            return response.json();
                        })
                        .then(data => {
                            console.log("Nearby locations data:", JSON.stringify(data));
                            if (data.locations && data.locations.length > 1) {
                                console.log("Multiple locations found, redirecting to options page");
                                window.location.href = `/location-options?lat=${latitude}&lon=${longitude}`;
                            } else {
                                console.log("Single location found, redirecting to weather page");
                                window.location.href = `/weather?lat=${latitude}&lon=${longitude}`;
                            }
                        })
                        .catch(error => {
                            console.error("Error checking nearby locations:", error);
                            console.log("Error occurred, falling back to direct weather display");
                            window.location.href = `/weather?lat=${latitude}&lon=${longitude}`;
                        });
                },
                // Error callback
                function(error) {
                    console.error("Error getting location:", error.message, "(Code:", error.code, ")");

                    // Error codes:
                    // 1: PERMISSION_DENIED
                    // 2: POSITION_UNAVAILABLE
                    // 3: TIMEOUT

                    // Redirect to manual location selection
                    window.location.href = "/location-options";
                },
                // Options
                {
                    enableHighAccuracy: true,
                    timeout: 10000,
                    maximumAge: 0
                }
            );
        } else {
            console.error("Browser doesn't support geolocation");
            alert("Your browser doesn't support geolocation. Please try a different browser or enter your location manually.");
            // Redirect to manual location entry page
            window.location.href = "/location-options";
        }
    });

    // Handle "Deny" button click
    denyBtn.addEventListener('click', function() {
        console.log("Deny button clicked");

        const popup = document.querySelector('.location-popup');
        if (popup) {
            popup.classList.remove('show');
        }

        // Redirect to manual location selection
        window.location.href = "/location-options";
    });
});