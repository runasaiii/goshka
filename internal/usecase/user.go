package usecase
import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
	"practice-7/internal/entity"
	"practice-7/internal/usecase/repo"
	"practice-7/pkg/mail"
	"practice-7/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailNotVerified    = errors.New("email not verified")
	ErrInvalidVerification = errors.New("invalid verification code")
	ErrVerificationExpired = errors.New("verification code expired")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type UserUseCase struct {
	repo *repo.UserRepo
	mail mail.Sender
}

func NewUserUseCase(r *repo.UserRepo, m mail.Sender) *UserUseCase {
	return &UserUseCase{repo: r, mail: m}
}

func randomFourDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	if user.Email == "admin@cmail.com" {
		user.Role = "admin"
		user.Verified = true
	} else {
		user.Role = "user"
		user.Verified = false
	}

	code, err := randomFourDigitCode()
	if err != nil {
		return nil, "", fmt.Errorf("verification code: %w", err)
	}
	exp := time.Now().UTC().Add(15 * time.Minute)
	user.EmailVerificationCode = code
	user.EmailVerificationExpiresAt = &exp

	user, err = u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}

	if !user.Verified {
		if err := u.mail.SendVerificationCode(user.Email, code); err != nil {
			return nil, "", fmt.Errorf("send verification email: %w", err)
		}
	}

	sessionID := uuid.New().String()
	return user, sessionID, nil
}

func (u *UserUseCase) VerifyEmail(email, code string) error {
	user, err := u.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidVerification
		}
		return fmt.Errorf("get user by email: %w", err)
	}
	if user.Verified {
		return nil
	}
	if user.EmailVerificationCode != code {
		return ErrInvalidVerification
	}
	if user.EmailVerificationExpiresAt == nil || time.Now().UTC().After(*user.EmailVerificationExpiresAt) {
		return ErrVerificationExpired
	}
	if err := u.repo.MarkEmailVerified(user.ID); err != nil {
		return fmt.Errorf("mark verified: %w", err)
	}
	return nil
}

func (u *UserUseCase) Login(username, password string) (*entity.User, string, string, error) {
	user, err := u.repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", fmt.Errorf("get user: %w", err)
	}
	if err := utils.CheckPassword(password, user.Password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}
	if !user.Verified {
		return nil, "", "", ErrEmailNotVerified
	}
	access, err := utils.GenerateAccessToken(user.ID.String(), user.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("access token: %w", err)
	}
	refresh, err := utils.GenerateRefreshToken(user.ID.String(), user.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("refresh token: %w", err)
	}
	return user, access, refresh, nil
}

func (u *UserUseCase) RefreshAccessToken(refreshToken string) (string, error) {
	userIDStr, _, err := utils.ParseRefreshTokenString(refreshToken)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidRefreshToken, err)
	}
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", ErrInvalidRefreshToken
	}
	user, err := u.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidRefreshToken
		}
		return "", fmt.Errorf("get user: %w", err)
	}
	if !user.Verified {
		return "", ErrEmailNotVerified
	}
	return utils.GenerateAccessToken(user.ID.String(), user.Role)
}

func (u *UserUseCase) GetUserByID(id uuid.UUID) (*entity.User, error) {
	user, err := u.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (u *UserUseCase) PromoteUserToAdmin(targetID uuid.UUID) error {
	if err := u.repo.UpdateUserRole(targetID, "admin"); err != nil {
		return fmt.Errorf("promote user: %w", err)
	}
	return nil
}
