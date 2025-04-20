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

    // Message template
    const messageTemplate = document.getElementById('message-template');

    // Variables
    let selectedImage = null;
    let chatRefreshInterval;
    let weatherRefreshInterval;

    // Initialize
    init();

    function init() {
        // Load initial chat messages
        loadChatRoom();

        // Load weather data
        loadWeatherData();

        // Setup event listeners
        setupEventListeners();

        // Start refresh intervals
        startRefreshIntervals();
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
        chatMessages.addEventListener('DOMNodeInserted', () => {
            chatMessages.scrollTop = chatMessages.scrollHeight;
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

            // Update messages
            updateChatMessages(data.messages);

            // Update recent photos
            updateRecentPhotos(data.messages);
        } catch (error) {
            console.error('Error loading chat room:', error);
        }
    }

    // Load weather data
    async function loadWeatherData() {
        try {
            const response = await fetch(`/api/weather/current?city=${encodeURIComponent(cityName)}`);
            if (!response.ok) throw new Error('Failed to load weather data');

            const data = await response.json();

            // Update current weather
            document.getElementById('current-temp').textContent = `${Math.round(data.main.temp)}°`;
            document.getElementById('current-condition').textContent = data.weather[0].main;

            // Update weather summary
            weatherSummary.innerHTML = `
                <div class="weather-detail">
                    <i class="fas fa-temperature-high"></i>
                    <span>High: ${Math.round(data.main.temp_max)}°</span>
                </div>
                <div class="weather-detail">
                    <i class="fas fa-temperature-low"></i>
                    <span>Low: ${Math.round(data.main.temp_min)}°</span>
                </div>
                <div class="weather-detail">
                    <i class="fas fa-wind"></i>
                    <span>Wind: ${Math.round(data.wind.speed)} mph</span>
                </div>
                <div class="weather-detail">
                    <i class="fas fa-tint"></i>
                    <span>Humidity: ${data.main.humidity}%</span>
                </div>
            `;
        } catch (error) {
            console.error('Error loading weather data:', error);
        }
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

    // Create message element
    function createMessageElement(msg) {
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

        // Mark own messages
        if (msg.user_id === parseInt(userID)) {
            messageEl.classList.add('own-message');
        }

        return messageEl;
    }

    // Update recent photos
    function updateRecentPhotos(messages) {
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
            photoEl.innerHTML = `<img src="${msg.image_url}" alt="Weather photo by ${msg.username}">`;

            // Add click event to show full image
            photoEl.addEventListener('click', () => {
                const fullImg = document.createElement('div');
                fullImg.className = 'fullscreen-image';
                fullImg.innerHTML = `
                    <div class="fullscreen-image-content">
                        <img src="${msg.image_url}" alt="Weather photo by ${msg.username}">
                        <div class="fullscreen-caption">
                            <p>Shared by ${msg.username} at ${formatTime(new Date(msg.created_at))}</p>
                        </div>
                        <button class="close-fullscreen">&times;</button>
                    </div>
                `;

                document.body.appendChild(fullImg);

                // Close on click
                fullImg.addEventListener('click', e => {
                    if (e.target === fullImg || e.target.className === 'close-fullscreen') {
                        document.body.removeChild(fullImg);
                    }
                });
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