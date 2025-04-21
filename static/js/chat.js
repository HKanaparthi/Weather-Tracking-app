// Chat functionality for Weather Chats

document.addEventListener('DOMContentLoaded', function() {
    // Elements
    const chatMessages = document.getElementById('chat-messages');
    const messageForm = document.getElementById('message-form');
    const messageText = document.getElementById('message-text');
    const attachImageBtn = document.getElementById('attach-image');
    const imageUpload = document.getElementById('image-upload');
    const imagePreview = document.getElementById('image-preview');
    const previewImg = document.getElementById('preview-img');
    const removeImageBtn = document.getElementById('remove-image');
    const activeUsersCount = document.getElementById('active-users-count');
    const weatherSummary = document.getElementById('weather-summary-content');
    const recentPhotosContainer = document.getElementById('recent-photos-container');
    const createPollBtn = document.getElementById('create-poll');
    const pollModal = document.getElementById('poll-modal');
    const pollForm = document.getElementById('poll-creator-form');
    const pollOptionsContainer = document.getElementById('poll-options-container');
    const addOptionBtn = document.getElementById('poll-option-add');
    const pollCloseBtn = document.querySelector('.poll-modal-close');
    const pollCancelBtn = document.getElementById('poll-cancel');

    // Message template
    const messageTemplate = document.getElementById('message-template');
    // Poll templates
    const pollTemplate = document.getElementById('poll-template');
    const pollOptionTemplate = document.getElementById('poll-option-template');

    // Variables
    let selectedImage = null;
    let chatRefreshInterval;
    let weatherRefreshInterval;
    // Make sure polls is defined globally
    window.polls = {};

    // Initialize
    init();

    function init() {
        // Verify templates exist
        validateTemplates();

        // Update city name in the header
        updateCityName();

        // Load initial chat messages
        loadChatRoom();

        // Load weather data immediately
        loadWeatherData();

        // Setup event listeners
        setupEventListeners();

        // Setup poll event listeners
        setupPollEventListeners();

        // Start refresh intervals
        startRefreshIntervals();

        // Debug poll rendering
        console.log("Init complete, checking poll templates:", {
            messageTemplate: !!messageTemplate,
            pollTemplate: !!pollTemplate,
            pollOptionTemplate: !!pollOptionTemplate,
            pollsObject: window.polls
        });
    }

    // Validate that all required templates exist
    function validateTemplates() {
        const requiredTemplates = [
            { id: 'message-template', name: 'Message Template' },
            { id: 'poll-template', name: 'Poll Template' },
            { id: 'poll-option-template', name: 'Poll Option Template' }
        ];

        for (const template of requiredTemplates) {
            if (!document.getElementById(template.id)) {
                console.error(`Required template missing: ${template.name} (${template.id})`);
            }
        }
    }

    // Make sure city name is displayed
    function updateCityName() {
        // Get all h2 elements that might contain the city name
        const cityHeaders = document.querySelectorAll('h2');
        cityHeaders.forEach(header => {
            // If the header appears to be for a city name (empty or contains a template variable)
            if (header.textContent.trim() === '' || header.textContent.includes('{{')) {
                header.textContent = cityName || 'Charlotte';
            }
        });

        // Also ensure the city name is visible in the header if it exists
        const cityInfo = document.querySelector('.city-info h2');
        if (cityInfo) {
            cityInfo.textContent = cityName || 'Charlotte';
        }
    }

    function setupEventListeners() {
        // Send message form
        messageForm.addEventListener('submit', sendMessage);

        // Image upload
        attachImageBtn.addEventListener('click', () => {
            imageUpload.click();
        });

        imageUpload.addEventListener('change', handleImageUpload);
        removeImageBtn.addEventListener('click', removeSelectedImage);

        // Auto scroll when new messages arrive
        // Replace DOMNodeInserted with MutationObserver for better performance
        const observer = new MutationObserver((mutations) => {
            chatMessages.scrollTop = chatMessages.scrollHeight;
        });

        observer.observe(chatMessages, {
            childList: true,
            subtree: true
        });
    }

    function setupPollEventListeners() {
        // Open poll modal
        createPollBtn.addEventListener('click', function() {
            pollModal.style.display = 'flex';
        });

        // Close poll modal
        function closePollModal() {
            pollModal.style.display = 'none';
            pollForm.reset();

            // Reset options to default (2 options)
            pollOptionsContainer.innerHTML = `
                <div class="poll-option-input">
                    <input type="text" class="poll-option-text" placeholder="Option 1" required>
                </div>
                <div class="poll-option-input">
                    <input type="text" class="poll-option-text" placeholder="Option 2" required>
                </div>
            `;
        }

        pollCloseBtn.addEventListener('click', closePollModal);
        pollCancelBtn.addEventListener('click', closePollModal);

        // Close modal when clicking outside
        window.addEventListener('click', function(e) {
            if (e.target === pollModal) {
                closePollModal();
            }
        });

        // Add option
        addOptionBtn.addEventListener('click', function() {
            const optionInputs = document.querySelectorAll('.poll-option-input');
            const newIndex = optionInputs.length + 1;

            const newOption = document.createElement('div');
            newOption.className = 'poll-option-input';
            newOption.innerHTML = `
                <input type="text" class="poll-option-text" placeholder="Option ${newIndex}" required>
                <button type="button" class="remove-option"><i class="fas fa-times"></i></button>
            `;

            // Add remove button functionality
            const removeBtn = newOption.querySelector('.remove-option');
            removeBtn.addEventListener('click', function() {
                newOption.remove();

                // Update placeholders
                document.querySelectorAll('.poll-option-text').forEach((input, index) => {
                    input.placeholder = `Option ${index + 1}`;
                });
            });

            pollOptionsContainer.appendChild(newOption);
        });

        // Submit poll form
        pollForm.addEventListener('submit', function(e) {
            e.preventDefault();

            const question = document.getElementById('poll-question').value.trim();
            const optionInputs = document.querySelectorAll('.poll-option-text');

            // Validate
            if (!question) {
                alert('Please enter a question');
                return;
            }

            if (optionInputs.length < 2) {
                alert('Please add at least 2 options');
                return;
            }

            // Create options array
            const options = [];
            optionInputs.forEach(input => {
                const text = input.value.trim();
                if (text) {
                    options.push({
                        id: 'option_' + Math.random().toString(36).substr(2, 9),
                        text: text,
                        votes: 0
                    });
                }
            });

            if (options.length < 2) {
                alert('Please add at least 2 non-empty options');
                return;
            }

            // Create poll object
            const pollId = 'poll_' + Math.random().toString(36).substr(2, 9);
            const pollData = {
                id: pollId,
                question: question,
                options: options,
                voters: [],
                createdBy: {
                    id: userID,
                    username: username
                },
                createdAt: new Date().toISOString()
            };

            // Store poll in memory
            window.polls[pollId] = pollData;

            // Log for debugging
            console.log("Created poll:", pollData);

            // Send poll message
            sendPollMessage(pollData);

            // Close modal
            closePollModal();
        });
    }

    function startRefreshIntervals() {
        // Refresh chat every 5 seconds
        chatRefreshInterval = setInterval(loadChatRoom, 5000);

        // Refresh weather every 10 minutes
        weatherRefreshInterval = setInterval(loadWeatherData, 600000);

        // Update activity status every minute
        setInterval(updateActivityStatus, 60000);

        // Handle page visibility changes
        document.addEventListener('visibilitychange', handleVisibilityChange);
    }

    function handleVisibilityChange() {
        if (document.visibilityState === 'visible') {
            // Page is visible,
            // Restart refresh intervals
            chatRefreshInterval = setInterval(loadChatRoom, 5000);
            weatherRefreshInterval = setInterval(loadWeatherData, 600000);

            // Force refresh
            loadChatRoom();
            updateActivityStatus();
        } else {
            // Page is hidden, clear intervals to save resources
            clearInterval(chatRefreshInterval);
            clearInterval(weatherRefreshInterval);
        }
    }

    // Load chat room data
    async function loadChatRoom() {
        try {
            const response = await fetch(`/api/chat/room?city=${encodeURIComponent(cityName)}`);
            if (!response.ok) throw new Error('Failed to load chat room');

            const data = await response.json();

            // Update active users count
            activeUsersCount.textContent = data.active_users;

            // Check for poll messages
            const pollMessages = data.messages.filter(msg =>
                msg.message && msg.message.startsWith('Poll:') && msg.poll_data
            );

            if (pollMessages.length > 0) {
                console.log("Found poll messages:", pollMessages.length);
            }

            // Update messages
            updateChatMessages(data.messages);

            // Update recent photos
            updateRecentPhotos(data.messages);

            // Process poll votes in messages
            processPollVotes(data.messages);
        } catch (error) {
            console.error('Error loading chat room:', error);
        }
    }

    // Load weather data with structure matching the weather.html template
    async function loadWeatherData() {
        // Show temporary loading state
        const currentTemp = document.getElementById('current-temp');
        const currentCondition = document.getElementById('current-condition');

        // If elements are empty, show loading indicators
        if (!currentTemp.textContent || currentTemp.textContent === '--°') {
            currentTemp.textContent = '--°';
        }

        if (!currentCondition.textContent || currentCondition.textContent === '--') {
            currentCondition.textContent = '--';
        }

        // Try all possible API endpoints until one works
        const apiEndpoints = [
            `/api/weather?city=${encodeURIComponent(cityName)}`,
            `/weather?city=${encodeURIComponent(cityName)}`,
            `/api/weather/current?city=${encodeURIComponent(cityName)}`
        ];

        let weatherData = null;
        let successfulEndpoint = '';

        // Try each endpoint until one works
        for (let endpoint of apiEndpoints) {
            try {
                console.log(`Trying to fetch weather data from ${endpoint}`);
                const response = await fetch(endpoint);
                if (response.ok) {
                    weatherData = await response.json();
                    successfulEndpoint = endpoint;
                    console.log(`Weather data loaded from ${endpoint}`);
                    break; // Exit the loop if successful
                }
            } catch (error) {
                console.log(`Failed to load from ${endpoint}:`, error);
                // Continue to next endpoint
            }
        }

        // If we got weather data from any endpoint
        if (weatherData) {
            console.log('Successfully loaded weather data structure:', weatherData);

            // Extract temperature using the same approach as weather.html
            let temp = null;

            // Try different possible locations for temperature data
            if (weatherData.Current && weatherData.Current.Temperature !== undefined) {
                temp = weatherData.Current.Temperature;
            } else if (weatherData.Current && weatherData.Current.Temp !== undefined) {
                temp = weatherData.Current.Temp;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.Temp !== undefined) {
                temp = weatherData.Current.Main.Temp;
            } else if (weatherData.temperature !== undefined) {
                temp = weatherData.temperature;
            } else if (weatherData.temp !== undefined) {
                temp = weatherData.temp;
            } else if (weatherData.main && weatherData.main.temp !== undefined) {
                temp = weatherData.main.temp;
            }

            // Format and display temperature
            if (temp !== null) {
                if (typeof temp === 'number') {
                    currentTemp.textContent = `${Math.round(temp)}°`;
                } else {
                    currentTemp.textContent = `${temp}°`;
                }
            }

            // Extract weather condition
            let condition = 'Unknown';

            // Try different possible locations for condition data
            if (weatherData.Current && weatherData.Current.Condition) {
                condition = weatherData.Current.Condition;
            } else if (weatherData.Current && weatherData.Current.Weather && weatherData.Current.Weather.length > 0) {
                condition = weatherData.Current.Weather[0].description || weatherData.Current.Weather[0].main;
            } else if (weatherData.condition) {
                condition = weatherData.condition;
            } else if (weatherData.weather && weatherData.weather.length > 0) {
                condition = weatherData.weather[0].description || weatherData.weather[0].main;
            }

            // Display condition
            currentCondition.textContent = condition;

            // Extract high temperature
            let highTemp = temp; // Default to current temp if high not available

            if (weatherData.Current && weatherData.Current.MaxTemp !== undefined) {
                highTemp = weatherData.Current.MaxTemp;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.TempMax !== undefined) {
                highTemp = weatherData.Current.Main.TempMax;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.temp_max !== undefined) {
                highTemp = weatherData.Current.Main.temp_max;
            } else if (weatherData.temp_max !== undefined) {
                highTemp = weatherData.temp_max;
            } else if (weatherData.main && weatherData.main.temp_max !== undefined) {
                highTemp = weatherData.main.temp_max;
            } else if (weatherData.high !== undefined) {
                highTemp = weatherData.high;
            }

            // Extract low temperature
            let lowTemp = temp; // Default to current temp if low not available

            if (weatherData.Current && weatherData.Current.MinTemp !== undefined) {
                lowTemp = weatherData.Current.MinTemp;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.TempMin !== undefined) {
                lowTemp = weatherData.Current.Main.TempMin;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.temp_min !== undefined) {
                lowTemp = weatherData.Current.Main.temp_min;
            } else if (weatherData.temp_min !== undefined) {
                lowTemp = weatherData.temp_min;
            } else if (weatherData.main && weatherData.main.temp_min !== undefined) {
                lowTemp = weatherData.main.temp_min;
            } else if (weatherData.low !== undefined) {
                lowTemp = weatherData.low;
            }

            // Extract wind speed
            let windSpeed = 0;

            if (weatherData.Current && weatherData.Current.WindSpeed !== undefined) {
                windSpeed = weatherData.Current.WindSpeed;
            } else if (weatherData.Current && weatherData.Current.Wind && weatherData.Current.Wind.Speed !== undefined) {
                windSpeed = weatherData.Current.Wind.Speed;
            } else if (weatherData.wind_speed !== undefined) {
                windSpeed = weatherData.wind_speed;
            } else if (weatherData.wind && weatherData.wind.speed !== undefined) {
                windSpeed = weatherData.wind.speed;
            }

            // Extract humidity
            let humidity = 0;

            if (weatherData.Current && weatherData.Current.Humidity !== undefined) {
                humidity = weatherData.Current.Humidity;
            } else if (weatherData.Current && weatherData.Current.Main && weatherData.Current.Main.Humidity !== undefined) {
                humidity = weatherData.Current.Main.Humidity;
            } else if (weatherData.humidity !== undefined) {
                humidity = weatherData.humidity;
            } else if (weatherData.main && weatherData.main.humidity !== undefined) {
                humidity = weatherData.main.humidity;
            }

            // Update weather summary
            if (weatherSummary) {
                weatherSummary.innerHTML = `
                    <div class="weather-detail">
                        <i class="fas fa-temperature-high"></i>
                        <span>High: ${Math.round(highTemp)}°</span>
                    </div>
                    <div class="weather-detail">
                        <i class="fas fa-temperature-low"></i>
                        <span>Low: ${Math.round(lowTemp)}°</span>
                    </div>
                    <div class="weather-detail">
                        <i class="fas fa-wind"></i>
                        <span>Wind: ${Math.round(windSpeed)} mph</span>
                    </div>
                    <div class="weather-detail">
                        <i class="fas fa-tint"></i>
                        <span>Humidity: ${humidity}%</span>
                    </div>
                `;
            }
        } else {
            // If all API calls failed, show error message
            console.error('All weather API endpoints failed');

            currentTemp.textContent = '--°';
            currentCondition.textContent = '--';

            if (weatherSummary) {
                weatherSummary.innerHTML = `
                    <div class="weather-detail">
                        <i class="fas fa-exclamation-circle"></i>
                        <span>Weather data unavailable</span>
                    </div>
                    <div class="weather-detail">
                        <i class="fas fa-sync"></i>
                        <span>Try refreshing the page</span>
                    </div>
                `;
            }
        }
    }

    // Process poll votes from messages
    function processPollVotes(messages) {
        messages.forEach(msg => {
            if (msg.message && msg.message.startsWith('Voted on poll:') && msg.poll_vote) {
                try {
                    const voteData = JSON.parse(msg.poll_vote);
                    const { pollId, optionId } = voteData;

                    // Update poll data if exists
                    if (window.polls[pollId]) {
                        // Check if user has already voted
                        if (!window.polls[pollId].voters.includes(msg.user_id)) {
                            window.polls[pollId].voters.push(msg.user_id);

                            // Update votes
                            window.polls[pollId].options.forEach(option => {
                                if (option.id === optionId) {
                                    option.votes++;
                                }
                            });

                            // Update UI
                            updatePollUI(pollId);
                        }
                    }
                } catch (error) {
                    console.error('Error processing poll vote:', error);
                }
            }
        });
    }

    // Update activity status
    async function updateActivityStatus() {
        try {
            await fetch('/api/chat/activity', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    city_name: cityName
                })
            });
        } catch (error) {
            console.error('Error updating activity status:', error);
        }
    }

    // Send a new message
    async function sendMessage(event) {
        event.preventDefault();

        const message = messageText.value.trim();
        if (!message && !selectedImage) return;

        // Prepare message data
        const messageData = {
            city_name: cityName,
            message: message
        };

        // Handle image upload
        if (selectedImage) {
            try {
                const imageUrl = await uploadImage(selectedImage);
                messageData.image_url = imageUrl;
            } catch (error) {
                console.error('Error uploading image:', error);
                alert('Failed to upload image. Please try again.');
                return;
            }
        }

        // Send message to server
        try {
            const response = await fetch('/api/chat/message', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(messageData)
            });

            if (!response.ok) throw new Error('Failed to send message');

            // Clear form
            messageText.value = '';
            removeSelectedImage();

            // Force refresh chat
            loadChatRoom();
        } catch (error) {
            console.error('Error sending message:', error);
            alert('Failed to send message. Please try again.');
        }
    }

    // Send poll message
    async function sendPollMessage(pollData) {
        // Create message data
        const messageData = {
            city_name: cityName,
            message: 'Poll: ' + pollData.question,
            poll_data: JSON.stringify(pollData)
        };

        console.log("Sending poll message:", messageData);

        try {
            const response = await fetch('/api/chat/message', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(messageData)
            });

            if (!response.ok) throw new Error('Failed to send poll');

            // Force refresh chat
            loadChatRoom();
        } catch (error) {
            console.error('Error sending poll:', error);
            alert('Failed to send poll. Please try again.');
        }
    }

    // Vote on a poll
    async function votePoll(pollId, optionId, messageId) {
        // Check if poll exists
        if (!window.polls) {
            window.polls = {};
        }

        const pollData = window.polls[pollId];
        if (!pollData) {
            console.error('Poll not found:', pollId);
            return;
        }

        // Check if user has already voted
        if (pollData.voters.includes(userID)) {
            console.log('User already voted');
            return;
        }

        // Update poll data locally first for immediate feedback
        pollData.voters.push(userID);
        pollData.options.forEach(option => {
            if (option.id === optionId) {
                option.votes++;
                // Track who voted for this option if not already tracking
                if (!option.voters) option.voters = [];
                option.voters.push(userID);
            }
        });

        // Update UI with local data
        updatePollUI(pollId);

        // Send vote to server
        try {
            const response = await fetch('/api/chat/message', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    city_name: cityName,
                    message: `Voted on poll: ${pollData.question}`,
                    poll_vote: JSON.stringify({
                        pollId: pollId,
                        optionId: optionId,
                        messageId: messageId
                    })
                })
            });

            if (!response.ok) throw new Error('Failed to send vote');

            // Force refresh chat to get updated poll data from server
            loadChatRoom();
        } catch (error) {
            console.error('Error sending vote:', error);

            // Revert vote if failed
            const index = pollData.voters.indexOf(userID);
            if (index > -1) {
                pollData.voters.splice(index, 1);
            }

            pollData.options.forEach(option => {
                if (option.id === optionId) {
                    option.votes = Math.max(0, option.votes - 1);
                    // Remove user from voters list
                    if (option.voters) {
                        const voterIndex = option.voters.indexOf(userID);
                        if (voterIndex > -1) {
                            option.voters.splice(voterIndex, 1);
                        }
                    }
                }
            });

            updatePollUI(pollId);
            alert('Failed to send vote. Please try again.');
        }
    }

    // Update poll UI after voting
    function updatePollUI(pollId) {
        // Find all poll elements with this ID
        const pollElements = document.querySelectorAll(`.poll-container[data-poll-id="${pollId}"]`);

        console.log(`Updating UI for poll ${pollId}, found ${pollElements.length} elements`);

        if (!window.polls || !window.polls[pollId]) {
            console.error(`Poll data not found for poll ID: ${pollId}`);
            return;
        }

        pollElements.forEach(pollElement => {
            const pollData = window.polls[pollId];

            // Update voters count
            const votersEl = pollElement.querySelector('.poll-voters');
            if (votersEl) {
                votersEl.textContent = `${pollData.voters.length} vote${pollData.voters.length !== 1 ? 's' : ''}`;
            }

            // Calculate total votes
            const totalVotes = pollData.options.reduce((sum, option) => sum + option.votes, 0);

            // Update each option
            pollData.options.forEach(option => {
                const optionElement = pollElement.querySelector(`.poll-option[data-option-id="${option.id}"]`);
                if (!optionElement) {
                    console.warn(`Option element not found for option ID: ${option.id}`);
                    return;
                }

                // Calculate percentage
                const percentage = totalVotes > 0 ? Math.round((option.votes / totalVotes) * 100) : 0;

                // Update percentage text
                const percentageEl = optionElement.querySelector('.poll-option-percentage');
                if (percentageEl) {
                    percentageEl.textContent = `${percentage}%`;
                }

                // Update progress bar
                const progressEl = optionElement.querySelector('.poll-option-progress');
                if (progressEl) {
                    progressEl.style.width = `${percentage}%`;
                }

                // Show this option as selected if the current user voted for it
                if (pollData.voters.includes(userID)) {
                    // Remove click handlers from all options
                    optionElement.style.cursor = 'default';

                    // Check if this was the user's choice
                    const voteData = pollData.options.find(o => o.id === option.id && o.voters && o.voters.includes(userID));
                    if (voteData) {
                        optionElement.classList.add('selected');
                    }
                }
            });
        });
    }

    // Handle image upload
    function handleImageUpload(event) {
        const file = event.target.files[0];
        if (!file) return;

        // Validate file
        if (!file.type.startsWith('image/')) {
            alert('Please select an image file.');
            return;
        }

        if (file.size > 5 * 1024 * 1024) {
            alert('Image size should be less than 5MB.');
            return;
        }

        // Store selected image
        selectedImage = file;

        // Show preview
        const reader = new FileReader();
        reader.onload = function(e) {
            previewImg.src = e.target.result;
            imagePreview.style.display = 'block';
        };
        reader.readAsDataURL(file);
    }

    // Remove selected image
    function removeSelectedImage() {
        selectedImage = null;
        imageUpload.value = '';
        imagePreview.style.display = 'none';
    }

    // Upload image to server
    async function uploadImage(file) {
        const formData = new FormData();
        formData.append('image', file);
        formData.append('purpose', 'chat');

        const response = await fetch('/api/upload/image', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) throw new Error('Failed to upload image');

        const data = await response.json();
        return data.url;
    }

    // Update chat messages
    function updateChatMessages(messages) {
        // Get current message IDs to avoid duplicates
        const currentMessageIds = Array.from(chatMessages.querySelectorAll('.message'))
            .map(el => parseInt(el.dataset.id));

        // Add new messages
        messages.forEach(msg => {
            if (!currentMessageIds.includes(msg.id)) {
                const messageEl = createMessageElement(msg);
                chatMessages.appendChild(messageEl);
            }
        });

        // Scroll to bottom if near the bottom
        if (chatMessages.scrollHeight - chatMessages.scrollTop < chatMessages.clientHeight + 100) {
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
    }

    // Create message element - improved version
    function createMessageElement(msg) {
        // Make sure the message template exists
        if (!messageTemplate) {
            console.error("Message template not found!");
            return document.createElement('div');
        }

        const messageEl = document.importNode(messageTemplate.content, true).firstElementChild;
        messageEl.dataset.id = msg.id;

        // Set user avatar
        const avatarImg = messageEl.querySelector('.message-avatar img');
        avatarImg.src = msg.avatar_url || '/static/profile_photos/default.jpg';
        avatarImg.alt = `${msg.username}'s avatar`;

        // Set message content
        messageEl.querySelector('.message-username').textContent = msg.username;
        messageEl.querySelector('.message-time').textContent = formatTime(new Date(msg.created_at));
        messageEl.querySelector('.message-text').textContent = msg.message;

        // Handle image if present
        if (msg.image_url) {
            const imageContainer = messageEl.querySelector('.message-image');
            const image = imageContainer.querySelector('img');
            image.src = msg.image_url;
            image.alt = `Shared by ${msg.username}`;
            imageContainer.style.display = 'block';
        }

        // Handle poll if present
        if (msg.message && msg.message.startsWith('Poll:') && msg.poll_data) {
            try {
                console.log("Rendering poll in message:", msg.id);

                const pollContainer = messageEl.querySelector('.message-poll');
                if (!pollContainer) {
                    console.error("Poll container not found in message element");
                    return messageEl;
                }

                // Make sure poll container is visible
                pollContainer.style.display = 'block';

                // Parse poll data
                let pollData;
                try {
                    pollData = JSON.parse(msg.poll_data);
                    console.log("Parsed poll data:", pollData);
                } catch (e) {
                    console.error("Error parsing poll data:", e);
                    return messageEl;
                }

                // Get the poll template
                if (!pollTemplate) {
                    console.error("Poll template not found!");
                    return messageEl;
                }

                // Clone the template
                const pollElement = document.importNode(pollTemplate.content, true).firstElementChild;
                if (!pollElement) {
                    console.error("Failed to clone poll template!");
                    return messageEl;
                }

                // Set poll data
                pollElement.dataset.pollId = pollData.id;

                // Set poll question
                const questionEl = pollElement.querySelector('.poll-question');
                if (questionEl) {
                    questionEl.textContent = pollData.question;
                }

                // Update voters count
                const votersEl = pollElement.querySelector('.poll-voters');
                if (votersEl) {
                    votersEl.textContent = `${pollData.voters.length} vote${pollData.voters.length !== 1 ? 's' : ''}`;
                }

                // Set author name
                const authorEl = pollElement.querySelector('.poll-author');
                if (authorEl && pollData.createdBy) {
                    authorEl.textContent = pollData.createdBy.username;
                }

                // Clear existing options (if any)
                const pollOptionsContainer = pollElement.querySelector('.poll-options');
                if (!pollOptionsContainer) {
                    console.error("Poll options container not found");
                    return messageEl;
                }

                pollOptionsContainer.innerHTML = '';

                // Calculate total votes
                const totalVotes = pollData.options.reduce((sum, option) => sum + option.votes, 0);

                // Check if current user has voted
                const hasVoted = pollData.voters.includes(userID);

                // Add options
                if (Array.isArray(pollData.options)) {
                    pollData.options.forEach(option => {
                        // Check if option is valid
                        if (!option || typeof option !== 'object') {
                            console.error("Invalid option:", option);
                            return;
                        }

                        // Make sure the poll option template exists
                        if (!pollOptionTemplate) {
                            console.error("Poll option template not found!");
                            return;
                        }

                        // Clone the option template
                        const optionElement = document.importNode(pollOptionTemplate.content, true).firstElementChild;
                        if (!optionElement) {
                            console.error("Failed to clone poll option template!");
                            return;
                        }

                        // Set option data
                        optionElement.dataset.optionId = option.id;

                        // Calculate percentage
                        const percentage = totalVotes > 0 ? Math.round((option.votes / totalVotes) * 100) : 0;

                        // Set option text
                        const labelEl = optionElement.querySelector('.poll-option-label');
                        if (labelEl) {
                            labelEl.textContent = option.text;
                        }

                        const percentageEl = optionElement.querySelector('.poll-option-percentage');
                        if (percentageEl) {
                            percentageEl.textContent = `${percentage}%`;
                        }

                        const progressEl = optionElement.querySelector('.poll-option-progress');
                        if (progressEl) {
                            progressEl.style.width = `${percentage}%`;
                        }

                        // Handle user voted state
                        if (hasVoted) {
                            optionElement.style.cursor = 'default';
                        } else {
                            // Add click event
                            optionElement.addEventListener('click', function() {
                                votePoll(pollData.id, option.id, msg.id);
                            });
                        }

                        // Add option to container
                        pollOptionsContainer.appendChild(optionElement);
                    });
                } else {
                    console.error("Poll options is not an array:", pollData.options);
                }

                // Add poll to message
                pollContainer.innerHTML = '';
                pollContainer.appendChild(pollElement);

                // Store poll in memory if not already there
                if (!window.polls) {
                    window.polls = {};
                }
                window.polls[pollData.id] = pollData;
            } catch (error) {
                console.error('Error rendering poll:', error);
            }
        }

        // Mark own messages
        if (msg.user_id === parseInt(userID)) {
            messageEl.classList.add('own-message');
        }

        return messageEl;
    }

    // Update recent photos
    function updateRecentPhotos(messages) {
        // Skip if the container doesn't exist
        if (!recentPhotosContainer) return;

        // Get messages with images, sorted by most recent
        const messagesWithImages = messages
            .filter(msg => msg.image_url)
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            .slice(0, 6); // Only show 6 most recent

        // Clear container
        recentPhotosContainer.innerHTML = '';

        // Add photos
        messagesWithImages.forEach(msg => {
            const photoEl = document.createElement('div');
            photoEl.className = 'photo-thumbnail';

            // Format time for alt text
            const msgDate = new Date(msg.created_at);
            const formattedTime = `${msgDate.getHours().toString().padStart(2, '0')}:${msgDate.getMinutes().toString().padStart(2, '0')}`;

            photoEl.innerHTML = `<img src="${msg.image_url}" alt="Weather photo by ${msg.username} at ${formattedTime}">`;

            // Add click event to show fullscreen image
            photoEl.addEventListener('click', function(event) {
                event.preventDefault();

                // Create the fullscreen container
                const fullscreenContainer = document.createElement('div');
                fullscreenContainer.className = 'fullscreen-image';

                // Create content container
                const contentContainer = document.createElement('div');
                contentContainer.className = 'fullscreen-image-content';

                // Add image and caption
                contentContainer.innerHTML = `
                    <img src="${msg.image_url}" alt="Weather photo by ${msg.username}">
                    <div class="fullscreen-caption">
                        <p>Shared by ${msg.username} at ${formatTime(new Date(msg.created_at))}</p>
                    </div>
                    <button class="close-fullscreen">&times;</button>
                `;

                // Add to document
                fullscreenContainer.appendChild(contentContainer);
                document.body.appendChild(fullscreenContainer);

                // Prevent scrolling on the body
                document.body.style.overflow = 'hidden';

                // Close on click outside or close button
                fullscreenContainer.addEventListener('click', function(e) {
                    if (e.target === fullscreenContainer || e.target.className === 'close-fullscreen') {
                        // Add exit animation
                        fullscreenContainer.style.animation = 'fadeIn 0.3s ease reverse';
                        contentContainer.style.animation = 'scaleIn 0.3s ease reverse';

                        // Remove after animation completes
                        setTimeout(() => {
                            document.body.removeChild(fullscreenContainer);
                            document.body.style.overflow = '';
                        }, 300);
                    }
                });

                // Close on escape key
                function handleEscKey(e) {
                    if (e.key === 'Escape') {
                        // Add exit animation
                        fullscreenContainer.style.animation = 'fadeIn 0.3s ease reverse';
                        contentContainer.style.animation = 'scaleIn 0.3s ease reverse';

                        // Remove after animation completes
                        setTimeout(() => {
                            document.body.removeChild(fullscreenContainer);
                            document.body.style.overflow = '';
                        }, 300);

                        // Remove event listener
                        document.removeEventListener('keydown', handleEscKey);
                    }
                }

                document.addEventListener('keydown', handleEscKey);
            });

            recentPhotosContainer.appendChild(photoEl);
        });
    }

    // Format time as HH:MM
    function formatTime(date) {
        const hours = date.getHours().toString().padStart(2, '0');
        const minutes = date.getMinutes().toString().padStart(2, '0');
        return `${hours}:${minutes}`;
    }
});