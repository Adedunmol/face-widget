package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Adedunmol/face-widget/api/db"
	"github.com/Adedunmol/face-widget/api/models"
	"github.com/Adedunmol/face-widget/core"
)

func VerifyUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.VerifyUserPayload
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Email == "" || thisRequest.EncodedImage == "" {
		respondWithError(w, "All fields are required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT
			first_name,
			last_name,
			facial_image
		FROM users
		WHERE email = $1`
	var firstName, lastName, baseImage string
	err = db.DB.QueryRow(query, thisRequest.Email).Scan(&firstName, &lastName, &baseImage)
	if err == sql.ErrNoRows {
		respondWithError(w, "User account doesn't exist", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondWithError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
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
	uniqueFilename := fmt.Sprintf(
		"%d_%s%s%s%s",
		time.Now().Unix(),
		firstName,
		lastName,
		"VerificationImage",
		".jpg",
	)
	filepath := fmt.Sprintf("./images/%s", uniqueFilename)

	// 4. Save the decoded data to a new file
	if err := os.WriteFile(filepath, decodedData, 0644); err != nil {
		respondWithError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	if err := core.CompareImages(baseImage, uniqueFilename); err == core.ErrNoMatch {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		os.Remove(filepath)
		respondWithError(w, "Failed to verify user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User verified successfully!"})
}
