package token

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type JWTMaker struct {
	secretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey, viper.GetDuration("auth.accessTokenTTL"), viper.GetDuration("auth.refreshTokenTTL")}
}

func (m *JWTMaker) CreateAccessToken(userID int, username string) (string, *UserClaims, error) {

	claims, err := NewUserClaims(userID, username, m.AccessTokenTTL)
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	accessToken, err := token.SignedString([]byte(m.secretKey))

	if err != nil {
		return "", nil, err
	}

	return accessToken, claims, nil
}

// func (m *JWTMaker) CreateRefreshToken(userID int, username string) (string, *UserClaims, error) {

// 	claims, err := NewUserClaims(userID, username, m.accessTokenTTL)
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

// 	accessToken, err := token.SignedString([]byte(m.secretKey))

// 	if err != nil {
// 		return "", nil, err
// 	}

// 	return accessToken, claims, nil
// }

func (m *JWTMaker) CreateRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	_, err := r.Read(b)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *JWTMaker) VerifyToken(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("invalid token signing method")
		}

		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)

	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
