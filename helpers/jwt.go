package helpers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var tokenDuration = 24 * time.Hour

var errSigningMethod = fmt.Errorf("there's a problem with the signing method")

type JwtClaim struct {
	UserID uint `json:"id"`
	jwt.RegisteredClaims
}

// TODO: Don't hardcode this, store it in the config file.
var jwtSecret = []byte("secret")

func GenerateJwt(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtClaim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("could not generate token string: %w", err)
	}

	return tokenString, nil
}

func ValidateJwt(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JwtClaim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errSigningMethod
		}

		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	return parsedToken, nil
}
