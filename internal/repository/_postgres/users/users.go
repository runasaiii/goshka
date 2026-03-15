package users

import (
	"fmt"
	"goshka/internal/repository/_postgres"
	"goshka/internal/repository/safequery"
	"goshka/pkg/modules"
	"time"
)

const userColumns = "id, name, email, age, gender, birth_date, created_at, deleted_at"

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
	return r.GetUsersFiltered("active", nil)
}

func (r *Repository) GetUsersFiltered(status string, filters []safequery.FilterSpec) ([]modules.User, error) {
	baseWhere := " WHERE 1=1"
	if status == "deleted" {
		baseWhere += " AND deleted_at IS NOT NULL"
	} else {
		baseWhere += " AND deleted_at IS NULL"
	}
	where, args := safequery.BuildWhereClause(filters)
	query := "SELECT " + userColumns + " FROM users" + baseWhere + where + " ORDER BY id"
	return r.selectUsers(r.db.DB, query, args...)
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	query := "SELECT " + userColumns + " FROM users WHERE id=$1 AND deleted_at IS NULL"
	err := r.db.DB.Get(&user, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateUser(user modules.User) (int, error) {
	var id int
	query := `INSERT INTO users (name, email, age, gender, birth_date) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.DB.QueryRow(query, user.Name, user.Email, user.Age, user.Gender, user.BirthDate).Scan(&id)
	return id, err
}

func (r *Repository) UpdateUser(user modules.User) error {
	query := `UPDATE users SET name=$1, email=$2, age=$3, gender=$4, birth_date=$5 WHERE id=$6 AND deleted_at IS NULL`
	res, err := r.db.DB.Exec(query, user.Name, user.Email, user.Age, user.Gender, user.BirthDate, user.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("update failed: no rows affected (user with ID %d not found)", user.ID)
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

// GetPaginatedUsers — пагинация с фильтрами, сортировкой и status (active|deleted).
func (r *Repository) GetPaginatedUsers(page, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	baseWhere := " WHERE 1=1"
	if status == "deleted" {
		baseWhere += " AND deleted_at IS NOT NULL"
	} else {
		baseWhere += " AND deleted_at IS NULL"
	}
	where, args := safequery.BuildWhereClause(filters)
	orderClause := safequery.BuildOrderBy(orderByCol, orderDir)

	countQuery := "SELECT COUNT(*) FROM users" + baseWhere + where
	var totalCount int
	err := r.db.DB.Get(&totalCount, countQuery, args...)
	if err != nil {
		return modules.PaginatedResponse{}, err
	}

	n := len(args)
	limitPh := fmt.Sprintf("$%d", n+1)
	offsetPh := fmt.Sprintf("$%d", n+2)
	query := "SELECT " + userColumns + " FROM users" + baseWhere + where + " ORDER BY " + orderClause + " LIMIT " + limitPh + " OFFSET " + offsetPh
	args = append(args, pageSize, offset)
	userList, err := r.selectUsers(r.db.DB, query, args...)
	if err != nil {
		return modules.PaginatedResponse{}, err
	}

	return modules.PaginatedResponse{
		Data:       userList,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetPaginatedUsersCursor — курсорная пагинация (cursor=0 — первая страница).
func (r *Repository) GetPaginatedUsersCursor(cursor int, pageSize int, status string, filters []safequery.FilterSpec, orderByCol, orderDir string) (modules.CursorPageResponse, error) {
	if pageSize < 1 {
		pageSize = 10
	}
	baseWhere := " WHERE 1=1"
	if status == "deleted" {
		baseWhere += " AND deleted_at IS NOT NULL"
	} else {
		baseWhere += " AND deleted_at IS NULL"
	}
	where, args := safequery.BuildWhereClause(filters)
	orderClause := safequery.BuildOrderBy(orderByCol, orderDir)
	dir := "ASC"
	if orderDir == "DESC" {
		dir = "DESC"
	}
	cursorCond := ""
	if cursor > 0 {
		if dir == "ASC" {
			cursorCond = " AND id > $" + fmt.Sprintf("%d", len(args)+1)
		} else {
			cursorCond = " AND id < $" + fmt.Sprintf("%d", len(args)+1)
		}
		args = append(args, cursor)
	}
	args = append(args, pageSize+1)
	limitPlaceholder := fmt.Sprintf("$%d", len(args))
	query := "SELECT " + userColumns + " FROM users" + baseWhere + where + cursorCond + " ORDER BY " + orderClause + " LIMIT " + limitPlaceholder
	userList, err := r.selectUsers(r.db.DB, query, args...)
	if err != nil {
		return modules.CursorPageResponse{}, err
	}
	hasMore := len(userList) > pageSize
	if hasMore {
		userList = userList[:pageSize]
	}
	var next, prev *int
	if hasMore && len(userList) > 0 {
		lastID := userList[len(userList)-1].ID
		next = &lastID
	}
	if cursor > 0 && len(userList) > 0 {
		firstID := userList[0].ID
		prev = &firstID
	}
	return modules.CursorPageResponse{
		Data:    userList,
		Next:    next,
		Prev:    prev,
		HasMore: hasMore,
	}, nil
}

func (r *Repository) GetCommonFriends(userID1, userID2 int) ([]modules.User, error) {
	if userID1 == userID2 {
		return nil, nil
	}
	query := `
		SELECT u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at, u.deleted_at
		FROM users u
		INNER JOIN user_friends f1 ON f1.friend_id = u.id AND f1.user_id = $1
		INNER JOIN user_friends f2 ON f2.friend_id = u.id AND f2.user_id = $2
		WHERE u.deleted_at IS NULL
		ORDER BY u.id`
	var list []modules.User
	err := r.db.DB.Select(&list, query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) selectUsers(db interface {
	Select(dest interface{}, query string, args ...interface{}) error
}, query string, args ...interface{}) ([]modules.User, error) {
	var users []modules.User
	err := db.Select(&users, query, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}
