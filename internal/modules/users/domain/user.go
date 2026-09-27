package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserInvalidID = errors.New("user invalid id")
)

type User struct {
	id        uuid.UUID
	createdAt time.Time
}

func NewUser(id uuid.UUID, createdAt time.Time) (*User, error) {
	if id == uuid.Nil() {
		return nil, ErrUserInvalidID
	}

	return &User{id: id, createdAt: createdAt}, nil
}

func RehydrateUser(id uuid.UUID, createdAt time.Time) *User {
	return &User{id: id, createdAt: createdAt}
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) CreatedAt() time.Time { return u.createdAt }
