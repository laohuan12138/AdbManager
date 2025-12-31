package ui

import (
	"adbmanager/internal/adb"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FileManagerUI 文件管理界面
type FileManagerUI struct {
	window         fyne.Window
	adbMgr         *adb.ADBManager
	getDevice      func() string
	currentPath    string
	files          []adb.FileInfo
	fileTable      *widget.List
	pathEntry      *widget.Entry
	selectedFileID int
}

// NewFileManagerUI 创建文件管理界面
func NewFileManagerUI(window fyne.Window, adbMgr *adb.ADBManager, getDevice func() string) *FileManagerUI {
	return &FileManagerUI{
		window:         window,
		adbMgr:         adbMgr,
		getDevice:      getDevice,
		currentPath:    "/sdcard",
		files:          make([]adb.FileInfo, 0),
		selectedFileID: -1,
	}
}

// Build 构建文件管理界面
func (f *FileManagerUI) Build() fyne.CanvasObject {
	// 导航栏
	backBtn := widget.NewButton("← 返回", func() {
		f.currentPath = filepath.Dir(f.currentPath)
		f.pathEntry.SetText(f.currentPath)
		f.refreshFileList()
	})

	refreshBtn := widget.NewButton("🔄 刷新", func() {
		f.refreshFileList()
	})

	f.pathEntry = widget.NewEntry()
	f.pathEntry.SetText(f.currentPath)
	f.pathEntry.OnSubmitted = func(s string) {
		f.currentPath = s
		f.refreshFileList()
	}

	navBar := container.NewBorder(
		nil, nil,
		container.NewHBox(backBtn, refreshBtn),
		nil,
		f.pathEntry,
	)

	// 表头
	header := container.NewHBox(
		widget.NewLabel("📁 名称"),
		widget.NewLabel(""),
		widget.NewLabel("📅 修改日期"),
		widget.NewLabel(""),
		widget.NewLabel("📏 大小"),
		widget.NewLabel(""),
		widget.NewLabel("🔐 权限"),
		widget.NewLabel(""),
		widget.NewLabel("⚙️ 操作"),
	)

	// 文件表格
	f.fileTable = widget.NewList(
		func() int {
			return len(f.files)
		},
		func() fyne.CanvasObject {
			// 创建表格行 - 使用固定宽度
			icon := widget.NewLabel("📁")

			name := widget.NewLabel("")
			name.Wrapping = fyne.TextTruncate

			date := widget.NewLabel("")
			size := widget.NewLabel("")
			perm := widget.NewLabel("")

			editBtn := widget.NewButton("重命名", nil)
			editBtn.Importance = widget.LowImportance

			deleteBtn := widget.NewButton("删除", nil)
			deleteBtn.Importance = widget.DangerImportance

			downloadBtn := widget.NewButton("下载", nil)
			downloadBtn.Importance = widget.SuccessImportance

			uploadBtn := widget.NewButton("上传", nil)
			uploadBtn.Importance = widget.MediumImportance

			// 使用 GridWrap 布局固定列宽
			return container.NewHBox(
				icon,
				container.NewPadded(
					container.NewStack(
						widget.NewLabel("____________________________________________"), // 占位空间
						name,
					),
				),
				date,
				widget.NewLabel("  "),
				size,
				widget.NewLabel("  "),
				perm,
				widget.NewLabel("  "),
				editBtn,
				deleteBtn,
				downloadBtn,
				uploadBtn,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[PANIC] 更新列表项 %d 时发生错误: %v\n", id, r)
				}
			}()

			if id >= len(f.files) {
				return
			}

			file := f.files[id]
			box := obj.(*fyne.Container)

			// HBox: [icon, nameStack, date, sp, size, sp, perm, sp, editBtn, deleteBtn, downloadBtn, uploadBtn]
			if len(box.Objects) < 12 {
				return
			}

			icon, _ := box.Objects[0].(*widget.Label)
			// nameStack -> Padded -> Stack -> [placeholder, name]
			nameContainer, ok := box.Objects[1].(*fyne.Container)
			if !ok || len(nameContainer.Objects) == 0 {
				return
			}
			nameStack, ok := nameContainer.Objects[0].(*fyne.Container)
			if !ok || len(nameStack.Objects) < 2 {
				return
			}
			name, _ := nameStack.Objects[1].(*widget.Label)

			date, _ := box.Objects[2].(*widget.Label)
			size, _ := box.Objects[4].(*widget.Label)
			perm, _ := box.Objects[6].(*widget.Label)
			editBtn, _ := box.Objects[8].(*widget.Button)
			deleteBtn, _ := box.Objects[9].(*widget.Button)
			downloadBtn, _ := box.Objects[10].(*widget.Button)
			uploadBtn, _ := box.Objects[11].(*widget.Button)

			if icon == nil || name == nil || date == nil || size == nil || perm == nil || editBtn == nil || deleteBtn == nil || downloadBtn == nil || uploadBtn == nil {
				return
			}

			// 设置图标
			if strings.HasPrefix(file.Permissions, "d") {
				icon.SetText("📁")
			} else {
				icon.SetText("📄")
			}

			// 设置文件名
			name.SetText(file.Name)

			// 设置日期、大小、权限
			date.SetText(file.Date)
			size.SetText(file.Size)
			perm.SetText(file.Permissions)

			// 编辑按钮
			editBtn.OnTapped = func() {
				f.showRenameDialog(int(id))
			}

			// 删除按钮
			deleteBtn.OnTapped = func() {
				f.deleteFile(int(id))
			}

			// 下载按钮
			downloadBtn.OnTapped = func() {
				f.downloadFile(int(id))
			}

			// 上传按钮（只对目录显示）
			if strings.HasPrefix(file.Permissions, "d") {
				uploadBtn.Show()
				uploadBtn.OnTapped = func() {
					f.uploadToDirectory(int(id))
				}
			} else {
				uploadBtn.Hide()
			}
		},
	)

	// 双击进入目录
	f.fileTable.OnSelected = func(id widget.ListItemID) {
		if id >= widget.ListItemID(len(f.files)) {
			return
		}
		file := f.files[id]
		// 如果是目录
		if strings.HasPrefix(file.Permissions, "d") {
			if file.Name == ".." {
				f.currentPath = filepath.Dir(f.currentPath)
			} else if file.Name != "." {
				f.currentPath = filepath.Join(f.currentPath, file.Name)
			}
			f.pathEntry.SetText(f.currentPath)
			f.refreshFileList()
		}
	}

	// 操作按钮栏
	uploadBtn := widget.NewButton("⬆️ 上传文件", func() {
		f.uploadFile()
	})

	downloadBtn := widget.NewButton("⬇️ 下载选中", func() {
		if f.fileTable.UnselectAll(); f.selectedFileID >= 0 {
			f.downloadFile(f.selectedFileID)
		}
	})

	chmodBtn := widget.NewButton("🔐 修改权限", func() {
		if f.selectedFileID >= 0 {
			f.showChmodDialog(f.selectedFileID)
		}
	})

	actionBar := container.NewHBox(
		uploadBtn,
		downloadBtn,
		chmodBtn,
	)

	// 整体布局
	content := container.NewBorder(
		container.NewVBox(
			navBar,
			widget.NewSeparator(),
			header,
		),
		container.NewVBox(
			widget.NewSeparator(),
			actionBar,
		),
		nil, nil,
		container.NewScroll(f.fileTable),
	)

	return content
}

// refreshFileList 刷新文件列表
func (f *FileManagerUI) refreshFileList() {
	device := f.getDevice()
	if device == "" {
		// 没有设备时清空列表
		f.files = make([]adb.FileInfo, 0)
		if f.fileTable != nil {
			f.fileTable.Refresh()
		}
		return
	}

	// 先尝试正常获取
	files, err := f.adbMgr.ListFiles(device, f.currentPath)
	if err != nil {
		// 只有失败时才尝试root（避免重启adb导致断开）
		// showError(f.window, "获取文件列表失败", err)

		// 错误时清空列表
		f.files = make([]adb.FileInfo, 0)
		if f.fileTable != nil {
			f.fileTable.Refresh()
		}
		return
	}

	f.files = files
	if f.fileTable != nil {
		f.fileTable.Refresh()
	}
}

// uploadFile 上传文件
func (f *FileManagerUI) uploadFile() {
	dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()

		device := f.getDevice()
		if device == "" {
			showError(f.window, "错误", fmt.Errorf("请先选择设备"))
			return
		}

		localPath := uc.URI().Path()
		remotePath := filepath.Join(f.currentPath, filepath.Base(localPath))

		err = f.adbMgr.PushFile(device, localPath, remotePath)
		if err != nil {
			showError(f.window, "上传失败", err)
			return
		}

		showInfo(f.window, "成功", "文件上传成功")
		f.refreshFileList()
	}, f.window)
}

// downloadFile 下载文件
func (f *FileManagerUI) downloadFile(id int) {
	if id < 0 || id >= len(f.files) {
		return
	}

	file := f.files[id]
	device := f.getDevice()
	if device == "" {
		showError(f.window, "错误", fmt.Errorf("请先选择设备"))
		return
	}

	remotePath := filepath.Join(f.currentPath, file.Name)

	// 选择保存位置
	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()

		localPath := uc.URI().Path()

		err = f.adbMgr.PullFile(device, remotePath, localPath)
		if err != nil {
			showError(f.window, "下载失败", err)
			return
		}

		showInfo(f.window, "成功", fmt.Sprintf("文件已下载到:\n%s", localPath))
	}, f.window)
}

// deleteFile 删除文件
func (f *FileManagerUI) deleteFile(id int) {
	if id < 0 || id >= len(f.files) {
		return
	}

	file := f.files[id]

	dialog.ShowConfirm("确认删除",
		"确定要删除 "+file.Name+" 吗？",
		func(confirmed bool) {
			if !confirmed {
				return
			}

			device := f.getDevice()
			if device == "" {
				showError(f.window, "错误", fmt.Errorf("请先选择设备"))
				return
			}

			remotePath := filepath.Join(f.currentPath, file.Name)
			err := f.adbMgr.DeleteFile(device, remotePath)
			if err != nil {
				showError(f.window, "删除失败", err)
				return
			}

			showInfo(f.window, "成功", "文件删除成功")
			f.refreshFileList()
		}, f.window)
}

// showRenameDialog 显示重命名对话框
func (f *FileManagerUI) showRenameDialog(id int) {
	if id < 0 || id >= len(f.files) {
		return
	}

	file := f.files[id]

	newNameEntry := widget.NewEntry()
	newNameEntry.SetText(file.Name)

	dialog.ShowCustomConfirm("重命名", "确定", "取消",
		newNameEntry,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			newName := newNameEntry.Text
			if newName == "" || newName == file.Name {
				return
			}

			device := f.getDevice()
			if device == "" {
				showError(f.window, "错误", fmt.Errorf("请先选择设备"))
				return
			}

			oldPath := filepath.Join(f.currentPath, file.Name)
			newPath := filepath.Join(f.currentPath, newName)

			err := f.adbMgr.RenameFile(device, oldPath, newPath)
			if err != nil {
				showError(f.window, "重命名失败", err)
				return
			}

			showInfo(f.window, "成功", "文件重命名成功")
			f.refreshFileList()
		}, f.window)
}

// showChmodDialog 显示权限修改对话框
func (f *FileManagerUI) showChmodDialog(id int) {
	if id < 0 || id >= len(f.files) {
		return
	}

	file := f.files[id]

	permEntry := widget.NewEntry()
	permEntry.SetPlaceHolder("例如: 755, 644")

	dialog.ShowCustomConfirm("修改权限", "确定", "取消",
		container.NewVBox(
			widget.NewLabel("文件: "+file.Name),
			widget.NewLabel("当前权限: "+file.Permissions),
			permEntry,
		),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			perm := permEntry.Text
			if perm == "" {
				return
			}

			device := f.getDevice()
			if device == "" {
				showError(f.window, "错误", fmt.Errorf("请先选择设备"))
				return
			}

			remotePath := filepath.Join(f.currentPath, file.Name)
			err := f.adbMgr.ChangePermissions(device, remotePath, perm)
			if err != nil {
				showError(f.window, "修改权限失败", err)
				return
			}

			showInfo(f.window, "成功", "权限修改成功")
			f.refreshFileList()
		}, f.window)
}

// uploadToDirectory 上传文件到指定目录
func (f *FileManagerUI) uploadToDirectory(id int) {
	if id < 0 || id >= len(f.files) {
		return
	}

	file := f.files[id]
	// 确保是目录
	if !strings.HasPrefix(file.Permissions, "d") {
		return
	}

	device := f.getDevice()
	if device == "" {
		showError(f.window, "错误", fmt.Errorf("请先选择设备"))
		return
	}

	targetDir := filepath.Join(f.currentPath, file.Name)

	// 选择要上传的文件
	dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()

		localPath := uc.URI().Path()
		fileName := filepath.Base(localPath)
		remotePath := filepath.Join(targetDir, fileName)

		err = f.adbMgr.PushFile(device, localPath, remotePath)
		if err != nil {
			showError(f.window, "上传失败", err)
			return
		}

		showInfo(f.window, "成功", fmt.Sprintf("文件已上传到:\n%s", remotePath))
		f.refreshFileList()
	}, f.window)
}
