package main

import (
	"fmt"
	"os"
	"flatly/internal/manager"
	"flatly/pkg/flatpak"
	"flatly/pkg/util"
)

func main() {
	args := os.Args[1:]

	// Handling commands (add, remove, etc.)
	if len(args) < 1 {
		fmt.Println("Usage: flatly <command> <package_name>")
		return
	}

	command := args[0]
	packageName := ""
	if len(args) > 1 {
		packageName = args[1]
	}

	// Define available commands
	switch command {
	case "add":
		err := flatpak.InstallFlatpakApplication(packageName)
		if err != nil {
			fmt.Printf("Error adding package: %v\n", err)
		}
	case "remove":
		err := flatpak.UninstallFlatpakApplication(packageName)
		if err != nil {
			fmt.Printf("Error removing package: %v\n", err)
		}
	case "daemon":
		manager.RunDaemon() // Daemon for managing packages in the background
	default:
		fmt.Println("Invalid command")
	}
}
