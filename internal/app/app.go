package app
import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	v1 "practice-7/internal/controller/http/v1"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/internal/usecase/repo"
	"practice-7/pkg/logger"
	"practice-7/pkg/mail"
	"practice-7/pkg/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)


func Run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	}

	pg, err := postgres.New(dsn)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	if err := pg.Conn.AutoMigrate(&entity.User{}); err != nil {
		if !isIgnorablePostgresMigrateErr(err) {
			return fmt.Errorf("migrate: %w", err)
		}
		log.Printf("migrate: skipped non-fatal schema drift: %v", err)
	}

	pg.Conn.Model(&entity.User{}).Where("username = ?", "Aruna").Update("role", "admin")

	userRepo := repo.NewUserRepo(pg)
	userUC := usecase.NewUserUseCase(userRepo, mail.NewFromEnv())
	appLogger := logger.New()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	api := r.Group("/api/v1")
	v1.RegisterRoutes(api, userUC, appLogger)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("listening on %s", addr)
	return r.Run(addr)
}

func isIgnorablePostgresMigrateErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42704" {
		msg := strings.ToLower(pgErr.Message)
		return strings.Contains(msg, "constraint") && strings.Contains(msg, "does not exist")
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "sqlstate 42704") &&
		strings.Contains(s, "constraint") &&
		strings.Contains(s, "does not exist")
}
