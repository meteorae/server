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
	docs, err := db.Query(UserCollection.String()).Where(clover.Field("username").Eq(&username)).FindAll()
	if err != nil {
		return nil, err
	}

	if len(docs) < 0 {
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

	userId, err := db.InsertOne(UserCollection.String(), document)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	newAccount.Id = userId

	return &newAccount, nil
}

// Returns the corresponding user from the specified user ID.
func GetUserById(id string) (*User, error) {
	var user User

	userDocument, err := db.Query(UserCollection.String()).Where(clover.Field("_id").Eq(id)).FindFirst()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userDocument.Unmarshal(&user)

	return &user, nil
}

// Returns the corresponding user from the specified username.
func GetUserByName(username string) (*User, error) {
	var user User

	userDocument, err := db.Query(UserCollection.String()).Where(clover.Field("username").Eq(username)).FindFirst()
	if err != nil {
		return nil, errInvalidCredentials
	}

	userDocument.Unmarshal(&user)

	return &user, nil
}

// Returns the requested fields for all users.
func GetUsers() ([]*User, error) {
	var users []*User //nolint:prealloc

	var user *User

	docs, err := db.Query(UserCollection.String()).FindAll()
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

	count, err := db.Query(UserCollection.String()).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}
