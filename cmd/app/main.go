package main
import (
	"log"
	"path/filepath"
	"practice-7/internal/app"
	"github.com/joho/godotenv"
)


func main() {
	loadDotEnv()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func loadDotEnv() {
	for _, p := range []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	} {
		if err := godotenv.Load(p); err == nil {
			return
		}
	}
}