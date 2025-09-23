package core

import (
	"fmt"
	"github.com/Kagami/go-face"
	"log"
	"path/filepath"
)

const ModelDir = "models"

func CompareImages() error {
	log.Println("comparing images")

	modelsPath := filepath.Join(".", ModelDir)
	rec, err := face.NewRecognizer(modelsPath)

	if err != nil {
		return fmt.Errorf("error creating NewRecognizer: %v", err)
	}
	defer rec.Close()

	// the saved image
	// change image path
	face1, err := rec.RecognizeSingleFile("./images/known.jpg")
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	face2, err := rec.RecognizeSingleFile("./images/unknown1.jpg")
	if err != nil {
		return fmt.Errorf("error recognizing file: %v", err)
	}

	// Load sample images (replace with actual file paths)
	unknownImg := "./images/jesse.jpg"

	// Add them to recognizer
	rec.SetSamples([]face.Descriptor{
		face1.Descriptor,
		face2.Descriptor,
	}, []int32{0, 1})

	// Now test with an unknown face
	testFace, err := rec.RecognizeSingleFile(unknownImg)
	if err != nil {
		log.Fatalf("No face found in %s", unknownImg)
	}

	// 1) Classify: always returns closest match
	closest := rec.Classify(testFace.Descriptor)
	fmt.Println("Classify result:", closest)

	// 2) ClassifyThreshold: returns -1 if not close enough
	threshold := 0.2
	match := rec.ClassifyThreshold(testFace.Descriptor, float32(threshold))
	if match < 0 {
		fmt.Println("ClassifyThreshold result: Unknown face")
	} else {
		fmt.Println("ClassifyThreshold result: Person index", match)
	}

	sameDist := face.SquaredEuclideanDistance(face1.Descriptor, face1.Descriptor)
	sameDist1 := face.SquaredEuclideanDistance(face2.Descriptor, face2.Descriptor)

	dist := face.SquaredEuclideanDistance(face1.Descriptor, testFace.Descriptor)

	fmt.Println("different distance: ", dist)
	fmt.Println("same distance: ", sameDist)
	fmt.Println("same distance1 : ", sameDist1)

	return nil
}
