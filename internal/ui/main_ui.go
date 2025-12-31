package ui

import (
	"adbmanager/internal/adb"
	"adbmanager/internal/batch"
	"adbmanager/internal/collector"
	"adbmanager/internal/scanner"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// MainUI 主界面
type MainUI struct {
	window    fyne.Window
	adbMgr    *adb.ADBManager
	batchMgr  *batch.BatchManager
	collector *collector.Collector
	scanner   *scanner.Scanner

	selectedDevices []string

	// UI 组件
	deviceList   *widget.List
	tabContainer *container.AppTabs
}

// NewMainUI 创建主界面
func NewMainUI(window fyne.Window) *MainUI {
	adbMgr := adb.NewADBManager()
	batchMgr := batch.NewBatchManager(adbMgr)
	collector := collector.NewCollector(adbMgr)
	scanner := scanner.NewScanner(adbMgr)

	return &MainUI{
		window:          window,
		adbMgr:          adbMgr,
		batchMgr:        batchMgr,
		collector:       collector,
		scanner:         scanner,
		selectedDevices: make([]string, 0),
	}
}

// Build 构建主界面
func (m *MainUI) Build() fyne.CanvasObject {
	// 创建各个功能标签页
	deviceTab := m.buildDeviceTab()
	shellTab := m.buildShellTab()
	fileTab := m.buildFileTab()
	infoTab := m.buildInfoTab()
	collectorTab := m.buildCollectorTab()
	appTab := m.buildAppTab()
	scannerTab := m.buildScannerTab()
	batchTab := m.buildBatchTab()

	// 创建标签页容器
	m.tabContainer = container.NewAppTabs(
		container.NewTabItem("设备管理", deviceTab),
		container.NewTabItem("命令执行", shellTab),
		container.NewTabItem("文件管理", fileTab),
		container.NewTabItem("设备信息", infoTab),
		container.NewTabItem("信息采集", collectorTab),
		container.NewTabItem("应用管理", appTab),
		container.NewTabItem("敏感信息", scannerTab),
		container.NewTabItem("批量操作", batchTab),
	)

	return m.tabContainer
}

// buildDeviceTab 构建设备管理标签页
func (m *MainUI) buildDeviceTab() fyne.CanvasObject {
	// 设备列表
	devices := make([]adb.Device, 0)
	deviceStrings := make([]string, 0)

	m.deviceList = widget.NewList(
		func() int {
			return len(deviceStrings)
		},
		func() fyne.CanvasObject {
			// 创建带状态指示器的列表项
			statusCircle := canvas.NewCircle(color.NRGBA{R: 0, G: 255, B: 0, A: 255})
			statusCircle.Resize(fyne.NewSize(20, 20))
			statusCircle.StrokeWidth = 0

			// 创建美化的状态标签（半透明、圆角）
			statusLabel := widget.NewLabel("在线")
			statusLabel.TextStyle = fyne.TextStyle{Bold: true}

			// 使用半透明绿色背景
			statusBg := canvas.NewRectangle(color.NRGBA{R: 76, G: 175, B: 80, A: 200})
			statusBg.CornerRadius = 4 // 圆角

			statusContainer := container.NewStack(
				statusBg,
				container.NewCenter(statusLabel),
			)

			// Shell 按钮
			shellBtn := widget.NewButton("💻 Shell", nil)
			shellBtn.Importance = widget.LowImportance

			return container.NewHBox(
				widget.NewCheck("", nil),
				statusCircle,
				widget.NewLabel("设备信息"),
				statusContainer,
				shellBtn,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(deviceStrings) {
				box := obj.(*fyne.Container)
				check := box.Objects[0].(*widget.Check)
				statusCircle := box.Objects[1].(*canvas.Circle)
				label := box.Objects[2].(*widget.Label)
				statusContainer := box.Objects[3].(*fyne.Container)
				shellBtn := box.Objects[4].(*widget.Button)
				statusBg := statusContainer.Objects[0].(*canvas.Rectangle)
				statusLabelContainer := statusContainer.Objects[1].(*fyne.Container)
				statusLabel := statusLabelContainer.Objects[0].(*widget.Label)

				label.SetText(deviceStrings[id])

				// 恢复勾选状态（修复滚动bug）
				serial := devices[id].Serial
				isSelected := false
				for _, s := range m.selectedDevices {
					if s == serial {
						isSelected = true
						break
					}
				}

				// 先设置回调为nil，避免触发事件
				check.OnChanged = nil
				check.SetChecked(isSelected)
				// 然后再绑定回调
				check.OnChanged = func(checked bool) {
					if checked {
						m.addSelectedDevice(serial)
					} else {
						m.removeSelectedDevice(serial)
					}
				}

				// 根据设备状态设置颜色和文本
				isOnline := devices[id].Status == "device"
				if isOnline {
					statusCircle.FillColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Material 绿
					statusBg.FillColor = color.NRGBA{R: 76, G: 175, B: 80, A: 200}     // 半透明绿
					statusLabel.SetText("在线")
					shellBtn.Show()
				} else if devices[id].Status == "offline" {
					statusCircle.FillColor = color.NRGBA{R: 244, G: 67, B: 54, A: 255} // Material 红
					statusBg.FillColor = color.NRGBA{R: 244, G: 67, B: 54, A: 200}     // 半透明红
					statusLabel.SetText("离线")
					shellBtn.Hide()
				} else {
					statusCircle.FillColor = color.NRGBA{R: 255, G: 152, B: 0, A: 255} // Material 橙
					statusBg.FillColor = color.NRGBA{R: 255, G: 152, B: 0, A: 200}     // 半透明橙
					statusLabel.SetText(devices[id].Status)
					shellBtn.Hide()
				}
				statusCircle.Refresh()
				statusBg.Refresh()

				// Shell 按钮点击事件
				shellBtn.OnTapped = func() {
					m.openShellWindow(devices[id].Serial, devices[id].Model)
				}
			}
		},
	)

	// 刷新设备列表
	refreshDevices := func() {
		devs, err := m.adbMgr.ListDevices()
		if err != nil {
			// 如果获取设备列表失败，不更新UI，保持之前的设备列表
			fmt.Printf("[UI] 获取设备列表失败: %v\n", err)
			fmt.Printf("[UI] 保持之前的 %d 台设备\n", len(devices))
			showError(m.window, "获取设备列表失败（保持上次结果）", err)
			return
		}

		// 更新设备列表（包括清空为0的情况）
		devices = devs
		deviceStrings = make([]string, len(devs))
		for i, dev := range devs {
			deviceStrings[i] = dev.Serial + " - " + dev.Status + " - " + dev.Model
		}

		// 调试信息
		fmt.Printf("刷新设备列表: 共 %d 台设备\n", len(devs))
		for _, dev := range devs {
			fmt.Printf("  - %s [%s]\n", dev.Serial, dev.Status)
		}

		m.deviceList.Refresh()
	}

	// 按钮
	refreshBtn := widget.NewButton("刷新设备列表", func() {
		refreshDevices()
	})

	// 诊断按钮
	diagnoseBtn := widget.NewButton("诊断 ADB 问题", func() {
		diagnosis, err := m.adbMgr.DiagnoseADB()
		if err != nil {
			showError(m.window, "诊断失败", err)
			return
		}
		// 显示诊断信息在消息窗口中
		showInfo(m.window, "ADB 诊断报告", diagnosis)
	})

	// 无线连接输入
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("输入 IP:PORT (例如: 192.168.1.100:5555)")

	connectBtn := widget.NewButton("连接", func() {
		address := ipEntry.Text
		if address == "" {
			showError(m.window, "连接失败", nil)
			return
		}

		err := m.adbMgr.Connect(address)
		if err != nil {
			showError(m.window, "连接失败", err)
			return
		}

		showInfo(m.window, "连接成功", "已成功连接到设备: "+address)
		refreshDevices()
	})

	disconnectBtn := widget.NewButton("断开选中设备", func() {
		if len(m.selectedDevices) == 0 {
			showInfo(m.window, "提示", "请先选择要断开的设备")
			return
		}

		for _, serial := range m.selectedDevices {
			m.adbMgr.Disconnect(serial)
		}

		m.selectedDevices = make([]string, 0)
		refreshDevices()
		showInfo(m.window, "成功", "已断开选中的设备")
	})

	// 从文件导入设备
	importBtn := widget.NewButton("从文件导入", func() {
		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()

			filePath := uc.URI().Path()
			err = m.batchMgr.ImportTargetsFromFile(filePath)
			if err != nil {
				showError(m.window, "导入失败", err)
				return
			}

			// 批量连接导入的设备
			targets := m.batchMgr.GetTargets()
			if len(targets) == 0 {
				showInfo(m.window, "提示", "文件中没有有效的设备地址")
				return
			}

			// 显示进度对话框
			resultText := fmt.Sprintf("正在连接 %d 个设备...\n\n", len(targets))
			resultEntry := widget.NewMultiLineEntry()
			resultEntry.SetText(resultText)
			resultEntry.Wrapping = fyne.TextWrapWord
			resultEntry.TextStyle = fyne.TextStyle{Monospace: true}

			resultDialog := dialog.NewCustom("批量连接结果", "关闭",
				container.NewScroll(resultEntry), m.window)
			resultDialog.Resize(fyne.NewSize(500, 400))
			resultDialog.Show()

			// 执行批量连接
			successCount := 0
			totalCount := len(targets)
			currentCount := 0

			m.batchMgr.BatchConnect(func(result batch.BatchConnectResult) {
				var status string
				if result.Success {
					status = "✓ 成功"
					successCount++
				} else {
					status = "✗ 失败: " + result.Error.Error()
				}

				resultText += fmt.Sprintf("%s - %s\n", result.Target, status)
				resultEntry.SetText(resultText)

				currentCount++
				// 所有连接完成后刷新
				if currentCount == totalCount {
					resultText += fmt.Sprintf("\n批量连接完成！成功 %d 个，失败 %d 个\n", successCount, totalCount-successCount)
					if successCount > 0 {
						resultText += "\n正在刷新设备列表..."
					}
					resultEntry.SetText(resultText)
				}
			})

			// BatchConnect现在是同步的，执行到这里时所有连接已完成
			if successCount > 0 {
				// 延迟1秒给ADB服务器留点时间
				time.AfterFunc(1*time.Second, func() {
					refreshDevices()
					resultEntry.SetText(resultEntry.Text + "\n✓ 设备列表已刷新")
				})
			}

			// 清空目标列表
			m.batchMgr.ClearTargets()
		}, m.window)
	})

	// 检查设备状态按钮
	checkStatusBtn := widget.NewButton("检查设备状态", func() {
		refreshDevices()

		// 统计在线设备数量
		onlineCount := 0
		offlineCount := 0
		otherCount := 0

		for _, dev := range devices {
			if dev.Status == "device" {
				onlineCount++
			} else if dev.Status == "offline" {
				offlineCount++
			} else {
				otherCount++
			}
		}

		message := fmt.Sprintf("设备状态已更新！\n\n"+
			"总计: %d 台设备\n"+
			"● 在线: %d 台\n"+
			"● 离线: %d 台\n"+
			"● 其他: %d 台",
			len(devices), onlineCount, offlineCount, otherCount)

		showInfo(m.window, "设备状态检查", message)
	})

	// 移除离线设备按钮
	removeOfflineBtn := widget.NewButton("移除离线设备", func() {
		refreshDevices()

		// 统计离线设备
		offlineDevices := make([]string, 0)
		for _, dev := range devices {
			if dev.Status == "offline" {
				offlineDevices = append(offlineDevices, dev.Serial)
			}
		}

		if len(offlineDevices) == 0 {
			showInfo(m.window, "提示", "没有离线设备")
			return
		}

		// 显示详细列表
		deviceList := "\n"
		for _, serial := range offlineDevices {
			deviceList += "  • " + serial + "\n"
		}

		dialog.ShowConfirm("警告：不可恢复操作",
			fmt.Sprintf("将移除以下 %d 台离线设备：%s\n移除后需重新连接，确定继续吗？", len(offlineDevices), deviceList),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				// 移除离线设备
				for _, serial := range offlineDevices {
					// 先断开连接
					m.adbMgr.Disconnect(serial)
					// 再从缓存中移除
					m.adbMgr.RemoveDevice(serial)
				}

				// 刷新设备列表并更新UI
				refreshDevices()
				showInfo(m.window, "成功", fmt.Sprintf("已移除 %d 台离线设备", len(offlineDevices)))
			}, m.window)
	})

	// 初始加载
	refreshDevices()

	// 布局
	connectBox := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("无线连接:"),
		connectBtn,
		ipEntry,
	)

	buttonBox := container.NewGridWithColumns(6,
		refreshBtn,
		disconnectBtn,
		importBtn,
		checkStatusBtn,
		removeOfflineBtn,
		diagnoseBtn,
	)

	return container.NewBorder(
		container.NewVBox(
			connectBox,
			buttonBox,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		m.deviceList,
	)
}

// buildShellTab 构建命令执行标签页
func (m *MainUI) buildShellTab() fyne.CanvasObject {
	// 命令输入
	commandEntry := widget.NewEntry()
	commandEntry.SetPlaceHolder("输入 ADB Shell 命令")

	// 输出显示 - 使用 Label 而不是 Entry，提高可读性
	outputText := widget.NewMultiLineEntry()
	outputText.Wrapping = fyne.TextWrapWord
	// 设置等宽字体和黑色文字
	outputText.TextStyle = fyne.TextStyle{Monospace: true}

	// Busybox 开关
	busyboxCheck := widget.NewCheck("启用 Busybox 模式", func(checked bool) {
		m.adbMgr.SetBusyboxEnabled(checked)
		if checked {
			outputText.SetText("✅ Busybox 模式已启用 - 命令执行时将自动加入 'busybox' 前缀\n")
		} else {
			outputText.SetText("❌ Busybox 模式已禁用\n")
		}
	})
	// 设置初始状态（不会触发回调）
	busyboxCheck.Checked = m.adbMgr.IsBusyboxEnabled()

	// 执行命令的通用函数
	executeCommand := func(command string) {
		if command == "" {
			return
		}

		if len(m.selectedDevices) == 0 {
			showError(m.window, "错误", nil)
			outputText.SetText("请先选择设备")
			return
		}

		outputText.SetText("正在执行命令...\n")

		// 在选中的设备上执行命令
		for _, device := range m.selectedDevices {
			output, err := m.adbMgr.ExecuteCommand(device, command)

			result := "\n========== " + device + " ==========\n"
			if err != nil {
				result += "错误: " + err.Error() + "\n"
			}
			result += output + "\n"

			outputText.SetText(outputText.Text + result)
		}
	}

	// 执行按钮
	executeBtn := widget.NewButton("执行命令", func() {
		executeCommand(commandEntry.Text)
	})

	clearBtn := widget.NewButton("清空输出", func() {
		outputText.SetText("")
	})

	// 快捷命令
	quickCommands := []struct {
		name    string
		command string
	}{
		{"获取屏幕分辨率", "wm size"},
		{"获取电池信息", "dumpsys battery"},
		{"列出进程", "ps"},
		{"获取网络连接", "netstat -an"},
		{"查看内存使用", "cat /proc/meminfo"},
		{"查看CPU信息", "cat /proc/cpuinfo"},
	}

	quickBtns := make([]fyne.CanvasObject, 0)
	for _, cmd := range quickCommands {
		cmdCopy := cmd // 避免闭包问题
		btn := widget.NewButton(cmdCopy.name, func() {
			// 填充命令到输入框
			commandEntry.SetText(cmdCopy.command)
			// 直接执行命令
			executeCommand(cmdCopy.command)
		})
		quickBtns = append(quickBtns, btn)
	}

	quickCmdBox := container.NewVBox(
		widget.NewLabel("快捷命令:"),
		container.NewGridWithColumns(3, quickBtns...),
	)

	// 控制面板：busybox开关 + 输入框 + 执行按钮
	controlPanel := container.NewVBox(
		busyboxCheck,
		widget.NewSeparator(),
		container.NewBorder(
			nil,
			nil,
			nil,
			container.NewHBox(executeBtn, clearBtn),
			commandEntry,
		),
	)

	return container.NewBorder(
		container.NewVBox(
			quickCmdBox,
			widget.NewSeparator(),
			controlPanel,
		),
		nil,
		nil,
		nil,
		container.NewScroll(outputText),
	)
}

// buildFileTab 构建文件管理标签页
func (m *MainUI) buildFileTab() fyne.CanvasObject {
	return NewFileManagerUI(m.window, m.adbMgr, m.getSelectedDevice).Build()
}

// buildInfoTab 构建设备信息标签页
func (m *MainUI) buildInfoTab() fyne.CanvasObject {
	return NewDeviceInfoUI(m.window, m.adbMgr, m.getSelectedDevice).Build()
}

// buildCollectorTab 构建信息采集标签页
func (m *MainUI) buildCollectorTab() fyne.CanvasObject {
	return NewCollectorUI(m.window, m.collector, m.getSelectedDevice).Build()
}

// buildAppTab 构建应用管理标签页
func (m *MainUI) buildAppTab() fyne.CanvasObject {
	return NewAppManagerUI(m.window, m.adbMgr, m.getSelectedDevice).Build()
}

// buildScannerTab 构建敏感信息扫描标签页
func (m *MainUI) buildScannerTab() fyne.CanvasObject {
	return NewScannerUI(m.window, m.scanner, m.adbMgr, m.getSelectedDevice).Build()
}

// buildBatchTab 构建批量操作标签页
func (m *MainUI) buildBatchTab() fyne.CanvasObject {
	return NewBatchUI(m.window, m.batchMgr, m.adbMgr, m.selectedDevices).Build()
}

// 辅助方法
func (m *MainUI) addSelectedDevice(serial string) {
	for _, s := range m.selectedDevices {
		if s == serial {
			return
		}
	}
	m.selectedDevices = append(m.selectedDevices, serial)
}

func (m *MainUI) removeSelectedDevice(serial string) {
	for i, s := range m.selectedDevices {
		if s == serial {
			m.selectedDevices = append(m.selectedDevices[:i], m.selectedDevices[i+1:]...)
			return
		}
	}
}

func (m *MainUI) getSelectedDevice() string {
	if len(m.selectedDevices) > 0 {
		return m.selectedDevices[0]
	}
	return ""
}

// openShellWindow 打开交互式 Shell 窗口
func (m *MainUI) openShellWindow(serial, model string) {
	// 创建新窗口
	shellWindow := fyne.CurrentApp().NewWindow(fmt.Sprintf("ADB Shell - %s (%s)", model, serial))
	shellWindow.Resize(fyne.NewSize(800, 600))

	// 命令历史
	commandHistory := make([]string, 0)
	_ = commandHistory // 预留用于未来增加上下箭头浏览历史功能

	// 输出显示区
	outputText := widget.NewMultiLineEntry()
	outputText.Wrapping = fyne.TextWrapWord
	outputText.TextStyle = fyne.TextStyle{Monospace: true}
	// 不禁用，但设置为只读样式确保文本清晰可见

	// 添加欢迎信息
	outputText.SetText(fmt.Sprintf("已连接到设备: %s\n", serial))
	outputText.SetText(outputText.Text + fmt.Sprintf("型号: %s\n", model))
	outputText.SetText(outputText.Text + "\n输入 'exit' 退出 Shell\n")
	outputText.SetText(outputText.Text + "========================================\n\n")

	// 命令输入框
	cmdEntry := widget.NewEntry()
	cmdEntry.SetPlaceHolder("输入命令...")

	// 执行命令的函数
	executeCommand := func(command string) {
		if command == "" {
			return
		}

		// 处理 exit 命令
		if command == "exit" {
			shellWindow.Close()
			return
		}

		// 处理 clear 命令
		if command == "clear" || command == "cls" {
			outputText.SetText("")
			cmdEntry.SetText("")
			return
		}

		// 添加到历史
		commandHistory = append(commandHistory, command)

		// 显示命令
		outputText.SetText(outputText.Text + fmt.Sprintf("$ %s\n", command))

		// 执行命令
		result, err := m.adbMgr.ExecuteCommand(serial, command)
		if err != nil {
			outputText.SetText(outputText.Text + fmt.Sprintf("错误: %v\n\n", err))
		} else {
			outputText.SetText(outputText.Text + result + "\n\n")
		}

		// 滚动到底部
		outputText.CursorRow = len(outputText.Text)

		// 清空输入框
		cmdEntry.SetText("")
	}

	// 执行按钮
	execBtn := widget.NewButton("执行", func() {
		executeCommand(cmdEntry.Text)
	})
	execBtn.Importance = widget.HighImportance

	// 回车键执行命令
	cmdEntry.OnSubmitted = func(text string) {
		executeCommand(text)
	}

	// 快捷命令按钮
	quickBtn1 := widget.NewButton("📁 ls -la", func() {
		cmdEntry.SetText("ls -la")
		executeCommand("ls -la")
	})
	quickBtn1.Importance = widget.LowImportance

	quickBtn2 := widget.NewButton("📊 top -n 1", func() {
		cmdEntry.SetText("top -n 1")
		executeCommand("top -n 1")
	})
	quickBtn2.Importance = widget.LowImportance

	quickBtn3 := widget.NewButton("🔍 ps -A", func() {
		cmdEntry.SetText("ps -A")
		executeCommand("ps -A")
	})
	quickBtn3.Importance = widget.LowImportance

	quickBtn4 := widget.NewButton("⚡ su", func() {
		cmdEntry.SetText("su")
		executeCommand("su")
	})
	quickBtn4.Importance = widget.WarningImportance

	clearBtn := widget.NewButton("🧹 清屏", func() {
		outputText.SetText("")
	})

	// 布局
	quickBtnBox := container.NewHBox(
		quickBtn1, quickBtn2, quickBtn3, quickBtn4, clearBtn,
	)

	inputBox := container.NewBorder(
		nil, nil, nil, execBtn,
		cmdEntry,
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("💻 交互式 Shell 终端"),
			widget.NewSeparator(),
			quickBtnBox,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			inputBox,
		),
		nil, nil,
		container.NewScroll(outputText),
	)

	shellWindow.SetContent(content)
	shellWindow.Show()

	// 聚焦到输入框
	shellWindow.Canvas().Focus(cmdEntry)
}

// 通用对话框函数
func showError(w fyne.Window, title string, err error) {
	message := title
	if err != nil {
		message += ": " + err.Error()
	}
	dlg := dialog.NewInformation("错误", message, w)
	dlg.Show()
}

func showInfo(w fyne.Window, title, message string) {
	dlg := dialog.NewInformation(title, message, w)
	dlg.Show()
}
