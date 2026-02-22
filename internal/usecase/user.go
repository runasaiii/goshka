package usecase
import (
    "errors"
    "fmt"
    "golang/internal/repository"
    "golang/pkg/modules"
)


type UserUsecase struct {
    repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
    return &UserUsecase{repo: repo}
}

func (u *UserUsecase) CreateUser(user modules.User) (int, error) {
    if user.Name == "" {
        return 0, errors.New("name is required!")
    }
    if user.Age < 0 || user.Age > 150 {
        return 0, fmt.Errorf("invalid age: %d", user.Age)
    }
    return u.repo.CreateUser(user)
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
    if id <= 0 {
        return nil, errors.New("invalid ID! It must be greater than 0")
    }
    return u.repo.GetUserByID(id)
}

func (u *UserUsecase) UpdateUser(user modules.User) error {
    if user.ID <= 0 {
        return errors.New("invalid ID for update!")
    }
    return u.repo.UpdateUser(user)
}

func (u *UserUsecase) DeleteUser(id int) error {
    if id <= 0 {
        return errors.New("invalid ID! Cannot delete user with non-positive ID")
    }
    return u.repo.DeleteUser(id)
}

func (u *UserUsecase) GetAllUsers() ([]modules.User, error) {
    return u.repo.GetUsers()
}