package modules

import "time"

// User - апись пользователя. Фильтр/сортировка: id, name, email, gender, birth_date, age, created_at
type User struct {
	ID        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	Age       int        `db:"age" json:"age"`
	Gender    string     `db:"gender" json:"gender"`
	BirthDate time.Time  `db:"birth_date" json:"birth_date"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}

// PaginatedResponse - ответ пагинированного списка пользователей
type PaginatedResponse struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

// CursorPageResponse - ответ курсорной пагинации (next/prev - id для следующей/пред страницы)
type CursorPageResponse struct {
	Data    []User `json:"data"`
	Next    *int   `json:"next,omitempty"`
	Prev    *int   `json:"prev,omitempty"`
	HasMore bool   `json:"hasMore"`
}
