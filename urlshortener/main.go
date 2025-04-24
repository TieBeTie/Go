//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// URLEntry represents a URL and its shortened key
type URLEntry struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// URLShortener manages the URL shortening service
type URLShortener struct {
	mu         sync.Mutex
	urlToKey   map[string]string
	keyToURL   map[string]string
	keyLength  int
	keyCharset []rune
}

// NewURLShortener creates a new URL shortener service
func NewURLShortener() *URLShortener {
	return &URLShortener{
		urlToKey:   make(map[string]string),
		keyToURL:   make(map[string]string),
		keyLength:  8,
		keyCharset: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	}
}

// generateKey creates a random key of configured length
func (us *URLShortener) generateKey() string {
	b := make([]rune, us.keyLength)
	for i := range b {
		b[i] = us.keyCharset[rand.Intn(len(us.keyCharset))]
	}
	return string(b)
}

// ShortenHandler processes POST /shorten requests
func (us *URLShortener) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		URL string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if we already have this URL
	key, exists := us.urlToKey[requestBody.URL]
	if !exists {
		// Generate a new key
		key = us.generateKey()
		us.urlToKey[requestBody.URL] = key
		us.keyToURL[key] = requestBody.URL
	}

	// Prepare response
	response := URLEntry{
		URL: requestBody.URL,
		Key: key,
	}

	// Set content type and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// RedirectHandler processes GET /go/<KEY> requests
func (us *URLShortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract key from path
	key := r.URL.Path[len("/go/"):]

	us.mu.Lock()
	url, exists := us.keyToURL[key]
	us.mu.Unlock()

	if !exists {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	// Redirect to the original URL
	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create a new URL shortener
	shortener := NewURLShortener()

	// Set up HTTP routes
	http.HandleFunc("/shorten", shortener.ShortenHandler)
	http.HandleFunc("/go/", shortener.RedirectHandler)

	// Start the server
	serverAddr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting server on %s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
