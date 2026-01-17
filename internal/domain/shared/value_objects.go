package shared

import (
	"github.com/google/uuid"
	"time"
)

// ID represents a unique identifier
type ID string

// NewID generates a new unique ID
func NewID() ID {
	return ID(uuid.New().String())
}

// String returns the string representation of the ID
func (id ID) String() string {
	return string(id)
}

// IsEmpty checks if the ID is empty
func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// Timestamp represents a point in time
type Timestamp struct {
	value time.Time
}

// NewTimestamp creates a new timestamp with the current time
func NewTimestamp() Timestamp {
	return Timestamp{value: time.Now().UTC()}
}

// NewTimestampFromTime creates a timestamp from a time.Time value
func NewTimestampFromTime(t time.Time) Timestamp {
	return Timestamp{value: t.UTC()}
}

// Time returns the underlying time.Time value
func (t Timestamp) Time() time.Time {
	return t.value
}

// String returns the RFC3339 string representation
func (t Timestamp) String() string {
	return t.value.Format(time.RFC3339)
}
