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

func NewVerifyUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Unaccepted method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var thisRequest models.NewVerifyUserPayload
	err = json.Unmarshal(body, &thisRequest)
	if err != nil {
		respondWithError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if thisRequest.Email == "" || len(thisRequest.Frames) != 5 {
		log.Println("Email missing or frames < 5")
		respondWithError(w, "Request fields invalid", http.StatusBadRequest)
		return
	}

	query := `
		SELECT
			id,
			first_name,
			last_name,
			facial_image
		FROM users
		WHERE email = $1`
	var thisUser models.User
	var baseImageURL string
	err = db.DB.QueryRow(query, thisRequest.Email).Scan(
		&thisUser.ID,
		&thisUser.FirstName,
		&thisUser.LastName,
		&baseImageURL,
	)
	if err == sql.ErrNoRows {
		respondWithError(w, "User account doesn't exist", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondWithError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	thisUser.Email = thisRequest.Email

	var frames []core.FrameData
	var mainDecoded []byte
	for i, frame := range thisRequest.Frames {
		decodedData, err := base64.StdEncoding.DecodeString(frame)
		if err != nil {
			respondWithError(w, "Invalid base64 string for frame "+string(i+1), http.StatusBadRequest)
			return
		}

		fileType := http.DetectContentType(decodedData)
		if fileType != "image/jpeg" {
			respondWithError(w, "Unsupported image format for frame "+string(i+1), http.StatusBadRequest)
			return
		}

		detected, err := core.Rec.RecognizeSingle(decodedData)
		if detected == nil {
			log.Println("No face found on frame", i)
			respondWithError(w, "Failed to find a face", http.StatusUnprocessableEntity)
			return
		}

		if i == 0 {
			mainDecoded = decodedData
		}

		frames = append(frames, core.FrameData{
			Descriptor: detected.Descriptor,
			Rect:       detected.Rectangle,
		})
	}

	if len(frames) < 5 {
		log.Println("Valid frames < 5")
		respondWithError(w, "Failed to verify face", http.StatusUnprocessableEntity)
		return
	}

	// 1. Check for same identity
	samePerson := core.IsSamePerson(core.Rec, frames)
	log.Println("same person: ", samePerson)
	if !samePerson {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 2. Check for movement
	rectMotion := core.ComputeRectangleMotion(frames)
	descriptorShift := core.ComputeDescriptorShift(frames)
	log.Printf("rectMotion: %v, descriptorShift: %v\n", rectMotion, descriptorShift)

	// Thresholds (tune by experimentation)
	live := rectMotion < 10 && descriptorShift > 0.07
	if !live {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Get the data from the URL
	resp, err := http.Get(baseImageURL)
	if err != nil {
		log.Printf("Failed to download file from URL: %w", err)
		respondWithError(w, "Error downloading baseImage from Cloudinary", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		respondWithError(w, "Error downloading baseImage from Cloudinary", http.StatusInternalServerError)
		return
	}

	baseImageFilename := fmt.Sprintf(
		"%d_%s%s%s%s",
		time.Now().Unix(),
		thisUser.FirstName,
		thisUser.LastName,
		"BaseImage",
		".jpg",
	)
	// 3. Create a unique filename for the new file
	verificationImageFilename := fmt.Sprintf(
		"%d_%s%s%s%s",
		time.Now().Unix(),
		thisUser.FirstName,
		thisUser.LastName,
		"VerificationImage",
		".jpg",
	)

	baseFilepath := fmt.Sprintf("./images/%s", baseImageFilename)
	verificationFilepath := fmt.Sprintf("./images/%s", verificationImageFilename)

	baseFile, err := os.Create(baseFilepath)
	if err != nil {
		log.Printf("Failed to create temp file: %w", err)
		respondWithError(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer baseFile.Close()

	if _, err := io.Copy(baseFile, resp.Body); err != nil {
		os.Remove(baseFilepath)
		respondWithError(w, "Failed to save baseImage file"+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(verificationFilepath, mainDecoded, 0644); err != nil {
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

	respondWithJSON(w, http.StatusOK, thisUser)
}
