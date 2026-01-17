package user

import (
	"receipt-bot/internal/domain/shared"
)

// UserID represents a unique user identifier
type UserID = shared.ID

// User represents a bot user (Entity)
type User struct {
	id         UserID
	telegramID int64
	username   string
	createdAt  shared.Timestamp
}

// NewUser creates a new User
func NewUser(telegramID int64, username string) (*User, error) {
	if telegramID <= 0 {
		return nil, shared.ErrInvalidTelegramID
	}

	return &User{
		id:         shared.NewID(),
		telegramID: telegramID,
		username:   username,
		createdAt:  shared.NewTimestamp(),
	}, nil
}

// ReconstructUser reconstructs a user from stored data (for repository)
func ReconstructUser(id UserID, telegramID int64, username string, createdAt shared.Timestamp) *User {
	return &User{
		id:         id,
		telegramID: telegramID,
		username:   username,
		createdAt:  createdAt,
	}
}

// ID returns the user ID
func (u *User) ID() UserID {
	return u.id
}

// TelegramID returns the Telegram user ID
func (u *User) TelegramID() int64 {
	return u.telegramID
}

// Username returns the username
func (u *User) Username() string {
	return u.username
}

// CreatedAt returns the creation timestamp
func (u *User) CreatedAt() shared.Timestamp {
	return u.createdAt
}

// UpdateUsername updates the user's username
func (u *User) UpdateUsername(username string) {
	u.username = username
}
