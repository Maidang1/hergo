# Hergo website

独立的中英双语产品官网。原生 HTML、CSS 和 JavaScript，无构建步骤、npm 依赖或外部字体。页面根据浏览器语言显示中文或英文，保存用户的语言选择；`?lang=zh` / `?lang=en` 可指定初始语言。

## 本地预览

在仓库根目录执行：

```sh
python3 -m http.server 4173 --bind 127.0.0.1 --directory website
```

打开 <http://127.0.0.1:4173/?lang=zh> 或 <http://127.0.0.1:4173/?lang=en>。

## 静态部署

将此目录的内容作为网站根目录即可。所有资源使用相对路径，支持 GitHub Pages 项目子路径，也可直接托管在其他静态服务。

通过 `.github/workflows/pages.yml` 部署到 GitHub Pages。工作流会检查 JavaScript 语法，上传 `website` 目录并部署。推送到 `main` 的官网或部署配置变更会自动触发，也可在 Actions 中手动运行 **Deploy website**。

仓库 **Settings → Pages → Source** 使用 **GitHub Actions**。默认地址为 <https://maidang1.github.io/hergo/>。

GitHub Pages 会自动使用 `404.html`；其他静态服务器需要配置相应错误页。使用自定义域名和子目录组合部署时，请按实际挂载路径调整 `404.html` 的首页链接。

## 内容与设计

- `index.html`：默认中文内容、语义化结构与无需 JavaScript 的下载入口。
- `app.js`：英文文案、平台安装说明、语言切换、移动导航、配色切换、工作区切换、布局演示及滚动叙事。
- `style.css`：暖白与橙色、统一字号和间距、响应式布局、减少动态效果及高对比模式支持。
- `native-preview.css`：依据原生应用布局与配色源码实现的紧凑窗口，默认深色单终端，可切换分屏、Diff 和浅色。
- `assets/`：随站点分发的项目图标、Agent 品牌素材及相关说明和许可证。

演示区是静态前端交互示意，不连接 Herdr、不执行终端命令，也不是实机截图。展示的终端输出是示例内容。下载链接打开 GitHub 最新 Release，用户在那里选择实际平台安装包，不猜测版本化资源地址。

发布新版本时，同步修改 `index.html` 首屏的版本号，依据仓库 `README.md` 校对功能与安装限制。新增或调整文案时同步维护 HTML 中的中文和 `app.js` 中的英文。

正文字体使用系统无衬线字体，英文首屏以本机 Georgia 斜体强调第二行。没有外部字体请求。页面不会加载分析脚本、设置 Cookie 或发送表单；只在本地存储语言偏好。GitHub 托管和下载服务本身的隐私政策由 GitHub 提供。

样式与脚本带缓存版本参数（样式当前为 `?v=3`，脚本为 `?v=4`）。修改资源并发布时，同步递增 HTML 中对应的版本参数，避免已访问网站的浏览器继续使用旧资源。

应用预览采用 `internal/design/palette.go` 的原生窗口色值，以及 `internal/ui/sidebar.go` 和 `internal/ui/tabbar.go` 的侧栏与标签结构。界面标签保留原生应用的英文，终端内容使用不含用户信息的示例。窄屏保留窗口布局，支持横向滑动和键盘滚动；外层说明与交互控件正常适配窄屏。

桌面工作流程预览随滚动切换，也可使用下方按钮选择。控件有键盘焦点时保留用户选择；减少动态效果模式和窄屏只手动切换，不自动更改内容。所有正文默认可见，不依赖入场动画才能阅读。

应用图标来源与设计说明见 `assets/ICON-NOTICE.md`，Agent 图标保留各自品牌权利，相关开源许可证见 `assets/LICENSE-lobe`。
