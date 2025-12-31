package ui

import (
	"adbmanager/internal/collector"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CollectorUI 信息采集界面
type CollectorUI struct {
	window    fyne.Window
	collector *collector.Collector
	getDevice func() string
}

// NewCollectorUI 创建信息采集界面
func NewCollectorUI(window fyne.Window, collector *collector.Collector, getDevice func() string) *CollectorUI {
	return &CollectorUI{
		window:    window,
		collector: collector,
		getDevice: getDevice,
	}
}

// Build 构建信息采集界面
func (c *CollectorUI) Build() fyne.CanvasObject {
	// 结果显示区域
	resultText := widget.NewMultiLineEntry()
	resultText.Wrapping = fyne.TextWrapWord
	resultText.TextStyle = fyne.TextStyle{Monospace: true}

	// 采集联系人
	contactsBtn := widget.NewButton("采集联系人", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		contacts, err := c.collector.GetContacts(device)
		if err != nil {
			showError(c.window, "采集联系人失败", err)
			return
		}

		result := "========== 联系人信息 ==========\n"
		result += fmt.Sprintf("联系人总数: %d\n\n", len(contacts))
		for i, contact := range contacts {
			result += fmt.Sprintf("%d. 姓名: %s, 电话: %s\n",
				i+1, contact.Name, contact.Phone)
		}

		resultText.SetText(result)
	})

	// 采集短信
	smsBtn := widget.NewButton("采集短信 (最近100条)", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		smsList, err := c.collector.GetSMS(device, 100)
		if err != nil {
			showError(c.window, "采集短信失败", err)
			return
		}

		result := "========== 短信信息 ==========\n"
		result += fmt.Sprintf("短信总数: %d\n\n", len(smsList))
		for i, sms := range smsList {
			msgType := "接收"
			if sms.Type == "2" {
				msgType = "发送"
			}
			result += fmt.Sprintf("%d. [%s] %s: %s\n",
				i+1, msgType, sms.Address, sms.Body)
		}

		resultText.SetText(result)
	})

	// 采集位置信息
	locationBtn := widget.NewButton("采集位置信息", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		location, err := c.collector.GetLocation(device)
		if err != nil {
			showError(c.window, "采集位置信息失败", err)
			return
		}

		result := "========== 位置信息 ==========\n"
		result += fmt.Sprintf("经度: %s\n", location.Longitude)
		result += fmt.Sprintf("纬度: %s\n", location.Latitude)
		result += fmt.Sprintf("精度: %s\n", location.Accuracy)
		result += fmt.Sprintf("时间: %s\n", location.Time)

		resultText.SetText(result)
	})

	// 采集WiFi信息
	wifiBtn := widget.NewButton("采集WiFi信息", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		wifiList, err := c.collector.GetWiFiInfo(device)
		if err != nil {
			showError(c.window, "采集WiFi信息失败", err)
			return
		}

		result := "========== WiFi 信息 ==========\n"
		result += fmt.Sprintf("已保存网络数: %d\n\n", len(wifiList))
		for i, wifi := range wifiList {
			result += fmt.Sprintf("%d. SSID: %s\n", i+1, wifi.SSID)
			if wifi.BSSID != "" {
				result += fmt.Sprintf("   BSSID: %s\n", wifi.BSSID)
			}
			if wifi.Password != "" {
				result += fmt.Sprintf("   密码: %s\n", wifi.Password)
			}
			result += "\n"
		}

		resultText.SetText(result)
	})

	// 采集电池信息
	batteryBtn := widget.NewButton("采集电池信息", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		batteryInfo, err := c.collector.GetBatteryInfo(device)
		if err != nil {
			showError(c.window, "采集电池信息失败", err)
			return
		}

		result := "========== 电池信息 ==========\n"
		for key, value := range batteryInfo {
			result += fmt.Sprintf("%s: %s\n", key, value)
		}

		resultText.SetText(result)
	})

	// 采集系统属性
	propsBtn := widget.NewButton("采集系统属性", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		props, err := c.collector.GetSystemProperties(device)
		if err != nil {
			showError(c.window, "采集系统属性失败", err)
			return
		}

		result := "========== 系统属性 ==========\n"
		result += fmt.Sprintf("属性总数: %d\n\n", len(props))
		for key, value := range props {
			result += fmt.Sprintf("%s = %s\n", key, value)
		}

		resultText.SetText(result)
	})

	// 清空按钮
	clearBtn := widget.NewButton("清空", func() {
		resultText.SetText("")
	})

	// 全部采集
	collectAllBtn := widget.NewButton("全部采集", func() {
		device := c.getDevice()
		if device == "" {
			showError(c.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		result := "========== 设备信息全面采集 ==========\n\n"

		// 联系人
		if contacts, err := c.collector.GetContacts(device); err == nil {
			result += fmt.Sprintf("✓ 联系人: %d 条\n", len(contacts))
		} else {
			result += "✗ 联系人: 采集失败\n"
		}

		// 短信
		if smsList, err := c.collector.GetSMS(device, 100); err == nil {
			result += fmt.Sprintf("✓ 短信: %d 条\n", len(smsList))
		} else {
			result += "✗ 短信: 采集失败\n"
		}

		// 位置
		if _, err := c.collector.GetLocation(device); err == nil {
			result += "✓ 位置信息: 已采集\n"
		} else {
			result += "✗ 位置信息: 采集失败\n"
		}

		// WiFi
		if wifiList, err := c.collector.GetWiFiInfo(device); err == nil {
			result += fmt.Sprintf("✓ WiFi: %d 个网络\n", len(wifiList))
		} else {
			result += "✗ WiFi: 采集失败\n"
		}

		// 应用
		if apps, err := c.collector.GetInstalledApps(device); err == nil {
			result += fmt.Sprintf("✓ 已安装应用: %d 个\n", len(apps))
		} else {
			result += "✗ 已安装应用: 采集失败\n"
		}

		// 电池
		if _, err := c.collector.GetBatteryInfo(device); err == nil {
			result += "✓ 电池信息: 已采集\n"
		} else {
			result += "✗ 电池信息: 采集失败\n"
		}

		result += "\n采集完成！可以点击各个按钮查看详细信息。"

		resultText.SetText(result)
	})

	// 布局
	buttonBox := container.NewGridWithColumns(3,
		contactsBtn,
		smsBtn,
		locationBtn,
	)

	buttonBox2 := container.NewGridWithColumns(3,
		wifiBtn,
		batteryBtn,
		propsBtn,
	)

	buttonBox3 := container.NewGridWithColumns(2,
		collectAllBtn,
		clearBtn,
	)

	return container.NewBorder(
		container.NewVBox(
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
