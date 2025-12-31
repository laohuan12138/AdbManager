package scanner

import (
	"adbmanager/internal/adb"
	"fmt"
	"regexp"
	"strings"
)

// Scanner 敏感信息扫描器
type Scanner struct {
	adbMgr *adb.ADBManager
}

// NewScanner 创建扫描器
func NewScanner(adbMgr *adb.ADBManager) *Scanner {
	return &Scanner{
		adbMgr: adbMgr,
	}
}

// SensitiveInfo 敏感信息
type SensitiveInfo struct {
	Type     string // 类型：password, api_key, token, email, phone, url 等
	Value    string // 值
	FilePath string // 文件路径
	Line     string // 所在行
}

// 敏感信息匹配模式
var sensitivePatterns = map[string]*regexp.Regexp{
	"password":     regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*['"]?([^'"\s]+)`),
	"api_key":      regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*['"]?([A-Za-z0-9_\-]{20,})`),
	"access_token": regexp.MustCompile(`(?i)(access[_-]?token|token)\s*[=:]\s*['"]?([A-Za-z0-9_\-\.]{20,})`),
	"secret_key":   regexp.MustCompile(`(?i)(secret[_-]?key|secretkey)\s*[=:]\s*['"]?([A-Za-z0-9_\-]{20,})`),
	"email":        regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	"phone":        regexp.MustCompile(`(?:(?:\+|00)86)?1[3-9]\d{9}`),
	"ip_address":   regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	"url":          regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`),
	"jwt":          regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`),
	"private_key":  regexp.MustCompile(`-----BEGIN (?:RSA |EC )?PRIVATE KEY-----`),
}

// ScanConfigFiles 扫描配置文件中的敏感信息
func (s *Scanner) ScanConfigFiles(serial string) ([]SensitiveInfo, error) {
	results := make([]SensitiveInfo, 0)

	// 常见配置文件路径
	configPaths := []string{
		"/data/data/*/shared_prefs/*.xml",
		"/data/data/*/files/*.xml",
		"/data/data/*/files/*.json",
		"/data/data/*/databases/*.db",
		"/sdcard/*.xml",
		"/sdcard/*.json",
		"/sdcard/*.txt",
		"/sdcard/*.conf",
	}

	for _, pattern := range configPaths {
		// 查找匹配的文件
		files, err := s.findFiles(serial, pattern)
		if err != nil {
			continue
		}

		// 扫描每个文件
		for _, file := range files {
			fileResults, err := s.scanFile(serial, file)
			if err != nil {
				continue
			}
			results = append(results, fileResults...)
		}
	}

	return results, nil
}

// findFiles 查找文件
func (s *Scanner) findFiles(serial, pattern string) ([]string, error) {
	// 提取目录和文件模式
	lastSlash := strings.LastIndex(pattern, "/")
	if lastSlash == -1 {
		return nil, fmt.Errorf("无效的路径模式: %s", pattern)
	}

	dir := pattern[:lastSlash]
	filePattern := pattern[lastSlash+1:]

	// 使用 find 命令查找文件
	cmd := fmt.Sprintf("find %s -name '%s' 2>/dev/null", dir, filePattern)
	output, err := s.adbMgr.ExecuteCommand(serial, cmd)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Permission denied") {
			files = append(files, line)
		}
	}

	return files, nil
}

// scanFile 扫描单个文件
func (s *Scanner) scanFile(serial, filePath string) ([]SensitiveInfo, error) {
	// 读取文件内容
	content, err := s.adbMgr.ExecuteCommand(serial, fmt.Sprintf("cat %s", filePath))
	if err != nil {
		return nil, err
	}

	results := make([]SensitiveInfo, 0)
	lines := strings.Split(content, "\n")

	// 对每一行应用所有匹配模式
	for lineNum, line := range lines {
		for patternType, pattern := range sensitivePatterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				info := SensitiveInfo{
					Type:     patternType,
					FilePath: filePath,
					Line:     fmt.Sprintf("Line %d", lineNum+1),
				}

				if len(match) > 1 {
					info.Value = match[len(match)-1]
				} else {
					info.Value = match[0]
				}

				results = append(results, info)
			}
		}
	}

	return results, nil
}

// ScanSharedPreferences 扫描 SharedPreferences
func (s *Scanner) ScanSharedPreferences(serial, packageName string) ([]SensitiveInfo, error) {
	results := make([]SensitiveInfo, 0)

	// SharedPreferences 路径
	prefsPath := fmt.Sprintf("/data/data/%s/shared_prefs/", packageName)

	// 列出所有 XML 文件
	files, err := s.findFiles(serial, prefsPath+"*.xml")
	if err != nil {
		return nil, err
	}

	// 扫描每个文件
	for _, file := range files {
		fileResults, err := s.scanFile(serial, file)
		if err != nil {
			continue
		}
		results = append(results, fileResults...)
	}

	return results, nil
}

// ScanDatabases 扫描数据库
func (s *Scanner) ScanDatabases(serial, packageName string) ([]string, error) {
	dbPath := fmt.Sprintf("/data/data/%s/databases/", packageName)

	// 列出所有数据库文件
	files, err := s.findFiles(serial, dbPath+"*.db")
	if err != nil {
		return nil, err
	}

	return files, nil
}

// ExportDatabase 导出数据库
func (s *Scanner) ExportDatabase(serial, dbPath, localPath string) error {
	return s.adbMgr.PullFile(serial, dbPath, localPath)
}

// ScanLogFiles 扫描日志文件
func (s *Scanner) ScanLogFiles(serial string) ([]SensitiveInfo, error) {
	results := make([]SensitiveInfo, 0)

	// 常见日志文件路径
	logPaths := []string{
		"/sdcard/Android/data/*/files/*.log",
		"/sdcard/*.log",
		"/data/local/tmp/*.log",
	}

	for _, pattern := range logPaths {
		files, err := s.findFiles(serial, pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			fileResults, err := s.scanFile(serial, file)
			if err != nil {
				continue
			}
			results = append(results, fileResults...)
		}
	}

	return results, nil
}

// GetInstalledCertificates 获取已安装的证书
func (s *Scanner) GetInstalledCertificates(serial string) (string, error) {
	return s.adbMgr.ExecuteCommand(serial, "ls -la /system/etc/security/cacerts/")
}

// CheckRootStatus 检查设备是否已 root
func (s *Scanner) CheckRootStatus(serial string) (bool, error) {
	output, err := s.adbMgr.ExecuteCommand(serial, "su -c 'id'")
	if err != nil {
		return false, nil
	}

	return strings.Contains(output, "uid=0"), nil
}

// GetAppDataSize 获取应用数据大小
func (s *Scanner) GetAppDataSize(serial, packageName string) (map[string]string, error) {
	output, err := s.adbMgr.ExecuteCommand(serial,
		fmt.Sprintf("dumpsys package %s | grep -A 5 'dataDir'", packageName))
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				info[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return info, nil
}
