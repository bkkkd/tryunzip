# tryunzip Makefile - Docker 编译和多平台打包
# 参考 tafile 项目结构

.PHONY: all build clean help version docker-builder docker-build linux linux-arm64 windows package package-linux package-linux-arm64 package-windows

# 读取版本号
VERSION := $(shell cat VERSION)
APP_NAME := tryunzip
DIST_DIR := dist
BUILD_DIR := dist/package
DOCKER_IMAGE_NAME := $(APP_NAME)-builder:$(VERSION)

# 默认目标：编译所有平台
all: clean linux windows package

# 显示帮助
help:
	@echo "tryunzip v$(VERSION) - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  all             - Clean and build all platforms (default)"
	@echo "  build           - Build release version for current platform"
	@echo "  linux           - Build Linux AMD64 version"
	@echo "  linux-arm64     - Build Linux ARM64 version"
	@echo "  windows         - Build Windows AMD64 version"
	@echo "  package         - Package all platforms"
	@echo "  package-linux   - Package Linux AMD64 version"
	@echo "  package-windows - Package Windows version"
	@echo ""
	@echo "Docker Targets:"
	@echo "  docker-builder  - Build Docker compiler image"
	@echo "  docker-build    - Build all platforms using Docker"
	@echo "  docker-linux    - Build Linux AMD64 using Docker"
	@echo "  docker-windows  - Build Windows using Docker"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean           - Clean build artifacts"
	@echo "  version         - Show version"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Current version: $(VERSION)"

# 显示版本
version:
	@echo "$(VERSION)"

# 构建发行版本（当前平台）
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(DIST_DIR)
	go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date +%Y%m%d%H%M%S)" -o $(DIST_DIR)/$(APP_NAME) .
	@echo "✓ Build complete: $(DIST_DIR)/$(APP_NAME) (v$(VERSION))"

# 构建 Linux AMD64 版本（使用 Docker）
docker-linux:
	@echo "Building Linux AMD64 version with Docker..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(shell pwd):/build" \
		-w /build \
		$(DOCKER_IMAGE_NAME) \
		sh -c 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(date +%Y%m%d%H%M%S)" -trimpath -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 .'
	@echo "✓ Linux AMD64 build complete: $(DIST_DIR)/$(APP_NAME)-linux-amd64"

# 构建 Linux ARM64 版本（使用 Docker）
docker-linux-arm64:
	@echo "Building Linux ARM64 version with Docker..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(shell pwd):/build" \
		-w /build \
		$(DOCKER_IMAGE_NAME) \
		sh -c 'CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(date +%Y%m%d%H%M%S)" -trimpath -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 .'
	@echo "✓ Linux ARM64 build complete: $(DIST_DIR)/$(APP_NAME)-linux-arm64"

# 构建 Windows 版本（使用 Docker）
docker-windows:
	@echo "Building Windows version with Docker..."
	@mkdir -p $(DIST_DIR)
	docker run --rm \
		-v "$(shell pwd):/build" \
		-w /build \
		$(DOCKER_IMAGE_NAME) \
		sh -c 'CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(date +%Y%m%d%H%M%S)" -trimpath -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe .'
	@echo "✓ Windows build complete: $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

# 使用 Docker 构建所有平台
docker-build: docker-builder docker-linux docker-windows

# 构建 Linux 版本（本地）
linux: docker-linux
	@echo "Linux AMD64 build complete: $(DIST_DIR)/$(APP_NAME)-linux-amd64"

# 构建 Linux ARM64 版本（本地）
linux-arm64: docker-linux-arm64
	@echo "Linux ARM64 build complete: $(DIST_DIR)/$(APP_NAME)-linux-arm64"

# 构建 Windows 版本（本地）
windows: docker-windows
	@echo "Windows build complete: $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

# 构建 Docker 编译镜像
docker-builder:
	@echo "Building Docker compiler image..."
	docker build -t $(DOCKER_IMAGE_NAME) -f Dockerfile.build .
	docker build -t $(APP_NAME)-builder:latest -f Dockerfile.build .
	@echo "✓ Docker compiler image built: $(DOCKER_IMAGE_NAME)"

# 打包所有平台
package: package-linux package-windows
	@echo ""
	@echo "✓ All packages created:"
	@ls -lh $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip 2>/dev/null || true

# 打包 Linux 版本
package-linux:
	@echo "Packaging Linux AMD64..."
	@mkdir -p $(BUILD_DIR)/linux
	# 复制文件
	cp $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(BUILD_DIR)/linux/$(APP_NAME)
	cp README.md $(BUILD_DIR)/linux/ 2>/dev/null || true
	# 打包
	cd $(BUILD_DIR)/linux && tar -czf ../$(APP_NAME)-linux-amd64-$(VERSION).tar.gz $(APP_NAME) README.md 2>/dev/null || tar -czf ../$(APP_NAME)-linux-amd64-$(VERSION).tar.gz $(APP_NAME)
	# 清理
	rm -rf $(BUILD_DIR)/linux
	@echo "✓ Linux package created: $(DIST_DIR)/$(APP_NAME)-linux-amd64-$(VERSION).tar.gz"

# 打包 Windows 版本
package-windows:
	@echo "Packaging Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	# 复制文件
	cp $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(BUILD_DIR)/windows/$(APP_NAME).exe
	cp README.md $(BUILD_DIR)/windows/ 2>/dev/null || true
	# 打包
	cd $(BUILD_DIR)/windows && zip -r ../$(APP_NAME)-windows-amd64-$(VERSION).zip $(APP_NAME).exe README.md 2>/dev/null || zip -r ../$(APP_NAME)-windows-amd64-$(VERSION).zip $(APP_NAME).exe
	# 清理
	rm -rf $(BUILD_DIR)/windows
	@echo "✓ Windows package created: $(DIST_DIR)/$(APP_NAME)-windows-amd64-$(VERSION).zip"

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(DIST_DIR)
	rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"
