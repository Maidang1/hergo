# Hergo

[Herdr](https://github.com/herdrdev/herdr) 的原生桌面客户端，让工作区、Git 工作树、coding agent 和多窗格终端集中在一个窗口里。

Hergo 取名自 **Her**dr 与 my**go**。使用 Go 和 [mygo](https://github.com/egoist/mygo) 构建，支持 macOS、Windows x64 与 Linux amd64/arm64。界面与终端不依赖 WebView，运行时无需 Node.js。

[下载最新版本](https://github.com/Maidang1/hergo/releases/latest) · [更新记录](CHANGELOG.md) · [反馈问题](https://github.com/Maidang1/hergo/issues)

## 功能

- **工作区与工作树**：浏览已有会话，查看分支、ahead/behind 和本地监听端口，通过菜单管理工作区与 worktree。
- **多窗格终端**：切换标签、水平或垂直分屏、调整比例、方向导航，以及放大和恢复窗格。
- **代码改动查看**：Git 仓库的工具栏显示 Git 图标，不是 Git 仓库时不显示。可查看相对 HEAD 的未提交改动，也可点历史提交查看该次改动；支持单栏、左右对照、语法高亮、搜索和手动刷新。
- **Agent 状态与详情**：按工作区聚合 agent，查看检测诊断、进程与会话引用，独立命名 agent，配置通知、输出匹配提醒与 agent 集成。
- **Herdr 管理入口**：设置内管理插件、会话恢复配置与保存的 SSH 机器；通过独立命令终端交互安装插件、连接远程或打开本地命名会话。
- **键盘与鼠标入口**：保留 Herdr prefix 模式，提供命令面板、快捷键参考、右键菜单和双击重命名。
- **终端体验**：文字选择与复制粘贴、查找、清屏、字体缩放、文件路径打开和 Zen 模式；具体能力因平台而异。
- **Agent 交互**：在菜单或命令面板给 agent 发送 prompt、在 agent 停下（完成/等待/空闲）时提醒一次、在 pane 中启动 agent；侧栏按视图（全部 / 需要处理 / 工作中 / 当前工作区）筛选 agent，并与 Herdr 客户端共享。
- **整理与布局**：拖拽侧栏工作区和标签重新排序；把 pane 移到新标签、新工作区或其他标签；保存标签布局并以新标签重新打开。
- **纯净终端模式**：本地 shell 与标签。可从菜单打开，或在设置里选为下次启动方式。
- **原生外观与更新**：System / Light / Dark 主题、持久化偏好，以及支持 Ed25519 校验的应用内更新。

### Herdr 检测与启动

启动时 Hergo 在 PATH 与常见安装目录（`~/.local/bin`、`~/.cargo/bin`、Homebrew，以及 Windows 的 `%LOCALAPPDATA%\Programs\Herdr\bin`）中查找 `herdr`。从 Dock 或 Finder 打开时，也会补上登录 shell 的 PATH。

- **已安装，默认会话已在监听**：直接连接，显示工作区、标签与 agent。
- **已安装，默认会话没有在监听**：后台启动 `herdr server`，等到套接字可连接后再进入。已经在监听时不会再启动第二个。关闭窗口不停止这个服务。
- **未安装**：打开安装页，可以安装官方 Herdr、选择本机已有的可执行文件、查看安装说明，或改用纯净终端。
- 指定了 `-socket`、`HERDR_SOCKET_PATH` 或具名 `HERDR_SESSION` 时，只连接该会话，不启动服务。

**Settings → Herdr → Startup** 选择 Automatic（默认）、Herdr 或 Plain terminal。Automatic 与 Herdr 都打开 Herdr 窗口；未安装时停在安装页。Plain terminal 以及命令行 `-plain` 本次以纯净终端打开，`-connect` 本次强制打开 Herdr 窗口。纯净终端快捷键：macOS 为 `Cmd+T` / `Cmd+W` / `Cmd+Shift+[ ]` / `Cmd+1…9`，其他平台为 `Ctrl+Shift+` 同名键，`Ctrl+Tab` 切换标签；单独的 Ctrl 组合键始终交给 shell。Herdr 模式下也可从 **Hergo → New Plain Terminal Window** 打开纯净终端窗口。

Hergo 不携带 Herdr 可执行文件。它使用本机安装的 Herdr，并在默认会话未运行时启动它。普通浏览保持窗口本地选择；需要改变会话的操作按 Herdr 语义执行。关闭窗口只释放本应用的连接，不关闭 Herdr 会话，也不停止 Herdr 服务，也不接管已有终端 owner。

通过 **Pane → Agent and Process Details** 查看所选终端详情；**Pane → Notify When Output Matches** 设置一次输出提醒。提醒匹配最近 80 行，已有匹配会立即触发；Details 内可选 Rust 正则与 1–1440 分钟超时，默认 30 分钟。提醒可取消，连接或终端身份变化时结束，不自动重复。

**Settings → Session restore** 修改 Herdr 的 `config.toml`，保留其他配置与注释，仅在下次 Herdr 启动时生效。保存终端历史可能包含令牌或密码。**Settings → Machines** 的远程入口使用 Herdr 交互式客户端，不会把远程工作区投射进原生侧栏；远程连接和 Windows 命令终端尚未实机验证。

## 安装

从 [GitHub Releases](https://github.com/Maidang1/hergo/releases) 安装 Hergo 后即可打开。尚未安装 [Herdr](https://github.com/herdrdev/herdr) 时，窗口里的 **Install Herdr** 会运行官方安装脚本（Unix 为 [install.sh](https://herdr.dev/install.sh)，Windows 为 [install.ps1](https://herdr.dev/install.ps1)）。也可以选择本机已有的 `herdr`。Windows 和 Linux 的 CI 集成测试使用 Herdr 0.9.5；API 接入以实际运行的 Herdr 版本为准。

| 平台 | 安装包 | 说明 |
| --- | --- | --- |
| macOS | 通用 `.dmg` | 同时包含 Apple Silicon 和 Intel 架构 |
| Windows x64 | `-setup.exe` 或便携 `.zip` | 原生命名管道与 endpoint 终端单元格渲染 |
| Linux amd64 / arm64 | `.deb` 或便携 `.tar.gz` + `install.sh` | Ubuntu 22.04+ 为验证基线，需要 glibc、GTK 3 和图形会话 |

### macOS

打开 DMG，将 **Hergo.app** 拖入 Applications 后启动。当前应用采用 ad hoc 签名，未经 Apple 公证；首次启动被 Gatekeeper 阻止时，确认安装包来源后执行：

```sh
xattr -dr com.apple.quarantine "/Applications/Hergo.app"
```

### Windows

运行安装器，或将便携 ZIP 完整解压后启动应用。请保留解压目录中的资源文件。

Windows 使用原生 endpoint 渲染，不启动 `herdr terminal attach`。快捷键使用 Ctrl+Shift 等组合，避免占用常用终端按键。安装、构建和已知限制见 [Windows 支持说明](docs/windows-support.md)。

### Linux

下载与 CPU 架构匹配的安装包：`amd64` 对应 x86_64，`arm64` 对应 aarch64。

- **Debian / Ubuntu**：使用 `sudo apt install ./hergo_<version>_<arch>.deb` 安装，再运行 `hergo`。
- **用户级安装**：从同一 release 下载 `install.sh` 与对应便携归档，以普通用户执行 `sh install.sh ./hergo-<version>-linux-<arch>.tar.gz`。
- **便携运行**：完整解压归档后运行其中的 `./hergo`，保留终端动态库与其他资源。

用户级安装需要 `~/.local/bin` 位于 PATH，并支持可写安装目录中的应用内更新。系统级 `.deb` 安装请使用新版 `.deb` 更新，不要以 root 运行 GUI。

安装前建议使用同一 release 的 SHA-256 文件校验下载。GTK 依赖、完整安装命令和验证范围见 [Linux 支持说明](docs/linux-support.md)。

## 使用

启动后选择工作区与标签，即可连接当前标签的可见终端。默认 prefix 是 `Ctrl+B`，再按动作键；`?` 打开快捷键参考。自定义配置和平台差异见 [快捷键说明](docs/keyboard-parity.md)。

常用 macOS 快捷键：

| 操作 | 快捷键 |
| --- | --- |
| 命令与会话导航面板 | `Cmd+P` |
| 新建 / 关闭标签 | `Cmd+T` / `Cmd+W` |
| 垂直 / 水平分屏 | `Cmd+D` / `Cmd+Shift+D` |
| 查找终端文字 | `Cmd+F` |
| 显示 / 隐藏侧栏 | `Cmd+B` |
| 跳到等待中的 agent | `Cmd+J` |
| 设置 | `Cmd+,` |
| 放大 / 缩小 / 重置字号 | `Cmd+=` / `Cmd+-` / `Cmd+0` |
| 行首 / 行尾 / 删到行首 | `Cmd+←` / `Cmd+→` / `Cmd+⌫` |
| 按词左移 / 右移 | `Option+←` / `Option+→` |

终端已有 owner 时，应用保留诊断信息并提供 Retry，不自动 takeover。关闭窗口只释放本应用的连接，不关闭 Herdr 会话。

### 查看代码改动

在 Git 仓库中，工作区工具栏右侧是 Git 图标。不是 Git 仓库时不显示。点击图标，或选择 **View → Show Changes**。左侧上方是未提交改动和最近 100 条提交，点一条提交查看它相对第一个父提交的改动；左侧下方是文件列表。右侧按文件连续展示只读 diff，直接滚动即可查看所有改动。**Unified / Split** 切换单栏与左右对照，**Refresh** 重新读取磁盘上的最新改动。已暂存和未暂存改动合并显示，未跟踪文件按新增内容预览，忽略的文件不列入。

在 diff 中使用查找快捷键搜索当前文件。再次点击 Git 图标、按 Esc 或关闭按钮可返回终端；终端会话继续运行。显示状态、选中的提交、文件和对照方式按工作区记忆：切到其他工作区时显示该工作区自己的视图，切回来恢复原来的位置并重新读取改动。文件按可见范围加载；二进制或非 UTF-8 文件显示提示；超过 1 MiB 的文件在可用时显示有界 patch。只读，不提供暂存、撤销或写回文件。

### 配色

Hergo 默认自动跟随 Herdr 的 `config.toml` 配色，无需在设置中选择。启动时读取配置，运行期间每秒检测一次文件内容变化，自动同步窗口与终端的默认背景、文字和选区。已有偏好文件没有 `color_scheme` 时也使用此默认行为；明确保存的配色选择继续保留。

如需单独设置 Hergo，在 **Settings → Appearance** 关闭 **Follow Herdr** 开关，再在下方 **Theme** 中直接选择 System、Hergo Light、Hergo Dark、Herdr 内置主题或本地新增的 **Monokai Pro**（`monokai-pro`）和 **Monokai Pro Light**（`monokai-pro-light`）。这两个主题分别固定使用上游的暗色和亮色配色。System 根据系统外观切换 Hergo Light 与 Hergo Dark。重新打开开关即可恢复自动跟随；手动主题在跟随期间不可修改。

跟随配置时支持 `[theme]` 的 `name`、`auto_switch`、`dark_name`、`light_name`，以及 `[theme.custom]` 和亮暗覆盖。例如：

```toml
[theme]
name = "gruvbox"
auto_switch = true
dark_name = "gruvbox"
light_name = "gruvbox-light"
```

跟随配置时，`name = "monokai-pro"`（或 `"monokai-pro-light"`）配合 `auto_switch = true`，Hergo 会按系统外观在这两个本地配色之间切换；Herdr 本身是否支持这些名称取决于其版本。

Hergo Light 和 Hergo Dark 分别固定使用 Hergo 自带的浅色与深色；其他固定名称的主题保持自身的亮暗风格。修改 Herdr 配置后自动应用；配置暂时无法读取或解析时保留当前配色，下次修改后重试。Herdr 内置色值来自固定版本的 Herdr 源码，本地主题随 Hergo 一同提供。Monokai Pro 的亮暗色值、选区混色和终端 16 色 ANSI 调色板直接取自 `loctvl842/monokai-pro.nvim` 的 `pro.lua`、`light.lua` 和主题映射；其他主题的 ANSI 调色板继续由终端插件提供。

### 自动更新

正式打包且安装位置可写的应用在启动 10 秒后检查更新，运行期间每天再检查一次。可从 **Check for Updates** 菜单手动检查，在设置的 Updates 页控制自动检查与自动下载安装。

下载的更新经过 Ed25519 签名校验。开发构建和不可写的安装位置不会后台检查；`v0.0.6` 及更早版本需要手动安装一次新版才能使用应用内更新。

## 开发

需要 `go.mod` 指定的 Go 版本（当前为 1.27.1）、Python 3.10+，以及运行原生窗口所需的平台环境。本机已安装 Herdr 且没有用 `-socket` 指定会话时，Hergo 会在默认服务未运行时启动它。集成测试传入 `-socket`，因此不会启动用户自己的服务。

```sh
git clone https://github.com/Maidang1/hergo.git
cd hergo
make test
make build
make run
```

请使用 `make` 或 `scripts/go.sh`，不要直接用普通 `go build` 替代。包装脚本为固定版本的 mygo 生成带有必要终端补丁的工作区，不修改 Go module cache，也不依赖开发者本机的 `replace` 路径。

| 命令 | 用途 |
| --- | --- |
| `make test` | Python 发布契约测试、Go 测试和 vet |
| `make build` | 构建当前平台可执行文件，输出到 `bin/` |
| `make run` | 构建并打开 Hergo |
| `make probe` | 只读检查 Herdr、配置与布局，不占用终端 |
| `make window` | 本地 `/bin/cat` 输入测试窗口，不连接 Herdr |
| `make integration` | 临时隔离服务、真实终端插件与原生窗口验证 |
| `make package` | 打包 macOS `.app` |
| `make dmg` | 打包 macOS 通用 `.app` 与 `.dmg` |
| `make linux` | 打包 Linux 双架构 `.deb`、便携归档与安装脚本 |

Windows 使用 `python scripts/go.py` 包装器；具体命令见平台文档。开发期首次启动终端可能下载经过校验的原生库，正式包已包含动态库，不在启动时下载。集成测试使用独立临时服务，不关闭已有用户会话；Linux 和 Windows 的测试环境要求见各自平台文档。

指定 Herdr 可执行文件或 socket：

```sh
./bin/hergo -herdr /path/to/herdr
./bin/hergo -socket /path/to/herdr.sock
```

### 图标

普通构建不需要 npm。重新生成嵌入图标时执行：

```sh
npm ci
npm run icons
```

来源清单位于 `scripts/icons.json` 和 `internal/icons/SOURCES.txt`。应用图标将 Herdr 的公羊、卷角与终端提示符，以及 mygo 的五色感叹号融合成圆润的小羊造型，设计来源见 [应用图标说明](assets/app/ICON-NOTICE.md)。源文件为 `assets/app/icon.svg`；修改后可用 `resvg -w 1024 -h 1024 assets/app/icon.svg assets/app/icon.png` 更新打包用 PNG。

## 发布

1. 更新 `mygo.json` 的 `version`。
2. 在 `CHANGELOG.md` 添加同版本的更新说明，运行 `make test` 并提交。
3. 推送与版本一致的 `vX.Y.Z` 标签。

[Release 工作流](.github/workflows/release.yml)并行测试和打包 macOS、Windows 与 Linux 双架构，所有平台成功后统一创建 GitHub Release。手动触发只生成工作流产物，不发布 release。

仓库 Actions secret `MYGO_UPDATER_PRIVATE_KEY` 提供更新签名私钥，配置文件只保存公钥。缺少密钥或签名更新产物会停止发布。macOS 同时提供通用与兼容旧客户端的架构专用更新清单；私钥不得提交到仓库。

## 文档与验证范围

- [设计与实现](docs/implementation.md)：架构、接口、事件同步与设计决策。
- [功能对齐](docs/feature-parity-plan.md)与[快捷键说明](docs/keyboard-parity.md)：功能范围和平台差异。
- [Windows 支持](docs/windows-support.md)与[Linux 支持](docs/linux-support.md)：安装、构建、测试与已知限制。
- [API schema](reference/herdr-api-schema.json)：机器可读接口参考。
- [侧栏验证](docs/sidebar-validation.md)及 [M0](docs/m0-validation.md)、[M1](docs/m1-validation.md)、[M2](docs/m2-validation.md)、[M3](docs/m3-validation.md)、[M4](docs/m4-validation.md)：阶段验收记录。
- [设计稿](docs.pen/new.pen)：使用 Pencil 打开；`docs.pen/01.pen` 为历史素材。

CI 测试不能替代实机桌面验收。Windows IME、AltGr、DPI 与完整交互，Linux X11/Wayland 与 IME，以及 macOS 的系统 IME、真实鼠标与 VoiceOver 等仍需实机验证。Windows 的 endpoint 查找覆盖已发布屏幕，且不渲染内联图像；具体限制以平台文档为准。

## 许可证

项目代码采用 [MIT License](LICENSE)。第三方依赖、图标及品牌标识保留各自的许可证和权利声明，不因本项目采用 MIT 而改变：

- [设计图标授权](assets/icons/NOTICE.md)
- [Lucide 图标许可证](internal/icons/LICENSE-lucide)
- [Lobe 图标许可证](internal/icons/LICENSE-lobe)

名称迁移说明：应用标识为 `dev.hergo`，源码入口为 `cmd/hergo`，各平台启动命令为 `hergo`。偏好保存到系统配置目录下的 `hergo/prefs.json`；新路径不存在时读取旧版偏好，后续修改写入新路径。GitHub 仓库与更新源为 `Maidang1/hergo`，更新公钥保持不变。Linux 新安装位于 `~/.local/hergo.app`（用户级）或 `/opt/hergo`（系统级）。应用标识变更及跨名称更新仍需各平台实机验证。
