package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID        int        `db:"id"`
	Name      string     `db:"name"`
	Username  string     `db:"username"`
	Password  string     `db:"password"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

func (pg *Postgres) CreateUser(ctx context.Context, u *User) (*User, error) {
	query := `INSERT INTO users (name, username, password) VALUES (@name, @username, @password) RETURNING id`
	args := pgx.NamedArgs{
		"name":     u.Name,
		"username": u.Username,
		"password": u.Password,
	}

	userRow := pg.db.QueryRow(ctx, query, args)

	if err := userRow.Scan(&u.ID); err != nil {
		return nil, err
	}

	return u, nil
}

func (pg *Postgres) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User

	query := `SELECT * FROM users WHERE username = @username`
	args := pgx.NamedArgs{
		"username": username,
	}

	userRow := pg.db.QueryRow(ctx, query, args)

	if err := userRow.Scan(&u.ID, &u.Name, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (pg *Postgres) GetUserById(ctx context.Context, userID int) (*User, error) {
	var u User

	query := `SELECT * FROM users WHERE id = @userID`
	args := pgx.NamedArgs{
		"userID": userID,
	}

	userRow := pg.db.QueryRow(ctx, query, args)

	if err := userRow.Scan(&u.ID, &u.Name, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (pg *Postgres) CheckUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = @username)`
	args := pgx.NamedArgs{
		"username": username,
	}

	var b bool
	if err := pg.db.QueryRow(ctx, query, args).Scan(&b); err != nil {
		return false, err
	}

	return b, nil

}
