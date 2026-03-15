package handlers

import (
	"encoding/json"
	"net/http"
	"goshka/internal/repository/safequery"
	"goshka/internal/usecase"
	"goshka/pkg/modules"
	"strconv"
	"strings"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(u *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: u}
}

func (h *UserHandler) sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.usecase.GetAllUsers()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid ID format!")
		return
	}

	user, err := h.usecase.GetUserByID(id)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "User not found...")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u modules.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	id, err := h.usecase.CreateUser(u)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u modules.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.usecase.UpdateUser(u); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Update failed: user might not exist...")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"updated"}`))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.usecase.DeleteUser(id); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"deleted"}`))
}

func parseFilters(q map[string][]string) []safequery.FilterSpec {
	var out []safequery.FilterSpec
	for _, v := range q["filter"] {
		parts := strings.SplitN(v, ":", 3)
		if len(parts) != 3 {
			continue
		}
		out = append(out, safequery.FilterSpec{Column: parts[0], Operator: parts[1], Value: parts[2]})
	}
	return out
}

func (h *UserHandler) GetPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	status := q.Get("status")
	if status != "deleted" {
		status = "active"
	}
	orderBy := q.Get("order_by")
	if orderBy == "" {
		orderBy = "id"
	}
	orderDir := q.Get("order_dir")
	if orderDir == "" {
		orderDir = "asc"
	}
	filters := parseFilters(q)

	resp, err := h.usecase.GetPaginatedUsers(page, pageSize, status, filters, orderBy, orderDir)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetPaginatedUsersCursor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	q := r.URL.Query()
	cursor, _ := strconv.Atoi(q.Get("cursor"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	status := q.Get("status")
	if status != "deleted" {
		status = "active"
	}
	orderBy := q.Get("order_by")
	if orderBy == "" {
		orderBy = "id"
	}
	orderDir := q.Get("order_dir")
	if orderDir == "" {
		orderDir = "asc"
	}
	filters := parseFilters(q)

	resp, err := h.usecase.GetPaginatedUsersCursor(cursor, pageSize, status, filters, orderBy, orderDir)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "GET only")
		return
	}
	q := r.URL.Query()
	id1, err1 := strconv.Atoi(q.Get("user_id_1"))
	id2, err2 := strconv.Atoi(q.Get("user_id_2"))
	if err1 != nil || err2 != nil || id1 <= 0 || id2 <= 0 {
		h.sendError(w, http.StatusBadRequest, "user_id_1 and user_id_2 required and must be positive integers")
		return
	}

	list, err := h.usecase.GetCommonFriends(id1, id2)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"common_friends": list})
}
