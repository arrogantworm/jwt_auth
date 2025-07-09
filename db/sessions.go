package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type (
	Session struct {
		ID           int        `db:"id"`
		SessionID    string     `db:"session_id"`
		UserID       int        `db:"user_id"`
		RefreshToken string     `db:"refresh_token"`
		IsRevoked    bool       `db:"is_revoked"`
		UserAgent    string     `db:"user_agent"`
		IPAddress    string     `db:"ip_address"`
		CreatedAt    time.Time  `db:"created_at"`
		UpdatedAt    *time.Time `db:"updated_at"`
		ExpiresAt    time.Time  `db:"expires_at"`
	}
)

func (pg *Postgres) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var s Session
	query := `SELECT * FROM sessions WHERE session_id = @sessionID`
	args := pgx.NamedArgs{
		"sessionID": sessionID,
	}

	sessionRow := pg.db.QueryRow(ctx, query, args)
	if err := sessionRow.Scan(&s.ID, &s.SessionID, &s.UserID, &s.RefreshToken, &s.IsRevoked, &s.UserAgent, &s.IPAddress, &s.CreatedAt, &s.UpdatedAt, &s.ExpiresAt); err != nil {
		return nil, err
	}

	return &s, nil
}

func (pg *Postgres) CheckActiveSession(ctx context.Context, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE user_id = @userID AND is_revoked = false)`
	args := pgx.NamedArgs{
		"userID": userID,
	}

	var b bool
	if err := pg.db.QueryRow(ctx, query, args).Scan(&b); err != nil {
		return false, err
	}

	return b, nil
}

func (pg *Postgres) CreateSession(ctx context.Context, s *Session) (*Session, error) {
	query := `INSERT INTO sessions (session_id, user_id, refresh_token, user_agent, ip_address, expires_at) 
		VALUES (@UserID, @HashedRefreshToken, @UserAgent, @IPAddress, @ExpiresAt)
		RETURNING id`
	args := pgx.NamedArgs{
		"UserID":             s.UserID,
		"SessionID":          s.SessionID,
		"HashedRefreshToken": s.RefreshToken,
		"UserAgent":          s.UserAgent,
		"IPAddress":          s.IPAddress,
		"ExpiresAt":          s.ExpiresAt,
	}
	sessionRow := pg.db.QueryRow(ctx, query, args)

	if err := sessionRow.Scan(&s.ID); err != nil {
		return nil, err
	}
	return s, nil
}

// func (pg *Postgres) UpdateSession(ctx context.Context, s *Session) (*Session, error) {
// 	query := `UPDATE sessions SET refresh_token=@RefreshToken, is_revoked=@IsRevoked,
// 		user_agent=@UserAgent, ip_address=@IPAddress, expires_at=@ExpiresAt WHERE user_id = @userID
// 		RETURNING id`
// 	args := pgx.NamedArgs{
// 		"UserID":       s.UserID,
// 		"IsRevoked":    false,
// 		"RefreshToken": s.RefreshToken,
// 		"UserAgent":    s.UserAgent,
// 		"IPAddress":    s.IPAddress,
// 		"ExpiresAt":    s.ExpiresAt,
// 	}
// 	sessionRow := pg.db.QueryRow(ctx, query, args)

// 	if err := sessionRow.Scan(&s.ID); err != nil {
// 		return nil, err
// 	}
// 	return s, nil
// }

func (pg *Postgres) CreateOrUpdateSession(ctx context.Context, s *Session) (*Session, error) {
	query := `INSERT INTO sessions (session_id, user_id, refresh_token, user_agent, ip_address, expires_at) 
		VALUES (@SessionID, @UserID, @HashedRefreshToken, @UserAgent, @IPAddress, @ExpiresAt)
		ON CONFLICT (session_id) DO UPDATE SET refresh_token=EXCLUDED.refresh_token, is_revoked=false, 
		user_agent=EXCLUDED.user_agent, ip_address=EXCLUDED.ip_address, updated_at=now(), expires_at=EXCLUDED.expires_at
		RETURNING id`
	args := pgx.NamedArgs{
		"SessionID":          s.SessionID,
		"UserID":             s.UserID,
		"HashedRefreshToken": s.RefreshToken,
		"UserAgent":          s.UserAgent,
		"IPAddress":          s.IPAddress,
		"ExpiresAt":          s.ExpiresAt,
	}
	sessionRow := pg.db.QueryRow(ctx, query, args)

	if err := sessionRow.Scan(&s.ID); err != nil {
		return nil, err
	}
	return s, nil
}

func (pg *Postgres) RenewAccessToken(ctx context.Context, newSessionID, sessionID, hashedRefreshToken, userIP string) error {
	query := `UPDATE sessions 
		SET session_id=@newSessionID, refresh_token=@hashedRefreshToken, ip_address=@userIP, updated_at=now() 
		WHERE session_id = @sessionID`
	args := pgx.NamedArgs{
		"newSessionID":       newSessionID,
		"hashedRefreshToken": hashedRefreshToken,
		"userIP":             userIP,
		"sessionID":          sessionID,
	}
	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) RevokeSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET is_revoked=true WHERE session_id = @sessionID`
	args := pgx.NamedArgs{
		"sessionID": sessionID,
	}
	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = @sessionID`
	args := pgx.NamedArgs{
		"sessionID": sessionID,
	}
	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}
