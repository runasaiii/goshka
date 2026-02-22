package handlers
import (
    "encoding/json"
    "golang/internal/usecase"
    "golang/pkg/modules"
    "net/http"
    "strconv"
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