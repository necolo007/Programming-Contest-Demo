package fs

import "os"

// FileExist 检查文件是否存在
func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
