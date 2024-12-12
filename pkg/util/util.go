package util

import (
	"fmt"
	"os"
)

// Get the user's config directory (e.g., `~/.config`)
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return ""
	}

	configDir := fmt.Sprintf("%s/.config/flatly", homeDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			fmt.Println("Error creating config directory:", err)
		}
	}

	return configDir
}
