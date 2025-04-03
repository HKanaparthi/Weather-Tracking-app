package models

import "time"

// Activity represents a user action in the application
type Activity struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
	// Additional fields can be added as needed
	City    string `json:"city,omitempty"`
	Details string `json:"details,omitempty"`
}

// ActivityStore defines the interface for activity storage
type ActivityStore interface {
	RecordActivity(activity Activity) error
	GetActivitiesByUserID(userID int, limit int) ([]Activity, error)
	GetRecentActivities(limit int) ([]Activity, error)
}
