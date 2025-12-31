# GitHub 发布指南

本指南帮助你将 AdbManager 项目发布到 GitHub。

## 📋 前置准备

### 1. 安装 Git
如果还未安装 Git，请从 [git-scm.com](https://git-scm.com/) 下载安装。

### 2. 注册 GitHub 账户
访问 [github.com](https://github.com) 注册账户（如果还没有的话）。

### 3. 配置 Git
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

## 🚀 发布步骤

### 第1步：初始化本地仓库
```bash
cd D:\工具\连接工具\AdbManager
git init
```

### 第2步：添加所有文件
```bash
git add .
```

### 第3步：第一次提交
```bash
git commit -m "Initial commit: ADB Manager - Powerful Android device management tool

- Smart device management supporting 20+ devices
- Real-time command execution with Busybox support
- Complete file management system
- Screenshot and device info collection
- Batch operations on multiple devices
- Reliable ADB version conflict detection
- Cross-platform GUI using Fyne framework"
```

### 第4步：在 GitHub 上创建新仓库

1. 访问 [github.com/new](https://github.com/new)
2. 填写仓库信息：
   - **Repository name**: `AdbManager`
   - **Description**: `A powerful ADB device management tool with GUI, supporting 20+ devices and advanced features`
   - **Public** (选择公开)
   - **Initialize this repository with**: 不勾选（因为我们已有本地仓库）
3. 点击 "Create repository"

### 第5步：关联远程仓库
```bash
# 替换 YOUR_USERNAME 为你的 GitHub 用户名
git remote add origin https://github.com/YOUR_USERNAME/AdbManager.git

# 将本地 main 分支改为 master（如果需要）
git branch -M main
```

### 第6步：推送到 GitHub
```bash
# 第一次推送
git push -u origin main

# 或者如果你想用 main 分支
git push -u origin main
```

### 第7步：添加 SSH 密钥（可选，但推荐）

为了避免每次都输入密码，可以添加 SSH 密钥：

```bash
# 生成 SSH 密钥
ssh-keygen -t rsa -b 4096 -C "your.email@example.com"

# 复制公钥内容
type %userprofile%\.ssh\id_rsa.pub
```

然后：
1. 访问 [github.com/settings/keys](https://github.com/settings/keys)
2. 点击 "New SSH key"
3. 粘贴公钥内容
4. 更改远程 URL：
```bash
git remote set-url origin git@github.com:YOUR_USERNAME/AdbManager.git
```

## 📝 发布后的维护

### 添加 Release 版本
```bash
# 创建标签
git tag v1.0.0
git push origin v1.0.0

# 然后在 GitHub 上创建 Release：
# 1. 访问 https://github.com/YOUR_USERNAME/AdbManager/releases
# 2. 点击 "Create a new release"
# 3. 选择标签 v1.0.0
# 4. 填写发布说明
```

### 添加 GitHub Actions（自动化构建）

创建文件 `.github/workflows/build.yml`：

```yaml
name: Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      run: go build -v ./...
    
    - name: Test
      run: go test -v ./...
```

### 生成 Release 二进制文件

编译多平台可执行文件并在 Release 页面上传：

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o AdbManager.exe ./...

# Linux
GOOS=linux GOARCH=amd64 go build -o AdbManager ./...

# macOS
GOOS=darwin GOARCH=amd64 go build -o AdbManager ./...
GOOS=darwin GOARCH=arm64 go build -o AdbManager-arm64 ./...
```

## 🔍 验证推送成功

访问 `https://github.com/YOUR_USERNAME/AdbManager`，你应该能看到：
- 完整的代码文件
- README.md 内容已正确渲染
- 所有 commit 历史
- 文件和文件夹结构

## 📤 后续更新

每次代码更新后：

```bash
# 1. 检查状态
git status

# 2. 添加变更
git add .

# 3. 提交
git commit -m "描述你的更改"

# 4. 推送到 GitHub
git push origin main

# 5. 创建新的 Release（如果是重要更新）
git tag v1.1.0
git push origin v1.1.0
```

## ⚠️ 常见问题

### Q: 我需要用 GitHub Desktop 吗？
A: 不需要，命令行就足够了。但如果你习惯 GUI，也可以使用 GitHub Desktop 或其他 Git GUI 工具。

### Q: 我改错了什么怎么办？
A: 如果本地还没推送，可以用 `git reset` 撤销：
```bash
git reset HEAD~1  # 撤销最后一次提交（保留修改）
```

### Q: 如何删除已推送的文件？
A: 
```bash
git rm --cached filename
git commit -m "Remove filename"
git push origin main
```

### Q: 项目需要保密吗？
A: 可以将仓库设置为私有，但公开会让更多人能使用和贡献。

## 🎉 完成！

发布后，你可以：
- ⭐ 邀请朋友给 Star
- 🐛 收集用户的 Issue 反馈
- 🔀 接受贡献者的 Pull Request
- 📢 在社交媒体上分享项目

---

**更多帮助**: 访问 [GitHub Help](https://docs.github.com)
