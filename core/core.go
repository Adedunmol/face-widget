package core

import (
	"errors"
	"fmt"
	"github.com/Kagami/go-face"
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	ModelDir = "models"
	ImageDir = "images"
)

var (
	ErrFileNotExist  = errors.New("file does not exist")
	ErrNoMatch       = errors.New("faces do not match")
	ErrInvalidFormat = errors.New("invalid image format")
	ErrDecodingImage = errors.New("error decoding image")
)

func CompareImages(knownImage, candidateImage string) error {
	log.Println("comparing images")

	knownImagePath := filepath.Join(".", ImageDir, knownImage)
	candidateImagePath := filepath.Join(".", ImageDir, candidateImage)

	log.Println("known image path: ", knownImagePath)
	log.Println("candidate image path: ", candidateImagePath)

	if _, err := os.Stat(knownImagePath); os.IsNotExist(err) {
		log.Println("known image path not exist")
		return ErrFileNotExist
	}

	if _, err := os.Stat(candidateImagePath); os.IsNotExist(err) {
		log.Println("candidate image path not exist")

		return ErrFileNotExist
	}

	if err := ValidateImage(knownImagePath); err != nil {
		return err
	}

	if err := ValidateImage(candidateImagePath); err != nil {
		return err
	}

	currentTime := time.Now()
	modelsPath := filepath.Join(".", ModelDir)
	rec, err := face.NewRecognizer(modelsPath)

	if err != nil {
		return fmt.Errorf("error creating NewRecognizer: %v", err)
	}
	defer rec.Close()

	face1, err := rec.RecognizeSingleFile(knownImagePath)
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	// Add them to recognizer
	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
	}, []int32{0})

	// test with an unknown face
	testFace, err := rec.RecognizeSingleFile(candidateImagePath)
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	if testFace == nil {
		return fmt.Errorf("test face is nil")
	}

	// ClassifyThreshold: returns -1 if not close enough
	threshold := 0.25
	match := rec.ClassifyThreshold(testFace.Descriptor, float32(threshold))

	log.Println("time to classify: ", time.Since(currentTime).Seconds())
	fmt.Println("euclidean distance: ", face.SquaredEuclideanDistance(face1.Descriptor, testFace.Descriptor))

	if match < 0 {
		fmt.Println("ClassifyThreshold result: Unknown face")
		return ErrNoMatch
	}
	fmt.Println("ClassifyThreshold result: Person index", match)
	fmt.Println("euclidean distance: ", face.SquaredEuclideanDistance(face1.Descriptor, testFace.Descriptor))
	return nil
}

func ValidateImage(imagePath string) error {

	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	_, format, err := image.DecodeConfig(file)
	if err != nil {
		log.Println(err)
		return ErrDecodingImage
	}

	if format == "jpeg" {
		return nil
	}

	return ErrInvalidFormat
}
