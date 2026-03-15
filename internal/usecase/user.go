package usecase

import (
	"errors"
	"fmt"
	"goshka/internal/repository"
	"goshka/internal/repository/safequery"
	"goshka/pkg/modules"
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

func (u *UserUsecase) GetUsersFiltered(status string, filters []safequery.FilterSpec) ([]modules.User, error) {
	return u.repo.GetUsersFiltered(status, filters)
}

func (u *UserUsecase) GetPaginatedUsers(page, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.PaginatedResponse, error) {
	return u.repo.GetPaginatedUsers(page, pageSize, status, filters, orderByCol, orderDir)
}

func (u *UserUsecase) GetPaginatedUsersCursor(cursor, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.CursorPageResponse, error) {
	return u.repo.GetPaginatedUsersCursor(cursor, pageSize, status, filters, orderByCol, orderDir)
}

func (u *UserUsecase) GetCommonFriends(userID1, userID2 int) ([]modules.User, error) {
	if userID1 <= 0 || userID2 <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return u.repo.GetCommonFriends(userID1, userID2)
}