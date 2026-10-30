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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/image/webp"
)

//go:embed assets/default_avatar.png assets/badges/*.svg
var assets embed.FS

const (
	baseURL       = "https://fossunited.org"
	fetchTimeout  = 5 * time.Second
	maxImageBytes = 5 << 20 // 5MB
)

// httpClient never follows redirects, so a redirect can't be used to
// bypass isSafeHost's check of the original URL.
var httpClient = &http.Client{
	Timeout: fetchTimeout,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var defaultAvatarDataURI string

// badgeAssets maps a template param name (e.g. "badge_fosshack_winner",
// "badge_forklore") to the base64 data URI of the embedded badge sticker.
// Populated from assets/badges/*.svg at startup.
var badgeAssets = map[string]string{}

func init() {
	data, err := assets.ReadFile("assets/default_avatar.png")
	if err != nil {
		log.Fatal("failed to load embedded default avatar: " + err.Error())
	}
	defaultAvatarDataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)

	entries, err := assets.ReadDir("assets/badges")
	if err != nil {
		log.Fatal("failed to list embedded badges: " + err.Error())
	}
	for _, entry := range entries {
		svg, err := assets.ReadFile("assets/badges/" + entry.Name())
		if err != nil {
			log.Fatal("failed to load embedded badge " + entry.Name() + ": " + err.Error())
		}
		// "fosshack-winner.svg" -> "badge_fosshack_winner"
		name := strings.TrimSuffix(entry.Name(), ".svg")
		key := "badge_" + strings.ReplaceAll(name, "-", "_")
		badgeAssets[key] = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)
	}
}

func DefaultAvatar() string {
	return defaultAvatarDataURI
}

// InjectBadgeAssets adds the embedded badge sticker data URIs to the params
// map (as badge_* keys) so templates can reference them. Keys unused by a
// given template are harmless.
func InjectBadgeAssets(params map[string]string) {
	for key, uri := range badgeAssets {
		params[key] = uri
	}
}

func isInvalidPath(path string) bool {
	p := strings.TrimSpace(path)
	return p == "" || strings.EqualFold(p, "none")
}

// resolveURL turns a param value into the URL to fetch: an absolute URL
// is used as-is, anything else is treated as a path relative to
// fossunited.org (the historical behavior).
func resolveURL(path string) string {
	if u, err := url.Parse(path); err == nil && u.IsAbs() {
		return path
	}
	return baseURL + "/" + strings.TrimPrefix(path, "/")
}

// isSafeHost blocks loopback/private/link-local/unspecified addresses so
// this can't be used to probe the server's own internal network (SSRF).
func isSafeHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}

// FetchAndEncode downloads an image -- either from an absolute URL, or
// from the given path on fossunited.org -- converts it to PNG, and
// returns a base64 data URI.
func FetchAndEncode(path string) (string, error) {
	if isInvalidPath(path) {
		return DefaultAvatar(), nil
	}

	target := resolveURL(path)
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil || !isSafeHost(req.URL.Host) {
		return "", fmt.Errorf("fetch %s: not allowed", target)
	}

	resp, err := httpClient.Do(req)
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
