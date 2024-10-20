//go:build windows
// +build windows

package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func messageBoxWindows(title string, text string) {
	user32 := windows.NewLazySystemDLL("user32")
	MessageBox := user32.NewProc("MessageBoxW")

	var title16, text16 *uint16
	if title != "" {
		title16, _ = syscall.UTF16PtrFromString(title)
	}
	text16, _ = syscall.UTF16PtrFromString(text)

	MessageBox.Call(uintptr(0), uintptr(unsafe.Pointer(text16)), uintptr(unsafe.Pointer(title16)), uintptr(windows.MB_OK))
}

func checkSingletonWindows() bool {
	path, err := os.Executable()
	if err != nil {
		return false
	}
	hashName := md5.Sum([]byte(path))
	name, err := syscall.UTF16PtrFromString("Local\\" + hex.EncodeToString(hashName[:]))
	if err != nil {
		return false
	}
	_, err = windows.CreateMutex(nil, false, name)
	return err == syscall.ERROR_ALREADY_EXISTS
}

// 这些是Windows API相关的常量
const (
	OFN_EXPLORER        = 0x00080000
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_HIDEREADONLY    = 0x00000004
	OFN_OVERWRITEPROMPT = 0x00000002
)

// OPENFILENAME结构体
type OPENFILENAME struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
}

var (
	modcomdlg32         = syscall.NewLazyDLL("Comdlg32.dll")
	procGetSaveFileName = modcomdlg32.NewProc("GetSaveFileNameW")
)

func utf16Ptr(s string) *uint16 {
	return (*uint16)(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

func saveFileDialog(defaultFileName, filter string) (string, error) {
	ofn := OPENFILENAME{
		lStructSize:  uint32(unsafe.Sizeof(OPENFILENAME{})),
		hwndOwner:    0,
		hInstance:    0,
		lpstrFilter:  utf16Ptr(filter + "\\0\\0"),
		nFilterIndex: 1,
		lpstrFile:    utf16Ptr(defaultFileName),
		nMaxFile:     256,
		flags:        OFN_EXPLORER | OFN_HIDEREADONLY | OFN_OVERWRITEPROMPT,
	}
	r, _, _ := procGetSaveFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", fmt.Errorf("User canceled the operation or an error occurred.")
	}
	return syscall.UTF16ToString((*[256]uint16)(unsafe.Pointer(ofn.lpstrFile))[:]), nil
}

// SaveBytesToFile 保存字节数组到用户选择的文件
func SaveBytesWindows(defaultFileName, filter string, data []byte) error {
	filePath, err := saveFileDialog(defaultFileName, filter)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write to file: %v", err)
	}

	fmt.Printf("The content was written to %s\n", filePath)
	return nil
}
