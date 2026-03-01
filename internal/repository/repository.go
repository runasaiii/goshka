package repository

import (
	"practice2/internal/repository/_postgres"
	"practice2/internal/repository/_postgres/users"
	"practice2/pkg/modules"
)


type UserRepository interface {
    GetUsers() ([]modules.User, error)
    GetUserByID(id int) (*modules.User, error)
    CreateUser(user modules.User) (int, error)
    UpdateUser(user modules.User) error
    DeleteUser(id int) error
}

type Repositories struct {
    UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
    return &Repositories{
        UserRepository: users.NewUserRepository(db),
    }
}