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
	DEFAULT_THEME    = THEME_LIGHT
	DEFAULT_WIDTH    = 60
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

// form field names [ensure matches FormData struct]
const (
	FORM_THEME_NAME    = "theme"
	FORM_WIDTH_NAME    = "width"
	FORM_HEIGHT_NAME   = "height"
	FORM_INVERT_NAME   = "invert"
	FORM_EXPOSURE_NAME = "exposure"
	FORM_STYLE_NAME    = "style"
	FORM_IMAGE_NAME    = "image"
)

// CheckboxBool struct for form checkboxes
type CheckboxBool string

func (cb CheckboxBool) Bool() bool {
	return cb == "on"
}

// FormData struct to parse form body
type FormData struct {
	Theme    *string      `form:"theme"`
	Width    *int         `form:"width"`
	Height   *int         `form:"height"`
	IsInvert CheckboxBool `form:"invert"`
	Exposure *float64     `form:"exposure"`
	Style    *string      `form:"style"`
}

// Relative Position struct for DitherNode
type RelativePosition struct {
	Dx int
	Dy int
}

// Dither node struct for dithering algorithms
type DitherNode struct {
	RelativePosition RelativePosition
	value            float64
}

// Encoding settings struct to describe how to encode image based on style
type EncodingSettings struct {
	UsePercievedBrightness bool
	DitherNodes            []DitherNode
}

// Point struct for representing position in image
type Point struct {
	X int
	Y int
}

// Option struct to represent an option tag in HTML
type Option struct {
	Value string
	Label string
}

// getThemes returns the valid web themes.
func getThemes() []string {
	return []string{THEME_LIGHT, THEME_DARK}
}

// getStyles returns the valid encoding styles.
func getStyles() []string {
	return []string{STYLE_NORMAL, STYLE_BRIGHTNESS, STYLE_HIGH_CONTRAST, STYLE_EDGE_CONTRAST, STYLE_SMOOTH}
}

// getInvalidStylesError returns an error that specifies to the user than the style is invalid
func getInvalidStylesError() error {
	return fmt.Errorf("invalid style: must be one of the following: %s", strings.Join(getStyles(), ", "))
}

// validateTheme ensures that the `theme` attribute of f is valid.
// Returns error if validation fails, nil otherwise.
// If theme is unset, update theme attribute to take on `defaultTheme`, return nil.
// If theme is set, and validated, return nil.
// If theme is set, but not validated, return error.
func validateTheme(f *FormData) error {
	if f.Theme != nil {
		theme := *f.Theme
		themes := getThemes()

		if !slices.Contains(themes, theme) {
			return fmt.Errorf("invalid theme: must be one of the following: %s", strings.Join(themes, ", "))
		}
	} else {
		defaultTheme := DEFAULT_THEME
		f.Theme = &defaultTheme
	}

	return nil
}

// validateExposure ensures that the `exposure` attribute of f is valid.
// Returns error if validation fails, nil otherwise.
// If exposure is unset, update exposure attribute to take on default value, return nil.
// If exposure is set, and validated, return nil.
// If exposure is set, but not validated, return error.
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

// validateWidthAndHeight ensures that the `width` and `height` attributes of f are valid.
// Returns error if validation fails, nil otherwise.
// If width / height is unset, update width / height attribute to take on default value, return nil.
// If width / height is set, and both are validated, return nil.
// If width / height is set, but one or both is not validated, return error.
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

// validateStyle ensures that the `style` attribute of f is valid.
// Returns error if validation fails, nil otherwise.
// If style is unset, update style attribute to take on default value, return nil.
// If style is set, and validated, return nil.
// If style is set, but not validated, return error.
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

// getDither returns the slice of DitherNodes associated with an encoding style.
// Generally, returns an non-empty slice of DitherNodes.
// However, if style does not use dithering, returns an empty slice of DitherNodes.
func getDither(style string) []DitherNode {
	switch style {
	case STYLE_NORMAL:
		return []DitherNode{
			// Floyd-Steinberg [https://en.wikipedia.org/wiki/Floyd%E2%80%93Steinberg_dithering]
			{value: 7.0 / 16.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 3.0 / 16.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 5.0 / 16.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
			{value: 1.0 / 16.0, RelativePosition: RelativePosition{Dx: 1, Dy: 1}},
		}
	case STYLE_HIGH_CONTRAST:
		// Atkinson [https://en.wikipedia.org/wiki/Atkinson_dithering]
		return []DitherNode{
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 2, Dy: 0}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 1, Dy: 1}},
			{value: 1.0 / 8.0, RelativePosition: RelativePosition{Dx: 0, Dy: 2}},
		}
	case STYLE_EDGE_CONTRAST:
		// Sierra Lite [https://tannerhelland.com/2012/12/28/dithering-eleven-algorithms-source-code.html#sierra-dithering]
		return []DitherNode{
			{value: 2.0 / 4.0, RelativePosition: RelativePosition{Dx: 1, Dy: 0}},
			{value: 1.0 / 4.0, RelativePosition: RelativePosition{Dx: -1, Dy: 1}},
			{value: 1.0 / 4.0, RelativePosition: RelativePosition{Dx: 0, Dy: 1}},
		}
	case STYLE_SMOOTH:
		// Minimized Average Error [https://en.wikipedia.org/wiki/Error_diffusion#minimized_average_error]
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

// getEncodingSettings returns the encoding settings associated with an encoding style.
// Generally this function returns EncodingSettings struct with a `nil` error.
// If style has no encoding setting, we define error in our return.
func getEncodingSettings(style string) (EncodingSettings, error) {
	var encodingSettings EncodingSettings
	var err error
	isValidStyle := true

	switch style {
	case STYLE_NORMAL:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
	case STYLE_HIGH_CONTRAST:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
	case STYLE_BRIGHTNESS:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: true,
		}
	case STYLE_EDGE_CONTRAST:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
	case STYLE_SMOOTH:
		encodingSettings = EncodingSettings{
			DitherNodes:            getDither(style),
			UsePercievedBrightness: false,
		}
	default:
		isValidStyle = false
	}

	if !isValidStyle {
		err = getInvalidStylesError()
	}

	return encodingSettings, err
}

// getLinearizedChannel takes a standard, 8-bit color channel, and converts it to a linearized value between 0.0 and 1.0.
// For more information, see: https://en.wikipedia.org/wiki/SRGB#Transfer_function_(%22gamma%22)
func getLinearizedChannel(colorChannel uint8) float64 {
	v := float64(colorChannel) / 255.0

	if v <= 0.04045 {
		return v / 12.92
	} else {
		return math.Pow((v+0.055)/1.055, 2.4)
	}
}

// getLuminance takes lineralized r, g, and b values, and determines a pixels luminance, a value between 0.0 and 1.0, where
// 0 represents most dark, and 1.0 represents most bright.
// For more information, see: https://en.wikipedia.org/wiki/Relative_luminance#Relative_luminance_and_%22gamma_encoded%22_colorspaces
func getLuminance(r float64, g float64, b float64) float64 {
	return (0.2126 * r) + (0.7152 * g) + (0.0722 * b)
}

// getColor takes a full 32-bit color channel, and an opacity value between 0.0 and 1.0, and converts the color to the 8-bit
// representation with opaicty "applied" such that the full color can be represented as RGB without A.
func getColor(channel uint32, opacity float64) uint8 {
	color := uint8(channel >> 8)
	return uint8(math.Round(255.0 - opacity*float64(255-color)))
}

// getPixelLuminance takes a pixel, and returns it's luminance.
// At a high level, this converts a full-color pixel to a black-and-white value, represented as a number between 0.0 and 1.0.
func getPixelLuminance(pixel color.Color) float64 {
	r, g, b, a := pixel.RGBA()

	// convert to 8-bit value
	opacity := float64(uint8(a>>8) / 255.0)
	red := getColor(r, opacity)
	green := getColor(g, opacity)
	blue := getColor(b, opacity)

	lr, lg, lb := getLinearizedChannel(red), getLinearizedChannel(green), getLinearizedChannel(blue)

	return getLuminance(lr, lg, lb)
}

// clampLuminance takes a luminance value, and forces it to be a float between 0.0 and 1.0.
func clampLuminance(luminance float64) float64 {
	if luminance < 0 {
		return 0
	}
	if luminance > 1 {
		return 1
	}
	return luminance
}

// getPercievedLuminance takes a luminance value, and returns it's percieved brightness.
// For more information, see: https://en.wikipedia.org/wiki/Lightness#1976
func getPercievedBrightness(l float64) float64 {
	luminance := clampLuminance(l)
	if luminance <= 0.008856 {
		return luminance * 903.3
	} else {
		return math.Pow(luminance, 1.0/3.0)*116 - 16
	}
}

// getPixelNumber maps a relative pixel coordinate to it's bit position of a brail ASCII, a number between 0 and 7.
// x must be an int between 0 and 1.
// y must be an int between 0 and 3.
// ⣿ <- for a better understanding. You can see a brail character is 4x2 pixels.
func getPixelNumber(x int, y int) int {
	if y <= 2 {
		return 3*x + y
	}
	return 2*y + x
}

// getFormData takes a gin context, and returns the request body based on the FormData struct.
// Returns an error if request body is malformed.
func getFormData(c *gin.Context) (FormData, error) {
	var form FormData

	if err := c.ShouldBind(&form); err != nil {
		return form, err
	}

	return form, nil
}

// validateFormData validates each form field that requires it.
// If all validation tests pass, then this function will simply return nil.
// If at least one validation test fails, then return an error with more details.
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

// getGrayscaleMatrix takes an image, and returns it in a grayscaled matrix format, with dimensions `totalHeight` x `totalWidth`.
// Each element in the matrix represents a pixel, converted to grayscale (luminance).
// Note that grayscale[y][x] does not correspond to image.At(x, y), since the width and height of the image may not correspond
// to `totalWidth` & `totalHeight`.
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

// diffuseError performs the error diffusion operation of a dithering algorithm.
// For more information, see: [https://en.wikipedia.org/wiki/Error_diffusion]
func diffuseError(dither []DitherNode, grayscaleMatrix [][]float64, point Point, quantError float64) {
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

// getMaxExposure determines the maximum exposure we should use as a threshold, which is dependent on `usePercievedBrightness`.
// Generally, we want our exposure to be a number between 0.0 and 1.0. Since user provides a number between 0 and 100, we need
// to divide by 100.
// However, if we are using percieved brightness as our threshold, then we can keep it as a number between 0 and 100.
func getMaxExposure(exposure float64, usePercievedBrightness bool) float64 {
	if usePercievedBrightness {
		return exposure
	}
	return exposure / 100.0
}

// pixelsToAscii converts a set of 8 pixels, starting at `point` and forming a brail shape (⣿), into an ASCII character,
// by analysing each pixel invididually, based on the exposure of each pixel.
// This function will diffuse the error generated by each pixel on every iteration.
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

			diffuseError(encodingSettings.DitherNodes, grayscaleMatrix, Point{X: x, Y: y}, quantError)
		}
	}

	return rune(0x2800 + int(offset))
}

// isInvertNeeded determines if we need to invert the ascii matrix before returning the result.
// Depends on `isInverted` and `theme`, both being defined in the request body.
// General logic: IF isInverted XOR theme => Invert NOT NEEDED; ELSE => Invert NEEDED
func isInvertNeeded(isInverted bool, theme string) bool {
	return isInverted == (theme == THEME_LIGHT)
}

// invertBrail takes a brail rune, and "inverts" it, by flipping the bits that control the brail (final byte) with XOR.
func invertBrail(r rune) rune {
	return r ^ 0xFF
}

// invertAscii loops over the entire ascii, and inverts each brail element.
// Note that this function generates a copy of the original ascii slice.
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

// generateAscii takes our input image, as well as configuration settings, and generates an ASCII representation of the image
func generateAscii(img image.Image, form FormData, encodingSettings EncodingSettings) []string {
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

	return ascii
}

// getAscii is the function executed when a user does a POST request to "/".
// This function parses the request body, and if validated, will generate an ASCII representation of their image.
// In the event of a success, the server will return a simple JSON object containing an ASCII matrix.
// In the event of a failure, the server will return an error JSON object to the client.
func getAscii(c *gin.Context) {
	// attempt to open image, and validate it
	file, _, err := c.Request.FormFile(FORM_IMAGE_NAME)
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
	ascii := generateAscii(image, form, encodingSettings)
	c.IndentedJSON(http.StatusOK, ascii)
}

// getWebClient is the function executed when a user does a GET request to "/".
// This simply returns a templated HTML file.
func getWebClient(c *gin.Context) {
	styleOptions := []Option{
		{Value: STYLE_NORMAL, Label: "Normal"},
		{Value: STYLE_HIGH_CONTRAST, Label: "High Contrast"},
		{Value: STYLE_EDGE_CONTRAST, Label: "Edge Contrast"},
		{Value: STYLE_SMOOTH, Label: "Smooth"},
		{Value: STYLE_BRIGHTNESS, Label: "Brightness"},
	}

	data := gin.H{
		"styleOptions": styleOptions,
		"names": gin.H{
			"image":    FORM_IMAGE_NAME,
			"theme":    FORM_THEME_NAME,
			"width":    FORM_WIDTH_NAME,
			"height":   FORM_HEIGHT_NAME,
			"invert":   FORM_INVERT_NAME,
			"exposure": FORM_EXPOSURE_NAME,
			"style":    FORM_STYLE_NAME,
		},
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// main establishes our server, and listens for GET and POST requests.
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
