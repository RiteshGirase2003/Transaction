package utils

import (
	"errors"
	"fmt"
	"go-transaction/config"
	"go-transaction/entity"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

// getKey retrieves the secret key used for JWT signing.
func getKey() []byte {
	key, _ := config.GetKey()
	return []byte(key)
}

// getIssuer retrieves the issuer of the JWT token.
func getIssuer() string {
	_, issuer := config.GetKey()
	return issuer
}

// GenerateAuthToken generates a JWT token for a given user.
// It takes a user object and generates a signed JWT token with claims such as user ID, role, and expiration time.
func GenerateAuthToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"iss":  getIssuer(),
		"uid":  user.UserID,
		"role": user.Role,
		"exp":  time.Now().Add(100 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(getKey())
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return signedToken, nil
}

// ValidateToken verifies the JWT token and returns the parsed token.
// It checks if the token is valid and signed using the correct signing method.
func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getKey(), nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Token validation failed")
		return nil, err
	}

	return token, nil
}

// GetPayloadFromJWT extracts claims from a JWT token.
// It returns the claims if the token is valid or returns an error if the token is invalid.
func GetPayloadFromJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getKey(), nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse JWT token")
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ConvertStringToUnixTimestamp converts a date string to a Unix timestamp.
// The date string must be in the format "2006-01-02 15:04:05.999999 -0700 UTC".
func ConvertStringToUnixTimestamp(dateString string) (int64, error) {
	// Define the layout for the date format
	layout := "2006-01-02 15:04:05.999999 -0700 UTC"
	// Parse the date string
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		return 0, err
	}

	// Return Unix timestamp (seconds since Jan 1, 1970)
	return parsedTime.Unix(), nil
}

// WrapError wraps an error with a custom message, providing context for the action.
func WrapError(action string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}
