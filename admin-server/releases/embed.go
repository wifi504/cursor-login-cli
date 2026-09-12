package releases

import (
	"embed"
	"io/fs"
)

//go:embed all:bin
var binFS embed.FS

// FS 返回嵌入的 CLI 二进制目录（发版流水线写入 bin/ 后再编译服务端）。
func FS() (fs.FS, error) {
	return fs.Sub(binFS, "bin")
}

// Read 读取嵌入的 CLI 产物；不存在时返回错误。
func Read(name string) ([]byte, error) {
	root, err := FS()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(root, name)
}
