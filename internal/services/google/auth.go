package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/zepzeper/vulgar/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	callbackPort = "8085"
	callbackPath = "/callback"
)

var scopes = []string{
	drive.DriveScope,               // Full access to Drive (read/write)
	sheets.SpreadsheetsScope,       // Full access to Sheets (read/write)
	calendar.CalendarScope,         // Full access to Calendar (read/write)
	"https://www.googleapis.com/auth/userinfo.email",
}

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func GetOAuthConfig() (*oauth2.Config, error) {
	clientID, clientSecret, ok := config.GetGoogleOAuthCredentials()
	if !ok {
		return nil, fmt.Errorf("google OAuth credentials not configured: set client_id and client_secret in %s", config.ConfigPath())
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("http://localhost:%s%s", callbackPort, callbackPath),
	}, nil
}

func LoadToken() (*oauth2.Token, error) {
	tokenPath := config.GoogleTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func SaveToken(token *oauth2.Token) error {
	tokenPath := config.GoogleTokenPath()
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}

func IsLoggedIn() bool {
	token, err := LoadToken()
	if err != nil {
		return false
	}
	return token.Valid() || token.RefreshToken != ""
}

func GetClient(ctx context.Context) (*http.Client, error) {
	oauthConfig, err := GetOAuthConfig()
	if err != nil {
		return nil, err
	}

	token, err := LoadToken()
	if err != nil {
		return nil, fmt.Errorf("not logged in: run 'vulgar gdrive login' first")
	}

	tokenSource := oauthConfig.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w (try 'vulgar gdrive login' again)", err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := SaveToken(newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

func Login(ctx context.Context) (*oauth2.Token, error) {
	oauthConfig, err := GetOAuthConfig()
	if err != nil {
		return nil, err
	}

	state := fmt.Sprintf("%d", time.Now().UnixNano())

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{Addr: ":" + callbackPort}

	http.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>vulgar - Google Authentication</title></head>
<body style="font-family: system-ui; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #1a1a2e;">
<div style="text-align: center; color: #eee;">
<h1 style="color: #7C3AED;">Authentication Successful</h1>
<p>You can close this window and return to the terminal.</p>
</div>
</body>
</html>`)

		codeChan <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	if err := openBrowser(authURL); err != nil {
		server.Shutdown(ctx)
		return nil, fmt.Errorf("failed to open browser: %w\n\nManually open this URL:\n%s", err, authURL)
	}

	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Shutdown(ctx)
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return nil, fmt.Errorf("authentication timed out")
	}

	server.Shutdown(ctx)

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

func Logout() error {
	tokenPath := config.GoogleTokenPath()
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func GetUserInfo(ctx context.Context) (*UserInfo, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	oauth2Service, err := googleoauth2.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 service: %w", err)
	}

	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &UserInfo{
		Email: userinfo.Email,
		Name:  userinfo.Name,
	}, nil
}

func NewDriveService(ctx context.Context) (*drive.Service, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return drive.NewService(ctx, option.WithHTTPClient(client))
}

func NewSheetsService(ctx context.Context) (*sheets.Service, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

func NewCalendarService(ctx context.Context) (*calendar.Service, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
