package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	notificationEndpoint = "https://ntfy.sh/dapidi_alerts"
)

func notify(message string) error {
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	payload := strings.NewReader(message)
	resp, err := http.Post(notificationEndpoint, "text/plain", payload)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	if err := notify("Hi"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Notification sent successfully")
}
