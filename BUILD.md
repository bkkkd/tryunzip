# 构建和打包指南

本文档说明如何使用 Docker 和 Makefile 构建和打包 tryunzip 项目。

## 前置要求

- Docker Desktop (Windows/macOS) 或 Docker (Linux)
- Go 1.16+ (可选，本地编译时需要)

## 快速开始

### 1. 构建 Docker 编译镜像

```bash
make docker-builder
```

这将创建一个包含 Go 编译环境的 Docker 镜像。

### 2. 编译所有平台

```bash
# 编译 Linux AMD64 和 Windows 版本
make docker-build

# 或分步执行
make docker-linux
make docker-windows
```

编译产物将保存在 `dist/` 目录：
- `tryunzip-linux-amd64` - Linux AMD64 可执行文件
- `tryunzip-windows-amd64.exe` - Windows 可执行文件

### 3. 打包发行版

```bash
make package
```

这将生成包含版本号的压缩文件：
- `dist/tryunzip-linux-amd64-1.0.0.tar.gz` - Linux 发行包
- `dist/tryunzip-windows-amd64-1.0.0.zip` - Windows 发行包

## 完整构建流程

```bash
# 1. 清理旧的构建产物
make clean

# 2. 构建 Docker 编译镜像
make docker-builder

# 3. 编译所有平台
make docker-build

# 4. 打包所有平台
make package
```

或使用一键命令：

```bash
make all
```

## Makefile 目标详解

### 构建目标

| 目标 | 说明 |
|------|------|
| `make all` | 清理并编译所有平台（默认） |
| `make build` | 构建当前平台的版本 |
| `make linux` | 构建 Linux AMD64 版本 |
| `make windows` | 构建 Windows AMD64 版本 |

### Docker 目标

| 目标 | 说明 |
|------|------|
| `make docker-builder` | 构建编译镜像 |
| `make docker-build` | 使用 Docker 编译所有平台 |
| `make docker-linux` | 使用 Docker 编译 Linux 版本 |
| `make docker-windows` | 使用 Docker 编译 Windows 版本 |

### 打包目标

| 目标 | 说明 |
|------|------|
| `make package` | 打包所有平台 |
| `make package-linux` | 打包 Linux 版本 |
| `make package-windows` | 打包 Windows 版本 |

### 工具目标

| 目标 | 说明 |
|------|------|
| `make clean` | 清理构建产物 |
| `make version` | 显示版本号 |
| `make help` | 显示帮助信息 |

## 版本管理

版本号存储在 `VERSION` 文件中。修改版本号后重新构建即可：

```bash
# 查看当前版本
cat VERSION
# 输出: 1.0.0

# 修改版本号
echo "1.1.0" > VERSION

# 重新构建
make all
```

## Windows 环境

在 Windows 上可能没有 `make` 命令。可以：

### 选项 1: 安装 make

使用 Chocolatey:
```powershell
choco install make
```

或使用 MSYS2/Git Bash:
```bash
# 在 Git Bash 或 MSYS2 终端中运行 make 命令
```

### 选项 2: 直接使用 Docker 命令

```powershell
# 构建编译镜像
docker build -t tryunzip-builder:1.0.0 -f Dockerfile.build .

# 编译 Linux 版本
docker run --rm -v "$(pwd):/build" -w /build tryunzip-builder:1.0.0 sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-X main.version=1.0.0' -trimpath -o dist/tryunzip-linux-amd64 ."

# 编译 Windows 版本
docker run --rm -v "$(pwd):/build" -w /build tryunzip-builder:1.0.0 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-X main.version=1.0.0' -trimpath -o dist/tryunzip-windows-amd64.exe ."

# 打包
cd dist
tar -czf tryunzip-linux-amd64-1.0.0.tar.gz tryunzip-linux-amd64
powershell Compress-Archive -Path tryunzip-windows-amd64.exe -DestinationPath tryunzip-windows-amd64-1.0.0.zip
```

## Linux/macOS 环境

在 Unix 系统上直接使用 make 命令：

```bash
# 查看帮助
make help

# 完整构建
make all

# 只编译不打包
make docker-build
```

## 输出文件

构建完成后，`dist/` 目录将包含：

```
dist/
├── tryunzip-linux-amd64              # Linux 可执行文件
├── tryunzip-windows-amd64.exe        # Windows 可执行文件
├── tryunzip-linux-amd64-1.0.0.tar.gz # Linux 发行包
└── tryunzip-windows-amd64-1.0.0.zip  # Windows 发行包
```

## 常见问题

### Q: Docker 构建失败

确保 Docker Desktop 正在运行，并且有足够的磁盘空间。

### Q: 编译很慢

首次编译需要下载依赖，后续编译会快很多。可以预先构建编译镜像：

```bash
make docker-builder
```

### Q: Windows 打包失败

确保已安装 `zip` 工具。可以使用 PowerShell 的 `Compress-Archive` 替代。

### Q: 如何交叉编译？

Makefile 已配置交叉编译。只需运行相应的目标即可，无需额外配置。
