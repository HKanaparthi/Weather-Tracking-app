/**
 * Weather Impact App Integration - Complete Solution
 *
 * This script integrates the Weather Impact Assessment functionality with the weather application.
 * It handles weather impact analysis, filtering, and detailed modal views.
 */

document.addEventListener('DOMContentLoaded', function() {
    console.log('Weather Impact App Integration initialized');

    // DOM Elements
    const elements = {
        searchForm: document.querySelector('.search-form'),
        cityInput: document.querySelector('.search-input'),
        currentWeatherSummary: document.getElementById('current-weather-summary'),
        impactFilter: document.getElementById('impact-filter'),
        impactCards: document.getElementById('impact-cards'),
        weatherAlert: document.getElementById('weather-alert'),
        alertText: document.getElementById('alert-text'),
        modal: document.getElementById('impact-modal'),
        modalClose: document.getElementById('modal-close'),
        modalTitle: document.getElementById('modal-title'),
        modalBody: document.getElementById('modal-body'),
        noResults: document.getElementById('no-results')
    };

    // Initialize the app
    init();

    function init() {
        setupEventListeners();
        checkInitialState();
    }

    function setupEventListeners() {
        // Form submission
        if (elements.searchForm) {
            elements.searchForm.addEventListener('submit', handleFormSubmit);
        }

        // Category filtering
        const categories = document.querySelectorAll('.impact-category');
        categories.forEach(category => {
            category.addEventListener('click', handleCategoryClick);
        });

        // Modal interactions
        if (elements.modalClose) {
            elements.modalClose.addEventListener('click', closeModal);
        }

        window.addEventListener('click', function(event) {
            if (event.target === elements.modal) {
                closeModal();
            }
        });

        // Impact card clicks
        document.addEventListener('click', function(event) {
            const card = event.target.closest('.impact-card');
            if (card) {
                showDetailedImpact(card);
            }
        });
    }

    function checkInitialState() {
        const urlParams = new URLSearchParams(window.location.search);
        const city = urlParams.get('city');

        // Show/hide UI elements based on initial state
        if (elements.currentWeatherSummary) {
            elements.currentWeatherSummary.style.display = city ? 'flex' : 'none';
        }
        if (elements.impactFilter) {
            elements.impactFilter.style.display = city ? 'block' : 'none';
        }
        if (elements.impactCards) {
            elements.impactCards.style.display = city ? 'grid' : 'none';
        }
        if (elements.noResults) {
            elements.noResults.style.display = city ? 'none' : 'block';
        }

        // Check for weather alerts
        if (city && elements.weatherAlert) {
            checkWeatherAlerts(city);
        }
    }

    function handleFormSubmit(e) {
        const city = elements.cityInput?.value.trim();

        if (!city) {
            e.preventDefault();
            showError('Please enter a city name');
            return;
        }

        console.log(`Searching for weather impact in: ${city}`);
        // Additional form handling logic can go here
    }

    function handleCategoryClick() {
        const selectedCategory = this.getAttribute('data-category');
        console.log(`Filtering by category: ${selectedCategory}`);

        // Update UI
        document.querySelectorAll('.impact-category').forEach(cat => {
            cat.classList.toggle('active', cat === this);
        });

        // Filter cards
        filterImpactCards(selectedCategory);
    }

    function filterImpactCards(category) {
        const cards = document.querySelectorAll('.impact-card');
        let visibleCount = 0;

        cards.forEach(card => {
            const shouldShow = category === 'all' || card.getAttribute('data-category') === category;
            card.style.display = shouldShow ? 'block' : 'none';
            if (shouldShow) visibleCount++;
        });

        // Show/hide no results message
        if (elements.noResults) {
            elements.noResults.style.display = visibleCount === 0 ? 'block' : 'none';
        }
    }

    function checkWeatherAlerts(city) {
        // In a real app, this would fetch actual alert data
        const shouldShowAlert = Math.random() > 0.7; // 30% chance for demo

        if (shouldShowAlert && elements.weatherAlert && elements.alertText) {
            const alerts = [
                `Weather advisory for ${city}: Possible precipitation in the next few hours.`,
                `Heat advisory for ${city}: Stay hydrated and limit outdoor exposure.`,
                `Air quality alert for ${city}: Consider limiting outdoor activities if sensitive.`,
                `Wind advisory for ${city}: Secure outdoor objects and use caution.`
            ];

            elements.alertText.textContent = alerts[Math.floor(Math.random() * alerts.length)];
            elements.weatherAlert.classList.add('active');

            // Auto-hide after 10 seconds
            setTimeout(() => {
                elements.weatherAlert.classList.remove('active');
            }, 10000);
        }
    }

    function showDetailedImpact(card) {
        if (!elements.modal || !elements.modalTitle || !elements.modalBody) return;

        try {
            // Extract data from card
            const title = card.querySelector('.impact-card-title')?.textContent.trim() || 'Activity';
            const rating = card.querySelector('.impact-card-rating-text')?.textContent.trim() || 'Moderate';
            const ratingClass = card.querySelector('.impact-card-rating-text')?.classList[1] || 'moderate';
            const description = card.querySelector('.impact-card-description')?.textContent.trim() || 'No description available';

            // Build factors HTML
            let factorsHTML = '';
            const factorElements = card.querySelectorAll('.impact-card-factor');
            factorElements.forEach(factor => {
                const name = factor.querySelector('.impact-card-factor-name')?.textContent || 'Factor';
                const value = factor.querySelector('.impact-card-factor-value')?.textContent || 'N/A';
                const valueClass = factor.querySelector('.impact-card-factor-value')?.classList[1] || 'moderate';

                factorsHTML += `
                    <div class="modal-factor">
                        <span class="impact-card-factor-name">${name}</span>
                        <span class="impact-card-factor-value ${valueClass}">${value}</span>
                    </div>
                `;
            });

            // Get recommendations
            const recommendations = getRecommendationsForActivity(title, rating);

            // Set modal content
            elements.modalTitle.textContent = title;
            elements.modalBody.innerHTML = `
                <div class="modal-rating">
                    <i class="fas fa-circle ${ratingClass}"></i>
                    <span class="impact-card-rating-text ${ratingClass}">${rating}</span>
                </div>
                <div class="modal-description">
                    ${description}
                </div>
                <div class="modal-factors">
                    ${factorsHTML || '<p>No factor data available</p>'}
                </div>
                <div class="modal-recommendations">
                    <h3>Recommendations</h3>
                    ${recommendations.map(rec => `
                        <div class="recommendation-item">
                            <span class="recommendation-icon"><i class="fas fa-check-circle"></i></span>
                            <span>${rec}</span>
                        </div>
                    `).join('')}
                </div>
            `;

            // Show modal
            elements.modal.classList.add('active');
            document.body.style.overflow = 'hidden'; // Prevent scrolling behind modal
        } catch (error) {
            console.error('Error showing detailed impact:', error);
            showError('Could not load impact details');
        }
    }

    function closeModal() {
        if (elements.modal) {
            elements.modal.classList.remove('active');
            document.body.style.overflow = ''; // Restore scrolling
        }
    }

    function getRecommendationsForActivity(activity, rating) {
        const recommendations = [
            "Check the forecast regularly for weather changes",
            "Prepare appropriate clothing for current conditions"
        ];

        // Activity-specific recommendations
        if (activity.match(/running|jogging/i)) {
            recommendations.push(
                "Stay hydrated before, during, and after your run",
                "Wear reflective gear if running in low visibility",
                "Choose routes with available shelter"
            );
        }
        else if (activity.match(/cycling|biking/i)) {
            recommendations.push(
                "Ensure your bike is in good working condition",
                "Use appropriate tire pressure for the conditions",
                "Carry a basic repair kit"
            );
        }
        else if (activity.match(/hiking|trekking/i)) {
            recommendations.push(
                "Carry a map and compass even if using GPS",
                "Pack layers to adjust to changing temperatures",
                "Tell someone your route and expected return time"
            );
        }
        else if (activity.match(/garden|planting/i)) {
            recommendations.push(
                "Focus on tasks appropriate for current soil conditions",
                "Keep sensitive plants protected from extreme elements",
                "Take breaks to prevent overexertion"
            );
        }
        else if (activity.match(/construction|building/i)) {
            recommendations.push(
                "Prioritize safety in all weather conditions",
                "Schedule weather-sensitive tasks appropriately",
                "Ensure proper equipment for current conditions"
            );
        }
        else if (activity.match(/health|respiratory/i)) {
            recommendations.push(
                "Monitor air quality reports daily",
                "Keep medications easily accessible",
                "Limit exposure during peak pollen or pollution times"
            );
        }
        else if (activity.match(/driving|travel/i)) {
            recommendations.push(
                "Maintain safe speeds for current conditions",
                "Keep extra distance between vehicles in poor conditions",
                "Ensure headlights and windshield wipers are functioning"
            );
        }

        // Rating-specific recommendations
        if (rating === "Unfavorable") {
            recommendations.push(
                "Consider alternative indoor activities if possible",
                "Prepare for potential hazards specific to current conditions"
            );
        } else if (rating === "Moderate") {
            recommendations.push(
                "Proceed with caution and monitor conditions",
                "Have contingency plans ready"
            );
        }

        // Return 4-5 most relevant recommendations
        return recommendations.slice(0, 5);
    }

    function showError(message) {
        console.error(message);
        // In a real app, you would show this to the user
        // alert(message); or use a proper error display element
    }
});