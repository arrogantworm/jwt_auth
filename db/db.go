package db

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Postgres struct {
		db *pgxpool.Pool
	}

	Config struct {
		Host    string
		Port    int
		User    string
		Pass    string
		DBName  string
		SSLMode string
	}
)

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, cfg Config) (*Postgres, error) {
	var err error
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	pgOnce.Do(func() {
		var db *pgxpool.Pool
		db, err = pgxpool.New(ctx, connString)
		if err == nil {
			pgInstance = &Postgres{db}
		}
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %s", err.Error())
	}

	if err = pgInstance.Ping(ctx); err != nil {
		return nil, err
	}

	return pgInstance, nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() {
	log.Println("shutting down database")
	pg.db.Close()
}
