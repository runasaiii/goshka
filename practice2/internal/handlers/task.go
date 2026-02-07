package handlers
import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)


type Task struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
}

var (
	tasks = make(map[int]Task)
	nextID = 1
	mu sync.Mutex
)

func TaskHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    switch r.Method {
    case http.MethodGet:
        idStr := r.URL.Query().Get("id")
        if idStr == "" {
            mu.Lock()
            var allTasks []Task 
            for _, t := range tasks {
                allTasks = append(allTasks, t)
            }
            mu.Unlock()
            
            if allTasks == nil {
                allTasks = []Task{}
            }
            json.NewEncoder(w).Encode(allTasks)
            return
        }

        id, err := strconv.Atoi(idStr)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "id have to be a number"}) 
            return
        }
        
        mu.Lock()
        task, ok := tasks[id]
        mu.Unlock()
        
        if !ok {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"error": "task is not found"}) 
            return
        }
        json.NewEncoder(w).Encode(task)

    case http.MethodPost:
        var t Task
        if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        
        if t.Title == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "task title cannot be empty"})
            return
        }

        if len(t.Title) > 100 {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "too long title name"})
            return
}

        mu.Lock()
        t.ID = nextID
        nextID++
        t.Done = false 
        tasks[t.ID] = t
        mu.Unlock()
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(t)

    case http.MethodPatch:
        idStr := r.URL.Query().Get("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        
        var body struct {
            Done bool `json:"done"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        mu.Lock()
        task, ok := tasks[id]
        if ok {
            task.Done = body.Done
            tasks[id] = task
            mu.Unlock()
            json.NewEncoder(w).Encode(map[string]bool{"updated": true})
        } else {
            mu.Unlock()
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"error": "nothing to update, id not found"})
        }
    
    case http.MethodDelete:
        idStr := r.URL.Query().Get("id")
        id, _ := strconv.Atoi(idStr)

        mu.Lock()
        if _, ok := tasks[id]; ok {
            delete(tasks, id)
            mu.Unlock()
            w.WriteHeader(http.StatusNoContent)
            return
        }
        mu.Unlock()
        w.WriteHeader(http.StatusNotFound)    

    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(map[string]string{"error": "method isnt supported"})
    }
}