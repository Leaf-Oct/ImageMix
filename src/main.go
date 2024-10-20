package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"golang.design/x/clipboard"
)

const (
	MAX_PATH = 260
	WINDOWS  = "windows"
	NOGUI    = "nogui"
	LINUX    = "linux"
)

//go:embed icon.ico
var icon []byte

func main() {
	if checkSingletonWindows() {
		messageBoxWindows("错误", "程序已运行。看右下角系统托盘(Windows)")
		return
	}
	// 设置systray
	systray.Run(onReady, onExit)
}

func onReady() {

	systray.SetIcon(icon)
	systray.SetTitle("Mix")
	systray.SetTooltip("图片像素混淆器")

	mixItem := systray.AddMenuItem("混淆", "从剪切板获取图像，修改像素，写回剪切板")
	saveItem := systray.AddMenuItem("保存", "保存剪切板图像到本地")
	aboutItem := systray.AddMenuItem("关于", "关于软件")
	exitItem := systray.AddMenuItem("退出", "退出程序")

	go func() {
		for {
			select {
			case <-mixItem.ClickedCh:
				mix()
			case <-aboutItem.ClickedCh:
				messageBoxWindows("关于", "本程序功能为，从剪切板读取图像(jpg, png, bmp, webp)，随机修改若干像素，写回剪切板\n十月叶~Leaf Oct 开发")
			case <-exitItem.ClickedCh:
				systray.Quit()
			case <-saveItem.ClickedCh:
				save()
			}
		}
	}()
}

func onExit() {
	// 清理工作可以在这里进行
}

func save() {
	img_bytes := clipboard.Read(clipboard.FmtImage)
	if len(img_bytes) == 0 {
		messageBoxWindows("错误", "复制的不是图像")
		return
	}
	filename := fmt.Sprintf("%d", time.Now().UnixMilli()) + ".png"
	filter := "Image Files (*.png)\\0*.png\\0"
	SaveBytesWindows(filename, filter, img_bytes)

	img_bytes = nil
	runtime.GC()
}

func mix() {
	img_bytes := clipboard.Read(clipboard.FmtImage)
	if len(img_bytes) == 0 {
		messageBoxWindows("错误", "复制的不是图像")
		return
	}
	img, err := png.Decode(bytes.NewReader(img_bytes))
	if err != nil {
		messageBoxWindows("解析图像错误", "读取剪切板图像后转化的格式不是png。问题可能出在clipboard库")
		return
	}
	bound := img.Bounds()
	width := bound.Max.X
	height := bound.Max.Y
	new_image := image.NewRGBA64(bound)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			new_image.SetRGBA64(x, y, color.RGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)})
			fmt.Println(x, y, " : ", r, g, b)
		}
	}
	// 随机更改10个像素的值
	for i := 0; i < 10; i += 1 {
		random_x := rand.Intn(width)
		random_y := rand.Intn(height)
		pixel := new_image.At(random_x, random_y)
		r, g, b, a := pixel.RGBA()
		r = rand.Uint32()
		g = rand.Uint32()
		b = rand.Uint32()
		changed_pixel := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		new_image.Set(random_x, random_y, changed_pixel)
	}
	var modified_png bytes.Buffer
	err = png.Encode(&modified_png, new_image)
	if err != nil {
		messageBoxWindows("失败", "混淆后的图片无法编码")
		return
	}
	clipboard.Write(clipboard.FmtImage, modified_png.Bytes())

	// 不放心，手动释放下内存
	img_bytes = nil
	img = nil
	new_image = nil
	runtime.GC()
}

// debug使用，在正式发布版本中不会用到
func writeOut(bytes []byte) {
	file, _ := os.Create("output.png")
	file.Write(bytes)
	file.Close()
}
