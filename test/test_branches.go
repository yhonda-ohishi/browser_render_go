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
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get credentials from environment
	compID := os.Getenv("COMP_ID")
	userName := os.Getenv("USER_NAME")
	userPass := os.Getenv("USER_PASS")

	log.Printf("Starting browser test with credentials - Company: %s, User: %s", compID, userName)

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
	l = l.Headless(false).Devtools(true).Leakless(false)
	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()
	defer page.MustClose()

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

	// First, try to get all branches
	log.Println("Trying to get branch list...")
	branches := page.MustEval(`() => {
		if (typeof VenusBridgeService !== 'undefined' && typeof VenusBridgeService.BranchList === 'function') {
			return new Promise((resolve, reject) => {
				VenusBridgeService.BranchList(
					(data) => resolve(data),
					(error) => reject(error)
				);
			});
		} else {
			return 'BranchList API not found';
		}
	}`)

	log.Printf("Branches result: %v", branches)

	// Test different branch IDs
	testBranches := []string{"", "00000000", "00000001", "00000002", "00000003"}
	testFilters := []string{"", "0", "1", "2", "3"}

	for _, branchID := range testBranches {
		for _, filterID := range testFilters {
			log.Printf("\n=== Testing branchID='%s', filterID='%s' ===", branchID, filterID)

			result := page.MustEval(`(branchID, filterID) => {
				return new Promise((resolve, reject) => {
					if (typeof VenusBridgeService === 'undefined') {
						reject('VenusBridgeService not found');
						return;
					}
					VenusBridgeService.VehicleStateTableForBranchEx(branchID, filterID,
						(data) => {
							resolve({
								count: data ? data.length : 0,
								sample: data && data.length > 0 ? data[0] : null
							});
						},
						(error) => reject(error)
					);
				});
			}`, branchID, filterID)

			// Convert to JSON for display
			jsonData, _ := json.MarshalIndent(result, "", "  ")
			log.Printf("Result: %s", jsonData)

			time.Sleep(2 * time.Second)
		}
	}

	log.Println("\n=== Test completed ===")
}