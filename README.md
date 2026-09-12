# Cursor Login

企业内部 Cursor 上号平台：由管理员通过 **Cursor Login Admin** 集中维护一批已开通的 Cursor 账号，员工通过 **Cursor Login CLI** 领取并写入本机登录态，无需直接持有账号密码。中心服务为 **Cursor Login Server**。

## 产品定位

公司通常会预注册并按周期续费一批 Cursor 席位。本项目把这些账号放进中心服务统一管理与下发：

- **管理员（Cursor Login Admin）**：在后台维护账号池（展示名、邮箱、长期 AccessToken、上号码、领取配额等），并按需获取 / 更新 AccessToken。
- **员工（Cursor Login CLI）**：用安装脚本装好命令行工具后，凭上号码从服务端领取账号凭证，由 CLI 完成本机上号；日常不接触明文密码。命令名为 `cursor-login`。
- **部署（Cursor Login Server）**：单二进制服务端（内嵌 Cursor Login Admin 与 CLI 安装包）+ 旁路 SQLite；对外暴露域名，Admin 走自定义安全入口，根路径不开放。

当前仓库已具备管理端基建、账号管理与 AccessToken 获取、CLI 安装 / 探活等能力；员工凭上号码领取与写入 Cursor 本地登录态等能力会随版本继续完善。

## 仓库结构

```text
admin-server/   Cursor Login Server（Go / Gin / SQLite；embed Admin 与 CLI）
admin-web/      Cursor Login Admin（Vue 3 + TypeScript + Vite + Arco Design）
cli/            Cursor Login CLI（命令名 cursor-login）
scripts/        根目录一键安装依赖、发版构建、清理脚本
.vscode/        Cursor / VS Code 任务（推荐日常开发入口）
```

| 目录 | 组件 | 说明 |
|------|------|------|
| `admin-server` | Cursor Login Server | HTTP 服务、安全入口、会话登录、账号 CRUD、安装脚本与 `/download` |
| `admin-web` | Cursor Login Admin | 管理界面；开发时由 Vite 提供，发版后由服务端 embed |
| `cli` | Cursor Login CLI | 员工侧命令行；安装脚本会配置运行所需环境变量 |

## 环境要求

本机可直接调用：

- Go
- Node.js
- pnpm

## 开发

### 推荐：用 Cursor / VS Code 任务

打开仓库根目录 → `Terminal` → `Run Task…`（或命令面板 **Tasks: Run Task**）。

| 任务 | 作用 |
|------|------|
| **组合：安装全部依赖** | 安装 Admin / Server / CLI 依赖 |
| **组合：开发启动前后端** | 并行启动 Cursor Login Admin（Vite）+ Cursor Login Server（`--dev`） |
| **组合：构建发布全平台** | Admin embed → CLI 全平台 → Server 全平台 |
| **清理：构建产物** | 清理 dist / bin / logs / 本地 db 等（保留依赖） |

首次开发：

1. 跑 **组合：安装全部依赖**
2. 跑 **组合：开发启动前后端**
3. 浏览器只打开 **http://localhost:5173/**（Cursor Login Admin）
4. 不要用 `http://127.0.0.1:8080/` 做日常页面开发（根路径故意返回 403）

开发约定：

- 管理 API 经 Vite 代理到 `http://127.0.0.1:8080/__dev__/api/*`
- `admin-web/.env.development` 中 `VITE_ADMIN_ENTRY=/__dev__`，不要改成正式安全入口或完整 URL
- `go run` 时 SQLite / 日志落在 `admin-server/` 工作目录
- 新增 / 调整管理端视图后，如需重新生成路由：在 `admin-web/` 执行 `pnpm gen-routes`

### 命令行等价方式

权威命令以 [`.vscode/tasks.json`](.vscode/tasks.json) 为准。常用等价：

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

服务端启动参数：

- `--listen`：监听地址，默认 `:8080`
- `--dev`：启用固定管理入口 `/__dev__`，供 Vite 联调

## 构建与发版

推荐直接跑 **组合：构建发布全平台**，或：

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

也可分步对照 tasks：

- Admin：`admin-web` 下 `pnpm build` 后复制到 `admin-server/web/dist`
- CLI：`cli/scripts/build-all.ps1` / `build-all.sh`
- Server：`admin-server/scripts/build-all.ps1` / `build-all.sh`

清理：

```bash
# Windows
powershell -File scripts/clean.ps1

# macOS / Linux
bash scripts/clean.sh
```

## 部署与使用概要

1. 将对应平台的 Cursor Login Server 二进制放到服务器，保证进程可写同目录（SQLite 与日志旁路落盘）。
2. 首次访问临时安全入口完成初始化（管理员账号、自定义安全入口、对外域名）。
3. 之后只用自定义安全入口进入 Cursor Login Admin；根路径 `/` 不可访问。
4. 在概览页复制安装 / 卸载命令，员工机器执行后即可使用 `cursor-login`（Cursor Login CLI，需能访问已配置的对外地址）。

公开路径包括健康检查、安装 / 卸载脚本与 CLI 下载；管理 API 位于安全入口之下，需登录 Cookie。
