package core

import (
	"errors"
	"fmt"
	"github.com/Kagami/go-face"
	"image"
	_ "image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

const (
	ModelDir  = "models"
	ImageDir  = "images"
	Threshold = 0.12
)

var (
	ErrFileNotExist  = errors.New("file does not exist")
	ErrNoMatch       = errors.New("faces do not match")
	ErrInvalidFormat = errors.New("invalid image format")
	ErrDecodingImage = errors.New("error decoding image")
	ErrNoFaceFound   = errors.New("no face found")
	rec              *face.Recognizer
)

func Init() *face.Recognizer {
	log.Println("initializing face recognizer")
	var err error
	modelsPath := filepath.Join(".", ModelDir)
	rec, err = face.NewRecognizer(modelsPath)

	if err != nil {
		log.Fatalf("error creating NewRecognizer: %v", err)
	}
	log.Println("done initializing face recognizer")

	return rec
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

	//defer rec.Close()

	face1, err := CheckFace(knownImagePath)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Add them to recognizer
	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
	}, []int32{0})

	// test with an unknown face
	testFace, err := CheckFace(candidateImagePath)
	if err != nil {
		log.Println(err.Error())
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
		log.Println(err.Error())
		return nil, fmt.Errorf("error recognizing file: %v", err)
	}

	if face1 == nil {
		return nil, ErrNoFaceFound
	}

	return face1, nil
}

type FrameData struct {
	Descriptor face.Descriptor
	Rect       image.Rectangle
}

func IsSamePerson(rec *face.Recognizer, frames []FrameData) bool {
	face1 := frames[0]

	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
	}, []int32{0})

	for i := 1; i < len(frames); i++ {
		if rec.ClassifyThreshold(frames[i].Descriptor, Threshold) < 0 {
			return false
		}
	}
	return true
}

func ComputeRectangleMotion(frames []FrameData) float64 {
	totalShift := 0.0
	for i := 1; i < len(frames); i++ {
		dx := float64(frames[i].Rect.Min.X - frames[i-1].Rect.Min.X)
		dy := float64(frames[i].Rect.Min.Y - frames[i-1].Rect.Min.Y)
		totalShift += math.Sqrt(dx*dx + dy*dy)
	}
	return totalShift / float64(len(frames)-1)
}

func ComputeDescriptorShift(frames []FrameData) float64 {
	total := 0.0
	for i := 1; i < len(frames); i++ {
		total += DescriptorDistance(frames[i-1].Descriptor, frames[i].Descriptor)
	}
	return total / float64(len(frames)-1)
}

func DescriptorDistance(a, b face.Descriptor) float64 {
	sum := 0.0
	for i := range a {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
