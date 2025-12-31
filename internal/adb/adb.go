package adb

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Device 表示一个 ADB 设备
type Device struct {
	Serial       string    // 设备序列号或 IP:PORT
	Status       string    // 设备状态
	Model        string    // 设备型号
	LastSeen     time.Time // 最后一次看到该设备的时间
	FailedChecks int       // 连续失败检查次数
}

// ADBManager ADB 管理器
type ADBManager struct {
	adbPath              string
	managedDevices       map[string]*Device // 使用Serial作为key的设备缓存
	lastDeviceListTime   time.Time
	deviceOfflineTimeout time.Duration // 设备离线超时时间（超过此时间无响应则删除）
	deviceCacheLock      sync.RWMutex  // 缓存锁
	useBusybox           bool          // 是否使用busybox执行命令
	busyboxLock          sync.RWMutex  // busybox开关锁
}

// NewADBManager 创建 ADB 管理器
func NewADBManager() *ADBManager {
	return &ADBManager{
		adbPath:              "adb", // 假设 adb 在系统 PATH 中
		managedDevices:       make(map[string]*Device),
		deviceOfflineTimeout: 5 * time.Minute, // 5分钟内无响应的设备才删除
		useBusybox:           false,
	}
}

// SetBusyboxEnabled 设置是否使用busybox
func (m *ADBManager) SetBusyboxEnabled(enabled bool) {
	m.busyboxLock.Lock()
	defer m.busyboxLock.Unlock()
	m.useBusybox = enabled
	if enabled {
		fmt.Println("[ADB] ✅ Busybox 模式已启用 - 命令执行前将加入 'busybox' 前缀")
	} else {
		fmt.Println("[ADB] ❌ Busybox 模式已禁用")
	}
}

// IsBusyboxEnabled 检查是否启用了busybox
func (m *ADBManager) IsBusyboxEnabled() bool {
	m.busyboxLock.RLock()
	defer m.busyboxLock.RUnlock()
	return m.useBusybox
}

// wrapCommandWithBusybox 如果启用了busybox，则给命令加上busybox前缀
func (m *ADBManager) wrapCommandWithBusybox(command string) string {
	m.busyboxLock.RLock()
	enabled := m.useBusybox
	m.busyboxLock.RUnlock()
	
	if enabled && !strings.HasPrefix(command, "busybox ") {
		return "busybox " + command
	}
	return command
}

// ListDevices 列出所有已连接的设备（使用增量式更新，避免全量覆盖）
func (m *ADBManager) ListDevices() ([]Device, error) {
	m.deviceCacheLock.Lock()
	defer m.deviceCacheLock.Unlock()

	cmd := exec.Command(m.adbPath, "devices", "-l")
	output, err := cmd.CombinedOutput() // 改用CombinedOutput来同时捕获stdout和stderr
	outputStr := ensureUTF8(string(output))

	// 检查是否有版本冲突信息
	if strings.Contains(outputStr, "doesn't match") {
		fmt.Println("[ADB] ⚠️ 检测到版本冲突警告!")
		fmt.Println("[ADB] ADB客户端版本与服务器版本不匹配，ADB已自动重启")
		fmt.Println("[ADB] 输出内容:")
		fmt.Println(outputStr)
		fmt.Println("[ADB] 解决方案:")
		fmt.Println("  1. 关闭所有使用ADB的工具")
		fmt.Println("  2. 手动执行: adb kill-server")
		fmt.Println("  3. 重启ADB: adb start-server")
		fmt.Println("  4. 或者更新ADB到最新版本")
		// 保持缓存设备列表不变，不清空
		return m.getDeviceList(), fmt.Errorf("ADB版本冲突，服务已重启，请解决版本问题后重试")
	}

	if err != nil {
		fmt.Printf("[ADB] 执行 adb devices 失败: %v，使用缓存的 %d 台设备\n", err, len(m.managedDevices))
		return m.getDeviceList(), fmt.Errorf("执行 adb devices 失败: %v", err)
	}

	// outputStr已经是编码转换过的，直接使用
	lines := strings.Split(outputStr, "\n")
	currentDevices := make(map[string]*Device) // 本次查询发现的设备

	// 解析adb devices输出
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// 检测真正的异常状态：仅当 Serial 为 "adb" 且 Status 为 "[server]" 时
			if len(parts) == 2 && parts[0] == "adb" && parts[1] == "[server]" {
				fmt.Printf("[ADB] 检测到异常状态: %s\n", line)
				fmt.Println("[ADB] 尝试重启 ADB 服务...")

				// 重启 ADB
				exec.Command(m.adbPath, "kill-server").Run()
				time.Sleep(500 * time.Millisecond)
				exec.Command(m.adbPath, "start-server").Run()
				time.Sleep(1 * time.Second)

				fmt.Println("[ADB] ADB 服务已重启，请重新导入设备")
				return nil, fmt.Errorf("ADB 服务已重启，请重新导入设备")
			}

			// 检测版本冲突信息并处理
			if strings.Contains(line, "server version") && strings.Contains(line, "doesn't match") {
				fmt.Printf("[ADB] 检测到版本冲突: %s\n", line)
				fmt.Println("[ADB] 版本冲突，正在重启 ADB 服务...")

				// 重启 ADB
				exec.Command(m.adbPath, "kill-server").Run()
				time.Sleep(500 * time.Millisecond)
				exec.Command(m.adbPath, "start-server").Run()
				time.Sleep(1 * time.Second)

				fmt.Println("[ADB] ADB 服务已重启，请重新导入设备")
				return nil, fmt.Errorf("ADB 服务已重启，请重新导入设备")
			}

			// 过滤掉其他异常条目：* [daemon] 等
			if parts[0] == "*" || parts[0] == "adb" || strings.HasPrefix(parts[1], "[") {
				fmt.Printf("[ADB] 跳过异常条目: %s\n", line)
				continue
			}

			serial := parts[0]
			device := &Device{
				Serial:   serial,
				Status:   parts[1],
				LastSeen: time.Now(),
			}

			// 提取设备型号
			for _, part := range parts[2:] {
				if strings.HasPrefix(part, "model:") {
					device.Model = strings.TrimPrefix(part, "model:")
					break
				}
			}

			currentDevices[serial] = device
		}
	}

	// 增量式更新：仅更新有变化的设备
	// 1. 更新本次发现的设备
	for serial, dev := range currentDevices {
		if existingDev, exists := m.managedDevices[serial]; exists {
			// 设备已存在，更新状态和型号，重置失败计数
			existingDev.Status = dev.Status
			existingDev.Model = dev.Model
			existingDev.LastSeen = time.Now()
			existingDev.FailedChecks = 0
			fmt.Printf("[ADB] 更新设备: %s -> %s\n", serial, dev.Status)
		} else {
			// 新发现的设备，加入管理
			m.managedDevices[serial] = dev
			fmt.Printf("[ADB] 新增设备: %s -> %s\n", serial, dev.Status)
		}
	}

	// 2. 标记本次没有发现但之前有的设备为失败
	now := time.Now()
	for serial, dev := range m.managedDevices {
		if _, found := currentDevices[serial]; !found {
			dev.FailedChecks++
			fmt.Printf("[ADB] 设备离线: %s (连续失败 %d 次)\n", serial, dev.FailedChecks)

			// 仅当连续失败3次（约15秒）才标记为离线
			// 或者超过5分钟未见到该设备才删除
			if dev.FailedChecks >= 3 || now.Sub(dev.LastSeen) > m.deviceOfflineTimeout {
				fmt.Printf("[ADB] 删除离线设备: %s (失败次数:%d, 离线时长:%v)\n",
					serial, dev.FailedChecks, now.Sub(dev.LastSeen))
				delete(m.managedDevices, serial)
			}
		}
	}

	m.lastDeviceListTime = time.Now()
	return m.getDeviceList(), nil
}

// DiagnoseADB 诊断ADB版本和状态
func (m *ADBManager) DiagnoseADB() (string, error) {
	var result strings.Builder
	
	result.WriteString("=== ADB 诊断报告 ===\n\n")
	
	// 1. 获取客户端版本
	cmd := exec.Command(m.adbPath, "version")
	output, _ := cmd.CombinedOutput()
	result.WriteString("【客户端版本】\n")
	result.WriteString(ensureUTF8(string(output)))
	result.WriteString("\n")
	
	// 2. 获取服务器信息
	cmd = exec.Command(m.adbPath, "shell", "getprop", "ro.build.version.release")
	output, _ = cmd.CombinedOutput()
	result.WriteString("【服务器状态】\n")
	if len(output) == 0 {
		result.WriteString("ADB 服务器未正常运行\n")
	} else {
		result.WriteString("ADB 服务器正常运行\n")
	}
	result.WriteString("\n")
	
	// 3. 推荐解决方案
	result.WriteString("【解决步骤】\n")
	result.WriteString("1. 打开命令行（cmd 或 PowerShell）\n")
	result.WriteString("2. 执行: adb kill-server\n")
	result.WriteString("3. 执行: adb start-server\n")
	result.WriteString("4. 如果问题持续，更新 ADB:\n")
	result.WriteString("   - 下载最新 Platform-Tools\n")
	result.WriteString("   - 替换本地的 adb.exe\n")
	
	return result.String(), nil
}

// getDeviceList 获取当前设备列表（已排序）
func (m *ADBManager) getDeviceList() []Device {
	devices := make([]Device, 0, len(m.managedDevices))
	for _, dev := range m.managedDevices {
		devices = append(devices, Device{
			Serial:   dev.Serial,
			Status:   dev.Status,
			Model:    dev.Model,
			LastSeen: dev.LastSeen,
		})
	}
	return devices
}

// Connect 连接到指定的设备（无线连接）
func (m *ADBManager) Connect(address string) error {
	fmt.Printf("[ADB] 尝试连接: %s\n", address)
	cmd := exec.Command(m.adbPath, "connect", address)
	output, err := cmd.CombinedOutput()
	outputStr := ensureUTF8(string(output))

	if err != nil {
		fmt.Printf("[ADB] 连接失败: %s - %v\n", address, err)
		return fmt.Errorf("连接失败: %v, 输出: %s", err, outputStr)
	}

	if !strings.Contains(outputStr, "connected") {
		fmt.Printf("[ADB] 连接失败: %s - %s\n", address, outputStr)
		return fmt.Errorf("连接失败: %s", outputStr)
	}

	fmt.Printf("[ADB] 连接成功: %s\n", address)
	return nil
}

// Disconnect 断开设备连接
func (m *ADBManager) Disconnect(serial string) error {
	cmd := exec.Command(m.adbPath, "disconnect", serial)
	_, err := cmd.Output()
	
	// 断开连接后，也从管理的设备列表中移除该设备
	if err == nil {
		m.deviceCacheLock.Lock()
		delete(m.managedDevices, serial)
		m.deviceCacheLock.Unlock()
		fmt.Printf("[ADB] 已断开并移除设备: %s\n", serial)
	}
	
	return err
}

// RemoveDevice 从管理列表中移除指定设备
func (m *ADBManager) RemoveDevice(serial string) error {
	m.deviceCacheLock.Lock()
	defer m.deviceCacheLock.Unlock()
	
	if _, exists := m.managedDevices[serial]; exists {
		delete(m.managedDevices, serial)
		fmt.Printf("[ADB] 已移除设备: %s\n", serial)
		return nil
	}
	
	return fmt.Errorf("设备 %s 不存在", serial)
}

// ExecuteCommand 在指定设备上执行命令
func (m *ADBManager) ExecuteCommand(serial, command string) (string, error) {
	// 如果启用了busybox，给命令加上前缀
	wrappedCommand := m.wrapCommandWithBusybox(command)
	
	fmt.Printf("[ADB] 执行命令: %s -> %s\n", serial, wrappedCommand)

	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "shell", wrappedCommand)
	} else {
		cmd = exec.Command(m.adbPath, "shell", wrappedCommand)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := ensureUTF8(string(output))
		
		// 检查是否有版本冲突信息
		if strings.Contains(outputStr, "doesn't match") {
			fmt.Println("[ADB] ⚠️ 检测到版本冲突警告!")
			fmt.Println("[ADB] ADB客户端版本与服务器版本不匹配")
			fmt.Println("[ADB] 解决方案: 更新ADB到最新版本或统一版本")
			return "", fmt.Errorf("ADB版本冲突: %v", err)
		}
		
		// 检查错误输出是否包含设备相关错误
		if strings.Contains(outputStr, "device not found") ||
			strings.Contains(outputStr, "unauthorized") ||
			strings.Contains(outputStr, "offline") {
			fmt.Printf("[ADB] 设备错误: %s -> %s: %v\n", serial, command, err)
			return outputStr, err
		}

		// 注意：不再自动重启 ADB 服务，因为单个命令失败不应该影响其他设备的连接
		// 如果需要重启服务，应该由用户手动操作或通过专门的诊断程序进行
		if strings.Contains(outputStr, "daemon") || strings.Contains(outputStr, "server") {
			fmt.Printf("[ADB] 服务相关错误: %s -> %s: %v\n", serial, command, err)
			fmt.Println("[ADB] 提示: 如果问题持续存在，请手动执行 'adb kill-server' 和 'adb start-server'")
		}

		fmt.Printf("[ADB] 命令失败: %s -> %s: %v\n", serial, command, err)
		return outputStr, err
	}

	fmt.Printf("[ADB] 命令成功: %s\n", serial)
	return ensureUTF8(string(output)), nil
}

// ensureUTF8 确保字符串是有效的 UTF-8，尝试从 GBK 转换
func ensureUTF8(s string) string {
	// 如果已经是有效的 UTF-8，直接返回
	if utf8.ValidString(s) {
		return s
	}

	// 尝试从 GBK 转换为 UTF-8
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Bytes, _, err := transform.Bytes(decoder, []byte(s))
	if err == nil && utf8.ValidString(string(utf8Bytes)) {
		return string(utf8Bytes)
	}

	// 如果 GBK 转换失败，尝试 GB18030
	decoder = simplifiedchinese.GB18030.NewDecoder()
	utf8Bytes, _, err = transform.Bytes(decoder, []byte(s))
	if err == nil && utf8.ValidString(string(utf8Bytes)) {
		return string(utf8Bytes)
	}

	// 如果都失败，移除无效字符
	v := make([]rune, 0, len(s))
	for _, r := range s {
		if r != utf8.RuneError {
			v = append(v, r)
		}
	}
	return string(v)
}

// ExecuteCommandWithTimeout 执行命令带超时
func (m *ADBManager) ExecuteCommandWithTimeout(serial, command string, timeout time.Duration) (string, error) {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "shell", command)
	} else {
		cmd = exec.Command(m.adbPath, "shell", command)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		return "", err
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			return "", fmt.Errorf("命令超时且无法终止: %v", err)
		}
		return "", fmt.Errorf("命令执行超时")
	case err := <-done:
		if err != nil {
			return errBuf.String(), err
		}
		return outBuf.String(), nil
	}
}

// PullFile 从设备拉取文件
func (m *ADBManager) PullFile(serial, remotePath, localPath string) error {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "pull", remotePath, localPath)
	} else {
		cmd = exec.Command(m.adbPath, "pull", remotePath, localPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("拉取文件失败: %v, 输出: %s", err, ensureUTF8(string(output)))
	}

	return nil
}

// PushFile 推送文件到设备
func (m *ADBManager) PushFile(serial, localPath, remotePath string) error {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "push", localPath, remotePath)
	} else {
		cmd = exec.Command(m.adbPath, "push", localPath, remotePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("推送文件失败: %v, 输出: %s", err, ensureUTF8(string(output)))
	}

	return nil
}

// Screenshot 截屏
func (m *ADBManager) Screenshot(serial, localPath string) error {
	// 使用通道实现超时控制
	type result struct {
		err error
	}
	ch := make(chan result, 1)

	go func() {
		// 在设备上截屏
		remotePath := "/sdcard/screenshot.png"
		_, err := m.ExecuteCommand(serial, fmt.Sprintf("screencap -p %s", remotePath))
		if err != nil {
			ch <- result{err: fmt.Errorf("截屏失败: %v", err)}
			return
		}

		// 拉取到本地
		if err := m.PullFile(serial, remotePath, localPath); err != nil {
			ch <- result{err: fmt.Errorf("拉取截屏失败: %v", err)}
			return
		}

		// 删除设备上的临时文件
		m.ExecuteCommand(serial, fmt.Sprintf("rm %s", remotePath))

		ch <- result{err: nil}
	}()

	// 等待10秒超时
	select {
	case res := <-ch:
		return res.err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("截图超时（10秒），请检查设备连接")
	}
}

// InstallApp 安装应用
func (m *ADBManager) InstallApp(serial, apkPath string) error {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "install", "-r", apkPath)
	} else {
		cmd = exec.Command(m.adbPath, "install", "-r", apkPath)
	}

	output, err := cmd.CombinedOutput()
	outputStr := ensureUTF8(string(output))

	if err != nil {
		return fmt.Errorf("安装失败: %v, 输出: %s", err, outputStr)
	}

	if !strings.Contains(outputStr, "Success") {
		return fmt.Errorf("安装失败: %s", outputStr)
	}

	return nil
}

// UninstallApp 卸载应用
func (m *ADBManager) UninstallApp(serial, packageName string) error {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "uninstall", packageName)
	} else {
		cmd = exec.Command(m.adbPath, "uninstall", packageName)
	}

	output, err := cmd.CombinedOutput()
	outputStr := ensureUTF8(string(output))

	if err != nil {
		return fmt.Errorf("卸载失败: %v, 输出: %s", err, outputStr)
	}

	if !strings.Contains(outputStr, "Success") {
		return fmt.Errorf("卸载失败: %s", outputStr)
	}

	return nil
}

// StartApp 启动应用
func (m *ADBManager) StartApp(serial, packageName string) error {
	// 获取应用的启动 Activity
	launchActivity, err := m.ExecuteCommand(serial,
		fmt.Sprintf("cmd package resolve-activity --brief %s | tail -n 1", packageName))
	if err != nil {
		return fmt.Errorf("获取启动 Activity 失败: %v", err)
	}

	launchActivity = strings.TrimSpace(launchActivity)
	if launchActivity == "" {
		return fmt.Errorf("未找到启动 Activity")
	}

	// 启动应用
	_, err = m.ExecuteCommand(serial,
		fmt.Sprintf("am start -n %s", launchActivity))
	if err != nil {
		return fmt.Errorf("启动应用失败: %v", err)
	}

	return nil
}

// StopApp 停止应用
func (m *ADBManager) StopApp(serial, packageName string) error {
	_, err := m.ExecuteCommand(serial, fmt.Sprintf("am force-stop %s", packageName))
	if err != nil {
		return fmt.Errorf("停止应用失败: %v", err)
	}
	return nil
}

// ListPackages 列出所有已安装的应用包
func (m *ADBManager) ListPackages(serial string) ([]string, error) {
	output, err := m.ExecuteCommand(serial, "pm list packages")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	packages := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkg := strings.TrimPrefix(line, "package:")
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

// GetDeviceInfo 获取设备信息
func (m *ADBManager) GetDeviceInfo(serial string) (map[string]string, error) {
	info := make(map[string]string)

	// 获取设备型号
	if model, err := m.ExecuteCommand(serial, "getprop ro.product.model"); err == nil {
		info["model"] = strings.TrimSpace(model)
	}

	// 获取 Android 版本
	if version, err := m.ExecuteCommand(serial, "getprop ro.build.version.release"); err == nil {
		info["android_version"] = strings.TrimSpace(version)
	}

	// 获取 SDK 版本
	if sdk, err := m.ExecuteCommand(serial, "getprop ro.build.version.sdk"); err == nil {
		info["sdk_version"] = strings.TrimSpace(sdk)
	}

	// 获取设备制造商
	if manufacturer, err := m.ExecuteCommand(serial, "getprop ro.product.manufacturer"); err == nil {
		info["manufacturer"] = strings.TrimSpace(manufacturer)
	}

	// 获取设备品牌
	if brand, err := m.ExecuteCommand(serial, "getprop ro.product.brand"); err == nil {
		info["brand"] = strings.TrimSpace(brand)
	}

	// 获取 CPU 架构
	if abi, err := m.ExecuteCommand(serial, "getprop ro.product.cpu.abi"); err == nil {
		info["cpu_abi"] = strings.TrimSpace(abi)
	}

	// 获取 IP 地址
	if ip, err := m.ExecuteCommand(serial, "ip -f inet addr show wlan0 | grep inet | awk '{print $2}' | cut -d/ -f1"); err == nil {
		info["ip_address"] = strings.TrimSpace(ip)
	}

	// 获取 MAC 地址
	if mac, err := m.ExecuteCommand(serial, "cat /sys/class/net/wlan0/address"); err == nil {
		info["mac_address"] = strings.TrimSpace(mac)
	}

	return info, nil
}

// ListFiles 列出目录下的文件
func (m *ADBManager) ListFiles(serial, path string) ([]FileInfo, error) {
	output, err := m.ExecuteCommand(serial, fmt.Sprintf("ls -la %s", path))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	files := make([]FileInfo, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		fileInfo := FileInfo{
			Permissions: fields[0],
			Owner:       fields[1],
			Group:       fields[2],
			Size:        fields[3],
			Date:        strings.Join(fields[4:7], " "),
			Name:        strings.Join(fields[7:], " "),
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// FileInfo 文件信息
type FileInfo struct {
	Name        string
	Permissions string
	Owner       string
	Group       string
	Size        string
	Date        string
}

// DeleteFile 删除文件
func (m *ADBManager) DeleteFile(serial, path string) error {
	_, err := m.ExecuteCommand(serial, fmt.Sprintf("rm -rf %s", path))
	return err
}

// RenameFile 重命名文件
func (m *ADBManager) RenameFile(serial, oldPath, newPath string) error {
	_, err := m.ExecuteCommand(serial, fmt.Sprintf("mv %s %s", oldPath, newPath))
	return err
}

// ChangePermissions 修改文件权限
func (m *ADBManager) ChangePermissions(serial, path, permissions string) error {
	_, err := m.ExecuteCommand(serial, fmt.Sprintf("chmod %s %s", permissions, path))
	return err
}

// TryEnableRoot 尝试获取 root 权限
func (m *ADBManager) TryEnableRoot(serial string) error {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "root")
	} else {
		cmd = exec.Command(m.adbPath, "root")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启用 root 失败: %v, 输出: %s", err, ensureUTF8(string(output)))
	}

	return nil
}

// ExecuteAsRoot 以 root 权限执行命令
func (m *ADBManager) ExecuteAsRoot(serial, command string) (string, error) {
	// 尝试使用 su 执行命令
	rootCmd := fmt.Sprintf("su -c \"%s\"", command)
	output, err := m.ExecuteCommand(serial, rootCmd)

	// 如果 su 失败，尝试 adb root
	if err != nil {
		// 尝试 adb root
		if rootErr := m.TryEnableRoot(serial); rootErr == nil {
			// root 成功后重新执行
			return m.ExecuteCommand(serial, command)
		}
	}

	return output, err
}

// CheckRootAccess 检查是否有 root 权限
func (m *ADBManager) CheckRootAccess(serial string) bool {
	output, err := m.ExecuteCommand(serial, "su -c 'id'")
	if err != nil {
		// 如果 su 失败，尝试 adb root
		if m.TryEnableRoot(serial) == nil {
			output, err = m.ExecuteCommand(serial, "id")
		}
	}

	if err != nil {
		return false
	}

	// 检查输出是否包含 uid=0 (root)
	return strings.Contains(output, "uid=0")
}

// GetProcessList 获取进程列表
func (m *ADBManager) GetProcessList(serial string) (string, error) {
	// 尝试不同的 ps 命令参数（兼容不同 Android 版本）
	output, err := m.ExecuteCommand(serial, "ps -A")
	if err != nil {
		// 如果 -A 失败，尝试不带参数
		output, err = m.ExecuteCommand(serial, "ps")
		if err != nil {
			// 如果还失败，尝试 -ef
			output, err = m.ExecuteCommand(serial, "ps -ef")
		}
	}
	return output, err
}

// GetNetworkConnections 获取网络连接信息
func (m *ADBManager) GetNetworkConnections(serial string) (string, error) {
	return m.ExecuteCommand(serial, "netstat -anp")
}

// InteractiveShell 创建交互式 shell
func (m *ADBManager) InteractiveShell(serial string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "shell")
	} else {
		cmd = exec.Command(m.adbPath, "shell")
	}

	return cmd, nil
}

// ExecuteCommandStream 执行命令并返回输出流
func (m *ADBManager) ExecuteCommandStream(serial, command string) (string, error) {
	var cmd *exec.Cmd
	if serial != "" {
		cmd = exec.Command(m.adbPath, "-s", serial, "shell", command)
	} else {
		cmd = exec.Command(m.adbPath, "shell", command)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var output strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output.WriteString(scanner.Text())
		output.WriteString("\n")
	}

	if err := cmd.Wait(); err != nil {
		return output.String(), err
	}

	return output.String(), nil
}
