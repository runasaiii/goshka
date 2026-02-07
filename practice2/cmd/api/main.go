package main
import (
	"log"
	"net/http"
	"practice2/internal/handlers"
	"practice2/internal/middleware"
)


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", handlers.TaskHandler)
	wrappedMux := middleware.Middleware(mux)

	addr := ":8080"
	log.Println("======================================")
	log.Printf("Api server for tasks is starting...")
	log.Printf("Address is here: http://localhost%s", addr)
	log.Println("======================================")

	err := http.ListenAndServe(addr, wrappedMux)
	if err != nil {
		log.Fatal("Server couldnt start: ", err)
	}
}