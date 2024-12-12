# Set variables for binary and application name
GOBIN=./bin
GO=go
APP_NAME=flatly
VERSION=$(shell git describe --tags --always --dirty)

# Define the user for running the service
USER=flatly

# Check if the necessary files exist
check:
	@echo "Checking for required files..."
	@if [ ! -f "./main.go" ]; then echo "Error: main.go is missing!"; exit 1; fi
	@echo "All required files found!"

# Create the flatly user if it doesn't exist
create-user:
	@id -u $(USER) &>/dev/null || sudo useradd -r -m $(USER)

# Build the application from the main.go file
build: create-user
	$(GO) build -o $(GOBIN)/$(APP_NAME) ./main.go

# Install the application system-wide
install: build create-user
	@echo "Installing flatly to /usr/local/bin"
	sudo install -m 755 ./bin/flatly /usr/local/bin/flatly

	@echo "Installing systemd service"
	# Copy the service to the appropriate directory
	sudo cp flatly.service ~/.config/systemd/user/
	systemctl --user daemon-reload

# Clean up the build artifacts
clean:
	rm -rf $(GOBIN)

# Run the application directly from the bin directory
run:
	$(GOBIN)/$(APP_NAME)

# Package the application into a release tarball
release: build
	mkdir -p release
	cp $(GOBIN)/$(APP_NAME) release/
	tar -czvf release/$(APP_NAME)-$(VERSION).tar.gz -C release $(APP_NAME)

distcheck: clean check build install
	@echo "Distcheck: Clean environment, build and install process completed successfully."
