package entity
import (
	"time"
	"github.com/google/uuid"
)


type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	Role     string    `json:"role"`
	Verified bool      `json:"verified"`

	EmailVerificationCode      string     `json:"-" gorm:"size:8"`
	EmailVerificationExpiresAt *time.Time `json:"-"`
}
