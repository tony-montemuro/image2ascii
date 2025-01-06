package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/tony-montemuro/image2ascii/fileio"
)

const (
	CHAR_WIDTH  = 2
	CHAR_HEIGHT = 4
)

type BrightnessComparisonFunc func(a, target float64) bool

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

func getWidthAndHeight(bounds image.Rectangle) (int, int) {
	if len(os.Args) < 5 {
		return int(math.Ceil(float64(bounds.Max.X) / CHAR_WIDTH)), int(math.Ceil(float64(bounds.Max.Y) / CHAR_HEIGHT))
	}

	width, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	height, err := strconv.ParseInt(os.Args[4], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	return int(width), int(height)
}

func isLightInverted() bool {
	if len(os.Args) < 6 {
		return false
	}

	return os.Args[5] == "1"
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

func getPixelNumber(x int, y int) int {
	if y <= 2 {
		return 3*x + y
	}
	return 2*y + x
}

func getBrightnessComparisonFunc(isInverted bool) BrightnessComparisonFunc {
	if isInverted {
		return func(brightness, minBrightness float64) bool { return brightness <= minBrightness }
	}
	return func(brightness, minBrightness float64) bool { return brightness >= minBrightness }
}

func generateAscii(img image.Image) []string {
	ascii := []string{}
	bounds := img.Bounds()
	minBrightness := getMinBrightness()
	userWidth, userHeight := getWidthAndHeight(bounds)
	fmt.Println(userWidth, userHeight)
	pixelWidth, pixelHeight := CHAR_WIDTH*userWidth, CHAR_HEIGHT*userHeight
	originalWidth, originalHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	scaleX, scaleY := float64(pixelWidth)/float64(originalWidth), float64(pixelHeight)/float64(originalHeight)
	isInverted := isLightInverted()
	isPixelOn := getBrightnessComparisonFunc(isInverted)

	pixelsToAscii := func(baseX, baseY int) rune {
		transformedX, transformedY := baseX*CHAR_WIDTH, baseY*CHAR_HEIGHT
		var offset uint8 = 0

		getPixelBrightness := func(pixel color.Color) float64 {
			r, g, b, _ := pixel.RGBA()

			// convert to 8-bit value
			red := uint8(r >> 8)
			green := uint8(g >> 8)
			blue := uint8(b >> 8)

			// linearize red, green, blue
			lr, lg, lb := getLinearizedChannel(red), getLinearizedChannel(green), getLinearizedChannel(blue)

			// calculate luminance
			luminance := getLuminance(lr, lg, lb)

			// return percieved brightness
			return getPercievedBrightness(luminance)
		}

		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
				// map current pixel to original "pixel"
				x, y := transformedX+dx, transformedY+dy
				originalX, originalY := float64(x)/float64(scaleX), float64(y)/float64(scaleY)
				x1, x2 := int(math.Floor(originalX)), int(math.Ceil(originalX))
				y1, y2 := int(math.Floor(originalY)), int(math.Ceil(originalY))

				// take the average brightness of all 4 original pixels
				b1 := getPixelBrightness(img.At(x1, y1))
				b2 := getPixelBrightness(img.At(x1, y2))
				b3 := getPixelBrightness(img.At(x2, y1))
				b4 := getPixelBrightness(img.At(x2, y2))
				brightness := (b1 + b2 + b3 + b4) / 4

				// if brightness exceeds min brightness, render a pixel (update offset)
				if isPixelOn(brightness, minBrightness) {
					offset |= (1 << getPixelNumber(dx, dy))
				}
			}
		}

		return rune(0x2800 + int(offset))
	}

	for y := 0; y < userHeight; y++ {
		var builder strings.Builder
		for x := 0; x < userWidth; x++ {
			builder.WriteRune(pixelsToAscii(x, y))
		}

		if y < userHeight-1 {
			builder.WriteRune('\n')
		}
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
