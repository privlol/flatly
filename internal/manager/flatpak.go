package manager

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"flatly/pkg/flatpak"
	"flatly/pkg/util"
)

func RunDaemon() {
	// Path to active.json
	configDir := util.GetConfigDir()
	activeFile := fmt.Sprintf("%s/active.json", configDir)

	// Check for existing active.json or create it if not present
	if _, err := os.Stat(activeFile); os.IsNotExist(err) {
		// Create active.json with currently installed applications
		flatpak.CreateActiveJson(activeFile)
	}

	// Main loop to monitor changes and adjust
	for {
		flatpak.SyncPackages(activeFile)
		time.Sleep(10 * time.Second) // Sleep for a while before checking again
	}
}
