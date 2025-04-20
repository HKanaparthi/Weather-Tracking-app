package models

import (
	"time"
)

// ChatMessage represents a single message in a city chat
type ChatMessage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	CityName  string    `json:"city_name"`
	Message   string    `json:"message"`
	ImageURL  string    `json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	AvatarURL string    `json:"avatar_url"`
}

// ChatRoom represents a city-based chat room
type ChatRoom struct {
	CityName    string        `json:"city_name"`
	Messages    []ChatMessage `json:"messages"`
	ActiveUsers int           `json:"active_users"`
}
