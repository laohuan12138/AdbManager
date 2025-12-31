package ui

import (
	"adbmanager/internal/adb"
	"adbmanager/internal/scanner"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ScannerUI 敏感信息扫描界面
type ScannerUI struct {
	window    fyne.Window
	scanner   *scanner.Scanner
	adbMgr    *adb.ADBManager
	getDevice func() string
}

// NewScannerUI 创建敏感信息扫描界面
func NewScannerUI(window fyne.Window, scanner *scanner.Scanner, adbMgr *adb.ADBManager, getDevice func() string) *ScannerUI {
	return &ScannerUI{
		window:    window,
		scanner:   scanner,
		adbMgr:    adbMgr,
		getDevice: getDevice,
	}
}

// Build 构建敏感信息扫描界面
func (s *ScannerUI) Build() fyne.CanvasObject {
	// 结果显示区域
	resultText := widget.NewMultiLineEntry()
	resultText.Wrapping = fyne.TextWrapWord
	resultText.TextStyle = fyne.TextStyle{Monospace: true}

	// 扫描配置文件
	scanConfigBtn := widget.NewButton("扫描配置文件", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		resultText.SetText("正在扫描配置文件...\n")

		results, err := s.scanner.ScanConfigFiles(device)
		if err != nil {
			showError(s.window, "扫描失败", err)
			return
		}

		result := "========== 配置文件扫描结果 ==========\n"
		result += fmt.Sprintf("发现敏感信息: %d 条\n\n", len(results))

		for i, info := range results {
			result += fmt.Sprintf("%d. 类型: %s\n", i+1, info.Type)
			result += fmt.Sprintf("   文件: %s\n", info.FilePath)
			result += fmt.Sprintf("   位置: %s\n", info.Line)
			result += fmt.Sprintf("   内容: %s\n\n", info.Value)
		}

		if len(results) == 0 {
			result += "未发现敏感信息或无权限访问配置文件\n"
		}

		resultText.SetText(result)
	})

	// 扫描 SharedPreferences
	scanPrefsBtn := widget.NewButton("扫描 SharedPreferences", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		// 输入包名
		pkgEntry := widget.NewEntry()
		pkgEntry.SetPlaceHolder("输入应用包名")

		dialog.ShowCustomConfirm("扫描 SharedPreferences", "开始扫描", "取消",
			pkgEntry,
			func(confirmed bool) {
				if !confirmed {
					return
				}

				packageName := pkgEntry.Text
				if packageName == "" {
					return
				}

				resultText.SetText(fmt.Sprintf("正在扫描 %s 的 SharedPreferences...\n", packageName))

				results, err := s.scanner.ScanSharedPreferences(device, packageName)
				if err != nil {
					showError(s.window, "扫描失败", err)
					return
				}

				result := fmt.Sprintf("========== %s SharedPreferences 扫描结果 ==========\n", packageName)
				result += fmt.Sprintf("发现敏感信息: %d 条\n\n", len(results))

				for i, info := range results {
					result += fmt.Sprintf("%d. 类型: %s\n", i+1, info.Type)
					result += fmt.Sprintf("   文件: %s\n", info.FilePath)
					result += fmt.Sprintf("   内容: %s\n\n", info.Value)
				}

				if len(results) == 0 {
					result += "未发现敏感信息\n"
				}

				resultText.SetText(result)
			}, s.window)
	})

	// 扫描数据库
	scanDbBtn := widget.NewButton("列出数据库文件", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		// 输入包名
		pkgEntry := widget.NewEntry()
		pkgEntry.SetPlaceHolder("输入应用包名")

		dialog.ShowCustomConfirm("列出数据库", "查询", "取消",
			pkgEntry,
			func(confirmed bool) {
				if !confirmed {
					return
				}

				packageName := pkgEntry.Text
				if packageName == "" {
					return
				}

				databases, err := s.scanner.ScanDatabases(device, packageName)
				if err != nil {
					showError(s.window, "查询失败", err)
					return
				}

				result := fmt.Sprintf("========== %s 数据库文件 ==========\n", packageName)
				result += fmt.Sprintf("发现数据库: %d 个\n\n", len(databases))

				for i, db := range databases {
					result += fmt.Sprintf("%d. %s\n", i+1, db)
				}

				if len(databases) == 0 {
					result += "未发现数据库文件或无权限访问\n"
				} else {
					result += "\n提示: 可以使用 pull 命令下载数据库文件进行分析\n"
				}

				resultText.SetText(result)
			}, s.window)
	})

	// 扫描日志文件
	scanLogsBtn := widget.NewButton("扫描日志文件", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		resultText.SetText("正在扫描日志文件...\n")

		results, err := s.scanner.ScanLogFiles(device)
		if err != nil {
			showError(s.window, "扫描失败", err)
			return
		}

		result := "========== 日志文件扫描结果 ==========\n"
		result += fmt.Sprintf("发现敏感信息: %d 条\n\n", len(results))

		for i, info := range results {
			result += fmt.Sprintf("%d. 类型: %s\n", i+1, info.Type)
			result += fmt.Sprintf("   文件: %s\n", info.FilePath)
			result += fmt.Sprintf("   位置: %s\n", info.Line)
			result += fmt.Sprintf("   内容: %s\n\n", info.Value)
		}

		if len(results) == 0 {
			result += "未发现敏感信息\n"
		}

		resultText.SetText(result)
	})

	// 提升 Root 权限按钮
	enableRootBtn := widget.NewButton("⚡ 提升 Root 权限", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		resultText.SetText("正在尝试获取 root 权限...\n")

		if err := s.adbMgr.TryEnableRoot(device); err != nil {
			resultText.SetText(fmt.Sprintf("获取 root 权限失败: %v\n\n请确保:\n1. 设备已 root\n2. 或者使用 adb root 命令 (eng/userdebug 版本)", err))
			return
		}

		// 检查 root 状态
		if s.adbMgr.CheckRootAccess(device) {
			resultText.SetText("✓ Root 权限获取成功！\n\n现在可以执行需要 root 权限的扫描操作")
			showInfo(s.window, "成功", "Root 权限已启用")
		} else {
			resultText.SetText("✗ 未能获取 root 权限\n\n某些扫描功能可能无法正常使用")
		}
	})

	// 检查Root状态
	checkRootBtn := widget.NewButton("检查Root状态", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		isRooted, err := s.scanner.CheckRootStatus(device)
		if err != nil {
			showError(s.window, "检查失败", err)
			return
		}

		result := "========== Root 状态检查 ==========\n"
		if isRooted {
			result += "✓ 设备已 Root\n"
			result += "警告: Root 设备存在安全风险\n"
		} else {
			result += "✗ 设备未 Root\n"
		}

		resultText.SetText(result)
	})

	// 查看证书
	certBtn := widget.NewButton("查看已安装证书", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		certs, err := s.scanner.GetInstalledCertificates(device)
		if err != nil {
			showError(s.window, "获取证书失败", err)
			return
		}

		result := "========== 已安装证书 ==========\n"
		result += certs

		resultText.SetText(result)
	})

	// 获取应用数据大小
	appDataBtn := widget.NewButton("查看应用数据", func() {
		device := s.getDevice()
		if device == "" {
			showError(s.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		// 输入包名
		pkgEntry := widget.NewEntry()
		pkgEntry.SetPlaceHolder("输入应用包名")

		dialog.ShowCustomConfirm("查看应用数据", "查询", "取消",
			pkgEntry,
			func(confirmed bool) {
				if !confirmed {
					return
				}

				packageName := pkgEntry.Text
				if packageName == "" {
					return
				}

				info, err := s.scanner.GetAppDataSize(device, packageName)
				if err != nil {
					showError(s.window, "查询失败", err)
					return
				}

				result := fmt.Sprintf("========== %s 应用数据 ==========\n", packageName)
				for key, value := range info {
					result += fmt.Sprintf("%s: %s\n", key, value)
				}

				resultText.SetText(result)
			}, s.window)
	})

	// 清空按钮
	clearBtn := widget.NewButton("清空", func() {
		resultText.SetText("")
	})

	// 布局
	buttonBox := container.NewGridWithColumns(3,
		scanConfigBtn,
		scanPrefsBtn,
		scanDbBtn,
	)

	buttonBox2 := container.NewGridWithColumns(4,
		scanLogsBtn,
		checkRootBtn,
		enableRootBtn,
		certBtn,
	)

	buttonBox3 := container.NewGridWithColumns(2,
		appDataBtn,
		clearBtn,
	)

	infoLabel := widget.NewLabel("提示: 某些操作需要设备Root权限或应用调试权限")
	infoLabel.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		container.NewVBox(
			infoLabel,
			widget.NewSeparator(),
			buttonBox,
			buttonBox2,
			buttonBox3,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewScroll(resultText),
	)
}
