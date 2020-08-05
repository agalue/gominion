package tools

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GetHTTPClient returns an HTTP Client with a given timeout, and transport
func GetHTTPClient(useSSLFilter bool, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: useSSLFilter},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// ParseHTTPResponseRange parses a given response range
func ParseHTTPResponseRange(responseRange string) (min int, max int) {
	var err error
	rangeArray := strings.Split(responseRange, "-")
	if len(rangeArray) < 2 {
		return 100, 399
	}
	if min, err = strconv.Atoi(rangeArray[0]); err == nil {
		min = 100
	}
	if max, err = strconv.Atoi(rangeArray[1]); err == nil {
		max = 399
	}
	return min, max
}
