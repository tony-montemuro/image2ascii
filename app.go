package main

import (
	"image"
	"log"
	"math"
	"os"

	"github.com/tony-montemuro/image2ascii/fileio"
)

func process(img image.Image, file *os.File) error {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		line := ""
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.At(x, y)
			r, g, b, a := pixel.RGBA()

			// convert to 8-bit
			red := uint16(r >> 8)
			green := uint16(g >> 8)
			blue := uint16(b >> 8)
			alpha := uint16(a >> 8)

			var level uint16 = math.MaxUint16
			levels := []string{" ", "░", "▒", "▓", "█"}
			var levelWidth uint16 = 205
			score := red + green + blue + alpha

			var i uint16 = 0
			for ; i < uint16(len(levels)) && level == math.MaxUint16; i++ {
				begin := i * levelWidth
				end := (i + 1) * levelWidth
				if score >= begin && score < end {
					level = i
				}
			}

			line += levels[level]
		}
		line += "\n"
		_, err := file.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	image, err := fileio.ReadImageFromFile()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("output.txt")
	if err != nil {
		log.Fatal(err)
	}

	err = process(image, file)
	if err != nil {
		log.Fatal(err)
	}
}
