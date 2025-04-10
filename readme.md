
# 🌤️ GoWeather Premium - Weather Tracking Dashboard

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go&logoColor=white)](https://golang.org/)  
[![OpenWeatherMap](https://img.shields.io/badge/API-OpenWeatherMap-orange)](https://openweathermap.org/api)

**GoWeather Premium** is a modern, responsive weather monitoring web application built using **Go (Gin framework)** and **MySQL**. It integrates with **OpenWeatherMap APIs** to deliver real-time and historical weather data, air quality, UV index, and local time tracking — all wrapped in a sleek, animated dashboard.



## 🌟 Features

- 🔐 Secure user authentication (Signup/Login)
- 🌍 Real-time weather data by city or geolocation
- 🕒 World Clock widget showing city-local time
- 📅 Hourly and daily forecasts
- 🌬️ Wind, humidity, pressure, UV index, air quality
- 📍 Interactive map-based weather search
- 📈 Historical comparison & trend analysis
- 🔄 Background jobs to fetch daily reports
- 🎨 Animated, weather-themed UI



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
weather-tracking-app/
├── main.go
├── models/
│   └── user.go
├── handlers/
├── static/css
│   └── dashboard.css
├── templates/
│   └── dashboard.html
├── assets/
│   └── ui-screenshot.png
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




