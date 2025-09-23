package main

import "github.com/Adedunmol/face-widget/core"

func main() {

	err := core.CompareImages()
	if err != nil {
		panic(err)
	}
}
