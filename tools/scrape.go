package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

func main() {
	// Parse command-line arguments
	urlFlag := flag.String("u", "", "URL to scrape")
	flag.Parse()

	// Ensure the URL flag is provided
	if *urlFlag == "" {
		log.Fatal("URL is required")
	}

	// Remove leading/trailing spaces from the URL
	urlInput := strings.TrimSpace(*urlFlag)

	// Parse the main URL
	mainURL, err := url.Parse(urlInput)
	if err != nil {
		log.Fatal("Failed to parse URL:", err)
	}

	// Make HTTP GET request
	response, err := http.Get(mainURL.String())
	if err != nil {
		log.Fatal("Failed to fetch the URL:", err)
	}
	defer response.Body.Close()

	// Parse the HTML body
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Failed to parse HTML:", err)
	}

	// Find all anchor tags
	visitedURLs := make(map[string]bool) // Map to store visited URLs
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			// Normalize the URL
			normalizedURL := normalizeURL(href, mainURL)

			// Check if the URL is valid and not excluded
			if isValidURL(normalizedURL, mainURL) {
				// Add URL to the visited URLs map
				visitedURLs[normalizedURL] = true
			}
		}
	})

	// Extract the domain name from the main URL
	domainName := mainURL.Hostname()

	// Create or open the file to save URLs
	file, err := os.Create(domainName + ".txt")
	if err != nil {
		log.Fatal("Failed to create file:", err)
	}
	defer file.Close()

	// Write the main URL to the file
	_, err = file.WriteString(mainURL.String() + "\n")
	if err != nil {
		log.Fatal("Failed to write URL to file:", err)
	}

	// Write unique URLs to the file
	counter := 1 // Start counter from 1 for the main URL
	for u := range visitedURLs {
		// Clean up URL by removing leading/trailing spaces
		u = strings.TrimSpace(u)

		// Exclude the main URL from being written again
		if u != mainURL.String() {
			// Write URL to the file
			_, err := file.WriteString(u + "\n")
			if err != nil {
				log.Fatal("Failed to write URL to file:", err)
			}
			counter++
		}
	}

	// Print the result
	if counter > 0 {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Printf("%s ", cyan("[Console] =>"))
		fmt.Printf("%d URLs were found and saved to %s.txt\n", counter, domainName)
	} else {
		fmt.Println("No URLs were found to save.")
	}
}

// Normalize the URL by handling relative URLs and removing unnecessary components
func normalizeURL(u string, baseURL *url.URL) string {
	// Remove leading/trailing spaces
	u = strings.TrimSpace(u)

	// Handle relative URLs
	if !strings.HasPrefix(u, "http") {
		u = baseURL.Scheme + "://" + baseURL.Host + u
	}

	// Remove the trailing slash
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}

	// Parse the URL
	parsedURL, err := url.Parse(u)
	if err != nil {
		return u
	}

	// Remove unnecessary components from the URL
	parsedURL.Fragment = ""
	parsedURL.RawQuery = ""

	return parsedURL.String()
}

// Check if the URL is valid and not excluded
func isValidURL(u string, mainURL *url.URL) bool {
	u = strings.TrimSpace(u)
	if u == "#" || u == "" || u == mainURL.String() {
		return false
	}

	// Remove trailing slash from the URLs
	u = removeTrailingSlash(u)

	// Parse the URL
	parsedURL, err := url.Parse(u)
	if err != nil {
		return false
	}

	// Exclude URLs with different hosts or subdomains
	if parsedURL.Hostname() != mainURL.Hostname() && !strings.HasSuffix(parsedURL.Hostname(), "."+mainURL.Hostname()) {
		return false
	}

	// Exclude anchor links within the same page
	if parsedURL.Path == mainURL.Path && parsedURL.Fragment != "" {
		return false
	}

	// Exclude URLs with file extensions
	ext := filepath.Ext(parsedURL.Path)
	if ext != "" {
		return false
	}

	return true
}


// Remove trailing slash from the URL
func removeTrailingSlash(u string) string {
	if strings.HasSuffix(u, "/") {
		u = u[:len(u)-1]
	}
	return u
}
