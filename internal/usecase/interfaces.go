package usecase
import (
	"practice-7/internal/entity"

	"github.com/google/uuid"
)


type (
	UserInterface interface {
		RegisterUser(user *entity.User) (*entity.User, string, error)
		VerifyEmail(email, code string) error
		Login(username, password string) (*entity.User, string, string, error)
		RefreshAccessToken(refreshToken string) (accessToken string, err error)
		GetUserByID(id uuid.UUID) (*entity.User, error)
		PromoteUserToAdmin(targetID uuid.UUID) error
	}
)
