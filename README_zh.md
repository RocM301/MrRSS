# MrRSS

<a href="https://trendshift.io/repositories/15731" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15731" alt="WCY-dt%2FMrRSS | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

![Screenshot](imgs/og1.png)

<p>
   <a href="README.md">English</a> | <strong>简体中文</strong>
</p>

[![Version](https://img.shields.io/badge/version-1.3.23-blue.svg)](https://github.com/WCY-dt/MrRSS/releases)
[![License](https://img.shields.io/badge/license-GPLv3-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Wails](https://img.shields.io/badge/Wails-v3%20alpha-red)](https://wails.io/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.5+-4FC08D?logo=vue.js)](https://vuejs.org/)

> [!NOTE]
> **本仓库是 [WCY-dt/MrRSS](https://github.com/WCY-dt/MrRSS) 的修改版（fork）**，维护于 [RocM301/MrRSS](https://github.com/RocM301/MrRSS)，沿用相同的 GPL-3.0 协议。
>
> **关于维护者：** XP 并非程序员，只是一名 YouTuber（[@XiaoPengTech](https://www.youtube.com/@XiaoPengTech)），纯属编译好玩。因此请勿期待专业级的支持，issue 和 PR 可能无法及时回复。
>
> **修改内容：**
>
> - ✨ **图片拖出保存**：在文章视图中按住图片可直接拖到 Finder / 文件管理器保存到本地
> - ✨ **一键复制原文链接**：文章工具栏新增按钮，单击复制文章来源网址
> - ✨ **非打扰式自动刷新**：定时刷新不再在阅读时把列表滚动位置重置——新文章会在你切到后台、点击"N 篇新文章"提示条、或滚动回顶部时才更新
> - ✨ **本地翻译（[MTranServer](https://github.com/xxnuo/MTranServer)）**：可接入自建的 MTranServer 作为翻译源，实现快速、无限量、完全离线的英译中
> - 🔧 **AI 请求重试**：AI 中转的瞬时故障（超时、限流、5xx）会自动重试，大幅减少"摘要/翻译失败"报错
> - 🐛 修复每次刷新后文章 ID 变化（`INSERT OR REPLACE` → upsert）导致的"暂无内容"、摘要报错需重启的问题；AI 摘要与标题翻译缓存不再被刷新清空
> - 🐛 修复添加订阅失败时报错显示原始 i18n key / 空信息的问题，现在会显示 HTTP 状态码和真实错误
> - 🐛 文章内容被缓存清理删除时给出友好提示，不再抛出原始 SQL 错误
> - 🐛 修复代理图片下载文件名变成 `proxy.*` 的问题
>
> 完整改动见[提交历史](https://github.com/RocM301/MrRSS/commits/main)。

## 🐳 Docker 一键部署（手机 / 自部署）

在群晖 NAS 或任意装了 Docker 的机器上跑无头服务端，然后用任意浏览器（含手机）阅读订阅。已发布预构建的多架构镜像（amd64/arm64）到 GitHub 镜像仓库：

```bash
docker run -d --name mrrss -p 1234:1234 -v mrrss-data:/app/data --restart unless-stopped ghcr.io/rocm301/mrrss:latest
```

然后浏览器打开 `http://你的服务器IP:1234`。数据存在 `mrrss-data` 卷里，更新容器不丢。想配合本地翻译，自部署 [MTranServer](https://github.com/xxnuo/MTranServer) 后在 设置 → 内容 → 翻译 里填它的地址即可。

## ✨ 功能特性

- 🌐 **自动翻译与摘要**: 自动翻译文章标题与正文，并生成简洁的内容摘要，助你快速获取信息
- 🤖 **AI 增强功能**: 集成先进 AI 技术，赋能翻译、摘要、推荐等多种功能，让阅读更智能
- 🔌 **丰富的插件生态**: 支持 Obsidian、Notion、FreshRSS、RSSHub 等主流工具集成，轻松扩展功能
- 📡 **多样化订阅方式**: 支持 URL、XPath、脚本、Newsletter 等多种订阅源类型，满足不同需求
- 🏭 **自定义脚本与自动化**: 内置过滤器与脚本系统，支持高度自定义的自动化流程

## 🚀 快速开始

### 下载与安装

#### 选项 1: 下载预构建安装包（推荐）

从 [Releases](https://github.com/WCY-dt/MrRSS/releases/latest) 页面下载适合您平台的最新安装包。

<details>

<summary>点击查看可用的安装包列表</summary>

<div markdown="1">

**标准安装版：**

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

**便携版**（无需安装，所有数据在一个文件夹内）：

- **Windows:** `MrRSS-{version}-windows-{arch}-portable.zip`
- **Linux:** `MrRSS-{version}-linux-{arch}-portable.tar.gz`
- **macOS:** `MrRSS-{version}-darwin-{arch}-portable.zip`

</div>

</details>

#### 选项 2: 源码构建

<details>

<summary>点击展开源码构建指南</summary>

<div markdown="1">

##### 前置要求

在开始之前，请确保已安装以下环境：

- [Go](https://go.dev/) (1.25 或更高版本)
- [Node.js](https://nodejs.org/) (20 LTS 或更高版本，带 npm)
- [Wails v3](https://v3alpha.wails.io/getting-started/installation/) CLI

**平台特定要求：**

- **Linux**: GTK3、WebKit2GTK 4.1、libsoup 3.0、GCC、pkg-config
- **Windows**: MinGW-w64（用于 CGO 支持）、NSIS（用于安装包）
- **macOS**: Xcode 命令行工具

详细安装说明请参见[构建要求](docs/BUILD_REQUIREMENTS.md)

```bash
# Linux 快速设置（Ubuntu 24.04+）：
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev libsoup-3.0-dev gcc pkg-config
```

##### 安装步骤

1. **克隆仓库**

   ```bash
   git clone https://github.com/WCY-dt/MrRSS.git
   cd MrRSS
   ```

2. **安装前端依赖**

   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **安装 Wails v3 CLI**

   ```bash
   go install github.com/wailsapp/wails/v3/cmd/wails3@latest
   ```

4. **构建应用**

   ```bash
   # 使用 Task（推荐）
   task build

   # 或使用 Makefile
   make build

   # 或直接使用 wails3
   wails3 build
   ```

   可执行文件将在 `build/bin` 目录下生成。

5. **运行应用**

   - Windows: `build/bin/MrRSS.exe`
   - macOS: `build/bin/MrRSS.app`
   - Linux: `build/bin/MrRSS`

</div>

</details>

### 数据存储

<details>

<summary>点击展开数据存储说明</summary>

<div markdown="1">

**正常模式**（默认）：

- **Windows:** `%APPDATA%\MrRSS\` (例如 `C:\Users\YourName\AppData\Roaming\MrRSS\`)
- **macOS:** `~/Library/Application Support/MrRSS/`
- **Linux:** `~/.local/share/MrRSS/`

**便携模式**（当 `portable.txt` 文件存在时）：

- 所有数据存储在 `data/` 文件夹中

这确保了您的数据在应用更新和重新安装时得以保留。

</div>

</details>

## 🛠️ 开发指南

<details>

<summary>点击展开开发指南</summary>

<div markdown="1">

### 开发模式运行

启动带有热重载的应用：

```bash
# 使用 Wails v3
wails3 dev

# 或使用 Task
task dev
```

### 代码质量工具

#### 使用 Make

我们提供了 `Makefile` 来处理常见的开发任务（在 Linux/macOS/Windows 上都可用）：

```bash
# 显示所有可用命令
make help

# 运行完整检查（lint + 测试 + 构建）
make check

# 清理构建产物
make clean

# 设置开发环境
make setup
```

### Pre-commit Hooks

本项目使用 pre-commit hooks 来确保代码质量：

```bash
# 安装 hooks
pre-commit install

# 在所有文件上运行
pre-commit run --all-files
```

### 运行测试

```bash
make test
```

### 服务器模式（仅限 API）

对于服务器部署和 API 集成，请使用无界面服务器版本：

```bash
# 使用 Docker（推荐）
docker run -p 1234:1234 mrrss-server:latest

# 或从源码构建
go build -tags server -o mrrss-server .
./mrrss-server
```

本 fork 提供了预构建的多架构（amd64/arm64）服务器镜像，发布在 ghcr.io：

```bash
docker run -d --name mrrss -p 1234:1234 -v mrrss-data:/app/data --restart unless-stopped ghcr.io/rocm301/mrrss:latest
```

请参阅[服务器模式 API 文档](docs/SERVER_MODE/swagger.json)以获取完整的 API 参考。

</div>

</details>

## 🤝 贡献

我们欢迎贡献！详情请参阅我们的[贡献指南](CONTRIBUTING.md)。

<details>

<summary>点击展开贡献指南</summary>

<div markdown="1">

在贡献之前：

1. 阅读[行为准则](CODE_OF_CONDUCT.md)
2. 检查现有 issue 或创建一个新 issue
3. Fork 仓库并创建功能分支
4. 进行更改并添加测试
5. 提交 Pull Request

</div>

</details>

## 🔒 安全

如果您发现安全漏洞，请遵循我们的[安全策略](SECURITY.md)。

## 📝 许可证

本项目采用 GPL-3.0 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 📮 联系与支持

- **Issues**: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- **讨论**: [GitHub Discussions](https://github.com/WCY-dt/MrRSS/discussions)
- **仓库**: [github.com/WCY-dt/MrRSS](https://github.com/WCY-dt/MrRSS)

---

<div align="center">
  <img src="imgs/sponsor.png" alt="Sponsor MrRSS"/>
  <p>Made with ❤️ by the MrRSS Team</p>
  <p>⭐ 如果您觉得这个项目有用，请在 GitHub 上给我们点星！</p>
</div>
