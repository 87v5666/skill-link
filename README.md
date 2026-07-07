# @wdtt/skill-mgr

[English](#english) | [中文](#chinese)

---

## English

**skill-mgr** is a CLI tool for managing skill files across projects. Maintain a central skill repository and link only the skills each project needs — reducing AI context pressure and keeping skills organized.

### Features

- 🖥️ **TUI** — Interactive terminal UI for browsing and selecting skills
- 🔗 **Symlink management** — Link/unlink skills to your project
- 📂 **Custom categories** — Organize skills into your own categories (a skill can belong to multiple categories)
- 📝 **Skill notes** — Add personal notes to any skill
- 🎯 **Per-project management** — Each project gets only the skills it needs, reducing context overhead
- 🏷️ **Flat & hierarchical repos** — Supports both `category/skill/` and `skill/` directory structures

### Installation

#### Via npm (recommended)

```bash
npm install -g @wdtt/skill-mgr
```

#### Via Go

```bash
go install github.com/87v5666/skill-link@latest
```

#### Manual

Download the binary for your platform from [GitHub Releases](https://github.com/87v5666/skill-link/releases), then:

```bash
chmod +x skill-mgr
sudo mv skill-mgr /usr/local/bin/
```

### Quick Start

```bash
# 1. Point to your skill repository
skill-mgr path set ~/.config/opencode/skills-repo

# 2. Browse and select skills
skill-mgr tui

# 3. Or manage from the command line
skill-mgr list              # List linked skills
skill-mgr add react         # Link a skill
skill-mgr remove react      # Unlink a skill
skill-mgr info react        # Show skill details
```

### Skill Repository Structure

The scanner supports two directory layouts:

**Flat** (each directory is a skill):
```
skills-repo/
├── docx/              ← skill directory (has SKILL.md)
│   ├── SKILL.md
│   └── scripts/       ← ignored (no SKILL.md)
├── grill-me/
│   └── SKILL.md
└── pptx/
    ├── SKILL.md
    └── scripts/
```

**Hierarchical** (category → skills):
```
skills-repo/
├── frontend/
│   ├── react-component/
│   │   └── SKILL.md
│   └── css-layout/
│       └── SKILL.md
└── backend/
    ├── api-design/
    │   └── SKILL.md
    └── database-schema/
        └── SKILL.md
```

### CLI Commands

| Command | Description |
|---------|-------------|
| `tui` | Launch the interactive TUI |
| `list` | List linked skills |
| `add <name>` | Link a skill to the current project |
| `remove <name>` | Unlink a skill |
| `info <name>` | Show skill details and SKILL.md |
| `sync` | Refresh the repository cache |
| `path [get\|set <path>]` | View or configure the repository path |

### TUI Key Bindings

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate |
| `Tab` | Switch between category panel and skill list |
| `Space` | Select/deselect a skill |
| `Enter` | Preview skill / confirm action |
| `L` | Link all selected skills |
| `D` | Unlink the skill under cursor |
| `N` | Create a new custom category |
| `A` | Add/remove current skill to/from a category |
| `E` | Edit note for the current skill |
| `/` | Search skills |
| `Q` / `Ctrl+C` | Quit |

### Development

```bash
git clone git@github.com:87v5666/skill-link.git
cd skill-link

# Build
go build -o skill-mgr -ldflags="-X main.version=dev" .

# Test
go test ./...

# Build with version
go build -o skill-mgr -ldflags="-X main.version=v0.2.0 -s -w" .
```

### Release

```bash
# Update version
npm pkg set version=0.3.0

# Tag and push (triggers GitHub Actions)
git commit -am "release: v0.3.0"
git tag v0.3.0 && git push origin v0.3.0
```

GitHub Actions will build binaries for all platforms, create a GitHub Release, and publish to npm.

### License

MIT

---

## Chinese

**skill-mgr** 是 skill 链接管理工具。统一管理 skill 仓库，按需为每个项目添加所需的 skill，减小 AI 上下文压力。

### 功能

- 🖥️ **TUI** — 终端交互界面，浏览和选择 skill
- 🔗 **软链接管理** — 链接/取消链接 skill 到项目
- 📂 **自定义分类** — 创建自己的分类管理 skill（一个 skill 可属于多个分类）
- 📝 **skill 备注** — 为任何 skill 添加个人备注
- 🎯 **按需管理** — 每个项目只链接需要的 skill，减小上下文压力
- 🏷️ **平面和分层结构** — 同时支持 `分类/skill/` 和 `skill/` 两种目录结构

### 安装

#### 通过 npm（推荐）

```bash
npm install -g @wdtt/skill-mgr
```

#### 通过 Go

```bash
go install github.com/87v5666/skill-link@latest
```

#### 手动安装

从 [GitHub Releases](https://github.com/87v5666/skill-link/releases) 下载对应平台的二进制文件：

```bash
chmod +x skill-mgr
sudo mv skill-mgr /usr/local/bin/
```

### 快速开始

```bash
# 1. 指定 skill 仓库路径
skill-mgr path set ~/.config/opencode/skills-repo

# 2. 启动 TUI 浏览选择
skill-mgr tui

# 3. 或者用命令行管理
skill-mgr list              # 列出已链接的 skill
skill-mgr add react         # 链接 skill
skill-mgr remove react      # 取消链接
skill-mgr info react        # 查看 skill 详情
```

### Skill 仓库结构

支持两种目录布局：

**平面结构**（每层目录是一个 skill）：
```
skills-repo/
├── docx/              ← skill 目录（含 SKILL.md）
│   ├── SKILL.md
│   └── scripts/       ← 被忽略（没有 SKILL.md）
├── grill-me/
│   └── SKILL.md
└── pptx/
    ├── SKILL.md
    └── scripts/
```

**分层结构**（分类 → skill）：
```
skills-repo/
├── frontend/
│   ├── react-component/
│   │   └── SKILL.md
│   └── css-layout/
│       └── SKILL.md
└── backend/
    ├── api-design/
    │   └── SKILL.md
    └── database-schema/
        └── SKILL.md
```

### CLI 命令

| 命令 | 功能 |
|------|------|
| `tui` | 启动交互式 TUI |
| `list` | 列出已链接的 skill |
| `add <name>` | 链接 skill 到当前项目 |
| `remove <name>` | 取消链接 skill |
| `info <name>` | 查看 skill 详情和 SKILL.md |
| `sync` | 刷新仓库缓存 |
| `path [get\|set <路径>]` | 查看或配置仓库路径 |

### TUI 快捷键

| 快捷键 | 功能 |
|--------|------|
| `↑↓` / `jk` | 导航 |
| `Tab` | 切换分类面板和 skill 列表 |
| `Space` | 选择/取消选择 skill |
| `Enter` | 预览 skill / 确认操作 |
| `L` | 链接所有选中的 skill |
| `D` | 取消链接当前光标下的 skill |
| `N` | 新建自定义分类 |
| `A` | 将当前 skill 添加/移出分类 |
| `E` | 编辑当前 skill 的备注 |
| `/` | 搜索 skill |
| `Q` / `Ctrl+C` | 退出 |

### 开发

```bash
git clone git@github.com:87v5666/skill-link.git
cd skill-link

# 构建
go build -o skill-mgr -ldflags="-X main.version=dev" .

# 测试
go test ./...

# 带版本号构建
go build -o skill-mgr -ldflags="-X main.version=v0.2.0 -s -w" .
```

### 发布

```bash
# 更新版本号
npm pkg set version=0.3.0

# 打 tag 并推送（触发 GitHub Actions）
git commit -am "release: v0.3.0"
git tag v0.3.0 && git push origin v0.3.0
```

GitHub Actions 会自动构建各平台二进制、创建 GitHub Release 并发布到 npm。

### 许可证

MIT