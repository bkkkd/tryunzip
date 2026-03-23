# ZIP 密码破解工具

一个用 Go 语言编写的 ZIP 文件密码破解工具，支持暴力破解和词典破解两种模式。

**当前版本**: 1.0.0

## 功能特点

- **暴力破解模式**: 按指定字符集和长度范围生成密码进行尝试
- **词典破解模式**: 从密码文件读取候选密码
- **支持字符集**: 数字、小写字母、大写字母、自定义字符
- **字符范围**: 支持 `a-z`、`A-Z`、`0-9` 等范围表示法
- **并发处理**: 支持多线程加速（注意：部分系统可能有文件锁限制）
- **进度显示**: 实时显示尝试次数、速率和耗时
- **超时控制**: 可设置最大运行时间
- **密码组合计算**: 显示预估的密码组合数
- **版本管理**: 支持通过VERSION文件管理版本号

## 安装

### 从源码编译

**方式一：使用 Makefile（Unix/Linux/macOS）**

```bash
# 克隆项目
git clone <repository_url>
cd tryunzip

# 构建发行版本
make build

# 或一步完成（清理+构建）
make all

# 查看帮助
make help
```

**方式二：使用 PowerShell 脚本（Windows）**

```powershell
# 克隆项目
git clone <repository_url>
cd tryunzip

# 构建发行版本
.\build.ps1

# 查看帮助
.\build.ps1 -Help

# 清理构建产物
.\build.ps1 -Clean

# 构建调试版本
.\build.ps1 -Debug
```

**方式三：直接使用 Go 命令**

```bash
# 克隆项目
git clone <repository_url>
cd tryunzip

# 构建发行版本（带版本号）
$VERSION = Get-Content VERSION
go build -ldflags "-X main.version=$VERSION" -o tryunzip.exe main.go

# 或简化版（硬编码版本）
go build -o tryunzip.exe main.go
```

### 版本管理

版本号存储在 `VERSION` 文件中：

```bash
# 查看当前版本
cat VERSION
# 输出: 1.0.0

# 修改版本号
echo "1.1.0" > VERSION
```

构建时会自动从 VERSION 文件读取版本号并注入到程序中。

### 依赖

- Go 1.16+
- github.com/yeka/zip (第三方加密ZIP支持库)

自动下载：
```bash
go mod tidy
```

## 使用方法

### 版本信息

```bash
# 查看版本
tryunzip -v
tryunzip --version
```

### 基本用法

```bash
# 暴力破解 - 4位数字密码
tryunzip -f test.zip

# 指定密码长度范围
tryunzip -f test.zip -min 4 -max 6

# 使用自定义字符集
tryunzip -f test.zip -charset 0123456789

# 小写字母密码
tryunzip -f test.zip -charset a-z -min 3 -max 5

# 字母数字混合密码
tryunzip -f test.zip -charset a-zA-Z0-9 -min 6 -max 8

# 词典模式
tryunzip -f test.zip -w passwords.txt

# 设置超时
tryunzip -f test.zip -timeout 30m

# 调整工作线程
tryunzip -f test.zip -workers 2
```

### 命令行选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-f` | ZIP 文件路径 | (必需) |
| `-w` | 密码词典文件路径 | (不使用) |
| `-min` | 密码最小长度 | 1 |
| `-max` | 密码最大长度 | 4 |
| `-charset` | 密码字符集 | 0123456789 |
| `-workers` | 并发工作线程数 | 1 |
| `-timeout` | 最大运行时间 | 0 (无限制) |
| `-progress` | 显示进度 | true |

### 字符集示例

```
0123456789                 # 纯数字
abcdefghijklmnopqrstuvwxyz # 小写字母
a-z                        # 范围表示法
ABCDEFGHIJKLMNOPQRSTUVWXYZ# 大写字母
a-zA-Z0-9                  # 字母数字混合
a-zA-Z0-9!@#$%^&*()        # 包含特殊字符
```

## 示例

### 示例 1: 破解4位数字密码

```bash
tryunzip -f encrypted.zip -min 4 -max 4 -charset 0123456789
```

### 示例 2: 破解6位小写字母密码

```bash
tryunzip -f secret.zip -charset a-z -min 6 -max 6
```

### 示例 3: 使用词典破解

```bash
# 创建词典文件
echo -e "password\n123456\nadmin\nqwerty" > passwords.txt

# 使用词典
tryunzip -f target.zip -w passwords.txt
```

### 示例 4: 完整的字母数字密码破解

```bash
tryunzip -f archive.zip -charset a-zA-Z0-9 -min 6 -max 8
```

## 创建加密 ZIP 测试文件

### 使用 7-Zip (推荐)

```bash
# 安装 7-Zip 后
7z a -p'your_password' test.zip file.txt
```

### 使用 Python

```python
import zipfile

# Python 3.2+ 支持创建加密 ZIP
with zipfile.ZipFile('test.zip', 'w') as zf:
    zf.setpassword(b'your_password')
    zf.write('file.txt')
```

## 性能说明

- **暴力破解**: 性能取决于密码复杂度和系统性能
  - 4位纯数字: ~10,000 组合，秒级完成
  - 6位小写字母: ~3亿组合，可能需要数小时
  - 8位混合字符: ~200万亿组合，需要极长时间

- **建议**:
  1. 从最短长度开始尝试
  2. 优先使用词典模式
  3. 从常见密码模式开始（如纯数字→小写字母→混合）
  4. 设置合理的超时时间

## 注意事项

⚠️ **仅用于合法用途**：本工具仅适用于破解自己忘记密码的 ZIP 文件或获得授权的文件。

⚠️ **性能限制**: Windows 系统上 ZIP 文件可能有访问限制，多线程效果可能不明显。

⚠️ **密码强度**: 强密码（长+混合字符）的破解时间可能不现实。

## 许可证

MIT License
