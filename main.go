package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	notificationEndpoint = "https://ntfy.sh/dapidi_alerts"
)

func getNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = os.Getenv("HOSTNAME")
	}
	if nodeName == "" {
		nodeName = "unknown"
	}
	return nodeName
}

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
	nodeName := getNodeName()
	message := fmt.Sprintf("%s is online", nodeName)

	if err := notify(message); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Notification sent successfully")
}
