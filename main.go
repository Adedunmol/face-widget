package main

import (
	"fmt"
	"github.com/Adedunmol/face-widget/core"
	"log"
	"net/http"

	"github.com/Adedunmol/face-widget/api/db"
	"github.com/Adedunmol/face-widget/api/handlers"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	rec := core.Init()
	defer rec.Close()

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Assuming environment variables are set in the environment.")
	}

	db.RunMigrations()

	db.ConnectDB()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", handlers.RegisterUser)
	mux.HandleFunc("POST /verify", handlers.VerifyUser)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)
	serverPort := ":8080"

	fmt.Printf("Face Recognition API server starting on port %s...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, handler))

}
