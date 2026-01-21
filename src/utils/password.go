package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns []byte for consistency with model
func HashPassword(password string) ([]byte, error) {
	// Use cost 12 for better security (default is 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}

// CheckPassword compares hash and password
func CheckPassword(hashedPassword []byte, password string) error {
    return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}