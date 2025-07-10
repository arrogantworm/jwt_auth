package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/arrogantworm/jwt_auth/api/token"
	"github.com/arrogantworm/jwt_auth/api/utils"
	"github.com/arrogantworm/jwt_auth/db"
)

type Handler struct {
	ctx        context.Context
	db         *db.Postgres
	TokenMaker *token.JWTMaker
}

func NewHandler(db *db.Postgres, secretKey string) (*Handler, error) {

	if secretKey == "" {
		return nil, errors.New("secret key not found")
	}

	return &Handler{
		ctx:        context.Background(),
		db:         db,
		TokenMaker: token.NewJWTMaker(secretKey),
	}, nil
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	if message == "" {
		http.Error(w, "", status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorRes{message})
}

func (h *Handler) sendSuccess(w http.ResponseWriter, message string, status int) {

	if message == "" {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(SuccessRes{message})
}

// api/test/
func (h *Handler) testHandler(w http.ResponseWriter, r *http.Request) {
	// userAgent := r.Header.Get("User-Agent")
	// if userAgent == "" {
	// 	userAgent = "unknown"
	// }

	// ip, _, err := net.SplitHostPort(r.RemoteAddr)
	// if err != nil {
	// 	ip = r.RemoteAddr
	// }

	// w.Write([]byte(ip))

	h.sendError(w, "error updating session", http.StatusInternalServerError)
}

// new-ip/
func (h *Handler) newIpReciever(w http.ResponseWriter, r *http.Request) {
	var req utils.NewIpNotification

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received IP Change for user %d\nOldIP: %s\nNewIP: %s", req.UserID, req.OldIp, req.NewIp)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// api/user
// @Summary UserInfo
// @Tags user
// @Description Get user info
// @ID user-info
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserRes
// @Failure 401,500 {object} ErrorRes
// @Router /api/user [get]
func (h *Handler) getUserInfo(w http.ResponseWriter, r *http.Request) {

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	u, err := h.db.GetUserById(h.ctx, claims.ID)

	if err != nil {
		h.sendError(w, "error getting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toUserRes(u))

}

// auth/signup
// @Summary SignUp
// @Tags auth
// @Description Registration
// @ID auth-signup
// @Accept json
// @Param input body UserReq true "User info"
// @Produce json
// @Success 201 {object} SuccessRes
// @Failure 400,500 {object} ErrorRes
// @Router /auth/signup [post]
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var u UserReq

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.sendError(w, "", http.StatusBadRequest)
		return
	}

	if u.Name == "" || u.Username == "" || u.Password == "" {
		h.sendError(w, "all fields are required", http.StatusBadRequest)
		return
	}

	hashed, err := utils.HashPassword(u.Password)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Password = hashed

	check, err := h.db.CheckUsername(h.ctx, u.Username)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if check {
		h.sendError(w, "username is already taken", http.StatusBadRequest)
		return
	}

	_, err = h.db.CreateUser(h.ctx, toDBUser(u))
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// res := toUserRes(created)
	h.sendSuccess(w, "user registered", http.StatusCreated)
}

func toDBUser(u UserReq) *db.User {
	return &db.User{
		Name:     u.Name,
		Username: u.Username,
		Password: u.Password,
	}
}

func toUserRes(u *db.User) UserRes {
	return UserRes{
		Id:       u.ID,
		Name:     u.Name,
		Username: u.Username,
	}
}

// auth/signin
// @Summary SignIn
// @Tags auth
// @Description Login
// @ID auth-signin
// @Accept json
// @Param input body LoginUserReq true "Credentials"
// @Produce json
// @Success 200 {object} LoginUserRes
// @Failure 400,401,500 {object} ErrorRes
// @Router /auth/signin [post]
func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var u LoginUserReq

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.sendError(w, "", http.StatusBadRequest)
		return
	}

	if u.Username == "" || u.Password == "" {
		h.sendError(w, "all fields are required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	gu, err := h.db.GetUserByUsername(h.ctx, u.Username)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check is password is valid

	err = utils.CheckPassword(u.Password, gu.Password)
	if err != nil {
		h.sendError(w, "wrong password", http.StatusUnauthorized)
		return
	}

	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "unknown"
	}

	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		userIP = r.RemoteAddr
	}

	accessToken, accessClaims, err := h.TokenMaker.CreateAccessToken(gu.ID, gu.Username)

	if err != nil {
		h.sendError(w, "error creating token", http.StatusInternalServerError)
		return
	}

	// refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(gu.ID, gu.Username)
	refreshToken, err := h.TokenMaker.CreateRefreshToken()

	if err != nil {
		h.sendError(w, "error creating token", http.StatusInternalServerError)
		return
	}

	hashedRefreshToken, err := utils.HashRefreshToken(refreshToken)
	if err != nil {
		h.sendError(w, fmt.Sprintf("error encrypting refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	refreshTTL := time.Now().Add(h.TokenMaker.RefreshTokenTTL)

	// Create Session

	s, err := h.db.CreateOrUpdateSession(h.ctx, &db.Session{
		SessionID:    accessClaims.RegisteredClaims.ID,
		UserID:       gu.ID,
		RefreshToken: hashedRefreshToken,
		UserAgent:    userAgent,
		IPAddress:    userIP,
		ExpiresAt:    refreshTTL,
	})
	if err != nil {
		h.sendError(w, "error saving session", http.StatusInternalServerError)
		return
	}

	res := LoginUserRes{
		SessionID:       s.SessionID,
		AccessToken:     accessToken,
		AccessTokenTTL:  accessClaims.ExpiresAt.Time,
		RefreshToken:    refreshToken,
		RefreshTokenTTL: refreshTTL,
		User:            toUserRes(gu),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// auth/logout
// @Summary LogOut
// @Tags auth
// @Description Logout
// @ID auth-logout
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessRes
// @Failure 401,500 {object} ErrorRes
// @Router /auth/logout [post]
func (h *Handler) logoutUser(w http.ResponseWriter, r *http.Request) {

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	if err := h.db.DeleteSession(h.ctx, claims.RegisteredClaims.ID); err != nil {
		h.sendError(w, "error revoking a session", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, "logged out", http.StatusOK)
}

// auth/tokens/renew
// @Summary Renew JWT Token
// @Tags auth, tokens
// @Description Renew JWT Token
// @ID auth-renew
// @Security BearerAuth
// @Accept json
// @Param input body RenewAccessTokenReq true "JWT Token"
// @Produce json
// @Success 200 {object} RenewAccessTokenRes
// @Failure 400,401,500 {object} ErrorRes
// @Router /auth/tokens/renew [post]
func (h *Handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	var req RenewAccessTokenReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "", http.StatusBadRequest)
		return
	}

	// refreshClaims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	// if err != nil {
	// 	http.Error(w, "error verifying token", http.StatusUnauthorized)
	// 	return
	// }

	s, err := h.db.GetSession(h.ctx, claims.RegisteredClaims.ID)
	if err != nil {
		h.sendError(w, "error getting session", http.StatusInternalServerError)
		return
	}

	// Check refresh token
	if err := utils.CheckRefreshToken(req.RefreshToken, s.RefreshToken); err != nil {
		h.sendError(w, "refresh token does not match", http.StatusUnauthorized)
		return
	}

	// UserAgent check
	if r.UserAgent() != s.UserAgent {
		h.sendError(w, "user agent does not match", http.StatusUnauthorized)
		h.db.DeleteSession(h.ctx, claims.RegisteredClaims.ID)
		return
	}

	// IP check
	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		userIP = r.RemoteAddr
	}

	if userIP != s.IPAddress {
		log.Println("[IP NOT MATCH]", s.UserID, s.IPAddress, " -- ", userIP)
		go utils.ChangedIPRequest(utils.NewIpNotification{UserID: s.UserID, NewIp: userIP, OldIp: s.IPAddress})
	}

	if s.IsRevoked {
		h.sendError(w, "session revoked", http.StatusUnauthorized)
		return
	}

	if s.UserID != claims.ID {
		h.sendError(w, "invalid session", http.StatusUnauthorized)
		return
	}

	// u, err := h.db.GetUserByUsername(h.ctx, refreshClaims.Subject)
	// if err != nil {
	// 	http.Error(w, "error getting user", http.StatusInternalServerError)
	// 	return
	// }

	u := db.User{
		ID:       claims.ID,
		Username: claims.Subject,
	}

	accessToken, accessClaims, err := h.TokenMaker.CreateAccessToken(u.ID, u.Username)
	if err != nil {
		h.sendError(w, "error creating accessToken", http.StatusInternalServerError)
		return
	}

	// refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(u.ID, u.Username)

	// if err != nil {
	// 	http.Error(w, "error creating token", http.StatusInternalServerError)
	// 	return
	// }

	refreshToken, err := h.TokenMaker.CreateRefreshToken()

	if err != nil {
		h.sendError(w, "error creating token", http.StatusInternalServerError)
		return
	}

	refreshedTTL := time.Now().Add(h.TokenMaker.RefreshTokenTTL)

	hashedRefreshToken, err := utils.HashRefreshToken(refreshToken)
	if err != nil {
		h.sendError(w, "error encrypting refresh token", http.StatusInternalServerError)
		return
	}

	if err := h.db.RenewAccessToken(h.ctx, accessClaims.RegisteredClaims.ID, claims.RegisteredClaims.ID, hashedRefreshToken, userIP); err != nil {
		h.sendError(w, "error updating session", http.StatusInternalServerError)
		return
	}

	res := RenewAccessTokenRes{
		AccessToken:     accessToken,
		AccessTokenTTL:  accessClaims.ExpiresAt.Time,
		RefreshToken:    refreshToken,
		RefreshTokenTTL: refreshedTTL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// auth/tokens/revoke
// @Summary Revoke JWT Token
// @Tags auth, tokens
// @Description Revoke JWT Token
// @ID auth-revoke
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessRes
// @Failure 401,500 {object} ErrorRes
// @Router /auth/tokens/revoke [post]
func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request) {

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	if err := h.db.RevokeSession(h.ctx, claims.RegisteredClaims.ID); err != nil {
		h.sendError(w, "error revoking session", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, "session revoked", http.StatusOK)
}
