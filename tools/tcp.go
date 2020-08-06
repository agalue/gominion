package tools

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// NetMessageContains reads message from network connection and verify if contains banner
func NetMessageContains(conn net.Conn, timeout time.Duration, banner string) (bool, error) {
	if banner == "" || banner == "*" {
		return true, nil
	}
	conn.SetReadDeadline(time.Now().Add(timeout))
	payload := make([]byte, 1024)
	size, err := conn.Read(payload)
	if err != nil {
		return false, err
	}
	if size == 0 {
		return false, fmt.Errorf("empty response received")
	}
	payloadCut := make([]byte, size)
	copy(payloadCut, payload[0:size])
	received := string(payloadCut)
	if strings.HasPrefix(banner, "~") {
		result := string([]rune(banner)[1:])
		exp, err := regexp.Compile(result)
		if err == nil {
			if exp.MatchString(received) {
				return true, nil
			}
			return false, fmt.Errorf("response %s doesn't match banner %s", received, banner)
		}
		return false, err
	}
	if strings.Contains(received, banner) {
		return true, nil
	}
	return false, fmt.Errorf("response %s doesn't contain banner %s", received, banner)
}
