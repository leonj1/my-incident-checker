package network

import (
	"fmt"
	"net/http"
	"time"
)

const (
	connectivityCheck = "https://www.google.com"
	connectTimeout    = 10 * time.Second
)

// CheckConnectivity verifies internet connectivity by making a request to a known endpoint
func CheckConnectivity() error {
	client := &http.Client{
		Timeout: connectTimeout,
	}

	resp, err := client.Get(connectivityCheck)
	if err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connectivity check failed with status code: %d", resp.StatusCode)
	}

	return nil
}
