package service

import "golang.org/x/crypto/bcrypt"

const passlibBcryptCost = 12

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), passlibBcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func verifyPassword(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
