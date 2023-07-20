package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/PuerkitoBio/goquery"
)

const (
	cssDir     = "css"
	jsDir      = "js"
	imgDir     = "imgs"
	fontDir    = "fonts"
	pagesDir   = "pages"
	statsTitle = "\033[32mScraping Statistics:\033[0m"
)


type Statistics struct {
	CSSFiles int
	JSFiles  int
	Images   int
	Fonts    int
}

func main() {
	intro()

	cyan := color.New(color.FgCyan).SprintFunc()
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", cyan("[Console] =>"))
	fmt.Print("Do you want to start the tool? (y/n): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer != "y" && answer != "yes" {
		fmt.Printf("%s ", cyan("[Console] =>"))
		fmt.Println("Tool execution canceled.")
		return
	}

	// Flush the output buffer to ensure the new line is displayed
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	urls, err := readURLsFromFile("urls.txt")
	if err != nil {
		handleError(fmt.Sprintf("\nError reading URLs from file: %s\n", err))
		return
	}

	startTime := time.Now()

	// Variables to keep track of total counts
	var totalCSSFiles int
	var totalJSFiles int
	var totalImages int
	var totalFonts int

	for _, url := range urls {
		cssCount, jsCount, imgCount, fontCount := scrapeWebsite(url)
		totalCSSFiles += cssCount
		totalJSFiles += jsCount
		totalImages += imgCount
		totalFonts += fontCount
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fmt.Printf("\nTask completed in %s\n", color.New(color.FgCyan).Sprint(elapsedTime))
	fmt.Println()

	stats := Statistics{
		CSSFiles: totalCSSFiles,
		JSFiles:  totalJSFiles,
		Images:   totalImages,
		Fonts:    totalFonts,
	}

	printStatistics(stats)

	showHTMLFormatterNote()
}



func intro() {
	asciiArt := figure.NewFigure("justClone", "", true)
	fmt.Println(asciiArt.String())

	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Printf("%s %s\n", red("Dev Alert: There is a slight bug with fonts, but the cloned site should look 98% identical"), red("\n"))
	fmt.Printf("%s ", cyan("[Console] =>"))
	fmt.Printf("%s %s %s\n", white("Version"), cyan("0.0.2"), "| 20th July, 2023")
	fmt.Printf("%s ", cyan("[Console] =>"))
	fmt.Println("Coded by " + cyan("PoppingXanax"))
	fmt.Printf("%s ", red("[INFO] =>"))
	fmt.Printf("%s %s", red("YOU MUST RUN THE SCRAPER.GO TO SCRAPE THE ENTIRE WEBSITE"), red("\n"))
	fmt.Printf("%s ", red("[INFO] =>"))
	fmt.Printf("%s %s\n", red("OR IT WILL ONLY SCRAPE THE INDEX/MAIN PAGE!"), red("\n"))
}

func readURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	urls := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func scrapeWebsite(urlStr string) (int, int, int, int) {
	hostDirectory, err := createDirectory(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError creating directory: %s\n", err))
		return 0, 0, 0, 0
	}

	fmt.Printf("%sScrape in progress for %s...\n", color.New(color.FgYellow).Sprint("[Console] => "), color.New(color.FgCyan).Sprint(urlStr))

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	urlStr, err = followRedirects(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError following redirects: %s\n", err))
		return 0, 0, 0, 0
	}

	response, err := http.Get(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError scraping the website: %s\n", err))
		return 0, 0, 0, 0
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		handleError(fmt.Sprintf("\nError scraping the website.\nError code: %d\nReason: %s\nWebsite URL: %s\n", response.StatusCode, response.Status, urlStr))
		return 0, 0, 0, 0
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		handleError(fmt.Sprintf("\nError parsing HTML document: %s\n", err))
		return 0, 0, 0, 0
	}

	// Rest of the code for scraping and downloading assets

	htmlContent, err := doc.Html()
	if err != nil {
		handleError(fmt.Sprintf("\nError getting HTML content: %s\n", err))
		return 0, 0, 0, 0
	}

	filePath := filepath.Join(hostDirectory, "index.html")
	err = os.WriteFile(filePath, []byte(htmlContent), 0644)
	if err != nil {
		handleError(fmt.Sprintf("\nError saving index.html file: %s\n", err))
		return 0, 0, 0, 0
	}

	fmt.Printf("%sScrape completed\n", color.New(color.FgGreen).Sprint("[Console] => "))

	cssFiles := doc.Find("link[rel='stylesheet']")
	cssCount := cssFiles.Length()

	jsFiles := doc.Find("script[src]")
	jsCount := jsFiles.Length()

	imgTags := doc.Find("img[src]")
	imgCount := imgTags.Length()

	fontTags := doc.Find("link[rel='stylesheet'][href$='.woff'], link[rel='stylesheet'][href$='.woff2'], link[rel='stylesheet'][href$='.ttf'], link[rel='stylesheet'][href$='.otf']")
	fontCount := fontTags.Length()

	return cssCount, jsCount, imgCount, fontCount
}

func createDirectory(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	hostDirectory := parsedURL.Host
	pageDirectory := pagesDir
	urlPath := strings.TrimPrefix(parsedURL.Path, "/")

	directoryName := filepath.Join(hostDirectory, pageDirectory, urlPath)

	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		err := os.MkdirAll(directoryName, 0755)
		if err != nil {
			return "", err
		}
	}

	return directoryName, nil
}

func createSubdirectories(directoryName string, subdirectories ...string) error {
	for _, subdirectory := range subdirectories {
		dirPath := filepath.Join(directoryName, subdirectory)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadFile(urlStr, directory, fileType, fileExt string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	parsedURL.RawQuery = ""

	filename := path.Base(parsedURL.Path)
	filePath := filepath.Join(directory, filename)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return filePath, nil
	}

	response, err := http.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error downloading %s file: %s", fileType, urlStr)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func followRedirects(urlStr string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Request.URL.String(), nil
}

func resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}
	rel, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(rel).String()
}

func getFileExtension(url string) string {
	ext := filepath.Ext(url)
	if len(ext) > 1 {
		return strings.TrimPrefix(ext, ".")
	}
	return ""
}

func handleError(message string) {
	color.Red(message)
}

func printStatistics(stats Statistics) {
	fmt.Println(statsTitle)
	fmt.Printf("Total CSS files: %s\n", color.New(color.FgCyan).Sprint(stats.CSSFiles))
	fmt.Printf("Total JS files: %s\n", color.New(color.FgCyan).Sprint(stats.JSFiles))
	fmt.Printf("Total images: %s\n", color.New(color.FgCyan).Sprint(stats.Images))
	fmt.Printf("Total font files: %s\n", color.New(color.FgRed).Sprint(stats.Fonts))
	fmt.Println()
}

func showHTMLFormatterNote() {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s %s\n\n", yellow("Fonts should appear as normal, despite the tool stating 0 were found."), "")
	fmt.Printf("%s %s\n", red("Note: I'm working on adding an HTML formatter. For now,"), "")
	fmt.Printf("%s %s\n\n", red("please use"), cyan("https://www.freeformatter.com/html-formatter.html"))
}
