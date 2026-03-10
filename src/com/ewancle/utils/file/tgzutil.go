package fileutil

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// 使用
//	files := []string{"file1.txt", "dir1"}
//	err := tgzutil.Compress("output.tar.gz", files)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("压缩完成!")
//
//	err = tgzutil.Decompress("output.tar.gz", "./extract")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("解压完成!")

// Compress 压缩 files 到 output tar.gz 文件
func Compress(output string, files []string) error {
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			panic(err)
		}
	}(outFile)

	gzw := gzip.NewWriter(outFile)
	defer func(gzw *gzip.Writer) {
		err := gzw.Close()
		if err != nil {
			panic(err)
		}
	}(gzw)

	tw := tar.NewWriter(gzw)
	defer func(tw *tar.Writer) {
		err := tw.Close()
		if err != nil {
			panic(err)
		}
	}(tw)

	for _, file := range files {
		err := addFileToTar(tw, file, "")
		if err != nil {
			return err
		}
	}
	return nil
}

// addFileToTar 递归添加文件或目录到 tar.Writer
func addFileToTar(tw *tar.Writer, path, baseInTar string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// tar 内的路径
	var tarPath string
	if baseInTar != "" {
		tarPath = filepath.Join(baseInTar, filepath.Base(path))
	} else {
		tarPath = filepath.Base(path)
	}

	if info.IsDir() {
		// 添加目录头
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = tarPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 遍历子文件
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			subPath := filepath.Join(path, entry.Name())
			if err := addFileToTar(tw, subPath, tarPath); err != nil {
				return err
			}
		}
	} else {
		// 添加文件头
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = tarPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 写入文件内容
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)

		_, err = io.Copy(tw, f)
		return err
	}
	return nil
}

// Decompress 解压 tar.gz 到 destDir
func Decompress(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func(gzr *gzip.Reader) {
		err := gzr.Close()
		if err != nil {
			panic(err)
		}
	}(gzr)

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(outFile, tr)
			err12 := outFile.Close()
			if err12 != nil {
				return err12
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
