# 🌤️ GoWeather Premium - Weather Tracking Dashboard


**GoWeather Premium** is a modern, responsive weather monitoring web application built using **Go (Gin framework)** and **MySQL**. It integrates with **OpenWeatherMap APIs** to deliver real-time and historical weather data, air quality, UV index, and local time tracking — all wrapped in a sleek, animated dashboard.

---

## 🚀 Features

- 🔐 User authentication (Signup/Login)
- 🌎 Real-time weather data by city or geolocation
- ⏰ **World clock** widget showing city-local time
- 📊 Hourly and daily forecasts
- 🌬️ Wind, humidity, pressure, UV index, air quality
- 🎯 Location-aware weather with map search
- 📅 Historical comparison & trend analysis
- 📥 Background jobs to fetch daily reports
- 🎨 Beautiful, animated UI with weather-based theming

---

## 🛠️ Tech Stack

- **Backend:** Go (Gin Framework), MySQL
- **Frontend:** HTML, CSS (custom + responsive), JS
- **Weather API:** OpenWeatherMap (One Call, Air Pollution, UV)
- **Session & Security:** Cookies, bcrypt
- **Scheduler:** Go routines for periodic jobs

---

## 📦 Project Structure

weather-tracking-app/ ├── main.go ├── models/ │ └── user.go ├── handlers/ ├── static/ │ └── dashboard.css ├── templates/ │ └── dashboard.html ├── assets/ │ └── ui-screenshot.png └── README.md


## 2. 🧱 Install MySQL & Create Database

```bash CREATE DATABASE weather_app;
CREATE USER 'weather_user'@'localhost' IDENTIFIED BY 'password123';
GRANT ALL PRIVILEGES ON weather_app.* TO 'weather_user'@'localhost';
FLUSH PRIVILEGES;
```

Then create the users table:

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

## 3. 🔑 Set Your OpenWeatherMap API Key

In main.go or your .env:

OPENWEATHER_API_KEY=your_actual_key
Or hardcode it inside main.go for now while testing.

## 4. 🚀 Run the App

go run main.go
Visit http://localhost:8080 in your browser.