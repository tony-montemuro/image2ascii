package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CHAR_WIDTH     = 2
	CHAR_HEIGHT    = 4
	MAX_BRIGHTNESS = 100.0
)

type BrightnessComparisonFunc func(a, target float64) bool

func getMinBrightness(c *gin.Context) (float64, error) {
	brightnessVal, ok := c.GetQuery("brightness")

	if ok {
		brightness, err := strconv.ParseFloat(brightnessVal, 64)
		if err != nil {
			return -1.0, fmt.Errorf("invalid brightness: must be a number")
		}
		return MAX_BRIGHTNESS - brightness, nil
	}

	return 50.0, nil
}

func getWidthAndHeight(bounds image.Rectangle, c *gin.Context) (int, int, error) {
	widthVal, okWidth := c.GetQuery("width")
	heightVal, okHeight := c.GetQuery("height")
	errs := []string{}

	width, height := 0, 0
	if !okWidth {
		width = int(math.Ceil(float64(bounds.Max.X) / CHAR_WIDTH))
	} else {
		w, err := strconv.ParseInt(widthVal, 10, 32)
		if err != nil {
			errs = append(errs, "invalid width: must be a number")
		}
		width = int(w)
	}

	if !okHeight {
		height = int(math.Ceil(float64(bounds.Max.Y) / CHAR_HEIGHT))
	} else {
		h, err := strconv.ParseInt(heightVal, 10, 32)
		if err != nil {
			errs = append(errs, "invalid height: must be a number")
		}
		height = int(h)
	}

	var err error
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, ", "))
	}
	return width, height, err
}

func isLightInverted(c *gin.Context) (bool, error) {
	invertedVal, ok := c.GetQuery("is_inverted")

	if ok {
		isInverted, err := strconv.ParseBool(invertedVal)
		if err != nil {
			return false, fmt.Errorf("invalid inverted: must be either 'true' or 'false'")
		}
		return isInverted, nil
	}

	return false, nil
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

func generateAscii(img image.Image, c *gin.Context) ([]string, int, error) {
	ascii := []string{}
	bounds := img.Bounds()

	// get query params
	minBrightness, err := getMinBrightness(c)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}

	userWidth, userHeight, err := getWidthAndHeight(bounds, c)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}

	isInverted, err := isLightInverted(c)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}
	isPixelOn := getBrightnessComparisonFunc(isInverted)

	fmt.Println(minBrightness, userWidth, userHeight, isInverted)
	pixelWidth, pixelHeight := CHAR_WIDTH*userWidth, CHAR_HEIGHT*userHeight
	originalWidth, originalHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	scaleX, scaleY := float64(pixelWidth)/float64(originalWidth), float64(pixelHeight)/float64(originalHeight)

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
		ascii = append(ascii, builder.String())
	}

	return ascii, http.StatusOK, nil
}

// func main() {
// 	image, err := fileio.ReadImageFromFile()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	ascii := generateAscii(image)
// 	file, err := os.Create("output.txt")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()

// 	for _, row := range ascii {
// 		file.WriteString(row)
// 	}
// }

func getAscii(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
		return
	}
	defer file.Close()

	image, imageType, err := image.Decode(file)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Bad image format: %s", imageType)})
		return
	}

	ascii, code, err := generateAscii(image, c)
	if code != http.StatusOK || err != nil {
		fmt.Println(code, err)
		c.IndentedJSON(code, gin.H{"error": err.Error()})
	} else {
		c.IndentedJSON(http.StatusOK, ascii)
	}
}

func main() {
	router := gin.Default()
	router.GET("/", getAscii)

	router.Run("localhost:8080")
}
