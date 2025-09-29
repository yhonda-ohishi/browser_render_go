package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yourusername/browser_render_go/src/config"
	"github.com/yourusername/browser_render_go/src/storage"
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

func (r *Renderer) GetVehicleData(_ context.Context, sessionID, branchID, filterID string, forceLogin bool) ([]VehicleData, string, error) {
	log.Println("GetVehicleData called")
	log.Printf("Parameters - SessionID: %s, BranchID: %s, FilterID: %s, ForceLogin: %v", sessionID, branchID, filterID, forceLogin)

	page := r.browser.MustPage()
	defer page.MustClose()

	// Set timeout
	page = page.Timeout(r.config.BrowserTimeout)

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
			return nil, "", fmt.Errorf("login failed: %w", err)
		}
		sessionID = newSessionID
		log.Printf("Login successful, new session ID: %s", sessionID)

		// Navigate again after login
		if err := r.navigateToMain(page, branchID, filterID); err != nil {
			return nil, "", fmt.Errorf("navigation failed after login: %w", err)
		}
		log.Println("Navigation to main page successful after login")
	} else {
		log.Println("Navigation to main page successful without login")
	}

	// Extract vehicle data
	vehicleData, err := r.extractVehicleData(page, branchID, filterID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract vehicle data: %w", err)
	}

	// Cache the data
	for _, vehicle := range vehicleData {
		r.storage.CacheVehicleData(vehicle.VehicleCD, vehicle, 5*time.Minute)
	}

	return vehicleData, sessionID, nil
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
		completed := page.MustEval(`() => window.__vehicleDataCompleted`).Bool()
		if completed {
			// Check for error
			hasError := page.MustEval(`() => window.__vehicleDataError !== null`).Bool()
			if hasError {
				errorMsg := page.MustEval(`() => window.__vehicleDataError`).String()
				return nil, fmt.Errorf("service error: %s", errorMsg)
			}

			// Get result
			result = page.MustEval(`() => window.__vehicleDataResult`).Val()
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

	var rawData []map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vehicle data: %w", err)
	}

	// Convert to VehicleData struct
	vehicles := make([]VehicleData, 0, len(rawData))
	for _, item := range rawData {
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

func (r *Renderer) Close() error {
	if r.browser != nil {
		return r.browser.Close()
	}
	return nil
}