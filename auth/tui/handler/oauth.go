package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// Constants for GitHub OAuth process
const (
	ClerkAuthURL              = "https://clerk.projectdiscovery.io/v1/client/sign_ins"
	AuthCallback              = "https://cloud.projectdiscovery.io/sign-in/sso-callback"
	Strategy                  = "oauth_github"
	ActionCompleteRedirectURL = "/"
	ContentTypeJSON           = "application/json"
	RequestTimeout            = 10 * time.Second
	RedirectHost              = "localhost"
	RedirectPort              = "8085" // Port where our local server will listen for OAuth callback
	//RedirectURI              = "http://" + RedirectHost + ":" + RedirectPort + "/callback"
	RedirectURI    = "https://cloud.projectdiscovery.io?login_mode=cli"
	ClerkJSVersion = "5.21.2"
)

// Error messages
const (
	ErrEncodingJSON     = "error encoding JSON"
	ErrCreatingRequest  = "error creating request"
	ErrMakingRequest    = "error making request"
	ErrDecodingResponse = "error decoding response"
	ErrAPIKeyNotFound   = "API key not found in response"
)

// SignInResponse represents the JSON structure returned by Clerk
type SignInResponse struct {
	Response struct {
		FirstFactorVerification struct {
			ExternalVerificationRedirectURL string `json:"external_verification_redirect_url"`
		} `json:"first_factor_verification"`
	} `json:"response"`
}

var wg sync.WaitGroup

// initiateGitHubOAuth sends a POST request to initiate the GitHub OAuth process.
func InitiateGitHubOAuth() (map[string]interface{}, error) {

	proxyURL, _ := url.Parse("http://localhost:8080")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	// Create a cookie jar to handle cookies
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:       jar, // Attach the jar to the client
		Transport: transport,
	}

	// Prepare POST data (form data)
	data := url.Values{}
	data.Set("strategy", "oauth_github")
	data.Set("redirect_url", AuthCallback)
	data.Set("action_complete_redirect_url", "https://www.google.com")

	// Send POST request to Clerk
	req, err := http.NewRequest("POST", ClerkAuthURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://cloud.projectdiscovery.io")
	req.Header.Set("Referer", "https://cloud.projectdiscovery.io")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var signInResponse SignInResponse
	if err := json.Unmarshal(body, &signInResponse); err != nil {
		return nil, err
	}

	fmt.Println("response:", signInResponse.Response.FirstFactorVerification.ExternalVerificationRedirectURL)
	githubOAuthURL := signInResponse.Response.FirstFactorVerification.ExternalVerificationRedirectURL
	u, err := url.Parse(githubOAuthURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}
	params := u.Query()
	params.Set("redirect_uri", fmt.Sprintf("http://127.0.0.1:%s/callback", RedirectPort))
	u.RawQuery = params.Encode()
	githubOAuthURL = u.String()
	err = openBrowser(githubOAuthURL)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	wg.Add(1)

	go func() {
		// Start a local server to listen for the OAuth callback
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			// Read the response body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading response", http.StatusInternalServerError)
				return
			}

			fmt.Println("Response body:", string(body))
			wg.Done()
		})

		if http.ListenAndServe(":"+RedirectPort, nil); err != nil {
			log.Fatalf("Failed to listen %s", err)
		}
	}()
	wg.Wait()
	return nil, nil
}

// openBrowser opens the default web browser for the user to log in.
func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
