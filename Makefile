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
	@echo "Uninstallation complete"
