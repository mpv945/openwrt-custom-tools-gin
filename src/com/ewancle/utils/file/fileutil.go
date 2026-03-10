package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const uploadDir = "./uploads"

func SaveFile(reader io.Reader, filename string) (string, error) {

	// 1. 当前工作目录
	cwd, err17 := os.Getwd()
	if err17 != nil {
		panic(err17)
	}
	fmt.Println("当前工作目录:", cwd)

	// 2. HOME 目录
	home, err13 := os.UserHomeDir()
	if err13 != nil {
		panic(err13)
	}
	fmt.Println("用户 HOME 目录:", home)

	// 3. 临时目录
	tmpDir := os.TempDir()
	fmt.Println("系统临时目录:", tmpDir)

	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return "", err
	}

	path := filepath.Join(uploadDir, filename)

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			panic(err)
		}
	}(out)

	buf := make([]byte, 1024*64)

	_, err = io.CopyBuffer(out, reader, buf)
	if err != nil {
		return "", err
	}

	return path, nil
}

// 使用
// dir := "./testdata"
//	exts := []string{".txt", ".log"}
//
//	files, err := FindFilesByExt(dir, exts, true) // true 表示递归子目录
//	if err != nil {
//		fmt.Println("搜索失败:", err)
//		return
//	}
//
//	fmt.Println("找到文件:")
//	for _, f := range files {
//		fmt.Println(f)
//	}

// FindFilesByExt 搜索指定目录下指定后缀的文件
// dir: 要搜索的目录
// exts: 支持多个后缀，如 []string{".txt", ".log"}
// recursive: 是否递归子目录
func FindFilesByExt(dir string, exts []string, recursive bool) ([]string, error) {
	var result []string

	// 将后缀统一为小写，便于不区分大小写
	for i := range exts {
		exts[i] = strings.ToLower(exts[i])
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 忽略权限或读取错误
			return nil
		}
		if info.IsDir() {
			if path != dir && !recursive {
				// 如果不递归，跳过子目录
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		for _, e := range exts {
			if ext == e {
				result = append(result, path)
				break
			}
		}
		return nil
	}

	err := filepath.Walk(dir, walkFunc)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CreateTempFileAutoDel() error {
	tmpFile, err := os.CreateTemp("", "myapp-*.tmp")
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			panic(err)
		}
	}(tmpFile.Name()) // 自动删除

	// 写入数据
	_, err = tmpFile.Write([]byte("hello world"))
	if err != nil {
		return err
	}
	err34 := tmpFile.Close()
	if err34 != nil {
		return err34
	}

	// 可以在这里使用 tmpFile.Name() 做临时操作
	fmt.Println("使用临时文件:", tmpFile.Name())
	return nil
}
