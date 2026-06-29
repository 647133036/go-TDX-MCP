package scraper

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
)

// ImagePreprocessor handles image preprocessing for OCR enhancement.
// Pure Go implementation using image package - no OpenCV dependency.
type ImagePreprocessor struct{}

// NewImagePreprocessor creates a new image preprocessor.
func NewImagePreprocessor() *ImagePreprocessor {
	return &ImagePreprocessor{}
}

// PreprocessOptions controls the preprocessing pipeline.
type PreprocessOptions struct {
	Grayscale    bool // convert to grayscale
	Threshold    int  // binary threshold (0-255), 0=disabled
	Denoise      bool // median noise reduction
	Invert       bool // invert colors (black<->white)
	ScaleUp      int  // scale factor (1=none, 2=2x, 4=4x)
	Deskew       bool // auto deskew detection
	Contrast     float64 // contrast enhancement (1.0=none, >1.0=enhance)
}

// DefaultPreprocessOptions returns settings optimized for financial chart OCR.
func DefaultPreprocessOptions() PreprocessOptions {
	return PreprocessOptions{
		Grayscale: true,
		Threshold: 128,
		Denoise:   true,
		Invert:    false,
		ScaleUp:   2,
		Deskew:    false,
		Contrast:  1.5,
	}
}

// PreprocessFromFile reads an image, preprocesses it, and saves the result.
func (p *ImagePreprocessor) PreprocessFromFile(inputPath, outputPath string, opts PreprocessOptions) error {
	img, err := loadImage(inputPath)
	if err != nil {
		return err
	}
	return p.saveProcessed(img, outputPath, opts)
}

// PreprocessFromBytes reads image bytes, preprocesses, and returns result bytes.
func (p *ImagePreprocessor) PreprocessFromBytes(data []byte, opts PreprocessOptions) ([]byte, error) {
	img, err := loadImageBytes(data)
	if err != nil {
		return nil, err
	}
	return p.processToBytes(img, opts)
}

// PreprocessAndSave preprocesses an image in-place (saves over original).
func (p *ImagePreprocessor) PreprocessAndSave(path string, opts PreprocessOptions) error {
	img, err := loadImage(path)
	if err != nil {
		return err
	}
	return p.saveProcessed(img, path, opts)
}

// prepareImage runs the full preprocessing pipeline.
func (p *ImagePreprocessor) prepareImage(img image.Image, opts PreprocessOptions) image.Image {
	result := img

	// Step 1: Grayscale
	if opts.Grayscale {
		result = toGrayscale(result)
	}

	// Step 2: Contrast enhancement
	if opts.Contrast != 1.0 {
		result = enhanceContrast(result, opts.Contrast)
	}

	// Step 3: Denoise
	if opts.Denoise {
		result = medianFilter(result, 3)
	}

	// Step 4: Threshold (binary)
	if opts.Threshold > 0 {
		result = binaryThreshold(result, opts.Threshold)
	}

	// Step 5: Invert
	if opts.Invert {
		result = invertColors(result)
	}

	// Step 6: Scale up
	if opts.ScaleUp > 1 {
		result = scaleImage(result, opts.ScaleUp)
	}

	return result
}

func (p *ImagePreprocessor) saveProcessed(img image.Image, outputPath string, opts PreprocessOptions) error {
	result := p.prepareImage(img, opts)

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, result)
}

func (p *ImagePreprocessor) processToBytes(img image.Image, opts PreprocessOptions) ([]byte, error) {
	result := p.prepareImage(img, opts)

	var buf os.FileInfo
	_ = buf // placeholder

	f, err := os.CreateTemp("", "ocr_preprocess_*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())

	if err := png.Encode(f, result); err != nil {
		return nil, err
	}
	f.Close()

	data, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}
	return data, nil
}

// toGrayscale converts any image to grayscale RGBA.
func toGrayscale(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)

			// ITU-R BT.601 luma coefficients
			v := uint8(0.299*float64(rr) + 0.587*float64(gg) + 0.114*float64(bb))
			dst.Set(x, y, color.Gray{Y: v})
		}
	}
	return dst
}

// enhanceContrast applies contrast enhancement using linear scaling.
func enhanceContrast(src image.Image, factor float64) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := src.At(x, y)
			r, g, b, a := p.RGBA()
			rr := float64(uint8(r >> 8))
			gg := float64(uint8(g >> 8))
			bb := float64(uint8(b >> 8))

			// Center around 128, scale, clamp
			rr = math.Min(255, math.Max(0, 128+(rr-128)*factor))
			gg = math.Min(255, math.Max(0, 128+(gg-128)*factor))
			bb = math.Min(255, math.Max(0, 128+(bb-128)*factor))

			dst.Set(x, y, color.RGBA{
				R: uint8(rr),
				G: uint8(gg),
				B: uint8(bb),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// binaryThreshold converts image to black-and-white using threshold.
func binaryThreshold(src image.Image, thresh int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := src.At(x, y)
			r, g, b, _ := p.RGBA()
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)

			// Luminance
			lum := uint8(0.299*float64(rr) + 0.587*float64(gg) + 0.114*float64(bb))
			c := lum
			if c >= uint8(thresh) {
				c = 255
			} else {
				c = 0
			}
			dst.Set(x, y, color.Gray{Y: c})
		}
	}
	return dst
}

// invertColors inverts all pixel colors.
func invertColors(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := src.At(x, y)
			r, g, b, a := p.RGBA()
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)

			dst.Set(x, y, color.RGBA{
				R: ^rr,
				G: ^gg,
				B: ^bb,
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// medianFilter applies a 3x3 median filter for noise reduction.
func medianFilter(src image.Image, kernelSize int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	half := kernelSize / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < half || x >= bounds.Max.X-half || y < half || y >= bounds.Max.Y-half {
				dst.Set(x, y, src.At(x, y))
				continue
			}

			values := make([]uint8, 0, kernelSize*kernelSize)
			for dy := -half; dy <= half; dy++ {
				for dx := -half; dx <= half; dx++ {
					p := src.At(x+dx, y+dy)
					r, _, _, _ := p.RGBA()
					values = append(values, uint8(r>>8))
				}
			}

			// Sort and find median
			for i := 0; i < len(values); i++ {
				for j := i + 1; j < len(values); j++ {
					if values[i] > values[j] {
						values[i], values[j] = values[j], values[i]
					}
				}
			}
			median := values[len(values)/2]
			dst.Set(x, y, color.Gray{Y: median})
		}
	}
	return dst
}

// scaleImage scales an image by the given factor using nearest-neighbor.
func scaleImage(src image.Image, factor int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx() * factor
	h := bounds.Dy() * factor
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			p := src.At(x, y)
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					dst.Set(x*factor+dx, y*factor+dy, p)
				}
			}
		}
	}
	return dst
}

// loadImage loads an image from a file (PNG or JPEG).
func loadImage(path string) (image.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadImageBytes(data)
}

// loadImageBytes loads an image from bytes (PNG or JPEG).
func loadImageBytes(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	src, _, err := image.Decode(reader)
	if err != nil {
		reader.Seek(0, io.SeekStart)
		src, err = jpeg.Decode(reader)
		if err != nil {
			return nil, err
		}
	}
	return src, nil
}

// DeskewAngle estimates the skew angle of a binarized image using Hough-like approach.
// Returns angle in radians (positive = counter-clockwise).
func DeskewAngle(src image.Image) float64 {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Collect edge points (black pixels in binarized image)
	edges := make([]image.Point, 0, width*height/10)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := src.At(x, y)
			r, _, _, _ := p.RGBA()
			if uint8(r>>8) < 128 {
				edges = append(edges, image.Point{x, y})
			}
		}
	}

	if len(edges) < 100 {
		return 0
	}

	// Simple PCA-based skew estimation
	var cx, cy float64
	for _, e := range edges {
		cx += float64(e.X)
		cy += float64(e.Y)
	}
	cx /= float64(len(edges))
	cy /= float64(len(edges))

	var covXX, covXY, covYY float64
	for _, e := range edges {
		dx := float64(e.X) - cx
		dy := float64(e.Y) - cy
		covXX += dx * dx
		covXY += dx * dy
		covYY += dy * dy
	}

	// Eigenvector of covariance matrix gives principal axis
	det := covXX*covYY - covXY*covXY
	if det == 0 {
		return 0
	}

	eigenVal := (covXX + covYY) / 2 + math.Sqrt(((covXX-covYY)/2)*(covXX-covYY)+covXY*covXY)
	eigenVecX := covXY
	eigenVecY := eigenVal - covXX

	if math.Abs(eigenVecX) < 1e-10 {
		return 0
	}

	return math.Atan2(eigenVecY, eigenVecX)
}

// CreateTestImage creates a simple test image with text-like patterns.
func CreateTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill white background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}
	// Draw some black rectangles to simulate text
	for y := 10; y < 30; y++ {
		for x := 10; x < 60; x++ {
			img.Set(x, y, color.Black)
		}
	}
	for y := 10; y < 20; y++ {
		for x := 70; x < 90; x++ {
			img.Set(x, y, color.Black)
		}
	}
	// Encode as PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
