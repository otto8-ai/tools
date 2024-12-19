package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/pkg/browser"
)

type oauthResponse struct {
	TokenType    string            `json:"token_type"`
	Scope        string            `json:"scope"`
	ExpiresIn    int               `json:"expires_in"`
	ExtExpiresIn int               `json:"ext_expires_in"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	Extras       map[string]string `json:"extras"`
}

type cred struct {
	Env          map[string]string `json:"env"`
	ExpiresAt    *time.Time        `json:"expiresAt"`
	RefreshToken string            `json:"refreshToken"`
}

type oauthConfig struct {
	ServerURL string `json:"serverURL,omitempty"`
	ToName    string `json:"toName,omitempty"`
}

type cliConfig struct {
	ServerURL string                 `json:"serverURL,omitempty"`
	Mappings  map[string]oauthConfig `json:"mappings,omitempty"`
}

var (
	integration   = os.Getenv("INTEGRATION")
	token         = os.Getenv("TOKEN")
	scope         = os.Getenv("SCOPE")
	optionalScope = os.Getenv("OPTIONAL_SCOPE")
)

const publicGatewayURL = "https://gateway-api.gptscript.ai"

func normalizeForEnv(appName string) string {
	return strings.ToUpper(strings.ReplaceAll(appName, "-", "_"))
}

func getURLs(appName string) (string, string, string, error) {
	var (
		authorizeURL = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_AUTH_URL", normalizeForEnv(appName)))
		refreshURL   = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_REFRESH_URL", normalizeForEnv(appName)))
		tokenURL     = os.Getenv(fmt.Sprintf("GPTSCRIPT_OAUTH_%s_TOKEN_URL", normalizeForEnv(appName)))
		err          error
	)

	if authorizeURL != "" && refreshURL != "" && tokenURL != "" {
		return authorizeURL, refreshURL, tokenURL, nil
	}

	configPath := os.Getenv("GPTSCRIPT_OAUTH_CONFIG")
	if configPath == "" {
		configPath, err = xdg.ConfigFile("gptscript/oauth.json")
		if err != nil {
			return "", "", "", fmt.Errorf("getURLs: failed to get config file: %w", err)
		}
	}

	var cfg cliConfig
	if cfgBytes, err := os.ReadFile(configPath); errors.Is(err, fs.ErrNotExist) {
	} else if err != nil {
		return "", "", "", fmt.Errorf("getURLs: failed to read config file: %w", err)
	} else {
		if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
			return "", "", "", fmt.Errorf("getURLs: failed to unmarshal config: %w", err)
		}
	}

	mapping := cfg.Mappings[integration]
	if mapping.ServerURL == "" {
		mapping.ServerURL = cfg.ServerURL
	}
	if mapping.ServerURL == "" {
		mapping.ServerURL = publicGatewayURL
	}

	if mapping.ToName == "" {
		mapping.ToName = integration
	}

	authorizeURL = fmt.Sprintf("%s/oauth-apps/%s/authorize", mapping.ServerURL, mapping.ToName)
	refreshURL = fmt.Sprintf("%s/oauth-apps/%s/refresh", mapping.ServerURL, mapping.ToName)
	tokenURL = fmt.Sprintf("%s/api/oauth-apps/get-token", mapping.ServerURL)

	return authorizeURL, refreshURL, tokenURL, nil
}

func main() {
	authorizeURL, refreshURL, tokenURL, err := getURLs(integration)
	if err != nil {
		fmt.Printf("main: failed to get URLs: %v\n", err)
		os.Exit(1)
	}

	// Refresh existing credential if there is one.
	existing := os.Getenv("GPTSCRIPT_EXISTING_CREDENTIAL")
	if existing != "" {
		var c cred
		if err := json.Unmarshal([]byte(existing), &c); err != nil {
			fmt.Printf("main: failed to unmarshal existing credential: %v\n", err)
			os.Exit(1)
		}

		u, err := url.Parse(refreshURL)
		if err != nil {
			fmt.Printf("main: failed to parse refresh URL: %v\n", err)
			os.Exit(1)
		}

		q := u.Query()
		q.Set("refresh_token", c.RefreshToken)
		if scope != "" {
			q.Set("scope", strings.Join(strings.Fields(scope), " "))
		}
		if optionalScope != "" {
			q.Set("optional_scope", optionalScope)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			fmt.Printf("main: failed to create refresh request: %v\n", err)
			os.Exit(1)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("main: failed to send refresh request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("main: unexpected status code from refresh request: %d\n", resp.StatusCode)
			os.Exit(1)
		}

		var oauthResp oauthResponse
		if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
			fmt.Printf("main: failed to decode refresh response JSON: %v\n", err)
			os.Exit(1)
		}

		envVars := map[string]string{
			token: oauthResp.AccessToken,
		}

		for k, v := range oauthResp.Extras {
			envVars[k] = v
		}

		out := cred{
			Env:          envVars,
			RefreshToken: oauthResp.RefreshToken,
		}

		if oauthResp.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Second * time.Duration(oauthResp.ExpiresIn))
			out.ExpiresAt = &expiresAt
		}

		credJSON, err := json.Marshal(out)
		if err != nil {
			fmt.Printf("main: failed to marshal refreshed credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(string(credJSON))
		return
	}

	state, err := generateString()
	if err != nil {
		fmt.Printf("main: failed to generate state: %v\n", err)
		os.Exit(1)
	}

	verifier, err := generateString()
	if err != nil {
		fmt.Printf("main: failed to generate verifier: %v\n", err)
		os.Exit(1)
	}

	h := sha256.New()
	h.Write([]byte(verifier))
	challenge := hex.EncodeToString(h.Sum(nil))

	u, err := url.Parse(authorizeURL)
	if err != nil {
		fmt.Printf("main: failed to parse authorize URL: %v\n", err)
		os.Exit(1)
	}

	q := u.Query()
	q.Set("state", state)
	q.Set("challenge", challenge)
	if scope != "" {
		q.Set("scope", strings.Join(strings.Fields(scope), " "))
	}
	if optionalScope != "" {
		q.Set("optional_scope", optionalScope)
	}
	u.RawQuery = q.Encode()

	gs, err := gptscript.NewGPTScript(gptscript.GlobalOptions{})
	if err != nil {
		fmt.Printf("main: failed to create GPTScript: %v\n", err)
		os.Exit(1)
	}

	metadata := map[string]string{
		"authType":        "oauth",
		"toolContext":     "credential",
		"toolDisplayName": fmt.Sprintf("%s%s Integration", strings.ToTitle(integration[:1]), integration[1:]),
		"authURL":         u.String(),
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		fmt.Printf("main: failed to marshal metadata: %v\n", err)
		os.Exit(1)
	}

	run, err := gs.Run(context.Background(), "sys.prompt", gptscript.Options{
		Input: fmt.Sprintf(`{"metadata":%s,"message":%q}`, b, fmt.Sprintf("To authenticate please open your browser to %s.", u.String())),
	})
	if err != nil {
		fmt.Printf("main: failed to run sys.prompt: %v\n", err)
		os.Exit(1)
	}

	out, err := run.Text()
	if err != nil {
		fmt.Printf("main: failed to get text from sys.prompt: %v\n", err)
		//os.Exit(1)
	}

	var m map[string]string
	_ = json.Unmarshal([]byte(out), &m)

	if m["handled"] != "true" {
		// Don't let the browser library print anything.
		browser.Stdout = io.Discard

		// Open the user's browser so that they can authorize the app.
		_ = browser.OpenURL(u.String())
	}

	t := time.NewTicker(2 * time.Second)
	for range t.C {
		// Construct the request to get the token from the gateway.
		req, err := http.NewRequest("GET", tokenURL, nil)
		if err != nil {
			fmt.Printf("main: failed to create token request: %v\n", err)
			os.Exit(1)
		}

		q = req.URL.Query()
		q.Set("state", state)
		q.Set("verifier", verifier)
		req.URL.RawQuery = q.Encode()

		// Send the request to the gateway.
		now := time.Now()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "main: failed to send token request: %v\n", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_, _ = fmt.Fprintf(os.Stderr, "main: unexpected status code from token request: %d\n", resp.StatusCode)
			continue
		}

		// Parse the response from the gateway.
		var oauthResp oauthResponse
		if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
			fmt.Printf("main: failed to decode token response JSON: %v\n", err)
			_ = resp.Body.Close()
			os.Exit(1)
		}
		_ = resp.Body.Close()

		envVars := map[string]string{
			token: oauthResp.AccessToken,
		}

		for k, v := range oauthResp.Extras {
			envVars[k] = v
		}

		out := cred{
			Env:          envVars,
			RefreshToken: oauthResp.RefreshToken,
		}

		if oauthResp.ExpiresIn > 0 {
			expiresAt := now.Add(time.Second * time.Duration(oauthResp.ExpiresIn))
			out.ExpiresAt = &expiresAt
		}

		credJSON, err := json.Marshal(out)
		if err != nil {
			fmt.Printf("main: failed to marshal token credential: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(string(credJSON))
		os.Exit(0)
	}
}

func generateString() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 256)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateString: failed to read random bytes: %w", err)
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}
