"use strict";

(() => {
  const english = {
    skip: "Skip to content",
    navigation: "Main navigation",
    "nav.workflow": "Workflow",
    "nav.native": "Made native",
    "nav.docs": "Docs",
    "nav.menu": "Toggle navigation",
    "hero.badge": "Native desktop. Open source.",
    "hero.title": "Parallel work.<br><span>A clearer head.</span>",
    "hero.description": "Your terminals, Git worktrees, and coding agents.<br>Together in Hergo, a native desktop client for Herdr.",
    "hero.download": "Download Hergo",
    "platform.windows": "Windows (experimental)",
    "demo.caption": "One window. Your whole workspace.",
    "demo.disclaimer": "Interactive preview · Sample session",
    "demo.connected": "Herdr connected",
    "demo.workspaces": "Workspaces",
    "demo.agentsLabel": "Agents",
    "demo.working": "Working",
    "demo.waiting": "Ready for review",
    "demo.local": "Local workspace",
    "demo.theme": "Toggle preview dark theme",
    "demo.prompt": "Add quick workspace switching",
    "demo.response": "I'll check navigation, then add the keybindings.",
    "demo.checking": "Checking the changes",
    "demo.input": "Keep the idea going.",
    "demo.testLabel": "Tests",
    "demo.passed": "All checks passed",
    "demo.notification": "Task complete. Ready for your review.",
    "demo.diffSummary": "Uncommitted changes · 1 file",
    "demo.status": "Two agents. One place to work.",
    "demo.try": "A different view. The same context.",
    "demo.views": "Preview view",
    "demo.focus": "Focus",
    "demo.split": "Split",
    "demo.changes": "Changes",
    "demo.themeHint": "Try the theme toggle, too",
    "demo.viewport": "Horizontally scrollable native interface preview",
    "demo.search": "Filter preview workspaces",
    "demo.empty": "No matching workspaces. Clear the search to see them all.",
    "demo.tabs": "Preview terminal tabs",
    "demo.scrollHint": "Swipe to explore the full window ↔",
    "agents.title": "The agents you know. A place for every one.",
    "workflow.eyebrow": "Keep your context close",
    "workflow.title": "Separate the tasks.<br><span>Keep the context.</span>",
    "workflow.description": "From one branch to a few ideas in parallel.<br>Your workspaces, agents, and changes stay in view.",
    "workflow.preview": "Workflow preview",
    "workflow.controls": "Choose a workflow preview",
    "step.0.title": "A little space for every task.",
    "step.0.body": "Give each task its own Git worktree. Branches, terminals, and local ports stay together, so switching projects doesn't mean starting over.",
    "step.0.tags": "Workspaces / Git worktrees",
    "step.1.title": "One writes. Another checks.",
    "step.1.body": "Split horizontally or vertically to keep agents, tests, and shells side by side. Follow agent status and set a notification for your next move.",
    "step.1.tags": "Split terminals / Agent notifications",
    "step.2.title": "See the changes. Know what's next.",
    "step.2.body": "Browse uncommitted files and recent commits in the same window. Read unified or split diffs with syntax highlighting, then get back to work.",
    "step.2.tags": "Code diffs / Commit history",
    "scene.0.label": "Every branch, making progress.",
    "scene.0.main": "Main workspace",
    "scene.working": "Working",
    "scene.ready": "Review",
    "scene.1.label": "Move in parallel. Step in when needed.",
    "scene.task.0": "Workspace navigation",
    "scene.task.1": "Terminal focus review",
    "scene.2.label": "An idea, down to every line.",
    "scene.tab.0": "Spaces",
    "scene.tab.1": "Parallel",
    "scene.tab.2": "Changes",
    "native.eyebrow": "Considered down to the keys",
    "native.title": "Feeling at home is <span>the whole point.</span>",
    "native.description": "Built with Go and mygo. Native interface and terminal.<br>No WebView or Node.js runtime required.",
    "native.keyboard.label": "Keep your hands on the keys",
    "native.keyboard.title": "Your next move, already in reach.",
    "native.keyboard.body": "Open the command palette, focus a pane, or jump to a waiting agent. The usual moves, without reaching for the mouse.",
    "native.keyboard.note": "macOS shortcuts shown. Other platforms are in the docs.",
    "native.theme.label": "Make it feel like you",
    "native.theme.title": "Your terminal. Your kind of quiet.",
    "native.theme.body": "Follow your Herdr colors, or choose Hergo and Monokai Pro light and dark themes. Fonts and cursors are yours to set, too.",
    "native.plain.title": "Or, just a good terminal.",
    "native.plain.body": "Plain terminal mode opens a local shell with multiple tabs, without connecting to Herdr.",
    "download.eyebrow": "A little room for your next idea",
    "download.title": "Your workspace. <span>Ready when you are.</span>",
    "download.description": "Free and open source. Choose your system and get started.",
    "download.platforms": "Download platform",
    "download.guide": "Installation guide",
    "download.license": "MIT licensed · Free to use",
    "download.all": "All packages & SHA-256 checksums ↗",
    "faq.eyebrow": "A few more things",
    "faq.title": "Before you get going.",
    "faq.docs": "Read the project docs",
    "faq.0.q": "Do I need to install Herdr first?",
    "faq.0.a": "Hergo uses Herdr running on your machine and doesn't bundle its executable. On first launch, it finds your installation or helps you install Herdr or choose an existing executable. For a local shell, go straight to plain terminal mode.",
    "faq.1.q": "Will closing Hergo stop my work?",
    "faq.1.a": "Closing Hergo doesn't stop the Herdr service or close its sessions. Reopen the app to reconnect. If a terminal already has an owner, Hergo offers diagnostics and a retry action instead of taking over automatically.",
    "faq.2.q": "Which systems are supported?",
    "faq.2.a": "macOS on Apple Silicon and Intel, and Linux amd64 / arm64. Windows x64 support is experimental. Ubuntu 22.04+ is the Linux validation baseline; glibc, GTK 3, and a graphical session are required. See the project docs for full requirements and known limitations.",
    "faq.3.q": "Can I contribute?",
    "faq.3.a": "Please do. Hergo is MIT licensed, with source code and build instructions on GitHub. Report an issue, share your experience, or contribute code.",
    "faq.source": "Explore the source ↗",
    "community.title": "Good tools are built together.",
    "community.description": "Built in the open. Made for your everyday.",
    "community.link": "Contribute to Hergo",
    "footer.credit": "A native desktop client for Herdr.",
    "footer.navigation": "Footer",
    "footer.changelog": "Changelog",
    "footer.issues": "Report an issue",
  };
  const platformCopy = {
    macos: {
      architecture: "Apple Silicon + Intel",
      guide: "https://github.com/Maidang1/hergo#macos",
      zh: ["下载 macOS 版", "通用 DMG 安装包。打开后，将 Hergo 拖入 Applications。", "当前版本未经 Apple 公证。首次打开若被阻止，请按安装说明操作。"],
      en: ["Hergo for macOS", "One universal DMG. Open it and drag Hergo into Applications.", "The app is not currently notarized by Apple. If your first launch is blocked, follow the installation guide."],
    },
    windows: {
      architecture: "Windows / x64",
      guide: "https://github.com/Maidang1/hergo/blob/main/docs/windows-support.md",
      zh: ["下载 Windows 版（实验性）", "选择 .exe 安装器或便携 ZIP。解压后，请保留目录中的全部资源文件。", "Windows x64 为实验性支持，原生终端不支持内联图像。平台要求见安装说明。"],
      en: ["Hergo for Windows (experimental)", "Choose the .exe installer or portable ZIP. Keep all resource files together after extracting.", "Windows x64 support is experimental. The native terminal does not support inline images. See the guide for platform requirements."],
    },
    linux: {
      architecture: "Linux / amd64 + arm64",
      guide: "https://github.com/Maidang1/hergo/blob/main/docs/linux-support.md",
      zh: ["下载 Linux 版", "提供 amd64 和 arm64 的 .deb 与便携 .tar.gz，以及用户级安装脚本。", "以 Ubuntu 22.04+ 为验证基线，需要 glibc、GTK 3 和图形会话。"],
      en: ["Hergo for Linux", "Available as .deb and portable .tar.gz packages for amd64 and arm64, with a user-level installer.", "Ubuntu 22.04+ is the validation baseline. Requires glibc, GTK 3, and a graphical session."],
    },
  };
  const translatedElements = [...document.querySelectorAll("[data-i18n]")];
  const translatedAttributes = [...document.querySelectorAll("[data-i18n-aria]")];
  const chinese = new Map(translatedElements.map((element) => [element.dataset.i18n, element.innerHTML]));
  const chineseAttributes = new Map(translatedAttributes.map((element) => [element.dataset.i18nAria, element.getAttribute("aria-label")]));
  const product = document.getElementById("app-window");
  const navigation = document.getElementById("main-navigation");
  const menuToggle = document.getElementById("menu-toggle");
  const themeToggle = document.getElementById("theme-toggle");
  const status = document.getElementById("interaction-status");
  const reducedMotion = matchMedia("(prefers-reduced-motion: reduce)");
  const desktop = matchMedia("(min-width: 681px)");
  let language = "zh";
  let platform = "macos";
  let scene = 0;
  let queuedFrame = false;

  function announce(text) {
    status.textContent = text;
  }

  function updatePlatform() {
    const selected = platformCopy[platform];
    const copy = selected[language];
    document.getElementById("platform-architecture").textContent = selected.architecture;
    document.getElementById("platform-title").textContent = copy[0];
    document.getElementById("platform-description").textContent = copy[1];
    document.getElementById("platform-note").textContent = copy[2];
    document.getElementById("platform-button").textContent = language === "zh" ? "前往 GitHub 下载" : "Download on GitHub";
    document.getElementById("platform-guide").href = selected.guide;
    document.querySelectorAll("[data-platform]").forEach((button) => {
      button.setAttribute("aria-pressed", String(button.dataset.platform === platform));
    });
  }

  function setLanguage(next) {
    language = next === "en" ? "en" : "zh";
    document.documentElement.lang = language === "en" ? "en" : "zh-CN";
    translatedElements.forEach((element) => {
      const messageId = element.dataset.i18n;
      // Translation markup is trusted, bundled content, never user input.
      element.innerHTML = language === "zh" ? chinese.get(messageId) : english[messageId];
    });
    translatedAttributes.forEach((element) => {
      const messageId = element.dataset.i18nAria;
      element.setAttribute("aria-label", language === "zh" ? chineseAttributes.get(messageId) : english[messageId]);
    });
    const toggle = document.getElementById("language-toggle");
    toggle.innerHTML = language === "zh"
      ? '<span>EN</span><span class="language-divider" aria-hidden="true">/</span><span class="language-current" lang="zh-CN">中</span>'
      : '<span lang="zh-CN">中</span><span class="language-divider" aria-hidden="true">/</span><span class="language-current">EN</span>';
    toggle.setAttribute("aria-label", language === "zh" ? "Switch to English" : "切换到中文");
    document.title = language === "zh" ? "Hergo · 并行开发，思路清晰。" : "Hergo · Parallel work. A clearer head.";
    const description = language === "zh"
      ? "Hergo 是 Herdr 的原生桌面客户端，将终端、Git 工作树和 AI 编程助手集中在一个窗口。支持 macOS 和 Linux，Windows 为实验性支持。"
      : "Hergo brings your terminals, Git worktrees, and coding agents together in one native desktop client for Herdr. For macOS and Linux, with experimental Windows support.";
    document.querySelector('meta[name="description"]').content = description;
    document.querySelector('meta[property="og:title"]').content = document.title;
    document.querySelector('meta[property="og:description"]').content = description;
    updatePlatform();
    try { localStorage.setItem("hergo-language", language); } catch { /* Browser language remains a working fallback. */ }
  }

  function setMenu(open, restoreFocus = false) {
    navigation.dataset.open = String(open);
    menuToggle.setAttribute("aria-expanded", String(open));
    menuToggle.querySelector("use").setAttribute("href", open ? "#i-close" : "#i-menu");
    if (open && !desktop.matches) navigation.querySelector("a").focus();
    if (restoreFocus) menuToggle.focus();
  }
  menuToggle.addEventListener("click", () => setMenu(menuToggle.getAttribute("aria-expanded") !== "true"));
  navigation.addEventListener("click", (event) => {
    if (event.target.closest("a")) setMenu(false);
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && menuToggle.getAttribute("aria-expanded") === "true") setMenu(false, true);
  });
  document.addEventListener("click", (event) => {
    if (!event.target.closest(".header-shell")) setMenu(false);
  });
  desktop.addEventListener("change", () => setMenu(false));

  document.getElementById("language-toggle").addEventListener("click", () => setLanguage(language === "zh" ? "en" : "zh"));
  document.querySelectorAll("[data-platform]").forEach((button) => {
    button.addEventListener("click", () => {
      platform = button.dataset.platform;
      updatePlatform();
      announce(document.getElementById("platform-title").textContent);
    });
  });
  themeToggle.addEventListener("click", () => {
    const dark = product.dataset.theme !== "dark";
    product.dataset.theme = dark ? "dark" : "light";
    themeToggle.setAttribute("aria-pressed", String(dark));
    announce(language === "zh" ? `演示已切换为${dark ? "深色" : "浅色"}配色` : `Preview switched to ${dark ? "dark" : "light"} theme`);
  });

  function setDemoView(mode, tab = mode === "diff" ? "3" : "1") {
    const shellOnly = mode === "workspace" && tab === "2";
    product.dataset.mode = mode;
    product.querySelector(".native-primary").hidden = mode === "diff" || shellOnly;
    product.querySelector(".native-diff").hidden = mode !== "diff";
    product.querySelector(".native-shell").hidden = mode !== "split" && !shellOnly;
    document.querySelectorAll("button[data-mode]").forEach((item) => item.setAttribute("aria-pressed", String(item.dataset.mode === mode)));
    document.querySelectorAll("[data-terminal-tab]").forEach((item) => item.setAttribute("aria-pressed", String(item.dataset.terminalTab === tab)));
    product.querySelector(".native-pane-caption").innerHTML = mode === "diff"
      ? '<svg class="icon small" aria-hidden="true"><use href="#i-diff"/></svg>Changes'
      : shellOnly
        ? '<svg class="icon small" aria-hidden="true"><use href="#i-terminal"/></svg>terminal'
        : '<img src="./assets/droid.svg" width="13" height="13" alt="">droid';
  }
  document.querySelectorAll("button[data-mode]").forEach((button) => {
    button.addEventListener("click", () => {
      setDemoView(button.dataset.mode);
      announce(language === "zh" ? `演示视图：${button.textContent.trim()}` : `Preview: ${button.textContent.trim()}`);
    });
  });
  const branches = { hergo: "main", website: "feat/landing", api: "fix/auth", docs: "main", cli: "feat/completions", examples: "main" };
  document.querySelectorAll("[data-workspace]").forEach((button) => {
    button.addEventListener("click", () => {
      const project = button.dataset.workspace;
      document.querySelectorAll("[data-workspace]").forEach((item) => item.setAttribute("aria-pressed", String(item === button)));
      document.getElementById("demo-project").textContent = project;
      document.querySelectorAll(".demo-path").forEach((element) => { element.textContent = `~/code/${project}`; });
      document.querySelectorAll(".demo-branch").forEach((element) => { element.textContent = branches[project]; });
      announce(language === "zh" ? `演示工作区：${project}` : `Preview workspace: ${project}`);
    });
  });

  document.querySelectorAll("[data-terminal-tab]").forEach((button) => {
    button.addEventListener("click", () => {
      setDemoView(button.dataset.terminalTab === "3" ? "diff" : "workspace", button.dataset.terminalTab);
      announce(language === "zh" ? `演示标签：${button.dataset.terminalTab}` : `Preview tab: ${button.dataset.terminalTab}`);
    });
  });
  document.getElementById("demo-filter").addEventListener("input", (event) => {
    const query = event.target.value.trim().toLowerCase();
    let count = 0;
    document.querySelectorAll("[data-workspace]").forEach((button) => {
      const matches = button.textContent.toLowerCase().includes(query);
      button.hidden = !matches;
      if (matches) count++;
    });
    document.getElementById("space-count").textContent = count;
    document.getElementById("demo-empty").hidden = count !== 0;
    announce(language === "zh" ? `可见工作区：${count} / 6` : `Visible workspaces: ${count} / 6`);
  });

  function setScene(index, fromControl = false) {
    if (index < 0 || index > 2) return;
    const changed = scene !== index;
    scene = index;
    document.querySelectorAll(".workflow-scene").forEach((panel) => {
      const active = Number(panel.dataset.scene) === scene;
      panel.hidden = !active;
      panel.classList.toggle("scene-enter", active && changed && !fromControl);
    });
    document.querySelectorAll("[data-step]").forEach((step) => step.classList.toggle("active", Number(step.dataset.step) === scene));
    document.querySelectorAll("[data-story]").forEach((button) => button.setAttribute("aria-pressed", String(Number(button.dataset.story) === scene)));
    document.querySelector(".preview-count").textContent = `${String(scene + 1).padStart(2, "0")} / 03`;
    if (fromControl) {
      const label = document.querySelector(`[data-story="${scene}"]`).textContent.trim();
      announce(language === "zh" ? `工作流程：${label}` : `Workflow preview: ${label}`);
    }
  }
  document.querySelectorAll("[data-story]").forEach((button) => {
    button.addEventListener("click", () => setScene(Number(button.dataset.story), true));
  });
  const steps = [...document.querySelectorAll("[data-step]")];
  function updateScrollScene() {
    queuedFrame = false;
    if (!desktop.matches || reducedMotion.matches || document.activeElement?.closest(".preview-controls")) return;
    const bounds = document.querySelector(".workflow-layout").getBoundingClientRect();
    const target = innerHeight * .48;
    if (bounds.top > target || bounds.bottom < target) return;
    let closest = 0;
    let distance = Infinity;
    steps.forEach((step, index) => {
      const rect = step.getBoundingClientRect();
      const nextDistance = Math.abs(rect.top + rect.height / 2 - target);
      if (nextDistance < distance) {
        distance = nextDistance;
        closest = index;
      }
    });
    if (closest !== scene) setScene(closest);
  }
  function requestScrollScene() {
    if (!queuedFrame) {
      queuedFrame = true;
      requestAnimationFrame(updateScrollScene);
    }
  }
  window.addEventListener("scroll", requestScrollScene, { passive: true });
  window.addEventListener("resize", requestScrollScene, { passive: true });
  reducedMotion.addEventListener("change", requestScrollScene);

  if ("IntersectionObserver" in window) {
    const reveals = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("reveal-enter");
          reveals.unobserve(entry.target);
        }
      });
    }, { threshold: .12 });
    document.querySelectorAll(".reveal").forEach((element) => reveals.observe(element));
  }
  const userAgent = navigator.userAgent.toLowerCase();
  if (userAgent.includes("win")) platform = "windows";
  else if (userAgent.includes("linux") && !userAgent.includes("android")) platform = "linux";
  let preferred = navigator.language.toLowerCase().startsWith("zh") ? "zh" : "en";
  try {
    const saved = localStorage.getItem("hergo-language");
    if (saved === "en" || saved === "zh") preferred = saved;
  } catch { /* Storage is optional. */ }
  const requested = new URLSearchParams(location.search).get("lang");
  if (requested === "en" || requested === "zh") preferred = requested;
  setLanguage(preferred);
  setScene(0);
})();
