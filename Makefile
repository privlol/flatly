GOBIN=./bin
GO=go
APP_NAME=flatly
VERSION=$(shell git describe --tags --always --dirty)

# Check if the flatly user exists, if not, create it
USER=flatly

# Check for required files and directories
check:
	@echo "Checking for required files..."
	@if [ ! -f "./main.go" ]; then echo "Error: main.go is missing!"; exit 1; fi
	@if [ ! -f "./flatly.service" ]; then echo "Error: flatly.service is missing!"; exit 1; fi
	@if [ ! -d "./bin" ]; then echo "Error: bin directory is missing!"; exit 1; fi
	@echo "All required files found!"

all: check build

# Create the flatly user if it doesn't exist
create-user:
	@id -u $(USER) &>/dev/null || sudo useradd -r -m $(USER)

# Build the application
build: create-user
	$(GO) build -o $(GOBIN)/$(APP_NAME) ./main.go

# Install the application system-wide
install: build
	@echo "Installing flatly to /usr/local/bin"
	sudo install -m 755 ./bin/flatly /usr/local/bin/flatly

	@echo "Installing systemd service"
	sudo cp flatly.service /etc/systemd/user/
	systemctl --user daemon-reload

# Clean the build
clean:
	rm -rf $(GOBIN)

# Run the application
run:
	$(GOBIN)/$(APP_NAME)

# Release the application
release: build
	mkdir -p release
	cp $(GOBIN)/$(APP_NAME) release/
	tar -czvf release/$(APP_NAME)-$(VERSION).tar.gz -C release $(APP_NAME)
