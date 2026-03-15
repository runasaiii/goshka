package repository

import (
	"goshka/internal/repository/_postgres"
	"goshka/internal/repository/_postgres/users"
	"goshka/internal/repository/safequery"
	"goshka/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
	GetUsersFiltered(status string, filters []safequery.FilterSpec) ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(user modules.User) (int, error)
	UpdateUser(user modules.User) error
	DeleteUser(id int) error
	GetPaginatedUsers(page, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.PaginatedResponse, error)
	GetPaginatedUsersCursor(cursor, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.CursorPageResponse, error)
	GetCommonFriends(userID1, userID2 int) ([]modules.User, error)
}

type Repositories struct {
    UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
    return &Repositories{
        UserRepository: users.NewUserRepository(db),
    }
}