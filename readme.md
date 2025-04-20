🌤️GoWeather - Advanced Weather Tracking Web App


[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go&logoColor=white)](https://golang.org/)  
[![OpenWeatherMap](https://img.shields.io/badge/API-OpenWeatherMap-orange)](https://openweathermap.org/api)

A full-featured, visually rich weather application built with Go (Gin), MySQL, and OpenWeatherMap APIs. From real-time weather reports to historical analysis, activity-based suggestions, and chat integration — this app is your smart weather assistant with an intuitive UI.


## 🚀 Live Features

•  🔐 User Authentication: Secure login and signup with bcrypt-hashed passwords
•  📍 Live Weather: Real-time weather by city or device geolocation
•  📅 Hourly & Daily Forecasts: Temperature, humidity, pressure, wind, UV index, and AQI
•  🌐 World Clock: Track local time in multiple cities
•  📊 Historical Data: Weather comparison across dates and locations
•  🗺️ Travel Weather Planner: Weather-based trip planning assistant
•  💬 Real-Time Chat: In-app messaging for connected users
•  🔔 Notifications: Custom alerts based on user-defined thresholds
•  📸 Profile & Activity Tracking: Change different avatars for your Profile as you like, and plan what to do.
•  🎨 Animated UI: Weather-themed, responsive design with smooth transitions


## 🛠️ Tech Stack

| Layer        | Technology                               |
|--------------|------------------------------------------|
| **Backend**  | Go (Gin Framework), MySQL                |
| **Frontend** | HTML, CSS (custom & responsive), JS      |
| **APIs**     | OpenWeatherMap (One Call, AQI, UV Index) |
| **Security** | bcrypt, secure cookie sessions           |
| **Scheduler**| Go routines for periodic updates         |



## 📁 Project Structure
```bash
.
├── main.go                       # App entrypoint
├── emailtest.go                  # Email integration test
├── handlers/                     # Route handlers (auth, weather, chat, etc.)
├── middleware/                   # Auth & session middleware
├── models/                       # Data models (User, Chat, Activity)
├── services/                     # Business logic, API integrations
├── static/                       # Frontend assets (CSS, JS, images)
├── templates/                    # HTML templates (rendered via Gin)
├── go.mod / go.sum               # Go module dependencies
└── README.md
```




## 🧱 Database Setup

### 1. Create MySQL Database and User

```sql
CREATE DATABASE weather_app;

CREATE USER 'weather_user'@'localhost' IDENTIFIED BY 'password123';
GRANT ALL PRIVILEGES ON weather_app.* TO 'weather_user'@'localhost';
FLUSH PRIVILEGES;
```

### 2. Create `users` Table

```sql
USE weather_app;
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(100) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  home_city VARCHAR(100),
  notifications_enabled BOOLEAN DEFAULT FALSE,
  alert_threshold VARCHAR(50) DEFAULT 'all',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```



## 🔑 Configure OpenWeatherMap API Key

In `main.go`, either:

```go
const OPENWEATHER_API_KEY = "your_actual_api_key"
```

Or better yet, store the key in an environment variable and use `os.Getenv("OPENWEATHER_API_KEY")`.



## 🚀 Run the App

```bash
go run main.go
```

Then open your browser and navigate to:  
👉 `http://localhost:8080`

---




