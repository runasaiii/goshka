package postgres
import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


type Postgres struct {
	Conn *gorm.DB
}

func New(dsn string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return &Postgres{Conn: db}, nil
}
