package misc

import (
	"fmt"
	"net/http"
	"strconv"
)

type spotifyAPIError struct {
	StatusCode int
	Response   *http.Response
}

func (e *spotifyAPIError) Error() string {
	return fmt.Sprintf("Spotify API error: %d %s", e.StatusCode, e.Response.Status)
}

func isRateLimitError(err error) bool {
	if respErr, ok := err.(*spotifyAPIError); ok && respErr.StatusCode == 429 {
		return true
	}
	return false
}

func getRetryAfterValue(err error) (int, error) {
	if respErr, ok := err.(*spotifyAPIError); ok {
		retryAfterStr := respErr.Response.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			return 0, err
		}
		return retryAfter, nil
	}
	return 0, fmt.Errorf("error is not a Spotify API error")
}
