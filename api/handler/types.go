package handler

import "time"

type (
	ErrorRes struct {
		Message string `json:"error"`
	}

	SuccessRes struct {
		Message string `json:"success"`
	}
)

type (
	UserReq struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	UserRes struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}

	LoginUserReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	LoginUserRes struct {
		SessionID       string    `json:"sessionID"`
		AccessToken     string    `json:"accessToken"`
		AccessTokenTTL  time.Time `json:"accessTokenTTL"`
		RefreshToken    string    `json:"refreshToken"`
		RefreshTokenTTL time.Time `json:"refreshTokenTTL"`
		User            UserRes   `json:"user"`
	}

	RenewAccessTokenReq struct {
		RefreshToken string `json:"refreshToken"`
	}

	RenewAccessTokenRes struct {
		AccessToken     string    `json:"accessToken"`
		AccessTokenTTL  time.Time `json:"accessTokenTTL"`
		RefreshToken    string    `json:"refreshToken"`
		RefreshTokenTTL time.Time `json:"refreshTokenTTL"`
	}
)
