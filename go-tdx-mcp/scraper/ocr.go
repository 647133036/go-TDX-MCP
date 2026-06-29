package scraper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// OCRResult holds the output from Tesseract OCR.
type OCRResult struct {
	Text       string
	Confidence float64
	Lang       string
	Duration   time.Duration
}

// OCRClient wraps Tesseract OCR for image/text extraction.
type OCRClient struct {
	languages []string
	timeout   time.Duration
	preproc   *ImagePreprocessor
}

// NewOCRClient creates a Tesseract OCR client.
// languages: e.g., ["chi_sim", "eng"] for Chinese + English.
func NewOCRClient(languages ...string) *OCRClient {
	if len(languages) == 0 {
		languages = []string{"chi_sim", "eng"}
	}
	return &OCRClient{
		languages: languages,
		timeout:   30 * time.Second,
		preproc:   NewImagePreprocessor(),
	}
}

// WithPreprocessOptions sets preprocessing options for OCR.
func (o *OCRClient) WithPreprocessOptions(opts PreprocessOptions) *OCRClient {
	o.preproc = NewImagePreprocessor()
	return o
}

// Recognize extracts text from an image file using Tesseract.
// Supported formats: PNG, JPG, BMP, TIFF, GIF.
func (o *OCRClient) Recognize(imagePath string) (*OCRResult, error) {
	start := time.Now()

	cmd := exec.Command("tesseract", imagePath, "-", "-l", strings.Join(o.languages, "+"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("tesseract not installed: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("tesseract failed: %w, stderr: %s", err, stderr.String())
		}
	case <-time.After(o.timeout):
		cmd.Process.Kill()
		return nil, fmt.Errorf("tesseract timed out after %v", o.timeout)
	}

	text := strings.TrimSpace(stdout.String())
	confidence := o.parseConfidence(text)

	return &OCRResult{
		Text:       text,
		Confidence: confidence,
		Lang:       strings.Join(o.languages, ","),
		Duration:   time.Since(start),
	}, nil
}

// RecognizeFromBytes extracts text from image bytes (PNG/JPG/BMP/TIFF).
// Applies automatic preprocessing: grayscale, denoise, threshold, scale-up.
func (o *OCRClient) RecognizeFromBytes(data []byte) (*OCRResult, error) {
	opts := DefaultPreprocessOptions()
	preprocessed, err := o.preproc.PreprocessFromBytes(data, opts)
	if err != nil {
		// Fall back to raw bytes if preprocessing fails
		preprocessed = data
	}

	tmpFile := fmt.Sprintf("/tmp/ocr_%d.png", time.Now().UnixNano())
	defer func() {
		os.Remove(tmpFile)
	}()

	if err := os.WriteFile(tmpFile, preprocessed, 0644); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}

	return o.Recognize(tmpFile)
}

// RecognizeBase64 extracts text from base64-encoded image data.
func (o *OCRClient) RecognizeBase64(base64Data string) (*OCRResult, error) {
	tmpFile := fmt.Sprintf("/tmp/ocr_%d.png", time.Now().UnixNano())
	defer os.Remove(tmpFile)

	// Decode base64 and write to temp file
	decoded, err := decodeBase64Image(base64Data)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	if err := os.WriteFile(tmpFile, decoded, 0644); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}

	return o.Recognize(tmpFile)
}

func decodeBase64Image(data string) ([]byte, error) {
	// Strip data URI prefix if present
	data = strings.TrimPrefix(data, "data:image/png;base64,")
	data = strings.TrimPrefix(data, "data:image/jpeg;base64,")
	data = strings.TrimPrefix(data, "data:image/jpg;base64,")
	data = strings.TrimSpace(data)

	buf := make([]byte, len(data)/4*3+10)
	n, err := decodeBase64To(buf, data)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func decodeBase64To(buf []byte, src string) (int, error) {
	// Simple base64 decoder for ASCII data
	dst := buf[:0]
	state := 0
	var val uint32
	for _, c := range src {
		if c == '=' || c == '\n' || c == '\r' || c == ' ' {
			continue
		}
		var v byte
		switch {
		case c >= 'A' && c <= 'Z':
			v = byte(c - 'A')
		case c >= 'a' && c <= 'z':
			v = byte(c - 'a' + 26)
		case c >= '0' && c <= '9':
			v = byte(c - '0' + 52)
		case c == '+':
			v = 62
		case c == '/':
			v = 63
		default:
			continue
		}
		val = val<<6 | uint32(v)
		state++
		if state == 4 {
			dst = append(dst, byte(val>>16), byte(val>>8), byte(val))
			state = 0
			val = 0
		}
	}
	return len(dst), nil
}

// parseConfidence extracts confidence score from Tesseract output.
func (o *OCRClient) parseConfidence(text string) float64 {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Confidence:") {
			var conf float64
			fmt.Sscanf(line, "Confidence: %f", &conf)
			return conf
		}
	}
	return 0
}

// IsAvailable checks if Tesseract is installed and functional.
func (o *OCRClient) IsAvailable() bool {
	cmd := exec.Command("tesseract", "--version")
	return cmd.Run() == nil
}
