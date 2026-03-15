package safequery

import (
	"fmt"
	"strings"
)

var allowedColumns = map[string]bool{
	"id": true, "name": true, "email": true, "gender": true,
	"birth_date": true, "age": true, "created_at": true, "deleted_at": true,
}

var allowedOps = map[string]bool{
	"=": true, "<": true, ">": true, "ilike": true,
	"<=": true, ">=": true,
}

type FilterSpec struct {
	Column   string
	Operator string
	Value    string
}

func ValidateColumn(col string) string {
	key := strings.ToLower(strings.TrimSpace(col))
	if allowedColumns[key] {
		return key
	}
	return ""
}

func ValidateOperator(op string) string {
	key := strings.ToLower(strings.TrimSpace(op))
	if allowedOps[key] {
		return key
	}
	return ""
}

func ValidateOrderColumn(col string) string {
	return ValidateColumn(col)
}

func BuildWhereClause(specs []FilterSpec) (where string, args []interface{}) {
	var parts []string
	args = make([]interface{}, 0, len(specs))
	idx := 0
	for _, s := range specs {
		col := ValidateColumn(s.Column)
		op := ValidateOperator(s.Operator)
		if col == "" || op == "" {
			continue
		}
		idx++
		placeholder := fmt.Sprintf("$%d", idx)
		if op == "ilike" {
			parts = append(parts, fmt.Sprintf("%s ILIKE %s", col, placeholder))
			args = append(args, "%"+s.Value+"%")
		} else {
			parts = append(parts, fmt.Sprintf("%s %s %s", col, op, placeholder))
			args = append(args, s.Value)
		}
	}
	if len(parts) == 0 {
		return "", args
	}
	return " AND " + strings.Join(parts, " AND "), args
}

func BuildOrderBy(orderByCol, orderDir string) string {
	col := ValidateOrderColumn(orderByCol)
	if col == "" {
		col = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(orderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	return col + " " + dir
}
