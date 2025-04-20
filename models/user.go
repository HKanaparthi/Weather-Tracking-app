package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID                   int       `json:"id" db:"id"`
	Username             string    `json:"username" db:"username"`
	Email                string    `json:"email" db:"email"`
	Password             string    `json:"-" db:"password"` // Password is excluded from JSON responses
	HomeCity             string    `json:"home_city" db:"home_city"`
	NotificationsEnabled bool      `json:"notifications_enabled" db:"notifications_enabled"`
	AlertThreshold       string    `json:"alert_threshold" db:"alert_threshold"` // e.g., "severe", "moderate", "all"
	ProfilePhoto         string    `json:"profile_photo" db:"profile_photo"`     // New field for profile photo
	AvatarColor          string    `json:"avatar_color" db:"avatar_color"`       // Added for avatar color
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// UserStore interface defines methods for user data storage
type UserStore interface {
	CreateUser(user *User) error
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	UpdateUser(user *User) error
	UpdateUserProfile(user *User) error // New method for profile updates
	GetUsersWithDailyReports() ([]*User, error)
	GetDB() *sql.DB // Add this method to get database connection
	Close() error
}

// MySQLStore implements UserStore with MySQL database
type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore creates a new MySQL user store
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	// Parse and validate DSN
	_, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &MySQLStore{db: db}, nil
}

// Close closes the database connection
func (s *MySQLStore) Close() error {
	return s.db.Close()
}

// CreateUser adds a new user to the store
func (s *MySQLStore) CreateUser(user *User) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Set default values if not provided
	if user.ProfilePhoto == "" {
		user.ProfilePhoto = "default.jpg" // Default profile photo
	}

	if user.AvatarColor == "" {
		user.AvatarColor = "blue" // Default avatar color
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if something fails

	// Insert the user
	result, err := tx.Exec(
		"INSERT INTO users (username, email, password, home_city, notifications_enabled, alert_threshold, profile_photo, avatar_color, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		user.Username, user.Email, string(hashedPassword), user.HomeCity, user.NotificationsEnabled, user.AlertThreshold, user.ProfilePhoto, user.AvatarColor, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		// Check for duplicate key error
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			// MySQL error 1062 is duplicate entry
			errMsg := mysqlErr.Message
			if strings.Contains(errMsg, "username") {
				return errors.New("username already exists")
			} else if strings.Contains(errMsg, "email") {
				return errors.New("email already exists")
			}
			return errors.New("duplicate entry error")
		}
		return err
	}

	// Get the ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)

	// Commit the transaction
	return tx.Commit()
}

// GetUserByUsername retrieves a user by username
func (s *MySQLStore) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, email, password, home_city, notifications_enabled, alert_threshold, profile_photo, avatar_color, created_at, updated_at FROM users WHERE username = ?",
		username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.HomeCity,
		&user.NotificationsEnabled, &user.AlertThreshold, &user.ProfilePhoto, &user.AvatarColor, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *MySQLStore) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, email, password, home_city, notifications_enabled, alert_threshold, profile_photo, avatar_color, created_at, updated_at FROM users WHERE email = ?",
		email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.HomeCity,
		&user.NotificationsEnabled, &user.AlertThreshold, &user.ProfilePhoto, &user.AvatarColor, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *MySQLStore) GetUserByID(id int) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, email, password, home_city, notifications_enabled, alert_threshold, profile_photo, avatar_color, created_at, updated_at FROM users WHERE id = ?",
		id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.HomeCity,
		&user.NotificationsEnabled, &user.AlertThreshold, &user.ProfilePhoto, &user.AvatarColor, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetUsersWithDailyReports retrieves all users with daily reports enabled
func (s *MySQLStore) GetUsersWithDailyReports() ([]*User, error) {
	users := []*User{}
	query := `
        SELECT id, username, email, password, home_city, notifications_enabled, alert_threshold, profile_photo, avatar_color, created_at, updated_at 
        FROM users 
        WHERE notifications_enabled = true 
        AND alert_threshold = 'all' 
        AND home_city != ''
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.HomeCity, &user.NotificationsEnabled, &user.AlertThreshold,
			&user.ProfilePhoto, &user.AvatarColor, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates an existing user
func (s *MySQLStore) UpdateUser(user *User) error {
	// Get existing user to confirm it exists
	_, err := s.GetUserByID(user.ID)
	if err != nil {
		return err
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Only update password if a new one is provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			"UPDATE users SET email = ?, password = ?, home_city = ?, notifications_enabled = ?, alert_threshold = ?, profile_photo = ?, avatar_color = ?, updated_at = ? WHERE id = ?",
			user.Email, string(hashedPassword), user.HomeCity, user.NotificationsEnabled, user.AlertThreshold, user.ProfilePhoto, user.AvatarColor, user.UpdatedAt, user.ID)

		if err != nil {
			return handleUpdateError(err)
		}
	} else {
		// Otherwise just update other fields
		_, err = tx.Exec(
			"UPDATE users SET email = ?, home_city = ?, notifications_enabled = ?, alert_threshold = ?, profile_photo = ?, avatar_color = ?, updated_at = ? WHERE id = ?",
			user.Email, user.HomeCity, user.NotificationsEnabled, user.AlertThreshold, user.ProfilePhoto, user.AvatarColor, user.UpdatedAt, user.ID)

		if err != nil {
			return handleUpdateError(err)
		}
	}

	return tx.Commit()
}

// UpdateUserProfile updates only the profile-related fields of an existing user
func (s *MySQLStore) UpdateUserProfile(user *User) error {
	// Get existing user to confirm it exists
	existingUser, err := s.GetUserByID(user.ID)
	if err != nil {
		return err
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Validate profile photo
	validPhotos := map[string]bool{
		"photo1.jpg":  true,
		"photo2.jpg":  true,
		"photo3.jpg":  true,
		"photo4.jpg":  true,
		"photo5.jpg":  true,
		"default.jpg": true,
	}

	// If profile photo is not provided or invalid, keep existing
	if user.ProfilePhoto == "" || !validPhotos[user.ProfilePhoto] {
		user.ProfilePhoto = existingUser.ProfilePhoto
	}

	// If avatar color is not provided, keep existing
	if user.AvatarColor == "" {
		user.AvatarColor = existingUser.AvatarColor
	}

	// Only update password if a new one is provided
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			"UPDATE users SET email = ?, password = ?, home_city = ?, notifications_enabled = ?, alert_threshold = ?, profile_photo = ?, avatar_color = ?, updated_at = ? WHERE id = ?",
			user.Email, string(hashedPassword), user.HomeCity, user.NotificationsEnabled, user.AlertThreshold, user.ProfilePhoto, user.AvatarColor, user.UpdatedAt, user.ID)

		if err != nil {
			return handleUpdateError(err)
		}
	} else {
		// Otherwise just update other fields
		_, err = tx.Exec(
			"UPDATE users SET email = ?, home_city = ?, notifications_enabled = ?, alert_threshold = ?, profile_photo = ?, avatar_color = ?, updated_at = ? WHERE id = ?",
			user.Email, user.HomeCity, user.NotificationsEnabled, user.AlertThreshold, user.ProfilePhoto, user.AvatarColor, user.UpdatedAt, user.ID)

		if err != nil {
			return handleUpdateError(err)
		}
	}

	return tx.Commit()
}

// handleUpdateError checks for duplicate key errors during updates
func handleUpdateError(err error) error {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok && mysqlErr.Number == 1062 {
		errMsg := mysqlErr.Message
		if strings.Contains(errMsg, "email") {
			return errors.New("email already exists")
		}
		return errors.New("duplicate entry error")
	}
	return err
}
func (s *MySQLStore) GetDB() *sql.DB {
	return s.db
}

// ValidatePassword checks if the provided password matches the user's password
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
