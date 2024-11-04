package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/gurkengewuerz/GitCodeJudge/internal/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

var (
	oauthCfg   *oauth2.Config
	oidcConfig *OpenIDConfiguration
)

type OpenIDConfiguration struct {
	Issuer                 string   `json:"issuer"`
	AuthorizationEndpoint  string   `json:"authorization_endpoint"`
	TokenEndpoint          string   `json:"token_endpoint"`
	UserinfoEndpoint       string   `json:"userinfo_endpoint"`
	JwksURI                string   `json:"jwks_uri"`
	ScopesSupported        []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

func fetchOpenIDConfiguration(issuerURL string) (*OpenIDConfiguration, error) {
	// Ensure the issuer URL ends with a slash
	if !strings.HasSuffix(issuerURL, "/") {
		issuerURL += "/"
	}

	wellKnownURL := issuerURL + ".well-known/openid-configuration"

	resp, err := http.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenID configuration: %v", err)
	}
	defer resp.Body.Close()

	var config OpenIDConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode OpenID configuration: %v", err)
	}

	return &config, nil
}

func InitOAuth2(cfg *config.Config) error {
	if cfg.OAuth2Issuer == "" {
		return nil
	}

	var err error
	oidcConfig, err = fetchOpenIDConfiguration(cfg.OAuth2Issuer)
	if err != nil {
		return fmt.Errorf("failed to initialize OAuth2: %v", err)
	}

	// Determine scopes
	defaultScopes := []string{"openid", "profile", "email"}
	scopes := make([]string, 0)
	for _, scope := range defaultScopes {
		if containsScope(oidcConfig.ScopesSupported, scope) {
			scopes = append(scopes, scope)
		}
	}

	oauthCfg = &oauth2.Config{
		ClientID:     cfg.OAuth2ClientID,
		ClientSecret: cfg.OAuth2Secret,
		RedirectURL:  fmt.Sprintf("%s/auth/callback", cfg.BaseURL),
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oidcConfig.AuthorizationEndpoint,
			TokenURL: oidcConfig.TokenEndpoint,
		},
	}

	log.WithFields(log.Fields{
		"issuer": oidcConfig.Issuer,
		"scopes": scopes,
	}).Info("OAuth2 initialized successfully")

	return nil
}

func containsScope(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func RequireAuth(cfg *config.Config) fiber.Handler {

	return func(c fiber.Ctx) error {
		if cfg.OAuth2Issuer == "" {
			return c.Next()
		}

		sess := session.FromContext(c)

		user := sess.Get("user")
		if user == nil {
			return c.Redirect().To("/auth/login")
		}

		return c.Next()
	}
}

func HandleLogin(c fiber.Ctx) error {
	if config.CFG.OAuth2Issuer == "" {
		return c.Redirect().To("/leaderboard")
	}

	state := generateRandomState() // Implement this helper function

	sess := session.FromContext(c)
	sess.Set("oauth2_state", state)

	url := oauthCfg.AuthCodeURL(state)
	return c.Redirect().To(url)
}

func HandleCallback(c fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state if it was saved in session
	sess := session.FromContext(c)

	savedState := sess.Get("oauth2_state")
	if savedState != nil && savedState.(string) != state {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}
	sess.Delete("oauth2_state")

	token, err := oauthCfg.Exchange(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	client := oauthCfg.Client(c.Context(), token)
	resp, err := client.Get(oidcConfig.UserinfoEndpoint)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode user info",
		})
	}

	sess.Set("user", userInfo["email"])

	return c.Redirect().To("/leaderboard")
}

func HandleLogout(c fiber.Ctx) error {
	sess := session.FromContext(c)

	sess.Delete("user")
	return c.Redirect().To("/")
}

func generateRandomState() string {
	random, err := uuid.NewRandom()
	if err != nil {
		log.WithError(err).Error("Failed to generate random state")
		return ""
	}
	return "state-" + random.String()
}
