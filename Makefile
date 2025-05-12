# Variables for paths and common configurations
SCRIPTS_DIR := ./scripts
MAIN_GO := ./cmd/pipecast/main.go

# Default target
.DEFAULT_GOAL := all

# Declare phony targets to prevent issues with files named like targets
.PHONY: all init build install clean run

# High-level targets
all: install

# Initialize environment (e.g., setting executable permissions)
init:
	@chmod +x $(SCRIPTS_DIR)/*.sh

# Build target
build: init
	@echo "Building project..."
	@$(SCRIPTS_DIR)/build.sh

# Install target
install: build
	@echo "Installing project..."
	@$(SCRIPTS_DIR)/install.sh

# Clean target
clean: init
	@echo "Cleaning project..."
	@$(SCRIPTS_DIR)/clean.sh

# Run target
run: build
	@echo "Running project..."
	@go run $(MAIN_GO) || true
