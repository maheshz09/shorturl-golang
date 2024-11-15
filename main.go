package main

import (
	"github.com/gin-gonic/gin"
	"log" // Import the log package
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	urls     = make(map[string]string) // Map to store URL mapping
	baseURL  = "http://localhost:8080"  // Base URL for short links
	urlMutex sync.Mutex                 // Mutex to handle concurrency
	urlLength = 6                        // Length of the generated short URL
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Function to generate a random short URL (mix of numbers and alphabets)
func generateShortURL() string {
	rand.Seed(time.Now().UnixNano())
	shortURL := make([]byte, urlLength)
	for i := range shortURL {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

// Function to handle creating a short URL
func createShortURL(c *gin.Context) {
	var input struct {
		LongURL string `json:"long_url"`
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	longURL := input.LongURL
	if len(longURL) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Long URL cannot be empty"})
		return
	}

	// Generate the short URL code
	shortURL := generateShortURL()

	// Lock for concurrency safety
	urlMutex.Lock()
	urls[shortURL] = longURL
	urlMutex.Unlock()

	// Respond with the short URL
	c.JSON(http.StatusOK, gin.H{
		"short_url": baseURL + "/" + shortURL,
	})
}

// Function to handle URL redirection
func redirectToLongURL(c *gin.Context) {
	shortURL := c.Param("shortURL")

	// Check if the short URL exists in the mapping
	urlMutex.Lock()
	longURL, exists := urls[shortURL]
	urlMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	// Redirect to the long URL
	c.Redirect(http.StatusFound, longURL)
}

// Serve static files (HTML, JS, CSS)
func serveFrontend(r *gin.Engine) {
	r.StaticFile("/", "./static/index.html") // Serve index.html at the root URL

	// Serve static assets like CSS, JS from the 'public' directory
	r.Static("/static", "./public")
}

func main() {
	// Create the gin router
	r := gin.Default()

	// Serve the frontend
	serveFrontend(r)

	// Endpoint to create a short URL
	r.POST("/shorten", createShortURL)

	// Endpoint to redirect a short URL to its long URL
	r.GET("/:shortURL", redirectToLongURL)

	// Start the web server
	err := r.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
}
