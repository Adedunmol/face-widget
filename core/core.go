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
	ModelDir  = "models"
	ImageDir  = "images"
	Threshold = 0.15
)

var (
	ErrFileNotExist  = errors.New("file does not exist")
	ErrNoMatch       = errors.New("faces do not match")
	ErrInvalidFormat = errors.New("invalid image format")
	ErrDecodingImage = errors.New("error decoding image")
	ErrNoFaceFound   = errors.New("no face found")
	rec              *face.Recognizer
)

func init() {
	var err error
	modelsPath := filepath.Join(".", ModelDir)
	rec, err = face.NewRecognizer(modelsPath)

	if err != nil {
		log.Fatalf("error creating NewRecognizer: %v", err)
	}
}

func CompareImages(knownImage, candidateImage string) error {
	log.Println("comparing images")

	knownImagePath := filepath.Join(".", ImageDir, knownImage)
	candidateImagePath := filepath.Join(".", ImageDir, candidateImage)

	log.Println("known image path: ", knownImagePath)
	log.Println("candidate image path: ", candidateImagePath)

	if _, err := os.Stat(candidateImagePath); os.IsNotExist(err) {
		log.Println("candidate image path not exist")

		return ErrFileNotExist
	}

	if _, err := os.Stat(knownImagePath); os.IsNotExist(err) {
		log.Println("known image path not exist")
		return ErrFileNotExist
	}

	if err := ValidateImage(knownImagePath); err != nil {
		return err
	}

	if err := ValidateImage(candidateImagePath); err != nil {
		return err
	}

	currentTime := time.Now()

	defer rec.Close()

	face1, err := CheckFace(knownImagePath)
	if err != nil {
		return err
	}

	// Add them to recognizer
	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
	}, []int32{0})

	// test with an unknown face
	testFace, err := CheckFace(candidateImagePath)
	if err != nil {
		return err
	}
	match := rec.ClassifyThreshold(testFace.Descriptor, float32(Threshold))

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

func CheckFace(imagePath string) (*face.Face, error) {
	face1, err := rec.RecognizeSingleFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error recognizing file: %v", err)
	}

	if face1 == nil {
		return nil, ErrNoFaceFound
	}

	return face1, nil
}
