package file

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/nfnt/resize"
)

// EncodeToBase64 encodes an image file to a base64 string after resizing it.
func EncodeToBase64(filePath string) (string, error) {
	imgFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return "", err
	}

	newImage := resize.Resize(800, 0, img, resize.Lanczos3)

	buf := new(bytes.Buffer)
	if err = jpeg.Encode(buf, newImage, &jpeg.Options{Quality: 80}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// ReadFile reads the content of the given file path and returns it as a byte slice.
func ReadFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

// IsHumanReadable checks if the file content is likely to be text based on MIME type analysis.
func IsHumanReadable(filePath string) bool {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}
	return isText(content) && isPrintableText(content)
}

// IsImage checks if the file extension indicates an image.
func IsImage(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff":
		return true
	default:
		return false
	}
}

// isText determines if the content's MIME type is textual.
func isText(data []byte) bool {
	mimeType := http.DetectContentType(data[:512])
	return strings.HasPrefix(mimeType, "text") || strings.Contains(mimeType, "charset=utf-8")
}

// isPrintableText checks if a significant portion of the content consists of printable characters.
func isPrintableText(data []byte) bool {
	printableCount := 0
	for _, b := range data {
		if unicode.IsPrint(rune(b)) || unicode.IsSpace(rune(b)) {
			printableCount++
		}
	}
	return float64(printableCount)/float64(len(data)) > 0.9
}
