package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang/internal/handlers"
	"golang/internal/middleware"
	"golang/internal/repository"
	"golang/internal/repository/_postgres"
	"golang/internal/usecase"
	"golang/pkg/modules"

	"github.com/joho/godotenv"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()

	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)
	fmt.Println("DB Connection Object:", _postgre)

	repos := repository.NewRepositories(_postgre)

	existingUsers, err := repos.GetUsers()
	if err != nil {
		fmt.Printf("Startup Check - Error: %v\n", err)
	} else {
		fmt.Printf("Startup Check - Users in DB: %+v\n", existingUsers)
	}

	userUsecase := usecase.NewUserUsecase(repos)
	userHandler := handlers.NewUserHandler(userUsecase)

	mux := http.NewServeMux()
	mux.HandleFunc("/users", userHandler.GetAllUsers)
	mux.HandleFunc("/user", userHandler.GetUserByID)
	mux.HandleFunc("/user/create", userHandler.CreateUser)
	mux.HandleFunc("/user/update", userHandler.UpdateUser)
	mux.HandleFunc("/user/delete", userHandler.DeleteUser)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	authStack := middleware.Auth(mux)
	finalHandler := middleware.Logger(authStack)

	serverAddr := ":8080"
	log.Printf("Starting server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, finalHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        os.Getenv("DB_HOST"),
		Port:        os.Getenv("DB_PORT"),
		Username:    os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		SSLMode:     os.Getenv("DB_SSLMODE"),
		ExecTimeout: 5 * time.Second,
	}
}