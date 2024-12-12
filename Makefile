GOBIN=./bin
GO=go
APP_NAME=flatly
VERSION=$(shell git describe --tags --always --dirty)
INSTALL_DIR=/usr/local/bin
SERVICE_FILE=service/flatly.service
SERVICE_DIR=/etc/systemd/system
USER_SERVICE_DIR=~/.config/systemd/user
SRC_DIR=./cmd/flatly
CONFIG_DIR=~/Documents/Projects/flatly/config

all: build

build:
	# Build the binary from the main.go file located in cmd/flatly/
	mkdir -p $(GOBIN)
	$(GO) build -o $(GOBIN)/$(APP_NAME) $(SRC_DIR)

run:
	$(GOBIN)/$(APP_NAME)

test:
	$(GO) test ./...

clean:
	rm -rf $(GOBIN)

install: build
	# Install the binary to /usr/local/bin for system-wide usage
	sudo cp $(GOBIN)/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	sudo chmod +x $(INSTALL_DIR)/$(APP_NAME)

	# (Optional) Install configuration files, if needed
	# sudo cp $(CONFIG_DIR)/config.yaml /etc/flatly/config.yaml

	# Install the systemd service file
	sudo cp $(SERVICE_FILE) $(SERVICE_DIR)/$(APP_NAME).service
	sudo systemctl daemon-reload
	sudo systemctl enable $(APP_NAME).service
	sudo systemctl start $(APP_NAME).service

	# Alternatively, install the service file for user-specific services
	# mkdir -p $(USER_SERVICE_DIR)
	# cp $(SERVICE_FILE) $(USER_SERVICE_DIR)/$(APP_NAME).service
	# systemctl --user daemon-reload
	# systemctl --user enable $(APP_NAME).service
	# systemctl --user start $(APP_NAME).service

release: build
	# Prepare release package
	mkdir -p release
	cp $(GOBIN)/$(APP_NAME) release/
	tar -czvf release/$(APP_NAME)-$(VERSION).tar.gz -C release $(APP_NAME)
