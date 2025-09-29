package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Get credentials from environment
	compID := os.Getenv("COMP_ID")
	userName := os.Getenv("USER_NAME")
	userPass := os.Getenv("USER_PASS")

	if compID == "" || userName == "" || userPass == "" {
		log.Fatal("Please set COMP_ID, USER_NAME, and USER_PASS in .env file")
	}

	fmt.Println("==============================================")
	fmt.Println("Browser Render Go - Console Test")
	fmt.Println("==============================================")
	fmt.Printf("Company ID: %s\n", compID)
	fmt.Printf("User Name: %s\n", userName)
	fmt.Println("==============================================\n")

	// Initialize browser
	fmt.Println("Starting browser...")

	// Try to find Chrome or Edge
	path, _ := launcher.LookPath()
	if path == "" {
		// Try common Chrome locations
		possiblePaths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		}

		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}

	fmt.Printf("Using browser: %s\n", path)

	l := launcher.New().
		Bin(path).
		Headless(false). // GUI mode for testing
		Devtools(false).
		NoSandbox(true).
		Leakless(false). // Disable leakless to avoid antivirus issues
		Set("disable-blink-features", "AutomationControlled")

	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()
	defer page.MustClose()

	// Set timeout
	page = page.Timeout(30 * time.Second)

	// Try to login
	fmt.Println("\n[1] Navigating to login page...")
	page.MustNavigate("https://theearth-np.com/F-OES1010[Login].aspx?mode=timeout")
	page.MustWaitLoad()

	// Important: Wait for page to stabilize
	fmt.Println("[2] Waiting for page to stabilize...")
	time.Sleep(3 * time.Second)

	// Check if login form exists
	if page.MustHas("#txtPass") {
		fmt.Println("[3] Login form found! Filling credentials...")

		// Handle popup if present
		if page.MustHas("#popup_1") {
			popup, _ := page.Element("#popup_1")
			if popup != nil {
				if visible, _ := popup.Visible(); visible {
					fmt.Println("   - Closing popup...")
					popup.MustClick()
					time.Sleep(1 * time.Second)
				}
			}
		}

		// Fill credentials
		fmt.Println("   - Filling Company ID...")
		page.MustElement("#txtID2").MustInput(compID)

		fmt.Println("   - Filling User Name...")
		page.MustElement("#txtID1").MustInput(userName)

		fmt.Println("   - Filling Password...")
		page.MustElement("#txtPass").MustInput(userPass)

		// Take screenshot
		fmt.Println("[4] Taking screenshot before login...")
		screenshot := page.MustScreenshot()
		os.WriteFile("login_form.png", screenshot, 0644)
		fmt.Println("   - Screenshot saved as login_form.png")

		// Click login button and wait for navigation
		fmt.Println("[5] Clicking login button...")
		loginBtn := page.MustElement("#imgLogin")
		loginBtn.MustClick()

		// Wait for navigation with proper timing
		fmt.Println("[6] Waiting for navigation...")
		page.MustWaitRequestIdle() // Wait for network to be idle
		time.Sleep(5 * time.Second) // Give enough time for page to load

		// Check if login was successful
		if page.MustHas("#Button1st_7") {
			fmt.Println("✅ Login successful!")
		} else {
			fmt.Println("⚠️ Login might have failed or already logged in")

			// Check for popup again (already logged in case)
			if page.MustHas("#popup_1") {
				fmt.Println("   - Found popup (already logged in), clicking to continue...")
				popup := page.MustElement("#popup_1")
				popup.MustClick()

				// Wait for proper navigation
				page.MustWaitRequestIdle()
				time.Sleep(5 * time.Second)
			}
		}

		// Save cookies
		fmt.Println("\n[7] Saving cookies...")
		cookies := page.MustCookies()
		cookieData, _ := json.MarshalIndent(cookies, "", "  ")
		os.WriteFile("cookies.json", cookieData, 0644)
		fmt.Printf("   - Saved %d cookies to cookies.json\n", len(cookies))

	} else {
		fmt.Println("❌ Login form not found!")
	}

	// Navigate to main page
	fmt.Println("\n[8] Navigating to Venus Main page...")
	page.MustNavigate("https://theearth-np.com/WebVenus/F-AAV0001[VenusMain].aspx")
	page.MustWaitLoad()

	// Wait for page to fully load
	fmt.Println("   - Waiting for page to fully load...")
	page.MustWaitRequestIdle() // Wait for all network requests
	time.Sleep(5 * time.Second) // Additional wait for JavaScript to initialize

	// Try to get vehicle data
	fmt.Println("[9] Attempting to get vehicle data...")

	// First check if VenusBridgeService exists
	hasService := page.MustEval(`() => {
		return typeof VenusBridgeService !== 'undefined' &&
		       typeof VenusBridgeService.VehicleStateTableForBranchEx === 'function';
	}`).Bool()

	fmt.Printf("   - VenusBridgeService available: %v\n", hasService)

	// Execute JavaScript and get the result
	var vehicleData interface{}
	var err error
	if hasService {
		err = rod.Try(func() {
			obj := page.MustEval(`() => {
				return new Promise((resolve, reject) => {
					VenusBridgeService.VehicleStateTableForBranchEx('00000000', '0',
						(data) => resolve(data),
						(error) => reject(error)
					);
				});
			}`)
			// Get the raw value directly
			vehicleData = obj
		})
	} else {
		err = fmt.Errorf("VenusBridgeService not found on page")
	}

	if err != nil {
		fmt.Printf("❌ Error getting vehicle data: %v\n", err)

		// Take error screenshot
		screenshot := page.MustScreenshot()
		os.WriteFile("error_screen.png", screenshot, 0644)
		fmt.Println("   - Error screenshot saved as error_screen.png")
	} else {
		fmt.Println("✅ Vehicle data retrieved successfully!")

		// Save data
		jsonData, _ := json.MarshalIndent(vehicleData, "", "  ")
		os.WriteFile("vehicle_data.json", jsonData, 0644)
		fmt.Println("   - Data saved to vehicle_data.json")

		// Show summary
		if vehicles, ok := vehicleData.([]interface{}); ok {
			fmt.Printf("   - Total vehicles found: %d\n", len(vehicles))

			// Show first few vehicles
			for i, item := range vehicles {
				if i >= 3 {
					break
				}
				if m, ok := item.(map[string]interface{}); ok {
					fmt.Printf("   - Vehicle %d: %v\n", i+1, m["VehicleCD"])
				}
			}
		}
	}

	fmt.Println("\n==============================================")
	fmt.Println("Test completed!")
	fmt.Println("Check the following files:")
	fmt.Println("  - login_form.png: Screenshot of login form")
	fmt.Println("  - cookies.json: Saved cookies")
	fmt.Println("  - vehicle_data.json: Retrieved vehicle data")
	fmt.Println("  - error_screen.png: Error screenshot (if any)")
	fmt.Println("==============================================")

	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}