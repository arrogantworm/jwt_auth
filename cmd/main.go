package main

import (
	"context"
	"log"
	"os"

	"github.com/arrogantworm/jwt_auth/api/handler"
	"github.com/arrogantworm/jwt_auth/api/server"
	_ "github.com/arrogantworm/jwt_auth/cmd/docs"
	"github.com/arrogantworm/jwt_auth/db"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// @title JWT Authorization
// @version 0.1
// @description JWT Authorization project

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {

	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env variables: %s", err.Error())
	}

	postgres, err := db.NewPG(context.Background(), db.Config{
		Host:    viper.GetString("db.host"),
		Port:    viper.GetInt("db.port"),
		User:    viper.GetString("db.user"),
		Pass:    os.Getenv("DB_PASS"),
		DBName:  viper.GetString("db.dbname"),
		SSLMode: viper.GetString("db.sslmode"),
	})
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}
	defer postgres.Close()

	handler, err := handler.NewHandler(postgres, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatalf("failed to initialize handlers: %s", err.Error())
	}
	router := handler.RegisterRoutes()

	srv := new(server.Server)
	if err := srv.Run(viper.GetString("port"), router); err != nil {
		log.Fatalf("error occured while running http server: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
