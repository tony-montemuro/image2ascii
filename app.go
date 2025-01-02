package main

import (
	"fmt"
	"log"

	"github.com/tony-montemuro/image2ascii/fileio"
)

func main() {
	image, err := fileio.ReadImageFromFile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(image.Bounds())
}
