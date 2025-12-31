package batch

import (
	"adbmanager/internal/adb"
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// BatchManager 批量操作管理器
type BatchManager struct {
	adbMgr  *adb.ADBManager
	targets []string
	mu      sync.Mutex
}

// NewBatchManager 创建批量操作管理器
func NewBatchManager(adbMgr *adb.ADBManager) *BatchManager {
	return &BatchManager{
		adbMgr:  adbMgr,
		targets: make([]string, 0),
	}
}

// ImportTargetsFromFile 从文件导入目标
func (bm *BatchManager) ImportTargetsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	bm.mu.Lock()
	defer bm.mu.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 验证 IP:PORT 格式
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				bm.targets = append(bm.targets, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	return nil
}

// AddTarget 添加单个目标
func (bm *BatchManager) AddTarget(target string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// 检查是否已存在
	for _, t := range bm.targets {
		if t == target {
			return
		}
	}

	bm.targets = append(bm.targets, target)
}

// RemoveTarget 移除目标
func (bm *BatchManager) RemoveTarget(target string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i, t := range bm.targets {
		if t == target {
			bm.targets = append(bm.targets[:i], bm.targets[i+1:]...)
			return
		}
	}
}

// GetTargets 获取所有目标
func (bm *BatchManager) GetTargets() []string {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	targets := make([]string, len(bm.targets))
	copy(targets, bm.targets)
	return targets
}

// ClearTargets 清空目标列表
func (bm *BatchManager) ClearTargets() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.targets = make([]string, 0)
}

// BatchConnectResult 批量连接结果
type BatchConnectResult struct {
	Target  string
	Success bool
	Error   error
}

// BatchConnect 批量连接设备
func (bm *BatchManager) BatchConnect(callback func(result BatchConnectResult)) {
	targets := bm.GetTargets()

	var wg sync.WaitGroup
	resultChan := make(chan BatchConnectResult, len(targets))

	// 启动协程处理回调
	var callbackWg sync.WaitGroup
	callbackWg.Add(1)
	go func() {
		defer callbackWg.Done()
		for result := range resultChan {
			if callback != nil {
				callback(result)
			}
		}
	}()

	// 并发连接所有目标
	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			err := bm.adbMgr.Connect(t)
			resultChan <- BatchConnectResult{
				Target:  t,
				Success: err == nil,
				Error:   err,
			}
		}(target)
	}

	wg.Wait()
	close(resultChan)
	// 等待回调处理完成
	callbackWg.Wait()
}

// CommandResult 命令执行结果
type CommandResult struct {
	Device string
	Output string
	Error  error
}

// BatchExecuteCommand 批量执行命令
func (bm *BatchManager) BatchExecuteCommand(devices []string, command string, callback func(result CommandResult)) {
	var wg sync.WaitGroup
	resultChan := make(chan CommandResult, len(devices))

	// 启动协程处理回调
	go func() {
		for result := range resultChan {
			if callback != nil {
				callback(result)
			}
		}
	}()

	// 并发在所有设备上执行命令
	for _, device := range devices {
		wg.Add(1)
		go func(dev string) {
			defer wg.Done()

			output, err := bm.adbMgr.ExecuteCommand(dev, command)
			resultChan <- CommandResult{
				Device: dev,
				Output: output,
				Error:  err,
			}
		}(device)
	}

	wg.Wait()
	close(resultChan)
}

// BatchInstallApp 批量安装应用
func (bm *BatchManager) BatchInstallApp(devices []string, apkPath string, callback func(device string, err error)) {
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go func(dev string) {
			defer wg.Done()

			err := bm.adbMgr.InstallApp(dev, apkPath)
			if callback != nil {
				callback(dev, err)
			}
		}(device)
	}

	wg.Wait()
}

// BatchUninstallApp 批量卸载应用
func (bm *BatchManager) BatchUninstallApp(devices []string, packageName string, callback func(device string, err error)) {
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go func(dev string) {
			defer wg.Done()

			err := bm.adbMgr.UninstallApp(dev, packageName)
			if callback != nil {
				callback(dev, err)
			}
		}(device)
	}

	wg.Wait()
}

// BatchPushFile 批量推送文件
func (bm *BatchManager) BatchPushFile(devices []string, localPath, remotePath string, callback func(device string, err error)) {
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go func(dev string) {
			defer wg.Done()

			err := bm.adbMgr.PushFile(dev, localPath, remotePath)
			if callback != nil {
				callback(dev, err)
			}
		}(device)
	}

	wg.Wait()
}

// BatchScreenshot 批量截屏
func (bm *BatchManager) BatchScreenshot(devices []string, outputDir string, callback func(device, filepath string, err error)) {
	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go func(dev string) {
			defer wg.Done()

			// 创建唯一的文件名
			filename := fmt.Sprintf("%s/%s_screenshot.png", outputDir, strings.ReplaceAll(dev, ":", "_"))
			err := bm.adbMgr.Screenshot(dev, filename)

			if callback != nil {
				callback(dev, filename, err)
			}
		}(device)
	}

	wg.Wait()
}

// ExportTargetsToFile 导出目标到文件
func (bm *BatchManager) ExportTargetsToFile(filePath string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, target := range bm.targets {
		_, err := writer.WriteString(target + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	return writer.Flush()
}
