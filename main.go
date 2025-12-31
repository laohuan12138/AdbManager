package main

import (
	"os"
	"path/filepath"

	"adbmanager/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

// customTheme 自定义主题，支持中文字体
type customTheme struct {
	fyne.Theme
	fontData fyne.Resource
}

func (t *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.fontData != nil {
		return t.fontData
	}
	return t.Theme.Font(style)
}

func main() {
	// Windows 中文字体路径列表（只使用 .ttf 格式，Fyne 不支持 .ttc）
	fonts := []string{
		"simhei.ttf",  // 黑体 - 最常用
		"simkai.ttf",  // 楷体
		"simfang.ttf", // 仿宋
		"simsunb.ttf", // 宋体粗体
		"SIMLI.TTF",   // 隶书
		"SIMYOU.TTF",  // 幼圆
		"STXIHEI.TTF", // 华文细黑
		"msyh.ttf",    // 微软雅黑
	}

	// 尝试设置字体
	fontsDir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	fontSet := false
	for _, font := range fonts {
		fontPath := filepath.Join(fontsDir, font)
		if _, err := os.Stat(fontPath); err == nil {
			os.Setenv("FYNE_FONT", fontPath)
			fontSet = true
			break
		}
	}

	// 如果没有找到可用字体，不设置 FYNE_FONT，让 Fyne 使用默认字体
	if !fontSet {
		os.Unsetenv("FYNE_FONT")
	}

	// 设置其他环境变量
	// 使用深色主题提高对比度，或者使用 light
	os.Setenv("FYNE_THEME", "dark") // 深色主题更易读
	os.Setenv("LANG", "zh_CN.UTF-8")

	// 启用 Windows 原生文件对话框
	os.Setenv("FYNE_USE_NATIVE_DIALOGS", "1")

	// 设置缩放比例，提高文字清晰度和大小
	os.Setenv("FYNE_SCALE", "1.3") // 增加到 1.3 倍

	myApp := app.New()

	// 加载中文字体
	var fontResource fyne.Resource
	for _, font := range fonts {
		fontPath := filepath.Join(fontsDir, font)
		if fontBytes, err := os.ReadFile(fontPath); err == nil {
			fontResource = &fyne.StaticResource{
				StaticName:    font,
				StaticContent: fontBytes,
			}
			break
		}
	}

	// 设置自定义主题
	if fontResource != nil {
		myApp.Settings().SetTheme(&customTheme{
			Theme:    theme.DarkTheme(),
			fontData: fontResource,
		})
	} else {
		// 如果没找到字体，使用默认深色主题
		myApp.Settings().SetTheme(theme.DarkTheme())
	}

	myWindow := myApp.NewWindow("ADB批量管理工具 By laohuan12138")

	// 创建主界面
	mainUI := ui.NewMainUI(myWindow)
	myWindow.SetContent(mainUI.Build())

	myWindow.Resize(fyne.NewSize(1200, 800))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}
