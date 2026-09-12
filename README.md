# Cursor Login

企业内部 Cursor 上号平台：管理员通过 **Cursor Login Admin** 维护账号池，员工通过 **Cursor Login CLI**（命令 `cursor-login`）凭上号码完成本机上号

中心服务为 **Cursor Login Server** 单二进制（内嵌管理端与 CLI）

## 下载

从 [最新 Release](https://github.com/wifi504/cursor-login-cli/releases/latest) 下载对应平台的 **Cursor Login Server**：


| 平台                  | 文件                               | 下载                                                                                                        |
| ------------------- | -------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Windows x64         | `admin-server-windows-amd64.exe` | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-windows-amd64.exe) |
| Windows ARM64       | `admin-server-windows-arm64.exe` | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-windows-arm64.exe) |
| macOS Intel         | `admin-server-darwin-amd64`      | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-darwin-amd64)      |
| macOS Apple Silicon | `admin-server-darwin-arm64`      | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-darwin-arm64)      |
| Linux x64           | `admin-server-linux-amd64`       | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-linux-amd64)       |
| Linux ARM64         | `admin-server-linux-arm64`       | [下载](https://github.com/wifi504/cursor-login-cli/releases/latest/download/admin-server-linux-arm64)       |


## 使用指南

### 1. 部署 Cursor Login Server

1. 下载上表对应文件，放到服务器目录（进程需对该目录可写，服务端启动会自动在当前路径下写 SQLite 与日志）

2. 非 Windows 先赋权：`chmod +x ./admin-server-<os>-<arch>`

3. 启动，例如：

   ```bash
   ./admin-server-linux-amd64 --listen :8080
   ```

4. 按终端日志中的 **临时安全入口** 打开浏览器，完成初始化（管理员账号、自定义安全入口、对外访问地址）

5. 之后只用自定义安全入口访问 Admin；根路径 `/` 不可访问；**生产环境建议前面加 HTTPS 反代**

常用参数：

- `--listen`：监听地址，默认 `:8080`
- `--dev`：仅开发联调使用（固定入口 `/__dev__`）

### 2. 管理员（Cursor Login Admin）

1. 用安全入口登录管理端
2. 在概览页复制 **安装 / 卸载** 命令发给员工（脚本由服务端按对外地址生成）
3. 在「账号」中创建账号：填写展示名、上号码、领取配额；核销前须补齐 **邮箱** 与 **AccessToken**（缺一不可领取）
4. 可用「获取 AccessToken」按钮来现场签发一个长期有效的 AccessToken，需要当前浏览器已经登录 Cursor 账号

### 3. 员工（Cursor Login CLI）

1. 本机已安装并**至少运行过一次** [Cursor](https://cursor.com) 客户端

2. 执行管理员提供的安装命令（会下载 CLI，并配置 `CURSOR_LOGIN_API` 与 PATH）。**重复执行安装会覆盖升级**，不会先拆掉旧版；下载失败时保留原有安装。

3. 向管理员索取上号码，在终端执行：

   ```bash
   cursor-login <上号码>
   ```

4. 按提示确认核销；请先关闭正在运行的 Cursor，保存好未保存的工作。成功后重启 Cursor 即可使用已上号账号。若提示找不到命令：新开终端，或关掉整个 Windows Terminal / 重启 IDE 后再试。

5. 卸载请使用管理端提供的**独立卸载命令**（删除 CLI、环境变量键与 shell 配置中的安装标记，不留空键 / 残留行）。

## 仓库结构

```text
admin-server/   Cursor Login Server（Go / Gin / SQLite；embed Admin 与 CLI）
admin-web/      Cursor Login Admin（Vue 3 + TypeScript + Vite + Arco Design）
cli/            Cursor Login CLI（命令名 cursor-login）
scripts/        根目录一键安装依赖、发版构建、清理脚本
.github/        GitHub Actions（tag v* 发版）
.vscode/        Cursor / VS Code 任务（推荐日常开发入口）
```


| 目录             | 组件                  | 说明                                        |
| -------------- | ------------------- | ----------------------------------------- |
| `admin-server` | Cursor Login Server | HTTP 服务、安全入口、会话、账号 CRUD、安装脚本与 `/download` |
| `admin-web`    | Cursor Login Admin  | 管理界面；开发时由 Vite 提供，发版后由服务端 embed           |
| `cli`          | Cursor Login CLI    | 员工侧命令行；安装脚本会配置运行所需环境变量                    |


## 开发

### 环境要求

- Go（见各模块 `go.mod`）
- Node.js
- pnpm

### 推荐：用 VS Code Task

打开仓库根目录 → `Terminal` → `Run Task…`（或命令面板 **Tasks: Run Task**）。


| 任务             | 作用                                                          |
| -------------- | ----------------------------------------------------------- |
| **组合：安装全部依赖**  | 安装 Admin / Server / CLI 依赖                                  |
| **组合：开发启动前后端** | 并行启动 Cursor Login Admin（Vite）+ Cursor Login Server（`--dev`） |
| **组合：构建发布全平台** | Admin embed → CLI 全平台 → Server 全平台                          |
| **清理：构建产物**    | 清理 dist / bin / logs / 本地 db 等（保留依赖）                        |


首次开发：

1. 跑 **组合：安装全部依赖**
2. 跑 **组合：开发启动前后端**
3. 浏览器只打开 [http://localhost:5173/](http://localhost:5173/)（Cursor Login Admin）
4. 不要用 `http://127.0.0.1:8080/` 做日常页面开发（根路径故意返回 403）

开发约定：

- 管理 API 经 Vite 代理到 `http://127.0.0.1:8080/__dev__/api/*`
- `admin-web/.env.development` 中 `VITE_ADMIN_ENTRY=/__dev__`，不要改成正式安全入口或完整 URL
- `go run` 时 SQLite / 日志落在 `admin-server/` 工作目录
- 新增 / 调整管理端视图后，如需重新生成路由：在 `admin-web/` 执行 `pnpm gen-routes`

### 命令行等价方式

可以参考 [.vscode/tasks.json](.vscode/tasks.json) 的实现，常用命令如下：

```bash
# 安装依赖
# Windows:  powershell -File scripts/install-all.ps1
# macOS / Linux:
bash scripts/install-all.sh

# 开发：Cursor Login Admin
cd admin-web && pnpm dev

# 开发：Cursor Login Server（开发固定入口 /__dev__）
cd admin-server && go run ./cmd/server --listen :8080 --dev
```

## 构建与发版

```bash
# Windows
powershell -File scripts/build-release.ps1

# macOS / Linux
bash scripts/build-release.sh
```

流水线顺序：

1. 构建 Cursor Login Admin，复制到 `admin-server/web/dist`（供 `go:embed`）
2. 交叉编译全平台 Cursor Login CLI，同步到 `admin-server/releases/bin/`
3. 交叉编译全平台 Cursor Login Server（内嵌 Admin 与 CLI）

清理：

```bash
# Windows
powershell -File scripts/clean.ps1

# macOS / Linux
bash scripts/clean.sh
```

## 致谢与引用

本项目在实现过程中参考并借鉴了以下开源项目：


| 项目                                                                | 说明                                                                                                                                                                                        | 协议                                                                                            |
| ----------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| [CursorPool_Client](https://github.com/Cloxl/CursorPool_Client)   | Cloxl / Sanyela — Cursor 账号池客户端（Tauri）；本项目参考其账号与本地登录相关思路。原项目另有品牌保留与署名约定，详见其 [开源协议声明](https://github.com/Cloxl/CursorPool_Client#-%E5%BC%80%E6%BA%90%E5%8D%8F%E8%AE%AE%E5%A3%B0%E6%98%8E)。 | [MIT](https://github.com/Cloxl/CursorPool_Client/blob/main/LICENSE)                           |
| [CursorTokenTool](https://github.com/wang-xi-lei/CursorTokenTool) | wang-xi-lei — Cursor AccessToken / 认证相关工具；本项目参考其 Token 获取与认证流程相关实现。                                                                                                                       | MIT（见其 [README](https://github.com/wang-xi-lei/CursorTokenTool#-%E8%AE%B8%E5%8F%AF%E8%AF%81)） |


感谢上述作者的开源贡献

## License

本仓库以 **MIT** 协议开源

```text
MIT License

Copyright (c) 2025 WIFI连接超时
```

