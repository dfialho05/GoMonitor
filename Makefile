# GoMonitor Makefile

APP_NAME = gom
INSTALL_PATH = /usr/local/bin

build:
	@echo "Building $(APP_NAME)..."
	@go build -ldflags="-s -w" -o $(APP_NAME) ./application
	@echo "Build complete: $(APP_NAME)"

install: build
	@echo "Installing $(APP_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(APP_NAME) $(INSTALL_PATH)/$(APP_NAME)
	@sudo chmod +x $(INSTALL_PATH)/$(APP_NAME)
	@echo "Installation complete!"

uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(APP_NAME)
	@if grep -q "^gom$$" $(HOME)/.bashrc || grep -q "# Run GoMonitor on terminal startup" $(HOME)/.bashrc; then \
		echo "Removing auto-start from ~/.bashrc..."; \
		sed -i '/# Run GoMonitor on terminal startup/d' $(HOME)/.bashrc; \
		sed -i '/^gom$$/d' $(HOME)/.bashrc; \
		echo "Auto-start removed from ~/.bashrc"; \
	fi
	@echo "Uninstallation complete"

help:
	@echo "Usage:"
	@echo "  make build       Build $(APP_NAME)"
	@echo "  make install     Install $(APP_NAME) to $(INSTALL_PATH)"
	@echo "  make uninstall   Uninstall $(APP_NAME) from $(INSTALL_PATH)"
	@echo "  make help        Display this help message"
