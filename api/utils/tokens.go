package utils

import "golang.org/x/crypto/bcrypt"

func HashRefreshToken(refreshToken string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckRefreshToken(refreshToken string, hashedRefreshToken string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedRefreshToken), []byte(refreshToken))
}
