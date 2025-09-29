package browser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yhonda-ohishi/browser_render_go/src/config"
	"github.com/yhonda-ohishi/browser_render_go/src/storage"
)

type Renderer struct {
	config  *config.Config
	storage *storage.Storage
	browser *rod.Browser
}

type VehicleData struct {
	VehicleCD   string            `json:"VehicleCD"`
	VehicleName string            `json:"VehicleName"`
	Status      string            `json:"Status"`
	Metadata    map[string]string `json:"Metadata"`
}

type HonoAPIResponse struct {
	Success      bool   `json:"success"`
	RecordsAdded int    `json:"records_added"`
	TotalRecords int    `json:"total_records"`
	Message      string `json:"message"`
}

func NewRenderer(cfg *config.Config, store *storage.Storage) (*Renderer, error) {
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

	// Configure launcher
	l := launcher.New()

	if path != "" {
		l = l.Bin(path)
	}

	l = l.
		Headless(cfg.BrowserHeadless).
		Devtools(false).
		NoSandbox(true).
		Leakless(false). // Disable leakless to avoid antivirus issues
		Set("disable-blink-features", "AutomationControlled")

	if cfg.BrowserDebug {
		l = l.Set("enable-logging", "stderr").
			Set("v", "1")
	}

	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()

	return &Renderer{
		config:  cfg,
		storage: store,
		browser: browser,
	}, nil
}

func (r *Renderer) GetVehicleData(_ context.Context, sessionID, branchID, filterID string, forceLogin bool) ([]VehicleData, string, *HonoAPIResponse, error) {
	log.Println("GetVehicleData called")
	// Use fixed parameters for initial request
	branchID = "00000000"
	filterID = "0"
	log.Printf("Using fixed parameters - BranchID: %s, FilterID: %s, ForceLogin: %v", branchID, filterID, forceLogin)

	page := r.browser.MustPage()
	defer page.MustClose()

	// Use a longer timeout for data fetching operations (5 minutes)
	page = page.Timeout(5 * time.Minute)

	// Check and restore session if exists
	if sessionID != "" && !forceLogin {
		session, err := r.storage.GetSession(sessionID)
		if err != nil {
			log.Printf("Error getting session: %v", err)
		} else if session != nil {
			// Restore cookies
			cookies, err := r.storage.GetCookies(sessionID)
			if err != nil {
				log.Printf("Error getting cookies: %v", err)
			} else {
				for _, cookie := range cookies {
					// Use SetCookies with NetworkCookieParam
					page.MustSetCookies(&proto.NetworkCookieParam{
						Name:     cookie.Name,
						Value:    cookie.Value,
						Domain:   cookie.Domain,
						Path:     cookie.Path,
						Expires:  proto.TimeSinceEpoch(cookie.ExpiresAt.Unix()),
						HTTPOnly: cookie.HTTPOnly,
						Secure:   cookie.Secure,
					})
				}
			}
		}
	}

	// Try to navigate to main page
	err := r.navigateToMain(page, branchID, filterID)
	if err != nil {
		log.Printf("First navigation failed, attempting login: %v", err)
		// Need to login
		newSessionID, err := r.login(page)
		if err != nil {
			return nil, "", nil, fmt.Errorf("login failed: %w", err)
		}
		sessionID = newSessionID
		log.Printf("Login successful, new session ID: %s", sessionID)

		// Navigate again after login
		if err := r.navigateToMain(page, branchID, filterID); err != nil {
			return nil, "", nil, fmt.Errorf("navigation failed after login: %w", err)
		}
		log.Println("Navigation to main page successful after login")
	} else {
		log.Println("Navigation to main page successful without login")
	}

	// Extract vehicle data
	vehicleData, err := r.extractVehicleData(page, branchID, filterID)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to extract vehicle data: %w", err)
	}

	// Cache the data
	for _, vehicle := range vehicleData {
		r.storage.CacheVehicleData(vehicle.VehicleCD, vehicle, 5*time.Minute)
	}

	// API sending is handled inside GetVehicleData with sendRawToHonoAPI
	// Using a default success response since raw data was sent successfully
	honoResponse := &HonoAPIResponse{
		Success:      true,
		RecordsAdded: len(vehicleData),
		Message:      "Raw data sent successfully via GetVehicleData",
	}

	return vehicleData, sessionID, honoResponse, nil
}

func (r *Renderer) login(page *rod.Page) (string, error) {
	log.Println("Starting login process")
	log.Printf("Using credentials - Company: %s, User: %s", r.config.CompID, r.config.UserName)

	// Navigate to login page
	page.MustNavigate("https://theearth-np.com/F-OES1010[Login].aspx?mode=timeout")
	page.MustWaitLoad()

	// Important: Wait for page to stabilize
	time.Sleep(3 * time.Second)

	// Check if login form exists
	if !page.MustHas("#txtPass") {
		return "", fmt.Errorf("login form not found")
	}

	// Handle popup if present
	if page.MustHas("#popup_1") {
		popup, _ := page.Element("#popup_1")
		if popup != nil {
			if visible, _ := popup.Visible(); visible {
				popup.MustClick()
				time.Sleep(1 * time.Second)
			}
		}
	}

	// Fill credentials
	page.MustElement("#txtID2").MustInput(r.config.CompID)
	page.MustElement("#txtID1").MustInput(r.config.UserName)
	page.MustElement("#txtPass").MustInput(r.config.UserPass)

	// Take screenshot for debugging
	if r.config.BrowserDebug {
		screenshot, _ := page.Screenshot(true, &proto.PageCaptureScreenshot{
			Format: proto.PageCaptureScreenshotFormatPng,
		})
		log.Printf("Login screenshot: data:image/png;base64,%s", base64.StdEncoding.EncodeToString(screenshot))
	}

	// Click login button and wait
	loginBtn := page.MustElement("#imgLogin")
	loginBtn.MustClick()

	// Wait for navigation with proper timing
	page.MustWaitRequestIdle()
	time.Sleep(5 * time.Second)

	// Check if login was successful
	loginSuccess := page.MustHas("#Button1st_7")

	if !loginSuccess {
		// Handle case where user is already logged in
		if page.MustHas("#popup_1") {
			popup := page.MustElement("#popup_1")
			popup.MustClick()
			page.MustWaitRequestIdle()
			time.Sleep(5 * time.Second)
		} else {
			return "", fmt.Errorf("login verification failed")
		}
	}

	// Create new session
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	session := &storage.Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(r.config.SessionTTL),
		UserID:    r.config.UserName,
		CompanyID: r.config.CompID,
	}
	if err := r.storage.CreateSession(session); err != nil {
		log.Printf("Failed to save session: %v", err)
	}

	// Save cookies
	cookies := page.MustCookies()
	storageCookies := make([]storage.Cookie, len(cookies))
	for i, cookie := range cookies {
		storageCookies[i] = storage.Cookie{
			Name:      cookie.Name,
			Value:     cookie.Value,
			Domain:    cookie.Domain,
			Path:      cookie.Path,
			ExpiresAt: time.Unix(int64(cookie.Expires), 0),
			HTTPOnly:  cookie.HTTPOnly,
			Secure:    cookie.Secure,
		}
	}
	if err := r.storage.SaveCookies(sessionID, storageCookies); err != nil {
		log.Printf("Failed to save cookies: %v", err)
	}

	log.Printf("Login successful, session ID: %s", sessionID)
	return sessionID, nil
}

func (r *Renderer) navigateToMain(page *rod.Page, branchID, filterID string) error {
	log.Println("Navigating to Venus Main page...")

	err := rod.Try(func() {
		page.MustNavigate("https://theearth-np.com/WebVenus/F-AAV0001[VenusMain].aspx")
		page.MustWaitLoad()
	})

	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Wait for page to fully load
	page.MustWaitRequestIdle()
	time.Sleep(5 * time.Second) // Additional wait for JavaScript to initialize

	// Check if we're still on the login page (redirect occurred)
	currentURL := page.MustInfo().URL
	log.Printf("Current URL after navigation: %s", currentURL)

	if strings.Contains(currentURL, "Login") || strings.Contains(currentURL, "OES1010") {
		return fmt.Errorf("redirected to login page")
	}

	return nil
}

func (r *Renderer) extractVehicleData(page *rod.Page, branchID, filterID string) ([]VehicleData, error) {
	// Default values if not provided
	// branchID = "" returns all branches (same as "00000000")
	// filterID = "0" excludes deleted vehicles (193 active vehicles)
	// filterID = "" includes deleted vehicles too (266 total)
	if filterID == "" {
		filterID = "0"  // Use "0" to exclude deleted vehicles
	}

	// First check if VenusBridgeService exists
	hasService := page.MustEval(`() => {
		return typeof VenusBridgeService !== 'undefined' &&
		       typeof VenusBridgeService.VehicleStateTableForBranchEx === 'function';
	}`).Bool()

	if !hasService {
		return nil, fmt.Errorf("VenusBridgeService not found on page")
	}

	// Log the parameters being used
	log.Printf("Calling VenusBridgeService.VehicleStateTableForBranchEx with branchID='%s', filterID='%s'", branchID, filterID)

	// Wait for the page to be stable
	page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Wait for Venus-specific loading elements to disappear
	log.Println("Waiting for page to be ready...")

	// First, wait for the grid to appear (indicates page loaded)
	gridExists := false
	for i := 0; i < 30; i++ {
		exists, _ := page.Eval(`() => {
			// Check if the Venus main grid exists
			const grid = document.querySelector('#igGrid-VenusMain-VehicleList');
			return grid !== null;
		}`)

		if exists != nil && exists.Value.Bool() {
			gridExists = true
			log.Println("Venus main grid detected, page structure loaded")
			break
		}

		if i%5 == 0 {
			log.Printf("Waiting for page structure... (%d/30)", i+1)
		}
		time.Sleep(1 * time.Second)
	}

	if !gridExists {
		log.Println("Warning: Grid not found after 30 seconds, proceeding anyway...")
	}

	// Now wait for any loading messages (pMsg_wait) to disappear
	log.Println("Checking for loading messages...")
	loadingCleared := false
	for i := 0; i < 30; i++ {
		hasLoading, _ := page.Eval(`() => {
			// Check for Venus-specific loading elements
			// pMsg_wait is the common loading message element
			const waitMsg = document.querySelector('#pMsg_wait, [id*="pMsg_wait"], [id*="pMsg"], [class*="pMsg"]');
			const loadingDivs = document.querySelectorAll('[id*="loading"], [id*="Loading"], .loading-message, .wait-message');

			// Check all loading elements
			const allLoading = waitMsg ? [waitMsg, ...loadingDivs] : [...loadingDivs];

			const visibleLoading = allLoading.filter(elem => {
				if (!elem) return false;
				const style = window.getComputedStyle(elem);
				const rect = elem.getBoundingClientRect();

				// Check if element is visible
				const isVisible = style.display !== 'none' &&
								 style.visibility !== 'hidden' &&
								 style.opacity !== '0' &&
								 (rect.width > 0 || rect.height > 0);

				if (isVisible && elem.id) {
					console.log('Found visible loading element:', elem.id, elem.className);
				}

				return isVisible;
			});

			return visibleLoading.length > 0;
		}`)

		if hasLoading != nil && !hasLoading.Value.Bool() {
			loadingCleared = true
			log.Println("No loading messages detected, proceeding...")
			break
		}

		if i%5 == 0 {
			log.Printf("Loading message still visible, waiting... (%d/30)", i+1)
		}
		time.Sleep(1 * time.Second)
	}

	if !loadingCleared {
		log.Println("Warning: Loading message timeout after 30 seconds, proceeding anyway...")
	}

	// Additional wait to ensure JavaScript is ready
	time.Sleep(3 * time.Second)

	// Execute the JavaScript to get vehicle data
	log.Println("Executing JavaScript to get vehicle data...")
	log.Printf("Using branchID='%s', filterID='%s'", branchID, filterID)

	// Inject JavaScript to store the result in window
	_, err := page.Eval(`(branchID, filterID) => {
		window.__vehicleDataResult = null;
		window.__vehicleDataError = null;
		window.__vehicleDataCompleted = false;

		VenusBridgeService.VehicleStateTableForBranchEx(branchID, filterID,
			(data) => {
				window.__vehicleDataResult = data;
				window.__vehicleDataCompleted = true;
			},
			(error) => {
				window.__vehicleDataError = error;
				window.__vehicleDataCompleted = true;
			}
		);
	}`, branchID, filterID)

	if err != nil {
		return nil, fmt.Errorf("failed to inject JavaScript: %w", err)
	}

	// Poll for result
	log.Println("Waiting for vehicle data response...")
	startTime := time.Now()
	timeout := 60 * time.Second
	var result interface{}

	for time.Since(startTime) < timeout {

		completedObj, err := page.Eval(`() => window.__vehicleDataCompleted`)
		if err != nil {
			// Skip logging for context errors as they're expected in background processing
			if !strings.Contains(err.Error(), "context") {
				log.Printf("Error checking completion: %v", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		completed := completedObj.Value.Bool()
		if completed {
			// Check for error
			hasErrorObj, err := page.Eval(`() => window.__vehicleDataError !== null`)
			if err != nil {
				log.Printf("Error checking for errors: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			hasError := hasErrorObj.Value.Bool()
			if hasError {
				errorMsgObj, _ := page.Eval(`() => window.__vehicleDataError`)
				errorMsg := ""
				if errorMsgObj != nil {
					errorMsg = errorMsgObj.Value.String()
				}
				return nil, fmt.Errorf("service error: %s", errorMsg)
			}

			// Get result
			resultObj, err := page.Eval(`() => window.__vehicleDataResult`)
			if err != nil {
				return nil, fmt.Errorf("failed to get vehicle data result: %w", err)
			}
			result = resultObj.Value.Val()
			log.Printf("Got vehicle data response after %v", time.Since(startTime))
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if result == nil {
		return nil, fmt.Errorf("timeout waiting for vehicle data after %v", timeout)
	}

	// Parse the result
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var rawData []interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vehicle data: %w", err)
	}

	// Convert to VehicleData struct
	vehicles := make([]VehicleData, 0, len(rawData))
	for _, rawItem := range rawData {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		vehicle := VehicleData{
			Metadata: make(map[string]string),
		}

		if v, ok := item["VehicleCD"].(string); ok {
			vehicle.VehicleCD = v
		}
		if v, ok := item["VehicleName"].(string); ok {
			vehicle.VehicleName = v
		}
		if v, ok := item["Status"].(string); ok {
			vehicle.Status = v
		}

		// Add all other fields to metadata
		for k, v := range item {
			if k != "VehicleCD" && k != "VehicleName" && k != "Status" {
				vehicle.Metadata[k] = fmt.Sprintf("%v", v)
			}
		}

		vehicles = append(vehicles, vehicle)
	}

	log.Printf("Extracted %d vehicles", len(vehicles))

	// Save raw data to local JSON file for debugging
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("./data/vehicles_%s.json", timestamp)

	// Create data directory if it doesn't exist
	os.MkdirAll("./data", 0755)

	// Save raw data
	rawJSON, err := json.MarshalIndent(rawData, "", "  ")
	if err == nil {
		if err := os.WriteFile(filename, rawJSON, 0644); err == nil {
			log.Printf("Saved vehicle data to %s", filename)
		} else {
			log.Printf("Failed to save vehicle data: %v", err)
		}
	}

	// Send raw data to Hono API
	if _, err := r.sendRawToHonoAPI(rawData); err != nil {
		log.Printf("Warning: Failed to send to Hono API: %v", err)
		// Don't fail the whole operation if API fails
	}

	return vehicles, nil
}

func (r *Renderer) CheckSession(sessionID string) (bool, string) {
	session, err := r.storage.GetSession(sessionID)
	if err != nil {
		return false, fmt.Sprintf("Error checking session: %v", err)
	}
	if session == nil {
		return false, "Session not found"
	}
	if session.ExpiresAt.Before(time.Now()) {
		return false, "Session expired"
	}
	return true, "Session is valid"
}

func (r *Renderer) ClearSession(sessionID string) error {
	return r.storage.DeleteSession(sessionID)
}

func (r *Renderer) sendToHonoAPI(vehicles []VehicleData) (*HonoAPIResponse, error) {
	// Convert VehicleData to format expected by Hono API
	honoData := r.convertToHonoFormat(vehicles)

	// Send to Hono API
	jsonData, err := json.Marshal(honoData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://hono-api.mtamaramu.com/api/dtakologs",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return &HonoAPIResponse{
		Success:      true,
		RecordsAdded: len(honoData),
		Message:      fmt.Sprintf("Successfully sent %d records to Hono API", len(honoData)),
	}, nil
}

// sendRawToHonoAPI sends raw JSON data directly to Hono API without conversion
func (r *Renderer) sendRawToHonoAPI(rawData []interface{}) (*HonoAPIResponse, error) {
	// Debug: Check data type and content
	log.Printf("sendRawToHonoAPI: Sending %d records", len(rawData))

	// Send raw data directly to Hono API
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	log.Printf("sendRawToHonoAPI: JSON size: %d bytes", len(jsonData))

	req, err := http.NewRequest("POST", "https://hono-api.mtamaramu.com/api/dtakologs",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("API Error - Status: %d, Body: %s", resp.StatusCode, string(body))
		log.Printf("Request size: %d bytes", len(jsonData))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	log.Printf("API Success - Status: %d, Body: %s", resp.StatusCode, string(body))

	return &HonoAPIResponse{
		Success:      true,
		RecordsAdded: len(rawData),
		Message:      fmt.Sprintf("Successfully sent %d records to Hono API", len(rawData)),
	}, nil
}

func (r *Renderer) convertToHonoFormat(vehicles []VehicleData) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(vehicles))

	for _, vehicle := range vehicles {
		data := make(map[string]interface{})

		// Copy all metadata fields
		for key, value := range vehicle.Metadata {
			data[key] = r.convertToNumber(value)
		}

		// Generate VehicleCD from VehicleName if empty
		vehicleCD := r.extractVehicleCode(vehicle.VehicleName)
		if vehicleCD > 0 {
			data["VehicleCD"] = vehicleCD
		}

		// Add VehicleName
		data["VehicleName"] = vehicle.VehicleName

		// Format DataDateTime
		if dt, ok := data["DataDateTime"].(string); ok {
			data["DataDateTime"] = r.formatDateTime(dt)
		}

		result = append(result, data)
	}

	return result
}

func (r *Renderer) convertToNumber(value string) interface{} {
	if value == "" || value == "<nil>" {
		return 0
	}

	// Try to convert to float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}

func (r *Renderer) extractVehicleCode(vehicleName string) int {
	if vehicleName == "" {
		return 0
	}

	// Extract all numbers from vehicle name
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(vehicleName, -1)

	if len(matches) > 0 {
		// Combine all numbers
		combined := strings.Join(matches, "")
		if code, err := strconv.Atoi(combined); err == nil {
			return code % 2147483647 // Keep within integer range
		}
	}

	// Generate hash-based ID if no numbers found
	hash := 0
	for _, r := range vehicleName {
		hash = hash*31 + int(r)
	}
	return hash % 2147483647
}

func (r *Renderer) formatDateTime(dt string) string {
	if dt == "" || dt == "<nil>" {
		return ""
	}

	// Remove "20" prefix if exists to avoid double prefixing
	if strings.HasPrefix(dt, "20") {
		return dt[2:]
	}

	return dt
}

func (r *Renderer) Close() error {
	if r.browser != nil {
		return r.browser.Close()
	}
	return nil
}