package main

import (
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
	cssDir  = "css"
	jsDir   = "js"
	imgDir  = "imgs"
	fontDir = "fonts"
)

func main() {
	intro()

	websiteURL := getInputURL()

	startTime := time.Now()
	scrapeWebsite(websiteURL)
	endTime := time.Now()

	elapsedTime := endTime.Sub(startTime)
	fmt.Printf("\nTask completed in %s\n", color.New(color.FgCyan).Sprint(elapsedTime))
	fmt.Println()

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
	fmt.Printf("%s %s %s\n", white("Version"), cyan("0.0.1"), "| 17th July, 2023")
	fmt.Printf("%s ", cyan("[Console] =>"))
	fmt.Println("Coded by " + cyan("PoppingXanax"))
}

func getInputURL() string {
	fmt.Print(color.CyanString("[Console] => "))
	fmt.Print("Enter a URL: ")
	var websiteURL string
	fmt.Scanln(&websiteURL)
	return websiteURL
}

func scrapeWebsite(urlStr string) {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	fmt.Printf("%s ", color.New(color.FgCyan).Sprint("[Console] =>"))
	fmt.Print("Starting Task Engine for site")
	fmt.Printf(" %s%s%s\n\n", color.New(color.FgWhite).Sprint("("), color.New(color.FgCyan).Sprint(urlStr), color.New(color.FgWhite).Sprint(")"))

	urlStr, err := followRedirects(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError following redirects: %s\n", err))
		return
	}

	response, err := http.Get(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError scraping the website: %s\n", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		handleError(fmt.Sprintf("\nError scraping the website.\nError code: %d\nReason: %s\nWebsite URL: %s\n", response.StatusCode, response.Status, urlStr))
		return
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		handleError(fmt.Sprintf("\nError parsing HTML document: %s\n", err))
		return
	}

	directoryName, err := createDirectory(urlStr)
	if err != nil {
		handleError(fmt.Sprintf("\nError creating directory: %s\n", err))
		return
	}

	err = createSubdirectories(directoryName, cssDir, jsDir, imgDir, fontDir)
	if err != nil {
		handleError(fmt.Sprintf("\nError creating subdirectories: %s\n", err))
		return
	}

	cssFiles := doc.Find("link[rel='stylesheet']")
	color.Yellow(fmt.Sprintf("Found %d CSS file(s).\n", cssFiles.Length()))
	cssFiles.Each(func(i int, s *goquery.Selection) {
		cssURL, _ := s.Attr("href")
		color.Cyan(fmt.Sprintf("Downloading CSS file: %s\n", cssURL))
		filePath, err := downloadFile(resolveURL(urlStr, cssURL), filepath.Join(directoryName, cssDir), "CSS", getFileExtension(cssURL))
		if err == nil {
			s.SetAttr("href", fmt.Sprintf("%s/%s", cssDir, filepath.Base(filePath)))
		}
	})

	jsFiles := doc.Find("script[src]")
	color.Yellow(fmt.Sprintf("Found %d JS file(s).\n", jsFiles.Length()))
	jsFiles.Each(func(i int, s *goquery.Selection) {
		jsURL, _ := s.Attr("src")
		color.Cyan(fmt.Sprintf("Downloading JS file: %s\n", jsURL))
		filePath, err := downloadFile(resolveURL(urlStr, jsURL), filepath.Join(directoryName, jsDir), "JS", getFileExtension(jsURL))
		if err == nil {
			s.SetAttr("src", fmt.Sprintf("%s/%s", jsDir, filepath.Base(filePath)))
		}
	})

	imgTags := doc.Find("img[src]")
	color.Yellow(fmt.Sprintf("Found %d image(s).\n", imgTags.Length()))
	imgTags.Each(func(i int, s *goquery.Selection) {
		imgURL, _ := s.Attr("src")
		color.Cyan(fmt.Sprintf("Downloading image: %s\n", imgURL))
		_, err := downloadFile(resolveURL(urlStr, imgURL), filepath.Join(directoryName, imgDir), "IMG", getFileExtension(imgURL))
		if err == nil {
			s.SetAttr("src", fmt.Sprintf("%s/%s", imgDir, filepath.Base(imgURL)))
		}
	})

	fontTags := doc.Find("link[rel='stylesheet'][href$='.woff'], link[rel='stylesheet'][href$='.woff2'], link[rel='stylesheet'][href$='.ttf'], link[rel='stylesheet'][href$='.otf']")
	color.Yellow(fmt.Sprintf("Found %d font file(s). \n", fontTags.Length()))
	fontTags.Each(func(i int, s *goquery.Selection) {
		fontURL, _ := s.Attr("href")
		color.Cyan(fmt.Sprintf("Downloading font file: %s\n", fontURL))
		_, err := downloadFile(resolveURL(urlStr, fontURL), filepath.Join(directoryName, fontDir), "FONT", getFileExtension(fontURL))
		if err == nil {
			s.SetAttr("href", fmt.Sprintf("%s/%s", fontDir, filepath.Base(fontURL)))
		}
	})

	htmlContent, err := doc.Html()
	if err != nil {
		handleError(fmt.Sprintf("\nError getting HTML content: %s\n", err))
		return
	}

	filePath := filepath.Join(directoryName, "index.html")
	err = os.WriteFile(filePath, []byte(htmlContent), 0644)
	if err != nil {
		handleError(fmt.Sprintf("\nError saving index.html file: %s\n", err))
		return
	}

	color.Green("\nScraping completed successfully!")
	printStatistics(cssFiles.Length(), jsFiles.Length(), imgTags.Length(), fontTags.Length())
}

func createDirectory(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	directoryName := parsedURL.Host

	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		err := os.Mkdir(directoryName, 0755)
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
		color.Yellow("%s file already exists: %s", fileType, filePath)
		color.Cyan("")
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

	color.Green("%s file was saved to --> %s", fileType, filePath)
	color.Cyan("")
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

func printStatistics(totalCSSFiles, totalJSFiles, totalImages, totalFonts int) {
	fmt.Println("\nScraping Statistics:")
	fmt.Printf("Total CSS files: %s\n", color.New(color.FgCyan).Sprint(totalCSSFiles))
	fmt.Printf("Total JS files: %s\n", color.New(color.FgCyan).Sprint(totalJSFiles))
	fmt.Printf("Total images: %s\n", color.New(color.FgCyan).Sprint(totalImages))
	fmt.Printf("Total font files: %s\n", color.New(color.FgRed).Sprint(totalFonts))
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

func showHTMLFormatterNote() {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s %s\n", red("Note: I'm working on adding an HTML formatter. For now,"), "")
	fmt.Println(red("Please use") + " " + cyan("https://www.freeformatter.com/html-formatter.html"))
	fmt.Println()
}
