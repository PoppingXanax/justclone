# Website Scraping Script

This script is designed to scrape a website and download its CSS files, JS files, and images. It also updates the HTML file with local references to the downloaded files.

## Features

- Downloads CSS files, JS files, and images from a website
- Replaces the URLs in the HTML with local file references
- Creates separate directories for CSS, JS, and images
- Handles redirects and follows them to the final destination
- Provides scraping statistics including the total number of CSS files, JS files, and images found

## Requirements

- Python 3.x
- Requests library: `pip install requests`
- Beautiful Soup library: `pip install beautifulsoup4`
- Colorama library: `pip install colorama`

## Usage

1. Clone the repository or download the script file.
2. Install the required libraries mentioned above, if not already installed.
3. Open a terminal or command prompt and navigate to the directory containing the script.
4. Run the script using the command: `python script.py`
5. Enter the URL of the website you want to scrape when prompted.
6. Wait for the script to complete the scraping process.
7. The downloaded files will be stored in separate directories (css, js, imgs) under the website's domain name.
8. The updated HTML file with local references will be saved as `index.html` in the website's directory.

## Limitations

- The script may encounter connection issues with certain websites, especially if they have strict security measures or block scraping activities. In such cases, it may fail to download certain files or raise connection errors.
- The script may not handle all possible edge cases or complex website structures. It is designed as a basic scraping tool and may require modifications for specific use cases.

## Disclaimer

This script is provided as-is without any warranty. Use it responsibly and make sure to comply with the website's terms of service and legal requirements when scraping websites.
