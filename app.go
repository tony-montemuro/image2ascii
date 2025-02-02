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

// themes
const (
	THEME_LIGHT = "light"
	THEME_DARK  = "dark"
)

// styles
const (
	STYLE_NORMAL        = "normal"
	STYLE_BRIGHTNESS    = "brightness"
	STYLE_HIGH_CONTRAST = "contrast"
	STYLE_EDGE_CONTRAST = "edge"
	STYLE_SMOOTH        = "smooth"
)

// defaults
const (
	DEFAULT_EXPOSURE = 50.0
	DEFAULT_INVERTED = false
	DEFAULT_STYLE    = STYLE_NORMAL
	DEFAULT_WIDTH    = 50
)

// ascii properties
const (
	CHAR_WIDTH  = 2
	CHAR_HEIGHT = 4
)

// limits
const (
	MIN_EXPOSURE = 0.0
	MAX_EXPOSURE = 100.0
	MIN_LENGTH   = 1
	MAX_LENGTH   = 500
)

type CheckboxBool string

func (cb CheckboxBool) Bool() bool {
	return cb == "on"
}

type FormData struct {
	Theme    *string      `form:"theme"`
	Width    *int         `form:"width"`
	Height   *int         `form:"height"`
	IsInvert CheckboxBool `form:"invert"`
	Exposure *float64     `form:"exposure"`
	Style    *string      `form:"style"`
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

func getThemes() []string {
	return []string{THEME_LIGHT, THEME_DARK}
}

func getStyles() []string {
	return []string{STYLE_NORMAL, STYLE_BRIGHTNESS, STYLE_HIGH_CONTRAST, STYLE_EDGE_CONTRAST, STYLE_SMOOTH}
}

func getInvalidStylesError() error {
	return fmt.Errorf("invalid style: must be one of the following: %s", strings.Join(getStyles(), ", "))
}

func validateTheme(f *FormData) error {
	if f.Theme != nil {
		theme := *f.Theme
		themes := getThemes()

		if !slices.Contains(themes, theme) {
			return fmt.Errorf("invalid theme: must be one of the following: %s", strings.Join(themes, ", "))
		}
	} else {
		defaultTheme := THEME_LIGHT
		f.Theme = &defaultTheme
	}

	return nil
}

func validateExposure(f *FormData) error {
	if f.Exposure != nil {
		exposure := *f.Exposure

		if exposure < MIN_EXPOSURE || exposure > MAX_EXPOSURE {
			return fmt.Errorf("invalid exposure: must be a number between %f & %f", MIN_EXPOSURE, MAX_EXPOSURE)
		}

		reversedExposure := MAX_EXPOSURE - exposure
		f.Exposure = &reversedExposure
	} else {
		defaultExposure := DEFAULT_EXPOSURE
		f.Exposure = &defaultExposure
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
	case STYLE_EDGE_CONTRAST:
		return []DitherNode{
			{value: 2.0 / 4.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 1.0 / 4.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 1.0 / 4.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
		}
	case STYLE_SMOOTH:
		return []DitherNode{
			{value: 7.0 / 48.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 5.0 / 48.0, RelativePosition: RelativePosition{Dx: 2, Dy: 0}},
			{value: 3.0 / 48.0, RelativePosition: RelativePosition{Dx: -2, Dy: 1}},
			{value: 5.0 / 48.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 7.0 / 48.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
			{value: 5.0 / 48.0, RelativePosition: RelativePosition{Dx: 1, Dy: 1}},
			{value: 3.0 / 48.0, RelativePosition: RelativePosition{Dx: 2, Dy: 1}},
			{value: 1.0 / 48.0, RelativePosition: RelativePosition{Dx: -2, Dy: 2}},
			{value: 3.0 / 48.0, RelativePosition: RelativePosition{Dx: -1, Dy: 2}},
			{value: 5.0 / 48.0, RelativePosition: RelativePosition{Dx: 0, Dy: 2}},
			{value: 3.0 / 48.0, RelativePosition: RelativePosition{Dx: 1, Dy: 2}},
			{value: 1.0 / 48.0, RelativePosition: RelativePosition{Dx: 2, Dy: 2}},
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
	case STYLE_EDGE_CONTRAST:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
		err = nil
	case STYLE_SMOOTH:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
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

func getFormData(c *gin.Context) (FormData, error) {
	var form FormData

	if err := c.ShouldBind(&form); err != nil {
		return form, err
	}

	return form, nil
}

func validateFormData(form *FormData, bounds image.Rectangle) error {
	if err := validateTheme(form); err != nil {
		return err
	}

	if err := validateExposure(form); err != nil {
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

func getMaxExposure(exposure float64, usePercievedBrightness bool) float64 {
	if usePercievedBrightness {
		return exposure
	}
	return exposure / 100.0
}

func pixelsToAscii(point Point, form FormData, grayscaleMatrix [][]float64, encodingSettings EncodingSettings) rune {
	var offset uint8 = 0
	transformedX, transformedY := point.X*CHAR_WIDTH, point.Y*CHAR_HEIGHT
	maxExposure := getMaxExposure(*form.Exposure, encodingSettings.UsePercievedBrightness)

	for dy := 0; dy < int(CHAR_HEIGHT); dy++ {
		for dx := 0; dx < int(CHAR_WIDTH); dx++ {
			x, y := transformedX+dx, transformedY+dy
			exposure := grayscaleMatrix[y][x]
			if encodingSettings.UsePercievedBrightness {
				exposure = getPercievedBrightness(exposure)
			}

			quantError := exposure
			if exposure < maxExposure {
				offset |= (1 << getPixelNumber(dx, dy))
			} else {
				quantError -= 1.0
			}

			ditherMatrix(encodingSettings.DitherNodes, grayscaleMatrix, Point{X: x, Y: y}, quantError)
		}
	}

	return rune(0x2800 + int(offset))
}

func isInvertNeeded(isInverted bool, theme string) bool {
	return isInverted == (theme == THEME_LIGHT)
}

func invertBrail(r rune) rune {
	return r ^ 0xFF
}

func invertAscii(ascii []string) []string {
	invertedAscii := make([]string, len(ascii))

	for y, row := range ascii {
		var builder strings.Builder
		for _, r := range row {
			builder.WriteRune(invertBrail(r))
		}
		invertedAscii[y] = builder.String()
	}

	return invertedAscii
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

	if isInvertNeeded(form.IsInvert.Bool(), *form.Theme) {
		ascii = invertAscii(ascii)
	}

	return ascii, http.StatusOK, nil
}

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
		return
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
