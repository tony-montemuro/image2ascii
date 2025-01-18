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
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// styles
const (
	STYLE_NORMAL     = "normal"
	STYLE_BRIGHTNESS = "brightness"
)

// defaults
const (
	DEFAULT_BRIGHTNESS = 50.0
	DEFAULT_INVERTED   = false
	DEFAULT_STYLE      = STYLE_NORMAL
)

// ascii properties
const (
	CHAR_WIDTH  = 2
	CHAR_HEIGHT = 4
)

// limits
const (
	MIN_BRIGHTNESS = 0.0
	MAX_BRIGHTNESS = 100.0
	MIN_LENGTH     = 1
	MAX_LENGTH     = 1000
)

type BrightnessComparisonFunc func(a, target float64) bool

type CheckboxBool string

func (cb CheckboxBool) Bool() bool {
	return cb == "on"
}

type FormData struct {
	Width      *int         `form:"width"`
	Height     *int         `form:"height"`
	IsInvert   CheckboxBool `form:"invert"`
	Brightness *float64     `form:"brightness"`
	Style      *string      `form:"style"`
}

func getMinBrightness(f FormData) (float64, error) {
	brightness := f.Brightness
	if brightness != nil {
		if *brightness < MIN_BRIGHTNESS || *brightness > MAX_BRIGHTNESS {
			return 0, fmt.Errorf("invalid brightness: must be a number between %f & %f", MIN_BRIGHTNESS, MAX_BRIGHTNESS)
		}
		return MAX_BRIGHTNESS - *brightness, nil
	}

	return DEFAULT_BRIGHTNESS, nil
}

func getWidthAndHeight(bounds image.Rectangle, f FormData) (int, int, error) {
	width, height := f.Width, f.Height
	widthErrMsg := fmt.Sprintf("invalid width: must be a number between %d and %d", MIN_LENGTH, MAX_LENGTH)
	heightErrMsg := fmt.Sprintf("invalid height: must be a number between %d and %d", MIN_LENGTH, MAX_LENGTH)
	errs := []string{}

	if width == nil {
		widthVal := int(math.Ceil(float64(bounds.Max.X) / CHAR_WIDTH))
		width = &widthVal
	} else {
		if *width < MIN_LENGTH || *width > MAX_LENGTH {
			errs = append(errs, widthErrMsg)
		}
	}

	if height == nil {
		heightVal := int(math.Ceil(float64(bounds.Max.Y) / CHAR_HEIGHT))
		height = &heightVal
	} else {
		if *height < MIN_LENGTH || *height > MAX_LENGTH {
			errs = append(errs, heightErrMsg)
		}
	}

	var err error
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, ", "))
	}
	return *width, *height, err
}

func getStyle(f FormData) (string, error) {
	style := f.Style
	if style != nil {
		styles := []string{STYLE_NORMAL, STYLE_BRIGHTNESS}
		if slices.Contains(styles, *style) {
			return "", fmt.Errorf("invalid style: must be a number between %f & %f", MIN_BRIGHTNESS, MAX_BRIGHTNESS)
		}
		return *style, nil
	}

	return DEFAULT_STYLE, nil
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

func getPixelLuminance(pixel color.Color) float64 {
	r, g, b, _ := pixel.RGBA()

	// convert to 8-bit value
	red := uint8(r >> 8)
	green := uint8(g >> 8)
	blue := uint8(b >> 8)

	// linearize red, green, blue
	lr, lg, lb := getLinearizedChannel(red), getLinearizedChannel(green), getLinearizedChannel(blue)

	// calculate luminance
	return getLuminance(lr, lg, lb)
}

func clampLuminance(luminance float64) float64 {
	if luminance < 0 {
		return 0
	}
	if luminance > 1 {
		return 1
	}
	return luminance
}

func getPercievedBrightness(l float64) float64 {
	luminance := clampLuminance(l)
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

	// get form data
	var form FormData
	if err := c.ShouldBind(&form); err != nil {
		return ascii, http.StatusBadRequest, err
	}

	minBrightness, err := getMinBrightness(form)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}

	userWidth, userHeight, err := getWidthAndHeight(bounds, form)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}

	style, err := getStyle(form)
	if err != nil {
		return ascii, http.StatusBadRequest, err
	}

	isInverted := form.IsInvert.Bool()
	isPixelOn := getBrightnessComparisonFunc(isInverted)

	totalWidth, totalHeight := CHAR_WIDTH*userWidth, CHAR_HEIGHT*userHeight
	originalWidth, originalHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	scaleX, scaleY := float64(totalWidth)/float64(originalWidth), float64(totalHeight)/float64(originalHeight)

	pixelsToAscii := func(baseX, baseY int) rune {
		transformedX, transformedY := baseX*CHAR_WIDTH, baseY*CHAR_HEIGHT
		var offset uint8 = 0

		getPixelBrightness := func(pixel color.Color) float64 {
			luminance := getPixelLuminance(pixel)

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

	// setup oldPixels storage, used by floyd-steinberg dithering
	oldPixels := make([][]float64, totalHeight)
	for y := range oldPixels {
		oldPixels[y] = make([]float64, totalWidth)
		for x := range oldPixels[y] {
			oldPixels[y][x] = -1.0
		}
	}
	oldPixels[0][0] = getPixelLuminance(img.At(0, 0))

	pixelsToAscii2 := func(baseX, baseY int) rune {
		getOriginalCoords := func(x, y int) (int, int) {
			originalX, originalY := float64(x)/float64(scaleX), float64(y)/float64(scaleY)
			return int(math.Round(originalX)), int(math.Round(originalY))
		}

		transformedX, transformedY := baseX*CHAR_WIDTH, baseY*CHAR_HEIGHT
		var offset uint8 = 0

		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
				x, y := transformedX+dx, transformedY+dy
				luminance := oldPixels[y][x]
				brightness := getPercievedBrightness(luminance)
				isPixelEnabled := isPixelOn(brightness, minBrightness)
				quantError := luminance
				if isPixelEnabled {
					offset |= (1 << getPixelNumber(dx, dy))
					quantError = luminance - 1.0
				}

				// dithering
				if x+1 < totalWidth {
					if oldPixels[y][x+1] == -1.0 {
						originalX, originalY := getOriginalCoords(x+1, y)
						oldPixels[y][x+1] = getPixelLuminance(img.At(originalX, originalY))
					}
					oldPixels[y][x+1] = oldPixels[y][x+1] + quantError*(7.0/16.0)
				}
				if x-1 >= 0 && y+1 < totalHeight {
					if oldPixels[y+1][x-1] == -1.0 {
						originalX, originalY := getOriginalCoords(x-1, y+1)
						oldPixels[y+1][x-1] = getPixelLuminance(img.At(originalX, originalY))
					}
					oldPixels[y+1][x-1] = oldPixels[y+1][x-1] + quantError*(3.0/16.0)
				}
				if y+1 < totalHeight {
					if oldPixels[y+1][x] == -1.0 {
						originalX, originalY := getOriginalCoords(x, y+1)
						oldPixels[y+1][x] = getPixelLuminance(img.At(originalX, originalY))
					}
					oldPixels[y+1][x] = oldPixels[y+1][x] + quantError*(5.0/16.0)
				}
				if y+1 < totalHeight && x+1 < totalWidth {
					if oldPixels[y+1][x+1] == -1.0 {
						originalX, originalY := getOriginalCoords(x+1, y+1)
						oldPixels[y+1][x+1] = getPixelLuminance(img.At(originalX, originalY))
					}
					oldPixels[y+1][x+1] = oldPixels[y+1][x+1] + quantError*(1.0/16.0)
				}
			}
		}

		return rune(0x2800 + int(offset))
	}

	switch style {
	case STYLE_BRIGHTNESS:
		for y := 0; y < userHeight; y++ {
			var builder strings.Builder
			for x := 0; x < userWidth; x++ {
				builder.WriteRune(pixelsToAscii(x, y))
			}
			ascii = append(ascii, builder.String())
		}
	case STYLE_NORMAL:
		for y := 0; y < userHeight; y++ {
			var builder strings.Builder
			for x := 0; x < userWidth; x++ {
				builder.WriteRune(pixelsToAscii2(x, y))
			}
			ascii = append(ascii, builder.String())
		}
		fmt.Println(oldPixels)
	}

	return ascii, http.StatusOK, nil
}

func getAscii(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "no image provided"})
		return
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "bad image format: must be either png or jpg/jpeg"})
		return
	}

	ascii, code, err := generateAscii(image, c)
	if code != http.StatusOK || err != nil {
		c.IndentedJSON(code, gin.H{"error": err.Error()})
	} else {
		c.IndentedJSON(http.StatusOK, ascii)
	}
}

func getWebClient(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
	router.Static("/assets", "./assets")

	// web client
	router.GET("/", getWebClient)

	// api
	router.POST("/", getAscii)

	router.Run("localhost:8080")
}
