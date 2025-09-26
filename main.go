package main

import (
	"github.com/Adedunmol/face-widget/core"
	"log"
)

func main() {

	err := core.CompareImages("known.jpg", "jesse.jpg")
	if err != nil {
		log.Println(err)
	}
}
