package users
import (
	"fmt"
	"golang/internal/repository/_postgres"
	"golang/pkg/modules"
	"time"
)


type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	query := "SELECT id, name, email, age FROM users WHERE deleted_at IS NULL"
	err := r.db.DB.Select(&users, query)
	if err != nil {
		return nil, err
	}

	fmt.Println(users)
	return users, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	query := "SELECT id, name, email, age FROM users WHERE id=$1 AND deleted_at IS NULL"
	err := r.db.DB.Get(&user, query, id)
	
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateUser(u modules.User) (int, error) {
	var id int
	query := "INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.DB.QueryRow(query, u.Name, u.Email, u.Age).Scan(&id)
	return id, err
}

func (r *Repository) UpdateUser(u modules.User) error {
	query := "UPDATE users SET name=$1, email=$2, age=$3 WHERE id=$4 AND deleted_at IS NULL"
	res, err := r.db.DB.Exec(query, u.Name, u.Email, u.Age, u.ID)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("update failed: no rows affected (user with ID %d not found)", u.ID)
	}
	return nil
}

func (r *Repository) DeleteUser(id int) error {
	query := "UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	res, err := r.db.DB.Exec(query, id)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("delete failed: user with ID %d not found or already deleted", id)
	}
	return nil
}