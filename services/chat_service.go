package services

import (
	"database/sql"
	"sync"
	"time"

	"weather-app/models"
)

// ChatService handles chat operations
type ChatService struct {
	db         *sql.DB
	chatRooms  map[string]*models.ChatRoom
	roomsMutex sync.RWMutex
}

// NewChatService creates a new instance of ChatService
func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{
		db:        db,
		chatRooms: make(map[string]*models.ChatRoom),
	}
}

// GetOrCreateChatRoom returns a chat room for a city
func (s *ChatService) GetOrCreateChatRoom(cityName string) *models.ChatRoom {
	s.roomsMutex.RLock()
	room, exists := s.chatRooms[cityName]
	s.roomsMutex.RUnlock()

	if !exists {
		room = &models.ChatRoom{
			CityName:    cityName,
			Messages:    make([]models.ChatMessage, 0),
			ActiveUsers: 0,
		}

		// Load recent messages from database
		messages, err := s.getRecentMessages(cityName, 50)
		if err == nil {
			room.Messages = messages
		}

		// Get active users count
		count, err := s.getActiveUsersCount(cityName)
		if err == nil {
			room.ActiveUsers = count
		}

		s.roomsMutex.Lock()
		s.chatRooms[cityName] = room
		s.roomsMutex.Unlock()
	}

	return room
}

// AddMessage adds a new message to a chat room
func (s *ChatService) AddMessage(msg models.ChatMessage) (models.ChatMessage, error) {
	// Save to database
	var id int64
	result, err := s.db.Exec(
		"INSERT INTO chat_messages (user_id, username, city_name, message, image_url, created_at, avatar_url) VALUES (?, ?, ?, ?, ?, NOW(), ?)",
		msg.UserID, msg.Username, msg.CityName, msg.Message, msg.ImageURL, msg.AvatarURL,
	)

	if err != nil {
		return models.ChatMessage{}, err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return models.ChatMessage{}, err
	}

	msg.ID = int(id)
	msg.CreatedAt = time.Now()

	// Add to in-memory room
	s.roomsMutex.Lock()
	room, exists := s.chatRooms[msg.CityName]
	if !exists {
		room = &models.ChatRoom{
			CityName:    msg.CityName,
			Messages:    make([]models.ChatMessage, 0),
			ActiveUsers: 0,
		}
		s.chatRooms[msg.CityName] = room
	}

	// Limit to last 100 messages in memory
	if len(room.Messages) >= 100 {
		room.Messages = room.Messages[1:]
	}

	room.Messages = append(room.Messages, msg)
	s.roomsMutex.Unlock()

	return msg, nil
}

// UpdateUserActivity marks a user as active in a chat room
func (s *ChatService) UpdateUserActivity(userID int, cityName string) error {
	_, err := s.db.Exec(
		`INSERT INTO chat_active_users (user_id, city_name, last_active) 
		VALUES (?, ?, NOW()) 
		ON DUPLICATE KEY UPDATE last_active = NOW()`,
		userID, cityName,
	)

	if err != nil {
		return err
	}

	// Update active users count in memory
	count, err := s.getActiveUsersCount(cityName)
	if err == nil {
		s.roomsMutex.Lock()
		if room, exists := s.chatRooms[cityName]; exists {
			room.ActiveUsers = count
		}
		s.roomsMutex.Unlock()
	}

	return nil
}

// RemoveInactiveUsers removes users who haven't been active for more than 10 minutes
func (s *ChatService) RemoveInactiveUsers() error {
	_, err := s.db.Exec(
		"DELETE FROM chat_active_users WHERE last_active < NOW() - INTERVAL 10 MINUTE",
	)

	if err != nil {
		return err
	}

	// Update all active rooms with new counts
	s.roomsMutex.Lock()
	for cityName, room := range s.chatRooms {
		count, err := s.getActiveUsersCount(cityName)
		if err == nil {
			room.ActiveUsers = count
		}
	}
	s.roomsMutex.Unlock()

	return nil
}

// getRecentMessages retrieves recent messages for a city from the database
func (s *ChatService) getRecentMessages(cityName string, limit int) ([]models.ChatMessage, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, username, message, image_url, created_at, avatar_url 
		FROM chat_messages 
		WHERE city_name = ? 
		ORDER BY created_at DESC 
		LIMIT ?`,
		cityName, limit,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.Message,
			&msg.ImageURL, &msg.CreatedAt, &msg.AvatarURL); err != nil {
			return nil, err
		}
		msg.CityName = cityName
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// getActiveUsersCount gets the number of active users in a city
func (s *ChatService) getActiveUsersCount(cityName string) (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM chat_active_users WHERE city_name = ? AND last_active > NOW() - INTERVAL 10 MINUTE",
		cityName,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetDB returns the database connection
func (s *ChatService) GetDB() *sql.DB {
	return s.db
}
