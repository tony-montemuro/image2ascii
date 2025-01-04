package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/tony-montemuro/image2ascii/fileio"
)

func getMinBrightness() float64 {
	if len(os.Args) < 3 {
		return 50.0
	}

	val, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatal(err)
	}
	return val
}

func getLinearizedChannel(colorChannel uint8) float64 {
	v := float64(colorChannel) / 255.0

	if v <= 0.04045 {
		return v / 12.92
	} else {
		return math.Pow((v+0.055)/1.055, 2.4)
	}
}

func getLuminance(r float64, g float64, b float64) float64 {
	return (0.2126 * r) + (0.7152 * g) + (0.0722 * b)
}

func getPercievedBrightness(luminance float64) float64 {
	if luminance <= 0.008856 {
		return luminance * 903.3
	} else {
		return math.Pow(luminance, 1.0/3.0)*116 - 16
	}
}

func getPixelBrightness(pixel color.Color) float64 {
	r, g, b, _ := pixel.RGBA()

	// convert to 8-bit value
	red := uint8(r >> 8)
	green := uint8(g >> 8)
	blue := uint8(b >> 8)

	// liearize red, green, blue
	lr, lg, lb := getLinearizedChannel(red), getLinearizedChannel(green), getLinearizedChannel(blue)

	// calculate luminance
	luminance := getLuminance(lr, lg, lb)

	// return percieved brightness
	return getPercievedBrightness(luminance)
}

func getPixelNumber(x int, y int) int {
	if y <= 2 {
		return 3*x + y
	}
	return 2*y + x
}

func generateAscii(img image.Image) []string {
	const CHAR_WIDTH = 2
	const CHAR_HEIGHT = 4
	bounds := img.Bounds()
	minBrightness := getMinBrightness()

	pixelsToAscii := func(startX int, startY int) rune {
		var offset uint8 = 0
		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
				x, y := startX+dx, startY+dy
				if x <= bounds.Max.X && y <= bounds.Max.Y {
					brightness := getPixelBrightness(img.At(x, y))
					if brightness > minBrightness {
						offset |= (1 << getPixelNumber(dx, dy))
					}
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
