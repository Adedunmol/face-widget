package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Adedunmol/face-widget/api/db"
	"github.com/Adedunmol/face-widget/api/models"

	"github.com/lib/pq"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.RegisterPayload
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Email == "" ||
		thisRequest.FirstName == "" ||
		thisRequest.LastName == "" ||
		thisRequest.EncodedImage == "" {
		respondWithError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// 1. Decode the Base64 string into bytes.
	decodedData, err := base64.StdEncoding.DecodeString(thisRequest.EncodedImage)
	if err != nil {
		respondWithError(w, "Invalid Base64 string", http.StatusBadRequest)
		return
	}

	// 2. Detect the content type (image format) from the decoded bytes.
	fileType := http.DetectContentType(decodedData)
	if fileType != "image/jpeg" {
		respondWithError(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	// 3. Create a unique filename for the new file
	uniqueFilename := fmt.Sprintf("%d_%s%s%s", time.Now().Unix(), thisRequest.FirstName, thisRequest.LastName, ".jpg")
	filepath := fmt.Sprintf("./images/%s", uniqueFilename)

	// 4. Save the decoded data to a new file
	if err := os.WriteFile(filepath, decodedData, 0644); err != nil {
		respondWithError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO users (
			email,
			first_name,
			last_name,
			facial_image
		) VALUES ($1, $2, $3, $4
		) RETURNING id`
	var userID int
	err = db.DB.QueryRow(
		query,
		thisRequest.Email,
		thisRequest.FirstName,
		thisRequest.LastName,
		filepath,
	).Scan(&userID)
	if err != nil {
		if dbError, ok := err.(*pq.Error); ok && dbError.Code.Name() == "unique_violation" {
			os.Remove(filepath)
			respondWithError(w, "Email already exists", http.StatusConflict)
			return
		}
		os.Remove(filepath)
		respondWithError(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "Registration successful!"})
}
