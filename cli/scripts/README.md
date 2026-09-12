# 安装 / 卸载脚本说明

正式环境中的安装与卸载脚本由 **admin-server 按已配置的对外地址动态生成**。

CLI 二进制在发版时由流水线写入 `admin-server/releases/bin/`，再被服务端 `go:embed` 打进包内；安装脚本通过 `/download/<平台文件名>` 拉取。

实现代码见：`admin-server/internal/install/scripts.go`、`admin-server/releases/`。
