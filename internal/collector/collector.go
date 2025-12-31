package collector

import (
	"adbmanager/internal/adb"
	"fmt"
	"strings"
)

// Collector 信息采集器
type Collector struct {
	adbMgr *adb.ADBManager
}

// NewCollector 创建信息采集器
func NewCollector(adbMgr *adb.ADBManager) *Collector {
	return &Collector{
		adbMgr: adbMgr,
	}
}

// ContactInfo 联系人信息
type ContactInfo struct {
	Name  string
	Phone string
	Email string
}

// GetContacts 获取联系人信息
func (c *Collector) GetContacts(serial string) ([]ContactInfo, error) {
	// 需要读取联系人权限
	output, err := c.adbMgr.ExecuteCommand(serial,
		"content query --uri content://contacts/phones --projection display_name:number")
	if err != nil {
		return nil, fmt.Errorf("获取联系人失败: %v", err)
	}

	contacts := make([]ContactInfo, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Row:") {
			continue
		}

		// 解析联系人信息
		var contact ContactInfo
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "display_name=") {
				contact.Name = strings.TrimPrefix(part, "display_name=")
			} else if strings.HasPrefix(part, "number=") {
				contact.Phone = strings.TrimPrefix(part, "number=")
			}
		}

		if contact.Name != "" || contact.Phone != "" {
			contacts = append(contacts, contact)
		}
	}

	return contacts, nil
}

// SMSInfo 短信信息
type SMSInfo struct {
	Address string
	Body    string
	Date    string
	Type    string // 1=接收, 2=发送
}

// GetSMS 获取短信信息
func (c *Collector) GetSMS(serial string, limit int) ([]SMSInfo, error) {
	cmd := fmt.Sprintf("content query --uri content://sms --projection address:body:date:type --sort \"date DESC\" --limit %d", limit)
	output, err := c.adbMgr.ExecuteCommand(serial, cmd)
	if err != nil {
		return nil, fmt.Errorf("获取短信失败: %v", err)
	}

	smsList := make([]SMSInfo, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Row:") {
			continue
		}

		var sms SMSInfo
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "address=") {
				sms.Address = strings.TrimPrefix(part, "address=")
			} else if strings.HasPrefix(part, "body=") {
				sms.Body = strings.TrimPrefix(part, "body=")
			} else if strings.HasPrefix(part, "date=") {
				sms.Date = strings.TrimPrefix(part, "date=")
			} else if strings.HasPrefix(part, "type=") {
				sms.Type = strings.TrimPrefix(part, "type=")
			}
		}

		if sms.Address != "" {
			smsList = append(smsList, sms)
		}
	}

	return smsList, nil
}

// LocationInfo 位置信息
type LocationInfo struct {
	Provider  string
	Latitude  string
	Longitude string
	Accuracy  string
	Time      string
}

// GetLocation 获取位置信息
func (c *Collector) GetLocation(serial string) (*LocationInfo, error) {
	// 尝试获取最后已知位置
	output, err := c.adbMgr.ExecuteCommand(serial,
		"dumpsys location | grep -A 10 'Last Known Locations'")
	if err != nil {
		return nil, fmt.Errorf("获取位置信息失败: %v", err)
	}

	location := &LocationInfo{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "lat=") && strings.Contains(line, "lon=") {
			// 解析位置信息
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "lat=") {
					location.Latitude = strings.TrimPrefix(part, "lat=")
				} else if strings.HasPrefix(part, "lon=") {
					location.Longitude = strings.TrimPrefix(part, "lon=")
				} else if strings.HasPrefix(part, "acc=") {
					location.Accuracy = strings.TrimPrefix(part, "acc=")
				}
			}
		}
	}

	return location, nil
}

// WiFiInfo WiFi 信息
type WiFiInfo struct {
	SSID     string
	BSSID    string
	Password string
	Security string
}

// GetWiFiInfo 获取 WiFi 信息
func (c *Collector) GetWiFiInfo(serial string) ([]WiFiInfo, error) {
	wifiList := make([]WiFiInfo, 0)

	// 获取当前连接的 WiFi
	currentSSID, _ := c.adbMgr.ExecuteCommand(serial,
		"dumpsys wifi | grep 'mWifiInfo' | awk '{print $4}'")
	currentSSID = strings.TrimSpace(currentSSID)

	// 获取已保存的 WiFi 配置（需要 root 权限）
	output, err := c.adbMgr.ExecuteCommand(serial,
		"cat /data/misc/wifi/wpa_supplicant.conf")
	if err != nil {
		// 如果没有 root 权限，尝试其他方法
		output, err = c.adbMgr.ExecuteCommand(serial, "dumpsys wifi")
		if err != nil {
			return nil, fmt.Errorf("获取 WiFi 信息失败: %v", err)
		}
	}

	lines := strings.Split(output, "\n")
	var currentWiFi *WiFiInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "network={") {
			currentWiFi = &WiFiInfo{}
		} else if currentWiFi != nil {
			if strings.HasPrefix(line, "ssid=") {
				currentWiFi.SSID = strings.Trim(strings.TrimPrefix(line, "ssid="), "\"")
			} else if strings.HasPrefix(line, "psk=") {
				currentWiFi.Password = strings.Trim(strings.TrimPrefix(line, "psk="), "\"")
			} else if strings.HasPrefix(line, "}") {
				if currentWiFi.SSID != "" {
					wifiList = append(wifiList, *currentWiFi)
				}
				currentWiFi = nil
			}
		}
	}

	return wifiList, nil
}

// GetInstalledApps 获取已安装应用列表
func (c *Collector) GetInstalledApps(serial string) ([]string, error) {
	return c.adbMgr.ListPackages(serial)
}

// AppPermission 应用权限信息
type AppPermission struct {
	Permission string
	Granted    bool
}

// GetAppPermissions 获取应用权限
func (c *Collector) GetAppPermissions(serial, packageName string) ([]AppPermission, error) {
	output, err := c.adbMgr.ExecuteCommand(serial,
		fmt.Sprintf("dumpsys package %s | grep permission", packageName))
	if err != nil {
		return nil, fmt.Errorf("获取应用权限失败: %v", err)
	}

	permissions := make([]AppPermission, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "android.permission.") {
			perm := AppPermission{}

			// 提取权限名称
			if idx := strings.Index(line, "android.permission."); idx != -1 {
				permStr := line[idx:]
				parts := strings.Fields(permStr)
				if len(parts) > 0 {
					perm.Permission = parts[0]
					perm.Granted = strings.Contains(line, "granted=true")
					permissions = append(permissions, perm)
				}
			}
		}
	}

	return permissions, nil
}

// GetBatteryInfo 获取电池信息
func (c *Collector) GetBatteryInfo(serial string) (map[string]string, error) {
	output, err := c.adbMgr.ExecuteCommand(serial, "dumpsys battery")
	if err != nil {
		return nil, fmt.Errorf("获取电池信息失败: %v", err)
	}

	info := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			info[key] = value
		}
	}

	return info, nil
}

// GetSystemProperties 获取系统属性
func (c *Collector) GetSystemProperties(serial string) (map[string]string, error) {
	output, err := c.adbMgr.ExecuteCommand(serial, "getprop")
	if err != nil {
		return nil, fmt.Errorf("获取系统属性失败: %v", err)
	}

	props := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]:") {
			// 格式: [key]: [value]
			parts := strings.Split(line, "]: [")
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "[]")
				value := strings.Trim(parts[1], "[]")
				props[key] = value
			}
		}
	}

	return props, nil
}
