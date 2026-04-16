package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goshka/internal/handlers"
	"goshka/internal/middleware"
	"goshka/internal/repository/_postgres"
	"goshka/internal/repository/_postgres/users"
	"goshka/internal/usecase"
	"goshka/pkg/modules"

	"github.com/joho/godotenv"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found, using system env")
	}
	cfg := &modules.PostgreConfig{
		Host:        os.Getenv("DB_HOST"),
		Port:        os.Getenv("DB_PORT"),
		Username:    os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		SSLMode:     envOrDefault("DB_SSLMODE", "disable"),
		ExecTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbDialect := _postgres.NewPGXDialect(ctx, cfg)
	log.Println("Connected to database successfully")

	userRepo := users.NewUserRepository(dbDialect)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handlers.NewUserHandler(userUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("/users/common-friends", userHandler.GetCommonFriends)
	mux.HandleFunc("/users/paginated", userHandler.GetPaginatedUsers)
	mux.HandleFunc("/users/cursor", userHandler.GetPaginatedUsersCursor)
	mux.HandleFunc("/users", userHandler.GetAllUsers)
	mux.HandleFunc("/user", userHandler.GetUserByID)
	mux.HandleFunc("/user/create", userHandler.CreateUser)
	mux.HandleFunc("/user/update", userHandler.UpdateUser)
	mux.HandleFunc("/user/delete", userHandler.DeleteUser)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	finalHandler := middleware.Logger(middleware.Auth(mux))

	server := &http.Server{
		Addr:    ":8080",
		Handler: finalHandler,
	}

	go func() {
		log.Printf("Server is starting on %s...", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	if dbDialect.DB != nil {
		dbDialect.DB.Close()
		log.Println("Database connection closed.")
	}

	log.Println("Server exited properly")
}
