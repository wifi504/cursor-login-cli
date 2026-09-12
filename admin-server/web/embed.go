// Package web 嵌入管理端前端静态资源。
//
// 仓库内保留 dist/index.html 占位，保证未跑前端构建时 go run / go build 仍可通过
// （开发联调走 Vite，不依赖真实 embed）。发版任务会用正式构建产物覆盖 dist/。
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS 返回嵌入的管理端前端 dist 文件系统。
func FS() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
