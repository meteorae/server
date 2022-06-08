package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/ostafen/clover"
	"github.com/rs/zerolog/log"
)

var errInvalidCredentials = errors.New("invalid credentials")

type User struct {
	Id        string    `clover:"_id"`
	Username  string    `clover:"username"`
	Password  string    `clover:"password"`
	CreatedAt time.Time `clover:"createdAt"`
	UpdatedAt time.Time `clover:"updatedAt"`
}

// Creates a new user with the specified username and password.
func CreateUser(username, password string) (*User, error) {
	var user User

	result, err := db.Query("todos").Where(clover.Field("username").Eq(username)).FindFirst()
	if err == nil {
		return nil, err
	}

	if result != nil {
		return nil, errors.New("Usernames must be unique")
	}

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Err(err).Msg("Could not create password hash")
	}

	newAccount := User{
		Username:  username,
		Password:  passwordHash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	document := clover.NewDocumentOf(&newAccount)

	_, err = db.InsertOne(string(UserCollection), document)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Returns the corresponding user from the specified user ID.
func GetUserById(id string) (*User, error) {
	var user User

	userDocument, err := db.Query(string(UserCollection)).Where(clover.Field("_id").Eq(id)).FindFirst()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userDocument.Unmarshal(&user)

	return &user, nil
}

// Returns the corresponding user from the specified username.
func GetUserByName(username string) (*User, error) {
	var user User

	userDocument, err := db.Query(string(UserCollection)).Where(clover.Field("username").Eq(username)).FindFirst()
	if err != nil {
		return nil, errInvalidCredentials
	}

	userDocument.Unmarshal(&user)

	return &user, nil
}

// Returns the requested fields for all users.
func GetUsers() ([]*User, error) {
	var users []*User

	var user *User

	docs, err := db.Query(string(UserCollection)).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(user)
		users = append(users, user)
	}

	return users, nil
}

// Returns the total number of users.
func GetUsersCount() (int64, error) {
	var count int

	count, err := db.Query(string(UserCollection)).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}
