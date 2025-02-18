package heartbeat

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	heartbeatEndpoint = "https://nosnch.in/2b7bdbea9e"
	heartbeatInterval = 5 * time.Minute
)

// sendHeartbeat sends a heartbeat signal to the monitoring service
func sendHeartbeat() error {
	payload := strings.NewReader("m=just checking in")
	resp, err := http.Post(heartbeatEndpoint, "application/x-www-form-urlencoded", payload)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat:: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("heartbeat failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// Run continuously sends heartbeat signals at regular intervals
func Run() {
	fmt.Printf("In runHeartbeat\n")
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		fmt.Printf("Sending heartbeat\n")
		if err := sendHeartbeat(); err != nil {
			fmt.Printf("Heartbeat error: %s\n", err.Error())
			log.Printf("Heartbeat error:: %s", err.Error())
		}
		<-ticker.C
	}
	fmt.Printf("Exiting runHeartbeat\n")
}
