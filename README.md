# 🚀 ADB Manager - 强大的ADB设备管理工具

[![Go Version](https://img.shields.io/badge/Go-1.16+-blue)]()
[![License](https://img.shields.io/badge/License-MIT-green)]()
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)]()

一个功能强大的 ADB (Android Debug Bridge) 图形化管理工具，采用 Go 语言开发，使用 Fyne 框架提供美观的跨平台界面。

## ✨ 核心特性

### 🎯 智能设备管理
- **实时设备列表** - 自动检测和管理多台设备（支持20+个设备）
- **增量式设备管理** - 避免因网络波动导致设备列表被清空
- **无线连接** - 支持 `IP:PORT` 方式无线连接Android设备
- **多设备选择** - 可多选设备，进行批量操作
- **设备离线检测** - 自动检测离线设备，3次失败后标记，5分钟后删除

### 📱 命令执行与Busybox支持
- **单设备命令执行** - Shell标签页支持实时输入和执行命令
- **批量命令执行** - 在多台设备上同时执行相同命令
- **Busybox模式** - 为物联网设备提供完整支持，一键启用自动加前缀
- **快捷命令** - 预设常用命令，如获取屏幕分辨率、电池信息等
- **命令历史** - 保存执行过的命令

### 📂 文件管理系统
- **文件浏览** - 浏览设备文件系统，支持目录导航
- **上传/下载** - 在设备和PC之间传输文件
- **文件编辑** - 删除、重命名文件和目录
- **权限管理** - 支持 `chmod` 修改文件权限
- **文件详情** - 查看文件大小、修改时间、权限等信息

### 📸 截屏与截图功能
- **一键截屏** - 快速获取设备当前屏幕
- **批量截屏** - 同时对多台设备进行截屏
- **自动保存** - 截图自动保存到本地

### 📊 设备信息采集
- **基础信息** - 型号、品牌、制造商、CPU架构等
- **系统信息** - Android版本、SDK版本、Build ID等
- **网络信息** - IP地址、MAC地址、WiFi连接状态
- **性能数据** - 内存使用、电池状态、进程列表
- **应用信息** - 已安装应用列表（系统应用+第三方应用）
- **存储信息** - 磁盘空间使用情况

### 🔍 高级功能
- **应用管理** - 安装、卸载、启动应用
- **设备扫描** - 支持子网扫描发现设备
- **数据采集** - 获取联系人、短信、位置信息等（需设备授权）
- **智能重连** - 自动检测并恢复设备连接
- **ADB诊断** - 检测版本冲突，提供解决方案

## 🔧 深层技术特性

### 可靠性保证
- **版本冲突检测** - 自动检测ADB客户端和服务器版本不匹配
- **智能设备缓存** - 不会因单个命令失败清除整个设备列表
- **故障自恢复** - 网络波动时保留设备列表，不丢失数据
- **线程安全** - 使用RWMutex保护并发访问

### 物联网友好
- **Busybox完整支持** - 为嵌入式设备优化
- **自定义命令前缀** - 灵活支持各种设备命令行工具
- **超时控制** - 防止长时间运行命令卡死应用

## 📥 安装使用

### 前置要求
- **Go 1.16 或更高版本**
- **ADB 工具**（Android SDK Platform-Tools）

### Windows 安装

1. **安装 ADB**
```bash
# 下载 Android SDK Platform-Tools
# 或使用 scoop/chocolatey
scoop install android-platform-tools
# 或
choco install androidstudio
# 下载地址
https://developer.android.com/tools/releases/platform-tools?hl=zh-cn
```
 
2. **克隆项目**
```bash
git clone https://github.com/yourusername/AdbManager.git
cd AdbManager
```

3. **编译运行**
```bash
# 编译
go build -ldflags="-s -w" -o AdbManager.exe

# 运行
./AdbManager.exe
```

### Linux/macOS 安装

```bash
# 安装 ADB
# Ubuntu/Debian
sudo apt-get install android-tools-adb

# macOS
brew install android-platform-tools

# 克隆和编译
git clone https://github.com/yourusername/AdbManager.git
cd AdbManager
go build -o AdbManager ./...

# 运行
./AdbManager
```

## 🎮 使用指南

### 基础操作

#### 1. 设备管理
```
1. 启动应用后，在左侧设备列表中查看所有已连接设备
2. 勾选要操作的设备
3. 使用各个功能标签进行操作
```

#### 2. 添加无线设备
```
1. 在"连接"输入框输入设备IP和端口 (例如: 192.168.1.100:5555)
2. 点击"连接"按钮
3. 等待连接成功
```

#### 3. 执行命令

**单设备命令执行：**
- Shell标签页 → 输入命令 → 点击"执行命令"

**批量命令执行：**
- 批量操作标签页 → 选择设备 → 输入命令 → 点击"批量执行命令"

**启用Busybox模式：**
- 勾选"启用 Busybox 模式"复选框
- 之后所有命令自动加入 `busybox` 前缀

### 快捷命令
以下快捷命令已预设（Shell标签页）：
- 获取屏幕分辨率
- 获取电池信息
- 列出进程
- 获取网络连接
- 查看内存使用
- 查看CPU信息

## 🐛 常见问题

### Q: 为什么提示"ADB服务器版本不匹配"？
A: 这是因为ADB客户端版本和服务器版本不一致。解决方案：
```bash
adb kill-server
adb start-server
```
或更新ADB到最新版本。点击"诊断 ADB 问题"按钮可查看详细信息。

### Q: 为什么连接的设备突然掉线了？
A: 常见原因：
- USB连接松动或故障
- WiFi网络不稳定
- 设备进入休眠或断开WiFi
- ADB服务异常

解决方案：
1. 检查物理连接或网络
2. 点击"刷新设备列表"重新扫描
3. 使用"诊断ADB问题"工具检查

### Q: 如何移除离线设备？
A: 点击"移除离线设备"按钮，确认后自动删除所有离线设备。设备被删除后需要重新连接才能使用。

### Q: Busybox模式有什么作用？
A: 许多物联网设备和嵌入式系统使用Busybox作为shell。启用此模式后，所有命令会自动加上 `busybox` 前缀，例如：
```
输入: ls -la
执行: busybox ls -la
```

## 🚀 高级功能

### 批量操作
- 支持多台设备同时执行命令
- 支持批量安装/卸载应用
- 支持批量截屏

### 设备扫描
- 自动扫描子网范围内的设备
- 发现新的无线连接设备

### 数据采集（需要设备授权）
- 联系人列表
- 短信记录
- 位置数据
- WiFi配置

## 📝 开发信息

### 项目结构
```
AdbManager/
├── main.go                 # 程序入口
├── internal/
│   ├── adb/               # ADB核心功能
│   │   └── adb.go
│   ├── ui/                # UI界面
│   │   ├── main_ui.go
│   │   ├── batch_ui.go
│   │   ├── device_info_ui.go
│   │   ├── file_manager_ui.go
│   │   ├── app_manager_ui.go
│   │   ├── scanner_ui.go
│   │   └── collector_ui.go
│   ├── batch/             # 批量操作
│   │   └── batch.go
│   ├── scanner/           # 设备扫描
│   │   └── scanner.go
│   └── collector/         # 数据采集
│       └── collector.go
├── go.mod
├── go.sum
└── README.md
```

### 技术栈
- **语言**: Go 1.16+
- **UI框架**: Fyne v2
- **特性**: 并发安全、跨平台、低延迟

### 编译命令
```bash
# 普通编译
go build ./...

# 生成可执行文件
go build -o AdbManager main.go

# 交叉编译（Windows）
GOOS=windows GOARCH=amd64 go build -o AdbManager.exe ./...

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -o AdbManager ./...

# 交叉编译（macOS）
GOOS=darwin GOARCH=amd64 go build -o AdbManager ./...
```

## 🔐 安全说明

- 本工具涉及设备访问敏感权限
- 建议仅在可信网络环境中使用
- 不收集或上传任何数据
- 所有操作均本地执行

## 🤝 贡献指南

欢迎提交Issue和Pull Request！

### 本地开发
```bash
git clone <your-fork>
cd AdbManager
go mod tidy
go build ./...
```

### 提交流程
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🎉 致谢

感谢以下开源项目的支持：
- [Fyne](https://fyne.io/) - 美观的Go GUI框架
- [Android SDK](https://developer.android.com/) - ADB工具

## 📞 联系方式

- 提交Issue：GitHub Issues
- 讨论功能：GitHub Discussions

## 📊 项目统计

- 支持设备数: **20+**
- 核心功能: **8+**
- 代码行数: **3000+**
- 响应延迟: **<100ms**

## 🗺️ 开发路线图

- [ ] 添加设备分组功能
- [ ] 支持设备日志导出
- [ ] 提供命令脚本编辑器
- [ ] 集成性能监控工具
- [ ] 支持设备录屏功能
- [ ] 添加Dark/Light主题切换
- [ ] 国际化语言支持

---

**最后更新**: 2025年12月31日  
**维护者**: AdbManager Team
