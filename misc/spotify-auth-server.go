package misc

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2" // Import the resty package
	"github.com/tidwall/gjson"     // Import the gjson package
)

var (
	redirectURI  = "http://localhost:3000"
	authorizeURL = "https://accounts.spotify.com/authorize"
	tokenURL     = "https://accounts.spotify.com/api/token"
)

func RunSpotifyAuthServer() {
	config, err := LoadConfig("../");
	if err != nil {
		log.Fatal(err)
	}
	r := gin.Default()

	r.GET("/login", func(c *gin.Context) {
		state, err := generateRandomString(16)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		scope := "user-read-playback-state user-modify-playback-state user-read-currently-playing app-remote-control streaming playlist-read-private playlist-read-collaborative playlist-modify-private playlist-modify-public user-read-playback-position user-top-read user-read-recently-played"

		url := fmt.Sprintf("%s?%s", authorizeURL, buildQueryParams(config, state, scope))
		c.Redirect(http.StatusTemporaryRedirect, url)
	})

	r.GET("/", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
	
		if state == "" {
			c.String(http.StatusBadRequest, "State parameter is missing")
			return
		}
	
		if code == "" {
			c.String(http.StatusBadRequest, "Code parameter is missing")
			return
		}
	
		// Exchange authorization code for access token
		resp, err := resty.New().R().
			SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(config.SpotifyClientId+":"+config.SpotifyClientSecret))).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetBody(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, redirectURI)).
			SetResult(map[string]interface{}{}).
			Post(tokenURL)
	
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	
		if resp.StatusCode() != http.StatusOK {
			c.String(resp.StatusCode(), string(resp.Body()))
			return
		}
	
		// Successful token exchange
		json := resp.Body()
		accessToken := gjson.Get(string(json), "access_token").String()
		tokenType := gjson.Get(string(json), "token_type").String()
		expiresIn := gjson.Get(string(json), "expires_in").Int()
	
		// Use the access token in your application
		// ...
	
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
			"token_type":   tokenType,
			"expires_in":   expiresIn,
		})
	})

	r.Run(":3000")
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func buildQueryParams(config Config, state string, scope string) string {
	params := map[string]string{
		"response_type": "code",
		"client_id":     config.SpotifyClientId,
		"scope":         scope,
		"redirect_uri":  redirectURI,
		"state":         state,
	}

	var parts []string
	for key, value := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(parts, "&")
}
