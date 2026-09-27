package media

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageProcessorRejectsDeclaredFormatMismatch(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 8, 8))))
	processor := NewImageProcessor(ImageLimits{})

	_, err := processor.Process("photo.jpg", "image/jpeg", encoded.Bytes())
	require.ErrorIs(t, err, ErrImageFormatMismatch)
}

func TestImageProcessorRejectsImagesOverPixelLimit(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 8, 8))))
	processor := NewImageProcessor(ImageLimits{MaxPixels: 32})

	_, err := processor.Process("photo.png", "image/png", encoded.Bytes())
	require.ErrorIs(t, err, ErrImageDimensionsExceeded)
}

func TestImageProcessorStripsJPEGMetadataAndStoresOnlyOriginal(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1600, 900))
	for y := 0; y < 900; y++ {
		for x := 0; x < 1600; x++ {
			source.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 80, A: 255})
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 92}))
	withExif := insertJPEGApplicationSegment(encoded.Bytes(), []byte("Exif\x00\x00FASTIMG-PRIVATE-METADATA"))
	processor := NewImageProcessor(ImageLimits{})

	result, err := processor.Process("photo.jpg", "image/jpeg", withExif)
	require.NoError(t, err)
	require.Equal(t, "jpeg", result.Format)
	require.Equal(t, int64(1600), result.Width)
	require.Equal(t, int64(900), result.Height)
	require.NotContains(t, string(result.Original), "FASTIMG-PRIVATE-METADATA")

	original, _, err := image.Decode(bytes.NewReader(result.Original))
	require.NoError(t, err)
	require.Equal(t, 1600, original.Bounds().Dx())
	require.Equal(t, 900, original.Bounds().Dy())
}

func TestImageProcessorAppliesPlanWatermarkToOriginal(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 320, 180))
	for y := 0; y < 180; y++ {
		for x := 0; x < 320; x++ {
			source.Set(x, y, color.RGBA{R: 30, G: 90, B: 150, A: 255})
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, source))
	processor := NewImageProcessor(ImageLimits{})

	plain, err := processor.Process("photo.png", "image/png", encoded.Bytes())
	require.NoError(t, err)
	watermarked, err := processor.ProcessWithOptions("photo.png", "image/png", encoded.Bytes(), ProcessOptions{
		WatermarkEnabled: true,
		WatermarkText:    "FastImg",
	})
	require.NoError(t, err)
	require.NotEqual(t, plain.Original, watermarked.Original)
}

func TestImageProcessorIncludesConfiguredDomainInWatermark(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 320, 180))
	for y := 0; y < 180; y++ {
		for x := 0; x < 320; x++ {
			source.Set(x, y, color.RGBA{R: 30, G: 90, B: 150, A: 255})
		}
	}
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, source))
	processor := NewImageProcessor(ImageLimits{})

	withoutDomain, err := processor.ProcessWithOptions("photo.png", "image/png", encoded.Bytes(), ProcessOptions{
		WatermarkEnabled: true,
		WatermarkText:    "FastImg",
	})
	require.NoError(t, err)
	withDomain, err := processor.ProcessWithOptions("photo.png", "image/png", encoded.Bytes(), ProcessOptions{
		WatermarkEnabled: true,
		WatermarkText:    "FastImg",
		WatermarkDomain:  "img.example.com",
	})
	require.NoError(t, err)
	require.NotEqual(t, withoutDomain.Original, withDomain.Original)
}

func insertJPEGApplicationSegment(jpegBytes, payload []byte) []byte {
	segment := []byte{0xff, 0xe1, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}
	segment = append(segment, payload...)
	result := append([]byte{}, jpegBytes[:2]...)
	result = append(result, segment...)
	return append(result, jpegBytes[2:]...)
}
