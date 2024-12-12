package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const activeFileName = "active.json"

// Read the active.json file and return the list of packages
func readActiveFile() ([]string, error) {
	data, err := ioutil.ReadFile(activeFileName)
	if err != nil {
		// If the file doesn't exist, return an empty list
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read active.json: %v", err)
	}

	var packages []string
	if len(data) > 0 {
		if err := json.Unmarshal(data, &packages); err != nil {
			return nil, fmt.Errorf("failed to parse active.json: %v", err)
		}
	}

	return packages, nil
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

// Compare the current and previous list of packages, and perform necessary installations/uninstallations
func syncPackages(prevPackages, currentPackages []string) error {
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

// Create the active.json file with the currently installed Flatpak applications
func createActiveFile() error {
	installedPackages, err := getInstalledFlatpakApplications()
	if err != nil {
		return err
	}

	// Write the current list of installed packages to active.json
	data, err := json.MarshalIndent(installedPackages, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal installed packages: %v", err)
	}

	err = ioutil.WriteFile(activeFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write active.json: %v", err)
	}

	fmt.Printf("active.json file created with %d installed packages.\n", len(installedPackages))
	return nil
}

func main() {
	// Check if the active.json file exists
	previousPackages, err := readActiveFile()
	if err != nil {
		log.Fatalf("Error reading active.json: %v", err)
	}

	// If the file doesn't exist, create it with the current installed packages
	if previousPackages == nil {
		if err := createActiveFile(); err != nil {
			log.Fatalf("Error creating active.json: %v", err)
		}
		// After creating the file, initialize previousPackages to the newly created list
		previousPackages, err = readActiveFile()
		if err != nil {
			log.Fatalf("Error reading active.json after creation: %v", err)
		}
	}

	for {
		// Read the current list of packages from active.json
		currentPackages, err := readActiveFile()
		if err != nil {
			log.Fatalf("Error reading active.json: %v", err)
		}

		// If the current file is different from the previous one, sync packages
		if len(currentPackages) > 0 && len(previousPackages) > 0 {
			if err := syncPackages(previousPackages, currentPackages); err != nil {
				log.Fatalf("Error syncing packages: %v", err)
			}
		}

		// Update the previous packages list to the current state
		previousPackages = currentPackages

		// Sleep for some time before checking again (e.g., 30 seconds)
		time.Sleep(30 * time.Second)
	}
}
