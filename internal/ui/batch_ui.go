package ui

import (
	"adbmanager/internal/adb"
	"adbmanager/internal/batch"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BatchUI 批量操作界面
type BatchUI struct {
	window          fyne.Window
	batchMgr        *batch.BatchManager
	adbMgr          *adb.ADBManager
	selectedDevices []string

	deviceCheckboxes map[string]*widget.Check
	devices          []adb.Device
}

// NewBatchUI 创建批量操作界面
func NewBatchUI(window fyne.Window, batchMgr *batch.BatchManager, adbMgr *adb.ADBManager, selectedDevices []string) *BatchUI {
	return &BatchUI{
		window:           window,
		batchMgr:         batchMgr,
		adbMgr:           adbMgr,
		selectedDevices:  selectedDevices,
		deviceCheckboxes: make(map[string]*widget.Check),
		devices:          make([]adb.Device, 0),
	}
}

// Build 构建批量操作界面
func (b *BatchUI) Build() fyne.CanvasObject {
	// 直接返回操作面板，不需要分割布局
	return b.buildOperationPanel()
}

// buildOperationPanel 构建操作面板
func (b *BatchUI) buildOperationPanel() fyne.CanvasObject {
	// 左侧：设备选择区域
	deviceContainer := container.NewVBox(
		widget.NewLabel("选择设备:"),
	)

	refreshDevicesBtn := widget.NewButton("刷新设备列表", func() {
		b.refreshDeviceList(deviceContainer)
	})

	// 全选按钮
	selectAllBtn := widget.NewButton("全选", func() {
		for _, checkbox := range b.deviceCheckboxes {
			checkbox.SetChecked(true)
		}
	})

	// 取消全选按钮
	unselectAllBtn := widget.NewButton("取消全选", func() {
		for _, checkbox := range b.deviceCheckboxes {
			checkbox.SetChecked(false)
		}
	})

	// 设备操作按钮组
	deviceButtonBox := container.NewGridWithColumns(3,
		refreshDevicesBtn,
		selectAllBtn,
		unselectAllBtn,
	)

	// 左侧面板
	leftPanel := container.NewBorder(
		container.NewVBox(
			deviceButtonBox,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewScroll(deviceContainer),
	)

	// 右侧：操作区域
	// 操作结果显示
	resultText := widget.NewMultiLineEntry()
	resultText.Wrapping = fyne.TextWrapWord
	resultText.TextStyle = fyne.TextStyle{Monospace: true}

	// Busybox 开关
	busyboxCheck := widget.NewCheck("启用 Busybox 模式", func(checked bool) {
		b.adbMgr.SetBusyboxEnabled(checked)
		if checked {
			resultText.SetText("✅ Busybox 模式已启用 - 命令执行时将自动加入 'busybox' 前缀\n")
		} else {
			resultText.SetText("❌ Busybox 模式已禁用\n")
		}
	})
	busyboxCheck.Checked = b.adbMgr.IsBusyboxEnabled()

	// 批量执行命令
	commandEntry := widget.NewEntry()
	commandEntry.SetPlaceHolder("输入要执行的命令")

	execBtn := widget.NewButton("批量执行命令", func() {
		command := commandEntry.Text
		if command == "" {
			showError(b.window, "错误", fmt.Errorf("请输入命令"))
			return
		}

		selectedDevs := b.getSelectedDevices()
		if len(selectedDevs) == 0 {
			showError(b.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		resultText.SetText(fmt.Sprintf("正在 %d 台设备上执行命令...\n\n", len(selectedDevs)))

		b.batchMgr.BatchExecuteCommand(selectedDevs, command, func(result batch.CommandResult) {
			output := resultText.Text
			output += fmt.Sprintf("========== %s ==========\n", result.Device)
			if result.Error != nil {
				output += fmt.Sprintf("错误: %s\n", result.Error.Error())
			}
			output += result.Output + "\n\n"
			resultText.SetText(output)
		})
	})

	// 批量安装APK
	installBtn := widget.NewButton("批量安装APK", func() {
		selectedDevs := b.getSelectedDevices()
		if len(selectedDevs) == 0 {
			showError(b.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()

			apkPath := uc.URI().Path()
			resultText.SetText(fmt.Sprintf("正在 %d 台设备上安装应用...\n\n", len(selectedDevs)))

			b.batchMgr.BatchInstallApp(selectedDevs, apkPath, func(device string, err error) {
				output := resultText.Text
				if err != nil {
					output += fmt.Sprintf("✗ %s: 安装失败 - %s\n", device, err.Error())
				} else {
					output += fmt.Sprintf("✓ %s: 安装成功\n", device)
				}
				resultText.SetText(output)
			})

			resultText.SetText(resultText.Text + "\n批量安装完成！")
		}, b.window)
	})

	// 批量卸载应用
	uninstallBtn := widget.NewButton("批量卸载应用", func() {
		selectedDevs := b.getSelectedDevices()
		if len(selectedDevs) == 0 {
			showError(b.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		pkgEntry := widget.NewEntry()
		pkgEntry.SetPlaceHolder("输入包名")

		dialog.ShowCustomConfirm("批量卸载", "确定", "取消",
			pkgEntry,
			func(confirmed bool) {
				if !confirmed || pkgEntry.Text == "" {
					return
				}

				packageName := pkgEntry.Text
				resultText.SetText(fmt.Sprintf("正在 %d 台设备上卸载应用...\n\n", len(selectedDevs)))

				b.batchMgr.BatchUninstallApp(selectedDevs, packageName, func(device string, err error) {
					output := resultText.Text
					if err != nil {
						output += fmt.Sprintf("✗ %s: 卸载失败 - %s\n", device, err.Error())
					} else {
						output += fmt.Sprintf("✓ %s: 卸载成功\n", device)
					}
					resultText.SetText(output)
				})

				resultText.SetText(resultText.Text + "\n批量卸载完成！")
			}, b.window)
	})

	// 批量截屏
	screenshotBtn := widget.NewButton("批量截屏", func() {
		selectedDevs := b.getSelectedDevices()
		if len(selectedDevs) == 0 {
			showError(b.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}

			outputDir := dir.Path()
			resultText.SetText(fmt.Sprintf("正在 %d 台设备上截屏...\n\n", len(selectedDevs)))

			b.batchMgr.BatchScreenshot(selectedDevs, outputDir, func(device, filepath string, err error) {
				output := resultText.Text
				if err != nil {
					output += fmt.Sprintf("✗ %s: 截屏失败 - %s\n", device, err.Error())
				} else {
					output += fmt.Sprintf("✓ %s: 已保存到 %s\n", device, filepath)
				}
				resultText.SetText(output)
			})

			resultText.SetText(resultText.Text + "\n批量截屏完成！")
		}, b.window)
	})

	// 清空结果
	clearResultBtn := widget.NewButton("清空", func() {
		resultText.SetText("")
	})

	// 初始加载设备列表
	b.refreshDeviceList(deviceContainer)

	// 右侧面板布局
	cmdBox := container.NewVBox(
		busyboxCheck,
		container.NewBorder(nil, nil, nil, execBtn, commandEntry),
	)

	buttonBox := container.NewGridWithColumns(3,
		installBtn,
		uninstallBtn,
		screenshotBtn,
	)

	rightPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("批量操作:"),
			cmdBox,
			buttonBox,
			clearResultBtn,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewScroll(resultText),
	)

	// 左右分栏布局
	return container.NewHSplit(
		leftPanel,
		rightPanel,
	)
}

// refreshDeviceList 刷新设备列表
func (b *BatchUI) refreshDeviceList(container *fyne.Container) {
	devices, err := b.adbMgr.ListDevices()
	if err != nil {
		return
	}

	// 只保留在线设备
	onlineDevices := make([]adb.Device, 0)
	for _, dev := range devices {
		if dev.Status == "device" {
			onlineDevices = append(onlineDevices, dev)
		}
	}

	b.devices = onlineDevices
	b.deviceCheckboxes = make(map[string]*widget.Check)

	// 清空容器
	container.Objects = []fyne.CanvasObject{
		widget.NewLabel("选择设备:"),
	}

	// 添加设备复选框（只显示在线设备）
	for _, dev := range onlineDevices {
		check := widget.NewCheck(dev.Serial+" - "+dev.Status, nil)
		b.deviceCheckboxes[dev.Serial] = check
		container.Add(check)
	}

	container.Refresh()
}

// getSelectedDevices 获取选中的设备
func (b *BatchUI) getSelectedDevices() []string {
	selected := make([]string, 0)
	for serial, check := range b.deviceCheckboxes {
		if check.Checked {
			selected = append(selected, serial)
		}
	}
	return selected
}
