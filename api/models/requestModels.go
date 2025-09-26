package models

type RegisterPayload struct {
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EncodedImage string `json:"encoded_facial_image"` // This will hold the Base64 string
}
