package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var errInvalidCredentials = errors.New("invalid credentials")

type User struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Creates a new user with the specified username and password.
func CreateUser(username, password string) (*User, error) {
	var user User

	result := db.Where("username = ?", username).First(&user)
	if result.Error == nil {
		// If the user exists, prevent registering one with the same name
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Err(err).Msg("Could not create password hash")
	}

	newAccount := User{
		Username: username,
		Password: passwordHash,
	}

	result = db.Create(&newAccount)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return &user, nil
}

// Returns the requested fields from the specified user ID.
func GetUserByID(id string) (*User, error) {
	var user User

	results := db.First(&user, id)
	if results.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", results.Error)
	}

	return &user, nil
}

// Returns the requested fields from the specified username.
func GetUserByName(username string) (*User, error) {
	var user User

	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to find user")

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errInvalidCredentials
		}

		return nil, result.Error
	}

	return &user, nil
}

// Returns the requested fields for all users.
func GetUsers() []*User {
	var users []*User

	db.Find(&users)

	return users
}

// Returns the total number of users.
func GetUsersCount() int64 {
	var count int64

	db.Model(&User{}).Count(&count)

	return count
}
