package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"strings"

	"github.com/tony-montemuro/image2ascii/fileio"
)

func coordinateToPixelNumber(x int, y int) int {
	if y <= 2 {
		return 3*x + y
	}
	return 2*y + x
}

func generateAscii(img image.Image) []string {
	const CHAR_WIDTH = 2
	const CHAR_HEIGHT = 4
	bounds := img.Bounds()

	pixelsToAscii := func(startX int, startY int) rune {
		getUpdatedOffset := func(pixel color.Color, offset uint8, pixelNumber int) uint8 {
			r, g, b, a := pixel.RGBA()
			newOffset := offset

			// convert to 8-bit
			red := uint16(r >> 8)
			green := uint16(g >> 8)
			blue := uint16(b >> 8)
			alpha := uint16(a >> 8)
			score := red + green + blue + alpha

			if score > 256*4/2 {
				newOffset |= (1 << pixelNumber)
			}
			return newOffset
		}

		var offset uint8 = 0
		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
				x, y := startX+dx, startY+dy
				if x <= bounds.Max.X && y <= bounds.Max.Y {
					pixel := img.At(x, y)
					pixelNumber := coordinateToPixelNumber(dx, dy)
					offset = getUpdatedOffset(pixel, offset, pixelNumber)
				}
			}
		}

		return rune(0x2800 + int(offset))
	}

	ascii := []string{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y += int(CHAR_HEIGHT) {
		var builder strings.Builder
		for x := bounds.Min.X; x < bounds.Max.X; x += int(CHAR_WIDTH) {
			builder.WriteRune(pixelsToAscii(x, y))
		}
		builder.WriteRune('\n')
		ascii = append(ascii, builder.String())
	}

	return ascii
}

func main() {
	image, err := fileio.ReadImageFromFile()
	if err != nil {
		log.Fatal(err)
	}

	ascii := generateAscii(image)
	file, err := os.Create("output.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, row := range ascii {
		file.WriteString(row)
	}
}
