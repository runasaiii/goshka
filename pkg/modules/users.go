package modules
import "time"


type User struct {
    ID        int        `db:"id" json:"id"`
    Name      string     `db:"name" json:"name"`
    Email     string     `db:"email" json:"email"`
    Age       int        `db:"age" json:"age"`
    DeletedAt *time.Time `db:"deleted_at" json:"-"` 
}