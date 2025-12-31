package ui

import (
	"adbmanager/internal/adb"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// DeviceInfoUI 设备信息界面
type DeviceInfoUI struct {
	window    fyne.Window
	adbMgr    *adb.ADBManager
	getDevice func() string
}

// NewDeviceInfoUI 创建设备信息界面
func NewDeviceInfoUI(window fyne.Window, adbMgr *adb.ADBManager, getDevice func() string) *DeviceInfoUI {
	return &DeviceInfoUI{
		window:    window,
		adbMgr:    adbMgr,
		getDevice: getDevice,
	}
}

// Build 构建设备信息界面
func (d *DeviceInfoUI) Build() fyne.CanvasObject {
	// 信息显示区域
	infoText := widget.NewMultiLineEntry()
	infoText.Wrapping = fyne.TextWrapWord
	infoText.TextStyle = fyne.TextStyle{Monospace: true}

	// 获取基本信息按钮
	basicInfoBtn := widget.NewButton("获取基本信息", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		info, err := d.adbMgr.GetDeviceInfo(device)
		if err != nil {
			showError(d.window, "获取信息失败", err)
			return
		}

		result := "========== 设备基本信息 ==========\n"
		for key, value := range info {
			result += fmt.Sprintf("%s: %s\n", key, value)
		}

		infoText.SetText(result)
	})

	// 获取网络连接信息
	networkBtn := widget.NewButton("获取网络连接", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, err := d.adbMgr.GetNetworkConnections(device)
		if err != nil {
			showError(d.window, "获取网络连接失败", err)
			return
		}

		infoText.SetText("========== 网络连接 ==========\n" + output)
	})

	// 获取进程列表
	processBtn := widget.NewButton("获取进程列表", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, err := d.adbMgr.GetProcessList(device)
		if err != nil {
			showError(d.window, "获取进程列表失败", err)
			return
		}

		infoText.SetText("========== 进程列表 ==========\n" + output)
	})

	// 获取已安装应用
	appsBtn := widget.NewButton("获取已安装应用", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		packages, err := d.adbMgr.ListPackages(device)
		if err != nil {
			showError(d.window, "获取应用列表失败", err)
			return
		}

		result := "========== 已安装应用 ==========\n"
		result += fmt.Sprintf("应用总数: %d\n\n", len(packages))
		for _, pkg := range packages {
			result += pkg + "\n"
		}

		infoText.SetText(result)
	})

	// 获取系统属性
	propBtn := widget.NewButton("获取系统属性", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, _ := d.adbMgr.ExecuteCommand(device, "getprop")
		infoText.SetText("========== 系统属性 ==========\n" + output)
	})

	// 获取存储信息
	storageBtn := widget.NewButton("获取存储信息", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, _ := d.adbMgr.ExecuteCommand(device, "df -h")
		infoText.SetText("========== 存储信息 ==========\n" + output)
	})

	// 获取电池信息
	batteryBtn := widget.NewButton("获取电池信息", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, _ := d.adbMgr.ExecuteCommand(device, "dumpsys battery")
		infoText.SetText("========== 电池信息 ==========\n" + output)
	})

	// 获取内存信息
	memoryBtn := widget.NewButton("获取内存信息", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		output, _ := d.adbMgr.ExecuteCommand(device, "cat /proc/meminfo")
		infoText.SetText("========== 内存信息 ==========\n" + output)
	})

	// 获取屏幕截图
	screenshotBtn := widget.NewButton("获取屏幕截图", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		// 显示进度提示
		infoText.SetText("正在截取屏幕...\n")

		// 使用时间戳+随机数避免冲突
		tempDir := os.TempDir()
		tempFile := filepath.Join(tempDir, fmt.Sprintf("adb_screenshot_%d_%d.png", time.Now().UnixNano(), os.Getpid()))

		// 执行截图
		err := d.adbMgr.Screenshot(device, tempFile)
		if err != nil {
			showError(d.window, "截图失败", err)
			infoText.SetText(fmt.Sprintf("截图失败: %v", err))
			return
		}

		// 延迟加载图片，确保文件写入完成
		time.Sleep(100 * time.Millisecond)

		// 读取图片数据到内存，避免文件被锁定
		imgData, err := os.ReadFile(tempFile)
		if err != nil {
			os.Remove(tempFile)
			showError(d.window, "读取截图失败", err)
			return
		}

		// 立即删除临时文件，释放句柄
		os.Remove(tempFile)

		// 从内存创建图片资源
		imgResource := &fyne.StaticResource{
			StaticName:    "screenshot.png",
			StaticContent: imgData,
		}
		img := canvas.NewImageFromResource(imgResource)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(400, 600))

		// 创建保存按钮
		saveBtn := widget.NewButton("保存截图", func() {
			dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
				if err != nil || uc == nil {
					return
				}
				defer uc.Close()

				_, err = uc.Write(imgData)
				if err != nil {
					showError(d.window, "保存失败", err)
					return
				}

				showInfo(d.window, "成功", fmt.Sprintf("截图已保存到:\n%s", uc.URI().Path()))
			}, d.window)
		})

		// 创建关闭按钮
		closeBtn := widget.NewButton("关闭", nil)

		// 创建预览对话框
		previewContent := container.NewBorder(
			widget.NewLabel(fmt.Sprintf("截图预览 - %s", time.Now().Format("2006-01-02 15:04:05"))),
			container.NewHBox(saveBtn, closeBtn),
			nil,
			nil,
			container.NewScroll(img),
		)

		previewDialog := dialog.NewCustom("屏幕截图", "关闭", previewContent, d.window)
		previewDialog.Resize(fyne.NewSize(500, 700))
		previewDialog.Show()

		infoText.SetText("截图成功！请在弹出窗口中查看")
	})

	// 获取启动项信息
	startupBtn := widget.NewButton("获取启动项", func() {
		device := d.getDevice()
		if device == "" {
			showError(d.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		infoText.SetText("正在获取启动项信息...\n")

		// 获取开机启动的应用
		output, err := d.adbMgr.ExecuteCommand(device, "pm list packages -e")
		if err != nil {
			showError(d.window, "获取失败", err)
			return
		}

		result := "========== 启动项信息 ==========\n\n"
		result += "[1] 已启用的应用包 (Enabled)\n"
		result += output + "\n\n"

		// 获取 init.rc 启动服务
		services, _ := d.adbMgr.ExecuteCommand(device, "getprop | grep init.svc")
		result += "[2] 系统启动服务\n"
		result += services + "\n\n"

		// 获取开机自启动 Activity
		boot, _ := d.adbMgr.ExecuteCommand(device, "dumpsys activity recents | grep 'Recent #'")
		result += "[3] 最近启动的 Activity\n"
		result += boot + "\n\n"

		// 获取开机广播接收器
		receivers, _ := d.adbMgr.ExecuteCommand(device, "dumpsys package | grep -A 5 'android.intent.action.BOOT_COMPLETED'")
		result += "[4] 开机广播接收器 (BOOT_COMPLETED)\n"
		result += receivers + "\n"

		infoText.SetText(result)
	})

	// 清空按钮
	clearBtn := widget.NewButton("清空", func() {
		infoText.SetText("")
	})

	// 布局
	buttonBox := container.NewGridWithColumns(3,
		basicInfoBtn,
		networkBtn,
		processBtn,
	)

	buttonBox2 := container.NewGridWithColumns(3,
		appsBtn,
		propBtn,
		storageBtn,
	)

	buttonBox3 := container.NewGridWithColumns(3,
		batteryBtn,
		memoryBtn,
		screenshotBtn,
	)

	buttonBox4 := container.NewGridWithColumns(3,
		startupBtn,
		clearBtn,
		widget.NewLabel(""), // 占位符
	)

	return container.NewBorder(
		container.NewVBox(
			buttonBox,
			buttonBox2,
			buttonBox3,
			buttonBox4,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewScroll(infoText),
	)
}
