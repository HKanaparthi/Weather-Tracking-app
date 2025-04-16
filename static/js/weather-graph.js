// weather-graph.js - Creates interactive charts for weather comparison visualization

// Global chart instances
let tempChart = null;
let humidityChart = null;
let windChart = null;

/**
 * Creates weather comparison graphs using Chart.js
 * @param {Object} data - The weather comparison data
 */
function createWeatherGraph(data) {
    if (!data || !data.historical) {
        console.error("Invalid data for graph creation");
        return;
    }

    // Get graph container
    const graphContainer = document.getElementById('weather-graph');
    if (!graphContainer) {
        console.error("Graph container not found");
        return;
    }

    // Prepare data for charts
    const labels = ['6 Months Ago', '3 Months Ago', '1 Month Ago', 'Current'];

    const tempData = [
        safeParse(data.historical['6months']?.temperature),
        safeParse(data.historical['3months']?.temperature),
        safeParse(data.historical['1month']?.temperature),
        safeParse(data.current?.temperature)
    ];

    const feelsLikeData = [
        safeParse(data.historical['6months']?.feelsLike),
        safeParse(data.historical['3months']?.feelsLike),
        safeParse(data.historical['1month']?.feelsLike),
        safeParse(data.current?.feelsLike)
    ];

    const humidityData = [
        safeParse(data.historical['6months']?.humidity),
        safeParse(data.historical['3months']?.humidity),
        safeParse(data.historical['1month']?.humidity),
        safeParse(data.current?.humidity)
    ];

    const windData = [
        safeParse(data.historical['6months']?.windSpeed),
        safeParse(data.historical['3months']?.windSpeed),
        safeParse(data.historical['1month']?.windSpeed),
        safeParse(data.current?.windSpeed)
    ];

    // Create chart container elements
    graphContainer.innerHTML = `
        <h3>Weather Data Visualization</h3>
        <div class="chart-controls">
            <button id="toggle-chart-type" class="chart-button">Switch to Bar Chart</button>
            <div class="chart-selector">
                <button id="show-temp-chart" class="chart-button active">Temperature</button>
                <button id="show-humidity-chart" class="chart-button">Humidity</button>
                <button id="show-wind-chart" class="chart-button">Wind Speed</button>
            </div>
        </div>
        <div class="chart-container">
            <canvas id="temp-chart"></canvas>
            <canvas id="humidity-chart" style="display: none;"></canvas>
            <canvas id="wind-chart" style="display: none;"></canvas>
        </div>
    `;

    // Create charts
    createTemperatureChart(labels, tempData, feelsLikeData);
    createHumidityChart(labels, humidityData);
    createWindChart(labels, windData);

    // Set up chart toggle event listeners
    document.getElementById('show-temp-chart').addEventListener('click', function() {
        toggleActiveChart('temp-chart');
        updateActiveButton(this);
    });

    document.getElementById('show-humidity-chart').addEventListener('click', function() {
        toggleActiveChart('humidity-chart');
        updateActiveButton(this);});

    document.getElementById('show-wind-chart').addEventListener('click', function() {
        toggleActiveChart('wind-chart');
        updateActiveButton(this);
    });

    // Set up chart type toggle (line/bar)
    let currentChartType = 'line';
    document.getElementById('toggle-chart-type').addEventListener('click', function() {
        if (currentChartType === 'line') {
            currentChartType = 'bar';
            this.textContent = 'Switch to Line Chart';
        } else {
            currentChartType = 'line';
            this.textContent = 'Switch to Bar Chart';
        }

        // Update all charts with new type
        updateChartType(tempChart, currentChartType);
        updateChartType(humidityChart, currentChartType);
        updateChartType(windChart, currentChartType);
    });
}

/**
 * Creates temperature comparison chart
 * @param {Array} labels - Chart labels
 * @param {Array} temperatureData - Temperature data points
 * @param {Array} feelsLikeData - Feels like temperature data points
 */
function createTemperatureChart(labels, temperatureData, feelsLikeData) {
    const ctx = document.getElementById('temp-chart').getContext('2d');

    // Destroy previous chart instance if it exists
    if (tempChart) {
        tempChart.destroy();
    }

    tempChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Temperature (°C)',
                    data: temperatureData,
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: 'rgba(54, 162, 235, 1)',
                    borderWidth: 2,
                    tension: 0.1
                },
                {
                    label: 'Feels Like (°C)',
                    data: feelsLikeData,
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    borderColor: 'rgba(255, 99, 132, 1)',
                    borderWidth: 2,
                    tension: 0.1
                }
            ]
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
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Temperature Comparison'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `${context.dataset.label}: ${context.raw.toFixed(1)}°C`;
                        }
                    }
                }
            }
        }
    });
}

/**
 * Creates humidity comparison chart
 * @param {Array} labels - Chart labels
 * @param {Array} humidityData - Humidity data points
 */
function createHumidityChart(labels, humidityData) {
    const ctx = document.getElementById('humidity-chart').getContext('2d');

    // Destroy previous chart instance if it exists
    if (humidityChart) {
        humidityChart.destroy();
    }

    humidityChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Humidity (%)',
                    data: humidityData,
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    borderColor: 'rgba(75, 192, 192, 1)',
                    borderWidth: 2,
                    tension: 0.1
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    suggestedMax: 100,
                    title: {
                        display: true,
                        text: 'Humidity (%)'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Humidity Comparison'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `Humidity: ${context.raw}%`;
                        }
                    }
                }
            }
        }
    });
}

/**
 * Creates wind speed comparison chart
 * @param {Array} labels - Chart labels
 * @param {Array} windData - Wind speed data points
 */
function createWindChart(labels, windData) {
    const ctx = document.getElementById('wind-chart').getContext('2d');

    // Destroy previous chart instance if it exists
    if (windChart) {
        windChart.destroy();
    }

    windChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Wind Speed (m/s)',
                    data: windData,
                    backgroundColor: 'rgba(153, 102, 255, 0.2)',
                    borderColor: 'rgba(153, 102, 255, 1)',
                    borderWidth: 2,
                    tension: 0.1
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Wind Speed (m/s)'
                    }
                }
            },
            plugins: {
                title: {
                    display: true,
                    text: 'Wind Speed Comparison'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `Wind Speed: ${context.raw.toFixed(1)} m/s`;
                        }
                    }
                }
            }
        }
    });
}

/**
 * Toggles visibility between charts
 * @param {string} activeChartId - The ID of the chart to show
 */
function toggleActiveChart(activeChartId) {
    const chartIds = ['temp-chart', 'humidity-chart', 'wind-chart'];

    chartIds.forEach(id => {
        const chartElement = document.getElementById(id);
        if (chartElement) {
            chartElement.style.display = id === activeChartId ? 'block' : 'none';
        }
    });
}

/**
 * Updates the active button in the chart selector
 * @param {HTMLElement} activeButton - The active button element
 */
function updateActiveButton(activeButton) {
    const buttons = document.querySelectorAll('.chart-selector .chart-button');
    buttons.forEach(button => {
        button.classList.remove('active');
    });

    activeButton.classList.add('active');
}

/**
 * Updates chart type between line and bar
 * @param {Chart} chart - Chart.js instance
 * @param {string} newType - New chart type ('line' or 'bar')
 */
function updateChartType(chart, newType) {
    if (!chart) return;

    chart.config.type = newType;
    chart.update();
}

/**
 * Safe way to parse numeric values with fallback
 * @param {*} value - Value to parse
 * @param {number} defaultValue - Default value if parsing fails
 * @returns {number} - Parsed value or default
 */
function safeParse(value, defaultValue = 0) {
    if (value === undefined || value === null || isNaN(value)) {
        return defaultValue;
    }
    return parseFloat(value);
}