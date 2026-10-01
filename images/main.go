package images

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/image/webp"
)

//go:embed assets/default_avatar.png
var assets embed.FS

const (
	baseURL       = "https://fossunited.org"
	fetchTimeout  = 5 * time.Second
	maxImageBytes = 5 << 20 // 5MB
)

var defaultAvatarDataURI string

func init() {
	data, err := assets.ReadFile("assets/default_avatar.png")
	if err != nil {
		log.Fatal("failed to load embedded default avatar: " + err.Error())
	}
	defaultAvatarDataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
}

func DefaultAvatar() string {
	return defaultAvatarDataURI
}

func isInvalidPath(path string) bool {
	p := strings.TrimSpace(path)
	return p == "" || strings.EqualFold(p, "none")
}

// resolveURL turns a param value into the fossunited.org URL to fetch. A
// relative path is appended to baseURL (the historical behavior); an
// absolute URL is used as-is only if it already points at fossunited.org,
// so this never fetches an arbitrary caller-supplied host.
func resolveURL(path string) (string, error) {
	u, err := url.Parse(path)
	if err == nil && u.IsAbs() {
		if u.Scheme != "http" && u.Scheme != "https" || !strings.EqualFold(u.Hostname(), "fossunited.org") {
			return "", fmt.Errorf("only fossunited.org URLs are allowed, got %q", path)
		}
		return path, nil
	}
	return baseURL + "/" + strings.TrimPrefix(path, "/"), nil
}

// FetchAndEncode downloads an image from fossunited.org at the given path
// (or a full fossunited.org URL), converts it to PNG, and returns a
// base64 data URI.
func FetchAndEncode(path string) (string, error) {
	if isInvalidPath(path) {
		return DefaultAvatar(), nil
	}

	target, err := resolveURL(path)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(target)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", target, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", target, err)
	}

	pngData, err := toPNG(body, resp.Header.Get("Content-Type"), path)
	if err != nil {
		return "", fmt.Errorf("convert %s: %w", target, err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData), nil
}

func toPNG(data []byte, contentType string, path string) ([]byte, error) {
	isWebP := strings.Contains(contentType, "webp") || strings.HasSuffix(strings.ToLower(path), ".webp")

	var img image.Image
	var err error

	if isWebP {
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			// webp decode failed, try generic decoder as fallback
			img, _, err = image.Decode(bytes.NewReader(data))
		}
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// ProcessImageParams scans the params map for keys ending in "_image",
// fetches each image, and replaces the value with a base64 data URI.
func ProcessImageParams(params map[string]string) {
	for key, val := range params {
		if !strings.HasSuffix(key, "_image") {
			continue
		}

		dataURI, err := FetchAndEncode(val)
		if err != nil {
			log.Printf("image fetch failed for %s=%s: %v (using default avatar)", key, val, err)
			dataURI = DefaultAvatar()
		}
		params[key] = dataURI
	}
}
