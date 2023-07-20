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

	"github.com/PuerkitoBio/goquery"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

const (
	cssDir  = "css"
	jsDir   = "js"
	imgDir  = "imgs"
	fontDir = "fonts"
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

	urls, err := readURLsFromFile("urls.txt")
	if err != nil {
		handleError(fmt.Sprintf("\nError reading URLs from file: %s\n", err))
		return
	}

	startTime := time.Now()

	stats := Statistics{}

	for _, url := range urls {
		cssCount, jsCount, imgCount, fontCount := scrapeWebsite(url)
		stats.CSSFiles += cssCount
		stats.JSFiles += jsCount
		stats.Images += imgCount
		stats.Fonts += fontCount
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fmt.Printf("\nTask completed in %s\n", color.New(color.FgCyan).Sprint(elapsedTime))

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
	fmt.Printf("%s %s %s\n", white("Version"), cyan("0.1.0"), "| 20th July, 2023")
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

	parsedURL, _ := url.Parse(urlStr)
	pageName := getPageName(parsedURL.Path)

	fmt.Printf("%sScraping %s...\n", color.New(color.FgYellow).Sprint("[Console] => "), color.New(color.FgCyan).Sprint(pageName))

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

	cssFiles := doc.Find("link[rel='stylesheet']")
	cssCount := cssFiles.Length()
	cssFiles.Each(func(i int, s *goquery.Selection) {
		cssURL, _ := s.Attr("href")
		filePath, err := downloadFile(resolveURL(urlStr, cssURL), filepath.Join(hostDirectory, cssDir), "CSS", getFileExtension(cssURL))
		if err == nil {
			s.SetAttr("href", fmt.Sprintf("%s/%s", cssDir, filepath.Base(filePath)))
		}
	})

	jsFiles := doc.Find("script[src]")
	jsCount := jsFiles.Length()
	jsFiles.Each(func(i int, s *goquery.Selection) {
		jsURL, _ := s.Attr("src")
		filePath, err := downloadFile(resolveURL(urlStr, jsURL), filepath.Join(hostDirectory, jsDir), "JS", getFileExtension(jsURL))
		if err == nil {
			s.SetAttr("src", fmt.Sprintf("%s/%s", jsDir, filepath.Base(filePath)))
		}
	})

	imgTags := doc.Find("img[src]")
	imgCount := imgTags.Length()
	imgTags.Each(func(i int, s *goquery.Selection) {
		imgURL, _ := s.Attr("src")
		filePath, err := downloadFile(resolveURL(urlStr, imgURL), filepath.Join(hostDirectory, imgDir), "IMG", getFileExtension(imgURL))
		if err == nil {
			s.SetAttr("src", fmt.Sprintf("%s/%s", imgDir, filepath.Base(filePath)))
		}
	})

	fontTags := doc.Find("link[rel='stylesheet'][href$='.woff'], link[rel='stylesheet'][href$='.woff2'], link[rel='stylesheet'][href$='.ttf'], link[rel='stylesheet'][href$='.otf']")
	fontCount := fontTags.Length()
	fontTags.Each(func(i int, s *goquery.Selection) {
		fontURL, _ := s.Attr("href")
		filePath, err := downloadFile(resolveURL(urlStr, fontURL), filepath.Join(hostDirectory, fontDir), "FONT", getFileExtension(fontURL))
		if err == nil {
			s.SetAttr("href", fmt.Sprintf("%s/%s", fontDir, filepath.Base(filePath)))
		}
	})

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

	fmt.Printf("%sScraping %s completed\n", color.New(color.FgGreen).Sprint("[Console] => "), color.New(color.FgCyan).Sprint(pageName))

	return cssCount, jsCount, imgCount, fontCount
}


func createDirectory(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	hostDirectory := parsedURL.Host
	pageName := getPageName(parsedURL.Path)
	directoryPath := filepath.Join(hostDirectory, "pages", pageName)

	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		err := os.MkdirAll(directoryPath, 0755)
		if err != nil {
			return "", err
		}
	}

	return directoryPath, nil
}

func getPageName(urlPath string) string {
	if urlPath == "" || urlPath == "/" {
		return "index"
	}

	base := path.Base(urlPath)
	pageName := strings.TrimSuffix(base, path.Ext(base))
	if pageName == "" {
		return "index"
	}

	return pageName
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

	// Create the nested directories if they don't exist
	err = os.MkdirAll(directory, 0755)
	if err != nil {
		return "", err
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


func resolveURL(baseURL, href string) string {
	base, _ := url.Parse(baseURL)
	relative, _ := url.Parse(href)
	return base.ResolveReference(relative).String()
}

func getFileExtension(filename string) string {
	ext := path.Ext(filename)
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

func printStatistics(stats Statistics) {
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()
	fmt.Printf("%s\n", green("Scraping Statistics:"))
	fmt.Printf("Total CSS files: %s\n", color.New(color.FgCyan).Sprint(stats.CSSFiles))
	fmt.Printf("Total JS files: %s\n", color.New(color.FgCyan).Sprint(stats.JSFiles))
	fmt.Printf("Total images: %s\n", color.New(color.FgCyan).Sprint(stats.Images))
	fmt.Printf("Total font files: %s\n", color.New(color.FgRed).Sprint(stats.Fonts))
	fmt.Println()
}



func handleError(message string) {
	color.Red(message)
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

func showHTMLFormatterNote() {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s %s\n", red("Note: I'm working on adding an HTML formatter. For now,"), "")
	fmt.Println(red("Please use") + " " + cyan("https://www.freeformatter.com/html-formatter.html"))
	fmt.Println()
}
