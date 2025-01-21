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
	STYLE_NORMAL        = "normal"
	STYLE_BRIGHTNESS    = "brightness"
	STYLE_HIGH_CONTRAST = "contrast"
)

// defaults
const (
	DEFAULT_BRIGHTNESS = 50.0
	DEFAULT_INVERTED   = false
	DEFAULT_STYLE      = STYLE_NORMAL
	DEFAULT_WIDTH      = 50
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

type RelativePosition struct {
	Dx int
	Dy int
}

type DitherNode struct {
	RelativePosition RelativePosition
	value            float64
}

type EncodingSettings struct {
	UsePercievedBrightness bool
	DitherNodes            []DitherNode
}

type Point struct {
	X int
	Y int
}

func getStyles() []string {
	return []string{STYLE_NORMAL, STYLE_BRIGHTNESS, STYLE_HIGH_CONTRAST}
}

func getInvalidStylesError() error {
	return fmt.Errorf("invalid style: must be one of the following: %s", strings.Join(getStyles(), ", "))
}

func validateBrightness(f *FormData) error {
	if f.Brightness != nil {
		brightness := *f.Brightness
		if brightness < MIN_BRIGHTNESS || brightness > MAX_BRIGHTNESS {
			return fmt.Errorf("invalid brightness: must be a number between %f & %f", MIN_BRIGHTNESS, MAX_BRIGHTNESS)
		}
	} else {
		defaultBrightness := DEFAULT_BRIGHTNESS
		f.Brightness = &defaultBrightness
	}

	return nil
}

func validateWidthAndHeight(f *FormData, bounds image.Rectangle) error {
	widthErrMsg := fmt.Sprintf("invalid width: must be a number between %d and %d", MIN_LENGTH, MAX_LENGTH)
	heightErrMsg := fmt.Sprintf("invalid height: must be a number between %d and %d", MIN_LENGTH, MAX_LENGTH)
	errs := []string{}

	if f.Width == nil {
		widthVal := DEFAULT_WIDTH
		f.Width = &widthVal
	} else if *f.Width < MIN_LENGTH || *f.Width > MAX_LENGTH {
		errs = append(errs, widthErrMsg)
	}

	if f.Height == nil {
		imgWidth, imgHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
		calculatedHeight := int(math.Round(float64(*f.Width*imgHeight) / float64(imgWidth) / 2.0))
		calculatedHeight = min(calculatedHeight, MAX_LENGTH)
		f.Height = &calculatedHeight
	} else if *f.Height < MIN_LENGTH || *f.Height > MAX_LENGTH {
		errs = append(errs, heightErrMsg)
	}

	var err error
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, ", "))
	}

	return err
}

func validateStyle(f *FormData) error {
	if f.Style != nil {
		style := *f.Style
		if !slices.Contains(getStyles(), style) {
			return getInvalidStylesError()
		}
	} else {
		defaultVal := DEFAULT_STYLE
		f.Style = &defaultVal
	}

	return nil
}

func getDither(style string) []DitherNode {
	switch style {
	case STYLE_NORMAL:
		return []DitherNode{
			{value: 7.0 / 16.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 3.0 / 16.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 5.0 / 16.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
			{value: 1.0 / 16.0, RelativePosition: RelativePosition{Dx: 1, Dy: 1}},
		}
	case STYLE_HIGH_CONTRAST:
		return []DitherNode{
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 2, Dy: 0}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 1, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 0, Dy: 2}},
		}
	}
	return []DitherNode{}
}

func getEncodingSettings(style string) (EncodingSettings, error) {
	var encodingSettings EncodingSettings
	err := getInvalidStylesError()

	switch style {
	case STYLE_NORMAL:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
		err = nil
	case STYLE_HIGH_CONTRAST:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
		err = nil
	case STYLE_BRIGHTNESS:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: true,
		}
		err = nil
	}

	return encodingSettings, err
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

func getColor(channel uint32, opacity float64) uint8 {
	color := uint8(channel >> 8)
	return uint8(math.Round(255.0 - opacity*float64(255-color)))
}

func getPixelLuminance(pixel color.Color) float64 {
	r, g, b, a := pixel.RGBA()

	// convert to 8-bit value
	opacity := float64(uint8(a>>8) / 255.0)
	red := getColor(r, opacity)
	green := getColor(g, opacity)
	blue := getColor(b, opacity)

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

func getFormData(c *gin.Context) (FormData, error) {
	var form FormData

	if err := c.ShouldBind(&form); err != nil {
		return form, err
	}

	return form, nil
}

func validateFormData(form *FormData, bounds image.Rectangle) error {
	if err := validateBrightness(form); err != nil {
		return err
	}

	if err := validateWidthAndHeight(form, bounds); err != nil {
		return err
	}

	if err := validateStyle(form); err != nil {
		return err
	}

	return nil
}

func getGrayscaleMatrix(img image.Image, totalWidth, totalHeight int) [][]float64 {
	bounds := img.Bounds()
	imageWidth, imageHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	scaleX, scaleY := float64(totalWidth)/float64(imageWidth), float64(totalHeight)/float64(imageHeight)

	getOriginalCoords := func(x, y int) (int, int) {
		originalX, originalY := float64(x)/float64(scaleX), float64(y)/float64(scaleY)
		return int(math.Round(originalX)), int(math.Round(originalY))
	}

	grayscale := make([][]float64, totalHeight)
	for y := range grayscale {
		grayscale[y] = make([]float64, totalWidth)
		for x := range grayscale[y] {
			originalX, originalY := getOriginalCoords(x, y)
			grayscale[y][x] = getPixelLuminance(img.At(originalX, originalY))
		}
	}

	return grayscale
}

func ditherMatrix(dither []DitherNode, grayscaleMatrix [][]float64, point Point, quantError float64) {
	compare := func(n, dn, length int) bool {
		if dn < n {
			return dn > 0
		}
		return dn < length
	}

	width, height := len(grayscaleMatrix[point.Y]), len(grayscaleMatrix)
	for _, node := range dither {
		dx, dy := point.X+node.RelativePosition.Dx, point.Y+node.RelativePosition.Dy
		if compare(point.X, dx, width) && compare(point.Y, dy, height) {
			grayscaleMatrix[dy][dx] = grayscaleMatrix[dy][dx] + quantError*node.value
		}
	}
}

func getMaxBrightness(brightness float64, usePercievedBrightness bool) float64 {
	if usePercievedBrightness {
		return brightness
	}
	return brightness / 100.0
}

func pixelsToAscii(point Point, form FormData, grayscaleMatrix [][]float64, encodingSettings EncodingSettings) rune {
	var offset uint8 = 0
	transformedX, transformedY := point.X*CHAR_WIDTH, point.Y*CHAR_HEIGHT
	maxBrightness := getMaxBrightness(*form.Brightness, encodingSettings.UsePercievedBrightness)

	for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
		for dx := 0; dx < int(CHAR_WIDTH); dx++ {
			x, y := transformedX+dx, transformedY+dy
			brightness := grayscaleMatrix[y][x]
			if encodingSettings.UsePercievedBrightness {
				brightness = getPercievedBrightness(brightness)
			}

			quantError := brightness
			if brightness < maxBrightness {
				offset |= (1 << getPixelNumber(dx, dy))
			} else {
				quantError -= 1.0
			}

			ditherMatrix(encodingSettings.DitherNodes, grayscaleMatrix, Point{X: x, Y: y}, quantError)
		}
	}

	return rune(0x2800 + int(offset))
}

func generateAscii(img image.Image, form FormData, encodingSettings EncodingSettings) ([]string, int, error) {
	ascii := []string{}
	grayscaleMatrix := getGrayscaleMatrix(img, CHAR_WIDTH**form.Width, CHAR_HEIGHT**form.Height)

	for y := 0; y < *form.Height; y++ {
		var builder strings.Builder
		for x := 0; x < *form.Width; x++ {
			builder.WriteRune(pixelsToAscii(Point{X: x, Y: y}, form, grayscaleMatrix, encodingSettings))
		}
		ascii = append(ascii, builder.String())
	}

	return ascii, http.StatusOK, nil
}

// func generateAscii(img image.Image, c *gin.Context) ([]string, int, error) {
// 	ascii := []string{}
// 	bounds := img.Bounds()

// 	// get form data
// 	var form FormData
// 	if err := c.ShouldBind(&form); err != nil {
// 		return ascii, http.StatusBadRequest, err
// 	}

// 	minBrightness, err := getMinBrightness(form)
// 	if err != nil {
// 		return ascii, http.StatusBadRequest, err
// 	}

// 	userWidth, userHeight, err := getWidthAndHeight(bounds, form)
// 	if err != nil {
// 		return ascii, http.StatusBadRequest, err
// 	}

// 	style, err := getStyle(form)
// 	if err != nil {
// 		return ascii, http.StatusBadRequest, err
// 	}

// 	isInverted := form.IsInvert.Bool()
// 	isPixelOn := getBrightnessComparisonFunc(isInverted)

// 	totalWidth, totalHeight := CHAR_WIDTH*userWidth, CHAR_HEIGHT*userHeight
// 	originalWidth, originalHeight := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
// 	scaleX, scaleY := float64(totalWidth)/float64(originalWidth), float64(totalHeight)/float64(originalHeight)

// 	pixelsToAscii := func(baseX, baseY int) rune {
// 		transformedX, transformedY := baseX*CHAR_WIDTH, baseY*CHAR_HEIGHT
// 		var offset uint8 = 0

// 		getPixelBrightness := func(pixel color.Color) float64 {
// 			luminance := getPixelLuminance(pixel)

// 			// return percieved brightness
// 			return getPercievedBrightness(luminance)
// 		}

// 		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
// 			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
// 				// map current pixel to original "pixel"
// 				x, y := transformedX+dx, transformedY+dy
// 				originalX, originalY := float64(x)/float64(scaleX), float64(y)/float64(scaleY)
// 				x1, x2 := int(math.Floor(originalX)), int(math.Ceil(originalX))
// 				y1, y2 := int(math.Floor(originalY)), int(math.Ceil(originalY))

// 				// take the average brightness of all 4 original pixels
// 				b1 := getPixelBrightness(img.At(x1, y1))
// 				b2 := getPixelBrightness(img.At(x1, y2))
// 				b3 := getPixelBrightness(img.At(x2, y1))
// 				b4 := getPixelBrightness(img.At(x2, y2))
// 				brightness := (b1 + b2 + b3 + b4) / 4

// 				// if brightness exceeds min brightness, render a pixel (update offset)
// 				if isPixelOn(brightness, minBrightness) {
// 					offset |= (1 << getPixelNumber(dx, dy))
// 				}
// 			}
// 		}

// 		return rune(0x2800 + int(offset))
// 	}

// 	// setup oldPixels storage, used by floyd-steinberg dithering
// 	oldPixels := make([][]float64, totalHeight)
// 	for y := range oldPixels {
// 		oldPixels[y] = make([]float64, totalWidth)
// 		for x := range oldPixels[y] {
// 			oldPixels[y][x] = -1.0
// 		}
// 	}
// 	oldPixels[0][0] = getPixelLuminance(img.At(0, 0))

// 	pixelsToAscii2 := func(baseX, baseY int) rune {
// 		getOriginalCoords := func(x, y int) (int, int) {
// 			originalX, originalY := float64(x)/float64(scaleX), float64(y)/float64(scaleY)
// 			return int(math.Round(originalX)), int(math.Round(originalY))
// 		}

// 		transformedX, transformedY := baseX*CHAR_WIDTH, baseY*CHAR_HEIGHT
// 		var offset uint8 = 0

// 		for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
// 			for dx := 0; dx < int(CHAR_WIDTH); dx++ {
// 				x, y := transformedX+dx, transformedY+dy
// 				luminance := oldPixels[y][x]
// 				brightness := getPercievedBrightness(luminance)
// 				isPixelEnabled := isPixelOn(brightness, minBrightness)
// 				quantError := luminance
// 				// if isInverted {
// 				// 	if brightness <= minBrightness {
// 				// 		offset |= (1 << getPixelNumber(dx, dy))
// 				// 		quantError = luminance - 1.0
// 				// 	}
// 				// } else {
// 				// 	if brightness >= minBrightness {
// 				// 		offset |= (1 << getPixelNumber(dx, dy))
// 				// 		quantError = luminance - 1.0
// 				// 	}
// 				// }
// 				if isPixelEnabled {
// 					offset |= (1 << getPixelNumber(dx, dy))
// 					if !isInverted {
// 						quantError = luminance - 1.0
// 					}
// 				} else {
// 					if isInverted {
// 						quantError = luminance - 1.0
// 					}
// 				}

// 				// dithering
// 				if x+1 < totalWidth {
// 					if oldPixels[y][x+1] == -1.0 {
// 						originalX, originalY := getOriginalCoords(x+1, y)
// 						oldPixels[y][x+1] = getPixelLuminance(img.At(originalX, originalY))
// 					}
// 					oldPixels[y][x+1] = oldPixels[y][x+1] + quantError*(7.0/16.0)
// 				}
// 				if x-1 >= 0 && y+1 < totalHeight {
// 					if oldPixels[y+1][x-1] == -1.0 {
// 						originalX, originalY := getOriginalCoords(x-1, y+1)
// 						oldPixels[y+1][x-1] = getPixelLuminance(img.At(originalX, originalY))
// 					}
// 					oldPixels[y+1][x-1] = oldPixels[y+1][x-1] + quantError*(3.0/16.0)
// 				}
// 				if y+1 < totalHeight {
// 					if oldPixels[y+1][x] == -1.0 {
// 						originalX, originalY := getOriginalCoords(x, y+1)
// 						oldPixels[y+1][x] = getPixelLuminance(img.At(originalX, originalY))
// 					}
// 					oldPixels[y+1][x] = oldPixels[y+1][x] + quantError*(5.0/16.0)
// 				}
// 				if y+1 < totalHeight && x+1 < totalWidth {
// 					if oldPixels[y+1][x+1] == -1.0 {
// 						originalX, originalY := getOriginalCoords(x+1, y+1)
// 						oldPixels[y+1][x+1] = getPixelLuminance(img.At(originalX, originalY))
// 					}
// 					oldPixels[y+1][x+1] = oldPixels[y+1][x+1] + quantError*(1.0/16.0)
// 				}
// 			}
// 		}

// 		return rune(0x2800 + int(offset))
// 	}

// 	switch style {
// 	case STYLE_BRIGHTNESS:
// 		for y := 0; y < userHeight; y++ {
// 			var builder strings.Builder
// 			for x := 0; x < userWidth; x++ {
// 				builder.WriteRune(pixelsToAscii(x, y))
// 			}
// 			ascii = append(ascii, builder.String())
// 		}
// 	case STYLE_NORMAL:
// 		for y := 0; y < userHeight; y++ {
// 			var builder strings.Builder
// 			for x := 0; x < userWidth; x++ {
// 				builder.WriteRune(pixelsToAscii2(x, y))
// 			}
// 			ascii = append(ascii, builder.String())
// 		}
// 	}

// 	return ascii, http.StatusOK, nil
// }

func getAscii(c *gin.Context) {
	// attempt to open image, and validate it
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

	// read form data, and validate it
	form, err := getFormData(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateFormData(&form, image.Bounds()); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// determine encoding settings
	encodingSettings, err := getEncodingSettings(*form.Style)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// attempt to generate ascii
	ascii, code, err := generateAscii(image, form, encodingSettings)
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
