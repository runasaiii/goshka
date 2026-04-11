package repo
import (
	"practice-7/internal/entity"
	"practice-7/pkg/postgres"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


type UserRepo struct {
	PG *postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{PG: pg}
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	err := u.PG.Conn.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := u.PG.Conn.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := u.PG.Conn.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) MarkEmailVerified(id uuid.UUID) error {
	return u.PG.Conn.Model(&entity.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"verified":                      true,
		"email_verification_code":       "",
		"email_verification_expires_at": gorm.Expr("NULL"),
	}).Error
}

func (u *UserRepo) GetUserByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := u.PG.Conn.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) UpdateUserRole(id uuid.UUID, role string) error {
	res := u.PG.Conn.Model(&entity.User{}).Where("id = ?", id).Update("role", role)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
