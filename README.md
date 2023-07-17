### Website Scraping Script

This script is designed to scrape a website and download its CSS files, JS files, and images. It also updates the HTML file with local references to the downloaded files.

FYI: Right now it only scrapes the index page. I will update the script to scrape all URLs, then save each CSS, JS, IMG, file correspondingly.

## Features

- Downloads CSS files, JS files, and images from a website
- Replaces the URLs in the HTML with local file references
- Creates separate directories for CSS, JS, and images
- Handles redirects and follows them to the final destination
- Provides scraping statistics including the total number of CSS files, JS files, and images found

## Requirements

- Go 1.16 or higher
- `go get github.com/common-nighthawk/go-figure`
- `go get github.com/PuerkitoBio/goquery`
- `go get github.com/fatih/color`

## Usage

1) Clone the repository or download the script file.
2) Build the project using the command go build.
3) Run the executable using ./main.
4) Enter the URL of the website you want to scrape when prompted.
5) Wait for the script to complete the scraping process.
6) The downloaded files will be stored in separate directories (css, js, imgs, etc) under the website's domain name.
7) The updated HTML file with local references will be saved as index.html in the website's directory.

You may need to run `chmod +x justclone`

## Todo List

- Proxy Support ❌
- Browser mode (for scraping sites with JS related challenges) ❌
- User-Agent use ❌
- HTML Parsing Improvements ❌
- Metadata Extraction ❌
- Interactive Mode (add an interactive mode where users can dynamically input URLs to scrape without relaunching the application each time) ❌
- Cache improvements ❌
- Better error logging ❌
- Rate limit bypassing ❌
- Authentication Support (if the website requires authentication or session management, add support for handling login credentials and maintaining authenticated sessions during the scraping process) ❌
- Pre-set cookie(s) ❌

## Limitations

- The script may encounter connection issues with certain websites, especially if they have strict security measures or block scraping activities. In such cases, it may fail to download certain files or raise connection errors.
- The script may not handle all possible edge cases or complex website structures. It is designed as a basic scraping tool and may require modifications for specific use cases.

## Disclaimer

This script is provided as-is without any warranty. Use it responsibly and make sure to comply with the website's terms of service and legal requirements when scraping websites.


