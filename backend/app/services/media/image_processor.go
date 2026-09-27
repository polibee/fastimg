package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"mime"
	"path/filepath"
	"strings"
)

var (
	ErrImageTooLarge           = errors.New("image exceeds upload size limit")
	ErrImageDimensionsExceeded = errors.New("image exceeds pixel limit")
	ErrImageFormatMismatch     = errors.New("image extension or content type does not match its data")
	ErrInvalidImage            = errors.New("invalid image data")
	ErrUnsupportedImage        = errors.New("unsupported image format")
	ErrTooManyFrames           = errors.New("animated image exceeds frame limit")
	ErrAnimationTooLarge       = errors.New("animated image exceeds decoded pixel limit")
	ErrImageProcessing         = errors.New("image processing failed")
)

const (
	defaultMaxImageBytes      int64 = 10_000_000
	defaultMaxImagePixels     int64 = 20_000_000
	defaultMaxAnimationPixels int64 = 40_000_000
	defaultMaxAnimationFrames       = 50
)

// ImageLimits bounds both the compressed upload and the decoded work performed
// by the synchronous MVP processor. Zero values use conservative defaults.
type ImageLimits struct {
	MaxBytes           int64
	MaxPixels          int64
	MaxAnimationPixels int64
	MaxAnimationFrames int
}

type ProcessedImage struct {
	Format      string
	ContentType string
	Width       int64
	Height      int64
	Original    []byte
}

// ProcessOptions contains plan-controlled transformations applied before the
// normalized original is stored.
type ProcessOptions struct {
	WatermarkEnabled bool
	WatermarkText    string
	WatermarkDomain  string
}

type ImageProcessor struct {
	limits ImageLimits
}

func NewImageProcessor(limits ImageLimits) *ImageProcessor {
	if limits.MaxBytes <= 0 {
		limits.MaxBytes = defaultMaxImageBytes
	}
	if limits.MaxPixels <= 0 {
		limits.MaxPixels = defaultMaxImagePixels
	}
	if limits.MaxAnimationPixels <= 0 {
		limits.MaxAnimationPixels = defaultMaxAnimationPixels
	}
	if limits.MaxAnimationFrames <= 0 {
		limits.MaxAnimationFrames = defaultMaxAnimationFrames
	}

	return &ImageProcessor{limits: limits}
}

// Process verifies image bytes against the supplied filename and MIME type and
// strips embedded metadata by decoding/re-encoding.
func (p *ImageProcessor) Process(filename, declaredContentType string, source []byte) (ProcessedImage, error) {
	return p.ProcessWithOptions(filename, declaredContentType, source, ProcessOptions{})
}

// ProcessWithOptions verifies and normalizes an uploaded image. The watermark
// is applied before the single stored original is encoded, so a plan cannot
// bypass it by requesting the public original URL.
func (p *ImageProcessor) ProcessWithOptions(filename, declaredContentType string, source []byte, options ProcessOptions) (ProcessedImage, error) {
	if int64(len(source)) == 0 {
		return ProcessedImage{}, ErrInvalidImage
	}
	if int64(len(source)) > p.limits.MaxBytes {
		return ProcessedImage{}, ErrImageTooLarge
	}

	extensionFormat, ok := formatForExtension(filepath.Ext(filename))
	if !ok {
		return ProcessedImage{}, ErrUnsupportedImage
	}
	if declaredContentType != "" {
		mediaType, _, err := mime.ParseMediaType(declaredContentType)
		if err != nil {
			return ProcessedImage{}, ErrImageFormatMismatch
		}
		mimeFormat, supported := formatForMIME(mediaType)
		if !supported || mimeFormat != extensionFormat {
			return ProcessedImage{}, ErrImageFormatMismatch
		}
	}

	config, actualFormat, err := image.DecodeConfig(bytes.NewReader(source))
	if err != nil {
		return ProcessedImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	if actualFormat != extensionFormat {
		return ProcessedImage{}, ErrImageFormatMismatch
	}
	width, height := int64(config.Width), int64(config.Height)
	if width <= 0 || height <= 0 || width > p.limits.MaxPixels/height {
		return ProcessedImage{}, ErrImageDimensionsExceeded
	}

	var original []byte
	switch actualFormat {
	case "gif":
		frameCount, totalPixels, scanErr := scanGIFFrames(source)
		if scanErr != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, scanErr)
		}
		if frameCount > p.limits.MaxAnimationFrames {
			return ProcessedImage{}, ErrTooManyFrames
		}
		if totalPixels > p.limits.MaxAnimationPixels {
			return ProcessedImage{}, ErrAnimationTooLarge
		}
		animation, decodeErr := gif.DecodeAll(bytes.NewReader(source))
		if decodeErr != nil || len(animation.Image) != frameCount {
			if decodeErr == nil {
				decodeErr = errors.New("frame count changed during decode")
			}
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, decodeErr)
		}
		if options.WatermarkEnabled {
			for index, frame := range animation.Image {
				watermarked := applyWatermark(frame, options.WatermarkText, options.WatermarkDomain)
				palette := append(color.Palette{}, frame.Palette...)
				if len(palette) < 256 {
					palette = append(palette, color.Black, color.White)
				}
				paletted := image.NewPaletted(frame.Bounds(), palette)
				draw.Draw(paletted, frame.Bounds(), watermarked, frame.Bounds().Min, draw.Src)
				animation.Image[index] = paletted
			}
		}
		var encoded bytes.Buffer
		if encodeErr := gif.EncodeAll(&encoded, animation); encodeErr != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrImageProcessing, encodeErr)
		}
		original = encoded.Bytes()
	case "jpeg":
		decoded, decodeErr := jpeg.Decode(bytes.NewReader(source))
		if decodeErr != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, decodeErr)
		}
		if options.WatermarkEnabled {
			decoded = applyWatermark(decoded, options.WatermarkText, options.WatermarkDomain)
		}
		original, err = encodeImage(decoded, actualFormat)
		if err != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrImageProcessing, err)
		}
	case "png":
		decoded, decodeErr := png.Decode(bytes.NewReader(source))
		if decodeErr != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrInvalidImage, decodeErr)
		}
		if options.WatermarkEnabled {
			decoded = applyWatermark(decoded, options.WatermarkText, options.WatermarkDomain)
		}
		original, err = encodeImage(decoded, actualFormat)
		if err != nil {
			return ProcessedImage{}, fmt.Errorf("%w: %v", ErrImageProcessing, err)
		}
	default:
		return ProcessedImage{}, ErrUnsupportedImage
	}

	return ProcessedImage{
		Format: actualFormat, ContentType: contentTypeForFormat(actualFormat),
		Width: width, Height: height, Original: original,
	}, nil
}

var watermarkGlyphs = map[rune][]string{
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'C': {"01111", "10000", "10000", "10000", "10000", "10000", "01111"},
	'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'I': {"11111", "00100", "00100", "00100", "00100", "00100", "11111"},
	'J': {"00111", "00010", "00010", "00010", "10010", "10010", "01100"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'Q': {"01110", "10001", "10001", "10001", "10101", "10010", "01101"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00010", "00100", "01000", "11111"},
	'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "10000", "11110", "00001", "00001", "11110"},
	'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
	'.': {"00000", "00000", "00000", "00000", "00000", "00110", "00110"},
	'-': {"00000", "00000", "00000", "11111", "00000", "00000", "00000"},
	':': {"00000", "00110", "00110", "00000", "00110", "00110", "00000"},
	'/': {"00001", "00010", "00010", "00100", "01000", "01000", "10000"},
	'?': {"01110", "10001", "00001", "00010", "00100", "00000", "00100"},
}

func applyWatermark(source image.Image, text, domain string) *image.NRGBA {
	bounds := source.Bounds()
	watermarked := image.NewNRGBA(bounds)
	draw.Draw(watermarked, bounds, source, bounds.Min, draw.Src)

	text = normalizedWatermarkText(text, domain)
	width, height, scale := watermarkDimensions(text, bounds)
	padding := max(4, scale*2)
	margin := max(8, scale*4)
	boxWidth := width + padding*2
	boxHeight := height + padding*2
	x := bounds.Max.X - margin - boxWidth
	y := bounds.Max.Y - margin - boxHeight
	if x < bounds.Min.X+margin {
		x = bounds.Min.X + margin
	}
	if y < bounds.Min.Y+margin {
		y = bounds.Min.Y + margin
	}
	box := image.Rect(x, y, x+boxWidth, y+boxHeight)
	draw.Draw(watermarked, box, &image.Uniform{C: color.NRGBA{R: 0, G: 0, B: 0, A: 105}}, image.Point{}, draw.Over)
	drawWatermarkText(watermarked, text, x+padding, y+padding, scale)
	return watermarked
}

func normalizedWatermarkText(text, domain string) string {
	var normalized strings.Builder
	combined := strings.TrimSpace(text)
	if strings.TrimSpace(domain) != "" {
		if combined != "" {
			combined += " "
		}
		combined += strings.TrimSpace(domain)
	}
	for _, char := range strings.ToUpper(combined) {
		if char == ' ' || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '-' || char == ':' || char == '/' {
			normalized.WriteRune(char)
		}
		if normalized.Len() >= 24 {
			break
		}
	}
	if normalized.Len() == 0 {
		return "FASTIMG"
	}
	return normalized.String()
}

func watermarkDimensions(text string, bounds image.Rectangle) (int, int, int) {
	scale := max(1, min(bounds.Dx(), bounds.Dy())/180)
	width := 0
	for _, char := range text {
		if char == ' ' {
			width += 3 * scale
		} else {
			width += 5 * scale
		}
		width += scale
	}
	if width > 0 {
		width -= scale
	}
	return width, 7 * scale, scale
}

func drawWatermarkText(destination *image.NRGBA, text string, x, y, scale int) {
	for _, char := range text {
		if char == ' ' {
			x += 4 * scale
			continue
		}
		glyph, ok := watermarkGlyphs[char]
		if !ok {
			glyph = watermarkGlyphs['?']
		}
		for row, pattern := range glyph {
			for column, pixel := range pattern {
				if pixel != '1' {
					continue
				}
				rect := image.Rect(x+column*scale, y+row*scale, x+(column+1)*scale, y+(row+1)*scale)
				draw.Draw(destination, rect, &image.Uniform{C: color.NRGBA{R: 255, G: 255, B: 255, A: 220}}, image.Point{}, draw.Over)
			}
		}
		x += 6 * scale
	}
}

func formatForExtension(extension string) (string, bool) {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "jpeg", true
	case ".png":
		return "png", true
	case ".gif":
		return "gif", true
	default:
		return "", false
	}
}

func formatForMIME(mediaType string) (string, bool) {
	switch strings.ToLower(mediaType) {
	case "image/jpeg", "image/jpg":
		return "jpeg", true
	case "image/png":
		return "png", true
	case "image/gif":
		return "gif", true
	default:
		return "", false
	}
}

func contentTypeForFormat(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func encodeImage(source image.Image, format string) ([]byte, error) {
	var output bytes.Buffer
	switch format {
	case "jpeg":
		err := jpeg.Encode(&output, source, &jpeg.Options{Quality: 90})
		return output.Bytes(), err
	case "png":
		err := png.Encode(&output, source)
		return output.Bytes(), err
	case "gif":
		err := gif.Encode(&output, source, nil)
		return output.Bytes(), err
	default:
		return nil, ErrUnsupportedImage
	}
}

func scanGIFFrames(data []byte) (int, int64, error) {
	if len(data) < 13 || (string(data[:6]) != "GIF87a" && string(data[:6]) != "GIF89a") {
		return 0, 0, errors.New("invalid GIF header")
	}
	position := 13
	packed := data[10]
	if packed&0x80 != 0 {
		position += 3 * (1 << ((packed & 0x07) + 1))
	}
	if position > len(data) {
		return 0, 0, errors.New("truncated global color table")
	}

	frames := 0
	var totalPixels int64
	for position < len(data) {
		switch data[position] {
		case 0x3b:
			if frames == 0 {
				return 0, 0, errors.New("GIF contains no image frames")
			}
			return frames, totalPixels, nil
		case 0x21:
			if position+2 > len(data) {
				return 0, 0, errors.New("truncated GIF extension")
			}
			position += 2
			var err error
			position, err = skipGIFSubBlocks(data, position)
			if err != nil {
				return 0, 0, err
			}
		case 0x2c:
			if position+10 > len(data) {
				return 0, 0, errors.New("truncated GIF image descriptor")
			}
			frameWidth := int64(data[position+5]) | int64(data[position+6])<<8
			frameHeight := int64(data[position+7]) | int64(data[position+8])<<8
			if frameWidth == 0 || frameHeight == 0 || frameWidth > math.MaxInt64/frameHeight {
				return 0, 0, errors.New("invalid GIF frame dimensions")
			}
			totalPixels += frameWidth * frameHeight
			if totalPixels < 0 {
				return 0, 0, errors.New("GIF frame pixel count overflow")
			}
			frames++
			position += 10
			imagePacked := data[position-1]
			if imagePacked&0x80 != 0 {
				position += 3 * (1 << ((imagePacked & 0x07) + 1))
			}
			if position >= len(data) {
				return 0, 0, errors.New("truncated GIF image data")
			}
			position++ // LZW minimum code size
			var err error
			position, err = skipGIFSubBlocks(data, position)
			if err != nil {
				return 0, 0, err
			}
		default:
			return 0, 0, errors.New("unexpected GIF block marker")
		}
	}
	return 0, 0, errors.New("missing GIF trailer")
}

func skipGIFSubBlocks(data []byte, position int) (int, error) {
	for {
		if position >= len(data) {
			return 0, errors.New("truncated GIF sub-block")
		}
		length := int(data[position])
		position++
		if length == 0 {
			return position, nil
		}
		if length > len(data)-position {
			return 0, errors.New("truncated GIF sub-block payload")
		}
		position += length
	}
}
