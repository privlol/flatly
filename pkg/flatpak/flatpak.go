package flatpak

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
)

// Install a Flatpak package
func InstallFlatpakApplication(packageName string) error {
	cmd := exec.Command("flatpak", "install", "--assumeyes", packageName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error installing %s: %w", packageName, err)
	}
	fmt.Printf("Successfully installed %s\n", packageName)
	return nil
}

// Uninstall a Flatpak package
func UninstallFlatpakApplication(packageName string) error {
	cmd := exec.Command("flatpak", "uninstall", "--assumeyes", packageName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error uninstalling %s: %w", packageName, err)
	}
	fmt.Printf("Successfully uninstalled %s\n", packageName)
	return nil
}

// Create or update the active.json file with currently installed packages
func CreateActiveJson(filePath string) error {
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error fetching Flatpak list: %w", err)
	}

	// Split output into lines
	packages := make([]string, 0)
	for _, line := range string(output) {
		packages = append(packages, line)
	}

	// Write to active.json
	data, err := json.MarshalIndent(packages, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling package list: %w", err)
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", filePath, err)
	}

	fmt.Printf("Created active.json with %d packages\n", len(packages))
	return nil
}

// Sync installed packages with active.json (install/remove based on changes)
func SyncPackages(filePath string) error {
	// Read current active.json
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading active.json: %w", err)
	}

	var currentPackages []string
	err = json.Unmarshal(data, &currentPackages)
	if err != nil {
		return fmt.Errorf("error unmarshalling active.json: %w", err)
	}

	// Get the list of installed packages again
	cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error fetching Flatpak list: %w", err)
	}

	var newPackages []string
	for _, line := range string(output) {
		newPackages = append(newPackages, line)
	}

	// Determine packages to add and remove
	toAdd := difference(newPackages, currentPackages)
	toRemove := difference(currentPackages, newPackages)

	// Add new packages
	for _, packageName := range toAdd {
		InstallFlatpakApplication(packageName)
	}

	// Remove removed packages
	for _, packageName := range toRemove {
		UninstallFlatpakApplication(packageName)
	}

	return nil
}

// Helper function to calculate the difference between two slices
func difference(a, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range b {
		m[item] = struct{}{}
	}

	var diff []string
	for _, item := range a {
		if _, found := m[item]; !found {
			diff = append(diff, item)
		}
	}
	return diff
}
