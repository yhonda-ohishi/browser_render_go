package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get credentials from environment
	compID := os.Getenv("COMP_ID")
	userName := os.Getenv("USER_NAME")
	userPass := os.Getenv("USER_PASS")

	log.Printf("Testing with credentials - Company: %s, User: %s", compID, userName)

	// Launch browser
	path, _ := launcher.LookPath()
	if path == "" {
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

	l := launcher.New()
	if path != "" {
		l = l.Bin(path)
	}
	l = l.Headless(false).Devtools(false).Leakless(false)
	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()
	defer page.MustClose()

	// Set timeout
	page = page.Timeout(60 * time.Second)

	// Login
	log.Println("Navigating to login page...")
	page.MustNavigate("https://theearth-np.com/F-OES1010[Login].aspx?mode=timeout")
	page.MustWaitLoad()
	time.Sleep(3 * time.Second)

	// Fill credentials
	page.MustElement("#txtID2").MustInput(compID)
	page.MustElement("#txtID1").MustInput(userName)
	page.MustElement("#txtPass").MustInput(userPass)

	// Login
	log.Println("Clicking login button...")
	page.MustElement("#imgLogin").MustClick()
	page.MustWaitRequestIdle()
	time.Sleep(5 * time.Second)

	// Navigate to Venus Main
	log.Println("Navigating to Venus Main page...")
	page.MustNavigate("https://theearth-np.com/WebVenus/F-AAV0001[VenusMain].aspx")
	page.MustWaitLoad()
	time.Sleep(5 * time.Second)

	// Test API call with default params (should get 193 vehicles)
	log.Println("Testing API call with branchID='', filterID='0'...")

	// First wait for VenusBridgeService to be available
	hasService := page.MustEval(`() => {
		return typeof VenusBridgeService !== 'undefined' &&
		       typeof VenusBridgeService.VehicleStateTableForBranchEx === 'function';
	}`).Bool()

	log.Printf("VenusBridgeService available: %v", hasService)

	if !hasService {
		log.Fatal("VenusBridgeService not found on page")
	}

	// Execute the API call
	start := time.Now()
	result := page.MustEval(`(branchID, filterID) => {
		return new Promise((resolve, reject) => {
			const startTime = Date.now();
			console.log('Calling API with branchID=' + branchID + ', filterID=' + filterID);

			VenusBridgeService.VehicleStateTableForBranchEx(branchID, filterID,
				(data) => {
					const elapsed = Date.now() - startTime;
					console.log('API returned ' + data.length + ' records in ' + elapsed + 'ms');
					resolve({
						success: true,
						count: data.length,
						elapsed: elapsed,
						sample: data.length > 0 ? data[0] : null
					});
				},
				(error) => {
					console.error('API error:', error);
					reject({
						success: false,
						error: error,
						elapsed: Date.now() - startTime
					});
				}
			);

			// Add a timeout fallback
			setTimeout(() => {
				reject({
					success: false,
					error: 'Timeout after 30 seconds',
					elapsed: 30000
				});
			}, 30000);
		});
	}`, "", "0")

	elapsed := time.Since(start)
	log.Printf("API call completed in %v", elapsed)

	// Convert to JSON for display
	jsonData, _ := json.MarshalIndent(result.Val(), "", "  ")
	log.Printf("Result: %s", jsonData)

	// Save full data
	os.WriteFile("../vehicle_test_result.json", jsonData, 0644)
	log.Println("Results saved to vehicle_test_result.json")
}