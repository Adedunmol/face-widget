package core

import (
	"errors"
	"fmt"
	"github.com/Kagami/go-face"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	ModelDir = "models"
	ImageDir = "images"
)

var ErrFileNotExist = errors.New("file does not exist")
var ErrNoMatch = errors.New("faces do not match")

func CompareImages(knownImage, candidateImage string) error {
	log.Println("comparing images")

	knownImagePath := filepath.Join(ImageDir, knownImage)
	candidateImagePath := filepath.Join(ImageDir, candidateImage)

	if _, err := os.Stat(knownImagePath); os.IsNotExist(err) {
		return ErrFileNotExist
	}

	if _, err := os.Stat(candidateImagePath); os.IsNotExist(err) {
		return ErrFileNotExist
	}

	currentTime := time.Now()
	modelsPath := filepath.Join(".", ModelDir)
	rec, err := face.NewRecognizer(modelsPath)

	if err != nil {
		return fmt.Errorf("error creating NewRecognizer: %v", err)
	}
	defer rec.Close()

	face1, err := rec.RecognizeSingleFile(knownImage) // "./images/known.jpg"
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	// Add them to recognizer
	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
	}, []int32{0})

	// test with an unknown face
	testFace, err := rec.RecognizeSingleFile(candidateImage)
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	// ClassifyThreshold: returns -1 if not close enough
	threshold := 0.25
	match := rec.ClassifyThreshold(testFace.Descriptor, float32(threshold))
	if match < 0 {
		fmt.Println("ClassifyThreshold result: Unknown face")
		return ErrNoMatch
	}
	log.Println("time to classify: ", time.Since(currentTime))
	fmt.Println("ClassifyThreshold result: Person index", match)
	return nil
}
