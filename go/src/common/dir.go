package common

import (
	"os"
	"path/filepath"
	"strings"
)

func GetExeDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir + "\\"
}
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

//! 返回的文件名，都是相对exe的
//获取指定目录下的所有文件
//	names, err := filepath.Glob("csv/*.csv")
//获取指定目录及子目录下的所有文件，可以匹配后缀过滤
//	names, err := WalkDir("csv/", ".csv")
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error {
		//if err != nil {
		// return err
		//}
		if fi.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}
