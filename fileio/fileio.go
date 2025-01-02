package fileio

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func ReadImageFromFile() (image.Image, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("no file path specified as argument")
	}

	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	image, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	fmt.Println(format)

	return image, nil
}
