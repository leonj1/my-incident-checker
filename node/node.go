package node

import "os"

// GetNodeName retrieves the node name from environment variables
func GetNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = os.Getenv("HOSTNAME")
	}
	if nodeName == "" {
		nodeName = "unknown"
	}
	return nodeName
}
