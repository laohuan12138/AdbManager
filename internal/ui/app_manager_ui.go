package ui

import (
	"adbmanager/internal/adb"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// AppManagerUI 应用管理界面
type AppManagerUI struct {
	window            fyne.Window
	adbMgr            *adb.ADBManager
	getDevice         func() string
	packages          []string
	packageList       *widget.List
	selectedPackageID int
}

// NewAppManagerUI 创建应用管理界面
func NewAppManagerUI(window fyne.Window, adbMgr *adb.ADBManager, getDevice func() string) *AppManagerUI {
	return &AppManagerUI{
		window:    window,
		adbMgr:    adbMgr,
		getDevice: getDevice,
		packages:  make([]string, 0),
	}
}

// Build 构建应用管理界面
func (a *AppManagerUI) Build() fyne.CanvasObject {
	// 应用列表
	a.packageList = widget.NewList(
		func() int {
			return len(a.packages)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("应用包名")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < widget.ListItemID(len(a.packages)) {
				label := obj.(*widget.Label)
				label.SetText(a.packages[id])
			}
		},
	)

	a.packageList.OnSelected = func(id widget.ListItemID) {
		a.selectedPackageID = int(id)
	}

	// 刷新应用列表
	refreshBtn := widget.NewButton("刷新应用列表", func() {
		a.refreshPackages()
	})

	// 安装应用
	installBtn := widget.NewButton("安装 APK", func() {
		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()

			device := a.getDevice()
			if device == "" {
				showError(a.window, "错误", fmt.Errorf("请先选择设备"))
				return
			}

			apkPath := uc.URI().Path()
			err = a.adbMgr.InstallApp(device, apkPath)
			if err != nil {
				showError(a.window, "安装失败", err)
				return
			}

			showInfo(a.window, "成功", "应用安装成功")
			a.refreshPackages()
		}, a.window)
	})

	// 卸载应用
	uninstallBtn := widget.NewButton("卸载应用", func() {
		if a.selectedPackageID < 0 || a.selectedPackageID >= len(a.packages) {
			showError(a.window, "错误", fmt.Errorf("请先选择应用"))
			return
		}

		packageName := a.packages[a.selectedPackageID]

		dialog.ShowConfirm("确认卸载",
			"确定要卸载 "+packageName+" 吗？",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				device := a.getDevice()
				if device == "" {
					showError(a.window, "错误", fmt.Errorf("请先选择设备"))
					return
				}

				err := a.adbMgr.UninstallApp(device, packageName)
				if err != nil {
					showError(a.window, "卸载失败", err)
					return
				}

				showInfo(a.window, "成功", "应用卸载成功")
				a.refreshPackages()
			}, a.window)
	})

	// 启动应用
	startBtn := widget.NewButton("启动应用", func() {
		if a.selectedPackageID < 0 || a.selectedPackageID >= len(a.packages) {
			showError(a.window, "错误", fmt.Errorf("请先选择应用"))
			return
		}

		packageName := a.packages[a.selectedPackageID]
		device := a.getDevice()
		if device == "" {
			showError(a.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		err := a.adbMgr.StartApp(device, packageName)
		if err != nil {
			showError(a.window, "启动失败", err)
			return
		}

		showInfo(a.window, "成功", "应用启动成功")
	})

	// 停止应用
	stopBtn := widget.NewButton("停止应用", func() {
		if a.selectedPackageID < 0 || a.selectedPackageID >= len(a.packages) {
			showError(a.window, "错误", fmt.Errorf("请先选择应用"))
			return
		}

		packageName := a.packages[a.selectedPackageID]
		device := a.getDevice()
		if device == "" {
			showError(a.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		err := a.adbMgr.StopApp(device, packageName)
		if err != nil {
			showError(a.window, "停止失败", err)
			return
		}

		showInfo(a.window, "成功", "应用已停止")
	})

	// 查看应用信息
	infoBtn := widget.NewButton("查看应用信息", func() {
		if a.selectedPackageID < 0 || a.selectedPackageID >= len(a.packages) {
			showError(a.window, "错误", fmt.Errorf("请先选择应用"))
			return
		}

		packageName := a.packages[a.selectedPackageID]
		device := a.getDevice()
		if device == "" {
			showError(a.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, err := a.adbMgr.ExecuteCommand(device,
			fmt.Sprintf("dumpsys package %s", packageName))
		if err != nil {
			showError(a.window, "获取应用信息失败", err)
			return
		}

		// 显示在对话框中
		infoText := widget.NewMultiLineEntry()
		infoText.SetText(output)
		infoText.Wrapping = fyne.TextWrapWord
		infoText.TextStyle = fyne.TextStyle{Monospace: true}

		infoDialog := dialog.NewCustom("应用信息: "+packageName, "关闭",
			container.NewScroll(infoText), a.window)
		infoDialog.Resize(fyne.NewSize(600, 400))
		infoDialog.Show()
	})

	// 清除数据
	clearDataBtn := widget.NewButton("清除数据", func() {
		if a.selectedPackageID < 0 || a.selectedPackageID >= len(a.packages) {
			showError(a.window, "错误", fmt.Errorf("请先选择应用"))
			return
		}

		packageName := a.packages[a.selectedPackageID]

		dialog.ShowConfirm("确认清除数据",
			"确定要清除 "+packageName+" 的数据吗？",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				device := a.getDevice()
				if device == "" {
					showError(a.window, "错误", fmt.Errorf("请先选择设备"))
					return
				}

				_, err := a.adbMgr.ExecuteCommand(device,
					fmt.Sprintf("pm clear %s", packageName))
				if err != nil {
					showError(a.window, "清除数据失败", err)
					return
				}

				showInfo(a.window, "成功", "应用数据已清除")
			}, a.window)
	})

	// 搜索框
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("搜索应用包名...")
	searchEntry.OnChanged = func(query string) {
		// 实现简单的搜索过滤
		// 这里可以改进为实时过滤
	}

	// 布局
	buttonBox := container.NewGridWithColumns(3,
		refreshBtn,
		installBtn,
		uninstallBtn,
	)

	buttonBox2 := container.NewGridWithColumns(4,
		startBtn,
		stopBtn,
		infoBtn,
		clearDataBtn,
	)

	return container.NewBorder(
		container.NewVBox(
			searchEntry,
			buttonBox,
			buttonBox2,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		a.packageList,
	)
}

// refreshPackages 刷新应用列表
func (a *AppManagerUI) refreshPackages() {
	device := a.getDevice()
	if device == "" {
		showError(a.window, "错误", fmt.Errorf("请先选择设备"))
		return
	}

	packages, err := a.adbMgr.ListPackages(device)
	if err != nil {
		showError(a.window, "获取应用列表失败", err)
		return
	}

	a.packages = packages
	a.packageList.Refresh()
}
