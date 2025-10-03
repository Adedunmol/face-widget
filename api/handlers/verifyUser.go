package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	var firstName, lastName, baseImageURL string
	err = db.DB.QueryRow(query, thisRequest.Email).Scan(&firstName, &lastName, &baseImageURL)
	if err == sql.ErrNoRows {
		respondWithError(w, "User account doesn't exist", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondWithError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the data from the URL
	resp, err := http.Get(baseImageURL)
	if err != nil {
		log.Fatalf("Failed to download file from URL: %w", err)
		respondWithError(w, "Error downloading baseImage from Cloudinary", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
		respondWithError(w, "Error downloading baseImage from Cloudinary", http.StatusInternalServerError)
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

	baseImageFilename := fmt.Sprintf(
		"%d_%s%s%s%s",
		time.Now().Unix(),
		firstName,
		lastName,
		"BaseImage",
		".jpg",
	)
	// 3. Create a unique filename for the new file
	verificationImageFilename := fmt.Sprintf(
		"%d_%s%s%s%s",
		time.Now().Unix(),
		firstName,
		lastName,
		"VerificationImage",
		".jpg",
	)

	baseFilepath := fmt.Sprintf("./images/%s", baseImageFilename)
	verificationFilepath := fmt.Sprintf("./images/%s", verificationImageFilename)

	baseFile, err := os.Create(baseFilepath)
	if err != nil {
		log.Fatalf("Failed to create temp file: %w", err)
		respondWithError(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer baseFile.Close()

	// 4. Save the decoded data to a new file
	if _, err := io.Copy(baseFile, resp.Body); err != nil {
		os.Remove(baseFilepath)
		respondWithError(w, "Failed to save baseImage file"+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(verificationFilepath, decodedData, 0644); err != nil {
		os.Remove(baseFilepath)
		respondWithError(w, "Failed to save verificationImage file"+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := core.CompareImages(baseImageFilename, verificationImageFilename); err == core.ErrNoMatch {
		os.Remove(baseFilepath)
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		os.Remove(baseFilepath)
		os.Remove(verificationFilepath)
		respondWithError(w, "Failed to verify user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User verified successfully!"})
}
