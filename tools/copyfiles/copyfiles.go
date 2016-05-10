package copyfiles

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lessgo/lessgo"
	"github.com/lessgo/lessgo/utils"
)

type FileInfo struct {
	RelPath string
	ModTime time.Time
	IsDir   bool
	Handle  *os.File
}

// 拷贝srcFolder文件夹下所有文件到dstFolder文件夹
// suffix不为空时，仅复制含有suffix后缀的文件
// copyFunc中可自定义复制操作，如附加一些字符串替换操作等，为nil时默认只有复制操作
// 注：在copyFunc中不可关闭文件句柄
func CopyFiles(srcFolder, dstFolder, suffix string, copyFunc func(srcHandle, dstHandle *os.File) error) {
	files_ch := make(chan *FileInfo, 100)
	go walkFiles(srcFolder, suffix, files_ch) //在一个独立的 goroutine 中遍历文件
	os.MkdirAll(dstFolder, os.ModePerm)
	writeFiles(dstFolder, files_ch, copyFunc)
}

//遍历目录，将文件信息传入通道
func walkFiles(srcDir, suffix string, c chan<- *FileInfo) {
	suffix = strings.ToUpper(suffix)
	filepath.Walk(srcDir, func(f string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil {
			lessgo.Logger().Error("%v", err)
			return err
		}
		fileInfo := &FileInfo{}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			if fh, err := os.OpenFile(f, os.O_RDONLY, os.ModePerm); err != nil {
				lessgo.Logger().Error("%v", err)
				return err
			} else {
				fileInfo.Handle = fh
				fileInfo.RelPath, _ = filepath.Rel(srcDir, f) //相对路径
				fileInfo.IsDir = fi.IsDir()
				fileInfo.ModTime = fi.ModTime()
			}
			c <- fileInfo
		}
		return nil
	})
	close(c) //遍历完成，关闭通道
}

//写目标文件
func writeFiles(dstDir string, c <-chan *FileInfo, copyFunc func(srcHandle, dstHandle *os.File) error) {
	if err := os.Chdir(dstDir); err != nil { //切换工作路径
		lessgo.Logger().Fatal("%v", err)
	}
	for f := range c {
		if fi, err := os.Stat(f.RelPath); os.IsNotExist(err) { //目标不存在
			if f.IsDir {
				if err := os.MkdirAll(f.RelPath, os.ModeDir); err != nil {
					lessgo.Logger().Error("%v", err)
				}
			} else {
				if err := ioCopy(f.Handle, f.RelPath, copyFunc); err != nil {
					lessgo.Logger().Error("%v", err)
				} else {
					lessgo.Logger().Info("CP: %v", f.RelPath)
				}
			}
		} else if !f.IsDir { //目标存在，而且源不是一个目录
			if fi.IsDir() != f.IsDir { //检查文件名被目录名占用冲突
				lessgo.Logger().Error("%v", "filename conflict:", f.RelPath)
			} else if !fi.ModTime().Equal(f.ModTime) { //源文件修改后重写
				if err := ioCopy(f.Handle, f.RelPath, copyFunc); err != nil {
					lessgo.Logger().Error("%v", err)
				} else {
					lessgo.Logger().Info("CP: %v", f.RelPath)
				}
			}
		}
	}

	//切换工作路径到自身所在目录下
	utils.SelfChdir()
}

//复制文件数据
func ioCopy(srcHandle *os.File, dstPth string, copyFunc func(srcHandle, dstHandle *os.File) error) (err error) {
	defer srcHandle.Close()
	dstHandle, err := os.OpenFile(dstPth, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer dstHandle.Close()

	if copyFunc != nil {
		return copyFunc(srcHandle, dstHandle)
	}

	_, err = io.Copy(dstHandle, srcHandle)
	return
}
