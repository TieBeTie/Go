//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Now(), nil
	}

	// Check if the time is in the format hh:mm:ss with regex
	// This ensures exactly 2 digits for hours, minutes, and seconds
	validFormat := regexp.MustCompile(`^[0-9]{2}:[0-9]{2}:[0-9]{2}$`)
	if !validFormat.MatchString(timeStr) {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	// Split the time string
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	// Parse each part
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return time.Time{}, fmt.Errorf("invalid hour")
	}

	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return time.Time{}, fmt.Errorf("invalid minute")
	}

	s, err := strconv.Atoi(parts[2])
	if err != nil || s < 0 || s > 59 {
		return time.Time{}, fmt.Errorf("invalid second")
	}

	// Set a fixed base date, only time matters
	return time.Date(2020, 1, 1, h, m, s, 0, time.UTC), nil
}

func parseK(kStr string) (int, error) {
	if kStr == "" {
		return 1, nil // Default value
	}

	k, err := strconv.Atoi(kStr)
	if err != nil {
		return 0, fmt.Errorf("invalid k")
	}

	if k < 1 || k > 30 {
		return 0, fmt.Errorf("invalid k")
	}

	return k, nil
}

// Function to get the width of a symbol
func symbolWidth(s string) int {
	return len(strings.SplitN(s, "\n", 2)[0])
}

func getSymbolHeight(s string) int {
	return len(strings.Split(s, "\n"))
}

func getDigitSymbol(digit int) string {
	switch digit {
	case 0:
		return Zero
	case 1:
		return One
	case 2:
		return Two
	case 3:
		return Three
	case 4:
		return Four
	case 5:
		return Five
	case 6:
		return Six
	case 7:
		return Seven
	case 8:
		return Eight
	case 9:
		return Nine
	default:
		return ""
	}
}

func drawSymbol(img *image.RGBA, symbol string, x, y, k int) {
	lines := strings.Split(symbol, "\n")
	for dy, line := range lines {
		for dx, c := range line {
			var pixelColor color.Color
			if c == '1' {
				pixelColor = Cyan
			} else {
				pixelColor = color.White
			}

			// Draw k*k pixels for each symbol pixel
			for i := 0; i < k; i++ {
				for j := 0; j < k; j++ {
					img.Set(x+dx*k+i, y+dy*k+j, pixelColor)
				}
			}
		}
	}
}

func drawDigitalClock(t time.Time, k int) *image.RGBA {
	digitWidth := symbolWidth(Zero)
	colonWidth := symbolWidth(Colon)
	height := getSymbolHeight(Zero)

	// Calculate total width: 6 digits + 2 colons
	totalWidth := digitWidth*6 + colonWidth*2

	// Create image with calculated dimensions
	img := image.NewRGBA(image.Rect(0, 0, totalWidth*k, height*k))

	// Fill the entire image with white background
	for x := 0; x < totalWidth*k; x++ {
		for y := 0; y < height*k; y++ {
			img.Set(x, y, color.White)
		}
	}

	// Format time to extract digits
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	// Draw hours
	h1 := hour / 10
	h2 := hour % 10
	drawSymbol(img, getDigitSymbol(h1), 0, 0, k)
	drawSymbol(img, getDigitSymbol(h2), digitWidth*k, 0, k)

	// Draw first colon
	drawSymbol(img, Colon, 2*digitWidth*k, 0, k)

	// Draw minutes
	m1 := minute / 10
	m2 := minute % 10
	drawSymbol(img, getDigitSymbol(m1), (2*digitWidth+colonWidth)*k, 0, k)
	drawSymbol(img, getDigitSymbol(m2), (3*digitWidth+colonWidth)*k, 0, k)

	// Draw second colon
	drawSymbol(img, Colon, (4*digitWidth+colonWidth)*k, 0, k)

	// Draw seconds
	s1 := second / 10
	s2 := second % 10
	drawSymbol(img, getDigitSymbol(s1), (4*digitWidth+2*colonWidth)*k, 0, k)
	drawSymbol(img, getDigitSymbol(s2), (5*digitWidth+2*colonWidth)*k, 0, k)

	return img
}

func handleTimeRequest(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	timeStr := query.Get("time")
	kStr := query.Get("k")

	// Parse time parameter
	t, err := parseTime(timeStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse k parameter
	k, err := parseK(kStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate the digital clock image
	img := drawDigitalClock(t, k)

	// Set content type header
	w.Header().Set("Content-Type", "image/png")

	// Encode the image as PNG and send it
	if err := png.Encode(w, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Parse command line flags
	portPtr := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Set up HTTP server
	http.HandleFunc("/", handleTimeRequest)

	// Start server
	port := fmt.Sprintf(":%d", *portPtr)
	log.Printf("Starting digital clock server on port %d", *portPtr)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
