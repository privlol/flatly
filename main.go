package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const activeFileName = "active.json"

// Get the user's home directory
func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	return homeDir
}

// Get the path for the flatly directory
func getFlatlyDir(debug bool) string {
	if debug {
		// Use the current working directory when debug is enabled
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		return cwd
	}

	// Default to ~/.config/flatly when debug is not enabled
	return filepath.Join(getHomeDir(), ".config", "flatly")
}

// Read the active.json file
func readActiveFile(debug bool) ([]string, error) {
	activeFilePath := filepath.Join(getFlatlyDir(debug), activeFileName)

	data, err := ioutil.ReadFile(activeFilePath)
	if err != nil {
		// If the file doesn't exist, return nil (indicating no active.json)
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %v", activeFilePath, err)
	}

	var packages []string
	if len(data) > 0 {
		if err := json.Unmarshal(data, &packages); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %v", activeFilePath, err)
		}
	}

	return packages, nil
}

// Create the active.json file with the current list of Flatpak applications
func createActiveFile(debug bool) error {
	// Get the path for the active.json file (either debug or default location)
	flatlyDir := getFlatlyDir(debug)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(flatlyDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", flatlyDir, err)
	}

	// Get the list of currently installed Flatpak applications
	installedPackages, err := getInstalledFlatpakApplications()
	if err != nil {
		return err
	}

	// Write the current list of installed packages to active.json
	data, err := json.MarshalIndent(installedPackages, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal installed packages: %v", err)
	}

	activeFilePath := filepath.Join(flatlyDir, activeFileName)

	err = ioutil.WriteFile(activeFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", activeFilePath, err)
	}

	fmt.Printf("active.json file created with %d installed packages at %s.\n", len(installedPackages), activeFilePath)
	return nil
}

// Get the list of currently installed Flatpak applications
func getInstalledFlatpakApplications() ([]string, error) {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get installed Flatpak applications: %v", err)
	}

	// Split the output into individual package names
	installedPackages := strings.Split(string(output), "\n")
	var packages []string
	for _, pkg := range installedPackages {
		if pkg != "" {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// Check if a Flatpak application is installed
func isFlatpakInstalled(packageName string) bool {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	installedPackages := strings.Split(string(output), "\n")
	for _, pkg := range installedPackages {
		if pkg == packageName {
			return true
		}
	}
	return false
}

// Install a new Flatpak package
func installFlatpakApplication(packageName string) error {
	if isFlatpakInstalled(packageName) {
		fmt.Printf("%s is already installed. Skipping.\n", packageName)
		return nil
	}

	cmd := exec.Command("flatpak", "install", "--assumeyes", packageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Flatpak package %s: %v", packageName, err)
	}

	fmt.Printf("%s successfully installed.\n", packageName)
	return nil
}

// Uninstall a Flatpak package
func uninstallFlatpakApplication(packageName string) error {
	if !isFlatpakInstalled(packageName) {
		fmt.Printf("%s is not installed. Skipping.\n", packageName)
		return nil
	}

	cmd := exec.Command("flatpak", "uninstall", "--assumeyes", packageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall Flatpak package %s: %v", packageName, err)
	}

	fmt.Printf("%s successfully uninstalled.\n", packageName)
	return nil
}

// Backup the current active.json file with a timestamp
func backupActiveFile(debug bool) error {
	flatlyDir := getFlatlyDir(debug)
	backupDir := filepath.Join(flatlyDir, "backups")

	// Create the backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %v", backupDir, err)
	}

	// Read the current active.json file
	activeFilePath := filepath.Join(flatlyDir, activeFileName)
	data, err := ioutil.ReadFile(activeFilePath)
	if err != nil {
		return fmt.Errorf("failed to read active.json for backup: %v", err)
	}

	// Create a backup file with a timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFilePath := filepath.Join(backupDir, fmt.Sprintf("active_backup_%s.json", timestamp))

	err = ioutil.WriteFile(backupFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to create backup %s: %v", backupFilePath, err)
	}

	fmt.Printf("Backup created: %s\n", backupFilePath)
	return nil
}

// Compare the current and previous list of packages, and perform necessary installations/uninstallations
func syncPackages(prevPackages, currentPackages []string) error {
	// Backup the current active.json before making changes
	if err := backupActiveFile(false); err != nil {
		log.Printf("Error creating backup: %v", err)
	}

	// Check for packages to uninstall (present in prevPackages but not in currentPackages)
	for _, prevPackage := range prevPackages {
		if !contains(currentPackages, prevPackage) {
			fmt.Printf("Package removed: %s\n", prevPackage)
			if err := uninstallFlatpakApplication(prevPackage); err != nil {
				return err
			}
		}
	}

	// Check for packages to install (present in currentPackages but not in prevPackages)
	for _, currentPackage := range currentPackages {
		if !contains(prevPackages, currentPackage) {
			fmt.Printf("Package added: %s\n", currentPackage)
			if err := installFlatpakApplication(currentPackage); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function to check if a package is in a list
func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

// Handle the add command
func handleAddCommand(packageName string, debug bool) {
	if err := installFlatpakApplication(packageName); err != nil {
		log.Fatalf("Error installing package: %v", err)
	}

	// Update active.json after installation
	updateActiveFile(debug)
}

// Handle the remove command
func handleRemoveCommand(packageName string, debug bool) {
	if err := uninstallFlatpakApplication(packageName); err != nil {
		log.Fatalf("Error uninstalling package: %v", err)
	}

	// Update active.json after removal
	updateActiveFile(debug)
}

// Update the active.json file after adding/removing packages
func updateActiveFile(debug bool) {
	installedPackages, err := getInstalledFlatpakApplications()
	if err != nil {
		log.Fatalf("Error getting installed packages: %v", err)
	}

	flatlyDir := getFlatlyDir(debug)
	activeFilePath := filepath.Join(flatlyDir, activeFileName)

	data, err := json.MarshalIndent(installedPackages, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling installed packages: %v", err)
	}

	if err := ioutil.WriteFile(activeFilePath, data, 0644); err != nil {
		log.Fatalf("Error writing to active.json: %v", err)
	}

	fmt.Printf("active.json updated with %d installed packages.\n", len(installedPackages))
}

// Run as a daemon to monitor package changes
func runDaemon(debug bool) {
	// Read the previous active.json file
	previousPackages, err := readActiveFile(debug)
	if err != nil {
		log.Fatalf("Error reading active.json: %v", err)
	}

	if previousPackages == nil {
		if err := createActiveFile(debug); err != nil {
			log.Fatalf("Error creating active.json: %v", err)
		}
		previousPackages, err = readActiveFile(debug)
		if err != nil {
			log.Fatalf("Error reading active.json after creation: %v", err)
		}
	}

	// Daemon loop
	for {
		// Read the current list of packages
		currentPackages, err := readActiveFile(debug)
		if err != nil {
			log.Fatalf("Error reading active.json: %v", err)
		}

		if len(currentPackages) > 0 && len(previousPackages) > 0 {
			// Sync packages if changes are detected
			if err := syncPackages(previousPackages, currentPackages); err != nil {
				log.Fatalf("Error syncing packages: %v", err)
			}
		}

		// Update the previous packages list
		previousPackages = currentPackages

		// Sleep before checking again (e.g., every 30 seconds)
		time.Sleep(30 * time.Second)
	}
}

func main() {
	// Handle commands
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Usage:")
		fmt.Println("  flatly add <package_name>    # Add a package")
		fmt.Println("  flatly remove <package_name> # Remove a package")
		fmt.Println("  flatly daemon                # Run as daemon")
		return
	}

	// Handle add command
	if args[0] == "add" && len(args) == 2 {
		handleAddCommand(args[1], false)
		return
	}

	// Handle remove command
	if args[0] == "remove" && len(args) == 2 {
		handleRemoveCommand(args[1], false)
		return
	}

	// Handle daemon mode
	if args[0] == "daemon" {
		runDaemon(false)
		return
	}

	fmt.Println("Invalid command.")
}
