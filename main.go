package main

import "github.com/Adedunmol/face-widget/core"

func main() {

	err := core.CompareImages("known.jpg", "jesse.jpg")
	if err != nil {
		panic(err)
	}
}
