import os
import requests
from bs4 import BeautifulSoup
from urllib.parse import urlparse, urljoin, urlunparse
from colorama import init, Fore

# Initialize colorama
init()

def intro():
    print(f"[{Fore.CYAN}Console{Fore.RESET}] {Fore.CYAN}Website Cloner{Fore.RESET}")
    print(f"[{Fore.CYAN}Console{Fore.RESET}] Coded by {Fore.CYAN}PoppingXanax{Fore.RESET}")

def create_directory(url):
    # Remove the protocol prefix from the URL
    parsed_url = urlparse(url)
    directory_name = parsed_url.netloc

    # Create the directory if it doesn't exist
    if not os.path.exists(directory_name):
        os.makedirs(directory_name)

    return directory_name

def download_file(url, directory, file_type):
    # Remove query parameters from the URL
    parsed_url = urlparse(url)
    updated_url = urlunparse(parsed_url._replace(query=''))

    # Extract the filename from the updated URL
    filename = os.path.basename(urlparse(updated_url).path)
    file_path = os.path.join(directory, filename)

    if os.path.isfile(file_path):
        print(f"{Fore.YELLOW}{file_type} file already exists: {file_path}{Fore.RESET}")
        print()
        return file_path
    else:
        try:
            response = requests.get(url, allow_redirects=True, timeout=10)
            response.raise_for_status()
        except requests.exceptions.RequestException as e:
            print(f"\n{Fore.RED}Error downloading {file_type} file: {url}")
            print(f"Exception: {str(e)}{Fore.RESET}\n")
            return None
        except IOError as e:
            print(f"\n{Fore.RED}Error saving {file_type} file: {file_path}")
            print(f"Exception: {str(e)}{Fore.RESET}\n")
            return None

        try:
            with open(file_path, "wb") as file:
                file.write(response.content)
        except IOError as e:
            print(f"\n{Fore.RED}Error saving {file_type} file: {file_path}")
            print(f"Exception: {str(e)}{Fore.RESET}\n")
            return None

        print(f"{Fore.GREEN}{file_type} file was saved to --> {file_path}{Fore.RESET}")
        print()

    return file_path


def follow_redirects(url):
    with requests.Session() as session:
        response = session.head(url, allow_redirects=True)

        if response.status_code in (301, 302):
            location = response.headers.get("Location")
            if location:
                return follow_redirects(urljoin(url, location))

        return url

def scrape_website(url):
    # Add 'https://' if scheme is missing
    if not url.startswith("http://") and not url.startswith("https://"):
        url = f"https://{url}"

    print(f"[{Fore.CYAN}Console{Fore.RESET}] Starting Task Engine for site ({Fore.CYAN}{url}{Fore.RESET})\n")

    # Follow redirects
    url = follow_redirects(url)

    response = requests.get(url)
    if response.status_code == 200:
        soup = BeautifulSoup(response.content, "html.parser")

        # Create the directory with the website name
        directory_name = create_directory(url)

        # Create subdirectories for CSS, JS, and images inside the website directory
        css_dir = os.path.join(directory_name, "css")
        js_dir = os.path.join(directory_name, "js")
        img_dir = os.path.join(directory_name, "imgs")
        os.makedirs(css_dir, exist_ok=True)
        os.makedirs(js_dir, exist_ok=True)
        os.makedirs(img_dir, exist_ok=True)

        # Download CSS files
        css_files = soup.find_all("link", {"rel": "stylesheet"})
        print(f"Found {Fore.YELLOW}{len(css_files)}{Fore.RESET} CSS file(s).")
        for css_file in css_files:
            css_url = urljoin(url, css_file["href"])
            print(f"Downloading CSS file: {Fore.BLUE}{css_url}{Fore.RESET}")
            download_file(css_url, css_dir, "CSS")

            # Replace the URL in the HTML with the local filename
            css_file["href"] = f"css/{os.path.basename(urlparse(css_url).path)}"

        # Download JS files
        js_files = soup.find_all("script", {"src": True})
        print(f"Found {Fore.YELLOW}{len(js_files)}{Fore.RESET} JS file(s).")
        for js_file in js_files:
            js_url = urljoin(url, js_file["src"])
            print(f"Downloading JS file: {Fore.BLUE}{js_url}{Fore.RESET}")
            download_file(js_url, js_dir, "JS")

            # Replace the URL in the HTML with the local filename
            js_file["src"] = f"js/{os.path.basename(urlparse(js_url).path)}"

        # Download images
        img_tags = soup.find_all("img", {"src": True})
        print(f"Found {Fore.YELLOW}{len(img_tags)}{Fore.RESET} image(s).")
        for img_tag in img_tags:
            img_url = urljoin(url, img_tag["src"])
            print(f"Downloading image: {Fore.BLUE}{img_url}{Fore.RESET}")
            download_file(img_url, img_dir, "IMG")

            # Replace the URL in the HTML with the local filename
            img_tag["src"] = f"imgs/{os.path.basename(urlparse(img_url).path)}"

        # Clean up HTML by prettifying the content
        cleaned_html = soup.prettify()

        # Update index.html with the cleaned and formatted HTML
        with open(os.path.join(directory_name, "index.html"), "w", encoding="utf-8") as file:
            file.write(cleaned_html)

        print("\nScraping completed successfully!")
        print_statistics(css_files, js_files, img_tags)
    else:
        print("\nError scraping the website.")
        print(f"Error code: {Fore.RED}{response.status_code}{Fore.RESET}")
        print(f"Reason: {Fore.RED}{response.reason}{Fore.RESET}")
        print(f"Website URL: {Fore.RED}{url}{Fore.RESET}")

def print_statistics(css_files, js_files, img_tags):
    total_css_files = len(css_files)
    total_js_files = len(js_files)
    total_images = len(img_tags)

    print("\nScraping Statistics:")
    print(f"Total CSS files: {Fore.YELLOW}{total_css_files}{Fore.RESET}")
    print(f"Total JS files: {Fore.YELLOW}{total_js_files}{Fore.RESET}")
    print(f"Total images: {Fore.YELLOW}{total_images}{Fore.RESET}")

intro()

# Prompt the user to enter a URL
website_url = input(f"[{Fore.CYAN}Console{Fore.RESET}] Enter a URL: {Fore.RESET}")

scrape_website(website_url)
