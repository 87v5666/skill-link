# slink — Skill 链接管理工具 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建 `skill-mgr` CLI 二进制工具 + OpenCode `/slink` 指令，支持从中央仓库浏览、选择、链接 skill 到项目的 `.opencode/skills/`。

**Architecture:** Go 单二进制，核心三层：仓库扫描(cache)、链接管理(symlink)、TUI/CLI/JSON 三种输出模式。通过 Cobra 管理子命令，Bubble Tea 驱动 TUI。

**Tech Stack:** Go 1.26, github.com/spf13/cobra, github.com/charmbracelet/bubbletea, github.com/charmbracelet/lipgloss

## Global Constraints

- 中央仓库默认路径：`~/.config/opencode/skills-repo/`
- 配置文件目录：`~/.config/slink/`
- 软链接目标：项目根目录下的 `.opencode/skills/`
- 项目根目录检测：从当前目录向上查找 `.opencode/` 目录
- 输出编码：所有文本输出使用 UTF-8
- 错误处理：CLI 命令返回非零 exit code，JSON 模式返回 `{"error": "..."}`

## 文件结构

```
skill-management/
├── main.go                        # 入口 + Cobra 根命令
├── go.mod
├── internal/
│   ├── config/
│   │   └── config.go              # 配置加载/保存
│   ├── repo/
│   │   ├── types.go               # Skill, Category 类型定义
│   │   └── scanner.go             # 仓库扫描 + 缓存
│   ├── linker/
│   │   └── linker.go              # 软链接创建/删除/列表
│   └── tui/
│       ├── model.go               # Bubble Tea 主模型
│       ├── browse.go              # 浏览视图（左右分栏）
│       └── preview.go             # 预览视图（SKILL.md）
├── .opencode/
│   └── agents/
│       └── slink.json             # OpenCode agent 注册
└── docs/superpowers/
    ├── specs/2026-07-06-slink-design.md
    └── plans/2026-07-06-slink-implementation.md
```

---

### Task 1: 项目框架 + 核心类型 + 配置管理

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `internal/config/config.go`
- Create: `internal/repo/types.go`
- Create: `internal/config/config_test.go`

**Interfaces:**
- Consumes: (无，第一个任务)
- Produces: `Config` 结构体 + `Load()/Save()` + `DefaultRepoPath()/ConfigDir()` / `Skill`/`Category` 类型

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd /home/zxf/projects/skill-management && go mod init skill-management
```

- [ ] **Step 2: 添加依赖**

```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
```

- [ ] **Step 3: 编写内部类型定义** `internal/repo/types.go`

```go
package repo

// Skill 代表仓库中的一个可链接 skill
type Skill struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Linked      bool   `json:"linked"`
}

// Category 代表一个分类
type Category struct {
	Name   string  `json:"name"`
	Skills []Skill `json:"skills"`
}
```

- [ ] **Step 4: 编写配置管理** `internal/config/config.go`

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	RepoPath string `json:"repo_path"`
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "slink")
}

func DefaultRepoPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode", "skills-repo")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Load() (*Config, error) {
	cfg := &Config{
		RepoPath: DefaultRepoPath(),
	}
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(configPath(), data, 0644); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	return nil
}

// FindProjectRoot 从当前目录向上查找包含 .opencode/ 的项目根目录
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 .opencode/ 目录（不在 OpenCode 项目中）")
		}
		dir = parent
	}
}

// ProjectSkillsDir 返回项目的 skill 链接目录
func ProjectSkillsDir() (string, error) {
	root, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".opencode", "skills"), nil
}
```

- [ ] **Step 5: 编写配置测试** `internal/config/config_test.go`

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// 确保没有配置文件干扰
	os.Unsetenv("HOME")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 不应返回错误，got: %v", err)
	}
	if cfg.RepoPath == "" {
		t.Fatal("默认 RepoPath 不应为空")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	// 覆盖 ConfigDir
	origConfigDir := ConfigDir
	ConfigDir = func() string { return tmpDir }
	defer func() { ConfigDir = origConfigDir }()

	cfg := &Config{RepoPath: "/tmp/test-repo"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() 失败: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}
	if loaded.RepoPath != "/tmp/test-repo" {
		t.Fatalf("期望 RepoPath=/tmp/test-repo, 实际: %s", loaded.RepoPath)
	}
}

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	subDir := filepath.Join(projectDir, "src", "components")

	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	origWd := FindProjectRoot
	FindProjectRoot = func() (string, error) {
		return projectDir, nil
	}
	defer func() { FindProjectRoot = origWd }()

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() 失败: %v", err)
	}
	if root != projectDir {
		t.Fatalf("期望 root=%s, 实际: %s", projectDir, root)
	}
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	origWd := os.Getwd
	os.Getwd = func() (string, error) { return tmpDir, nil }
	defer func() { os.Getwd = origWd }()

	_, err := FindProjectRoot()
	if err == nil {
		t.Fatal("期望错误，但返回了 nil")
	}
}
```

- [ ] **Step 6: 运行测试确认通过**

```bash
go test ./internal/config/ -v
```

Expected: PASS

- [ ] **Step 7: 编写 main.go 入口 + Cobra 根命令骨架**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skill-mgr",
	Short: "管理 OpenCode skill 链接的工具",
	Long: `skill-mgr — 从中央仓库浏览、选择、链接 skill 到当前项目的 .opencode/skills/

首次使用前请先配置仓库路径:
  skill-mgr path set ~/my-skills-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 无参数时打印帮助
		return cmd.Help()
	},
}

func init() {
	// 子命令将在后续任务中注册
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

- [ ] **Step 8: 编译验证**

```bash
go build -o skill-mgr . && ./skill-mgr --help
```

Expected: 显示帮助信息

- [ ] **Step 9: 提交**

```bash
git add -A && git commit -m "feat: project scaffold with config and types"
```

---

### Task 2: 仓库扫描器 + 缓存

**Files:**
- Create: `internal/repo/scanner.go`
- Create: `internal/repo/scanner_test.go`

**Interfaces:**
- Consumes: `Config.RepoPath`, `Skill`, `Category`
- Produces: `NewScanner(repoPath) *Scanner`, `Scan() ([]Category, error)`, `GetSkill(name) (*Skill, error)`, `BuildCache() error`, `CachedSkills() ([]Skill, error)`

- [ ] **Step 1: 编写 scanner_test.go**

```go
package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// frontend/react-component
	os.MkdirAll(filepath.Join(dir, "frontend", "react-component"), 0755)
	os.WriteFile(filepath.Join(dir, "frontend", "react-component", "SKILL.md"),
		[]byte("# React Component\n\nBuild React components with TypeScript"), 0644)
	// frontend/css-layout
	os.MkdirAll(filepath.Join(dir, "frontend", "css-layout"), 0755)
	os.WriteFile(filepath.Join(dir, "frontend", "css-layout", "SKILL.md"),
		[]byte("# CSS Layout\n\nResponsive layouts with Flexbox and Grid"), 0644)
	// backend/api-design
	os.MkdirAll(filepath.Join(dir, "backend", "api-design"), 0755)
	os.WriteFile(filepath.Join(dir, "backend", "api-design", "SKILL.md"),
		[]byte("# API Design\n\nREST API design patterns"), 0644)
	return dir
}

func TestScanner_Scan(t *testing.T) {
	repoDir := createTestRepo(t)
	s := NewScanner(repoDir)
	categories, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() 失败: %v", err)
	}

	if len(categories) != 2 {
		t.Fatalf("期望 2 个分类，实际: %d", len(categories))
	}

	found := false
	for _, c := range categories {
		if c.Name == "frontend" {
			found = true
			if len(c.Skills) != 2 {
				t.Fatalf("frontend 期望 2 个 skill，实际: %d", len(c.Skills))
			}
			if c.Skills[0].Name != "react-component" {
				t.Fatalf("期望 skill 名 react-component，实际: %s", c.Skills[0].Name)
			}
		}
	}
	if !found {
		t.Fatal("未找到 frontend 分类")
	}
}

func TestScanner_GetSkill(t *testing.T) {
	repoDir := createTestRepo(t)
	s := NewScanner(repoDir)

	skill, err := s.GetSkill("react-component")
	if err != nil {
		t.Fatalf("GetSkill() 失败: %v", err)
	}
	if skill.Name != "react-component" {
		t.Fatalf("期望 react-component，实际: %s", skill.Name)
	}
	if skill.Category != "frontend" {
		t.Fatalf("期望 category=frontend，实际: %s", skill.Category)
	}

	_, err = s.GetSkill("nonexistent")
	if err == nil {
		t.Fatal("期望查找不存在的 skill 返回错误")
	}
}

func TestScanner_Cache(t *testing.T) {
	repoDir := createTestRepo(t)
	s := NewScanner(repoDir)

	if err := s.BuildCache(); err != nil {
		t.Fatalf("BuildCache() 失败: %v", err)
	}

	skills, err := CachedSkills()
	if err != nil {
		t.Fatalf("CachedSkills() 失败: %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("期望 3 个 skill，实际: %d", len(skills))
	}
}

func TestScanner_SKILL_Description(t *testing.T) {
	repoDir := createTestRepo(t)
	s := NewScanner(repoDir)
	skill, _ := s.GetSkill("react-component")
	if skill.Description == "" {
		t.Fatal("期望从 SKILL.md 提取描述")
	}
}
```

- [ ] **Step 2: 运行测试 — 期望失败**

```bash
go test ./internal/repo/ -v
```

Expected: FAIL (scanner.go 不存在)

- [ ] **Step 3: 编写 scanner.go**

```go
package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skill-management/internal/config"
)

type Scanner struct {
	repoPath string
}

func NewScanner(repoPath string) *Scanner {
	return &Scanner{repoPath: repoPath}
}

// Scan 扫描仓库目录，返回分类树
func (s *Scanner) Scan() ([]Category, error) {
	entries, err := os.ReadDir(s.repoPath)
	if err != nil {
		return nil, fmt.Errorf("读取仓库目录失败: %w", err)
	}

	var categories []Category
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		catPath := filepath.Join(s.repoPath, entry.Name())
		cat := Category{Name: entry.Name()}

		skillDirs, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}
		for _, sd := range skillDirs {
			if !sd.IsDir() || strings.HasPrefix(sd.Name(), ".") {
				continue
			}
			skillPath := filepath.Join(catPath, sd.Name())
			skill := Skill{
				Name:     sd.Name(),
				Category: entry.Name(),
				Path:     skillPath,
			}
			skill.Description = readDescription(skillPath)
			cat.Skills = append(cat.Skills, skill)
		}
		if len(cat.Skills) > 0 {
			categories = append(categories, cat)
		}
	}
	return categories, nil
}

// GetSkill 按名称查找 skill（不区分分类）
func (s *Scanner) GetSkill(name string) (*Skill, error) {
	categories, err := s.Scan()
	if err != nil {
		return nil, err
	}
	for _, cat := range categories {
		for _, sk := range cat.Skills {
			if sk.Name == name {
				return &sk, nil
			}
		}
	}
	return nil, fmt.Errorf("未找到 skill: %s", name)
}

// BuildCache 扫描并缓存到 JSON 文件
func (s *Scanner) BuildCache() error {
	categories, err := s.Scan()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(config.ConfigDir(), "cache.json")
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(categories, "", "  ")
	return os.WriteFile(cachePath, data, 0644)
}

// CachedSkills 从缓存读取所有 skill（扁平列表）
func CachedSkills() ([]Skill, error) {
	cachePath := filepath.Join(config.ConfigDir(), "cache.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("读取缓存失败: %w", err)
	}
	var categories []Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("解析缓存失败: %w", err)
	}
	var skills []Skill
	for _, c := range categories {
		for _, s := range c.Skills {
			skills = append(skills, s)
		}
	}
	return skills, nil
}

// readDescription 从 skill 目录的 SKILL.md 提取第一行作为描述
func readDescription(skillPath string) string {
	mdPath := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(data))
	lines := strings.SplitN(text, "\n", 3)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "# ")
		if line != "" {
			return line
		}
	}
	return ""
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/repo/ -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add -A && git commit -m "feat: repo scanner with cache support"
```

---

### Task 3: 链接管理器

**Files:**
- Create: `internal/linker/linker.go`
- Create: `internal/linker/linker_test.go`

**Interfaces:**
- Consumes: `config.ProjectSkillsDir()` / `config.FindProjectRoot()`
- Produces: `NewLinker(projectRoot) *Linker`, `Link(skillName, skillPath) error`, `Unlink(skillName) error`, `ListLinked() ([]string, error)`, `IsLinked(skillName) bool`

- [ ] **Step 1: 编写 linker_test.go**

```go
package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkAndUnlink(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	// 模拟中央仓库中的 skill 目录
	repoSkillDir := filepath.Join(tmpDir, "repo", "react-component")
	os.MkdirAll(repoSkillDir, 0755)
	os.WriteFile(filepath.Join(repoSkillDir, "SKILL.md"), []byte("# test"), 0644)

	l := NewLinker(tmpDir)

	// 链接
	if err := l.Link("react-component", repoSkillDir); err != nil {
		t.Fatalf("Link() 失败: %v", err)
	}

	// 检查软链接是否存在
	linkPath := filepath.Join(skillsDir, "react-component")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("读取软链接失败: %v", err)
	}
	if target != repoSkillDir {
		t.Fatalf("期望链接到 %s，实际: %s", repoSkillDir, target)
	}

	// IsLinked
	if !l.IsLinked("react-component") {
		t.Fatal("IsLinked() 应返回 true")
	}

	// ListLinked
	linked, err := l.ListLinked()
	if err != nil {
		t.Fatalf("ListLinked() 失败: %v", err)
	}
	if len(linked) != 1 || linked[0] != "react-component" {
		t.Fatalf("期望 [react-component]，实际: %v", linked)
	}

	// 取消链接
	if err := l.Unlink("react-component"); err != nil {
		t.Fatalf("Unlink() 失败: %v", err)
	}
	if l.IsLinked("react-component") {
		t.Fatal("取消链接后 IsLinked() 应返回 false")
	}
}

func TestLink_AlreadyLinked(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	repoSkillDir := filepath.Join(tmpDir, "repo", "react-component")
	os.MkdirAll(repoSkillDir, 0755)

	l := NewLinker(tmpDir)
	l.Link("react-component", repoSkillDir)

	// 重复链接应返回错误
	err := l.Link("react-component", repoSkillDir)
	if err == nil {
		t.Fatal("重复链接应返回错误")
	}
}

func TestLink_InvalidSkillPath(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".opencode", "skills"), 0755)

	l := NewLinker(tmpDir)
	err := l.Link("invalid", "/nonexistent/path")
	if err == nil {
		t.Fatal("无效路径应返回错误")
	}
}

func TestUnlink_NotLinked(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".opencode", "skills"), 0755)

	l := NewLinker(tmpDir)
	err := l.Unlink("nonexistent")
	if err == nil {
		t.Fatal("取消未链接的 skill 应返回错误")
	}
}

func TestLink_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	names := []string{"react", "vue", "angular"}
	for _, n := range names {
		d := filepath.Join(tmpDir, "repo", n)
		os.MkdirAll(d, 0755)
	}

	l := NewLinker(tmpDir)
	for _, n := range names {
		if err := l.Link(n, filepath.Join(tmpDir, "repo", n)); err != nil {
			t.Fatalf("Link(%s) 失败: %v", n, err)
		}
	}

	linked, _ := l.ListLinked()
	if len(linked) != 3 {
		t.Fatalf("期望 3 个，实际: %d", len(linked))
	}
}
```

- [ ] **Step 2: 运行测试 — 期望失败**

```bash
go test ./internal/linker/ -v
```

Expected: FAIL

- [ ] **Step 3: 编写 linker.go**

```go
package linker

import (
	"fmt"
	"os"
	"path/filepath"
)

type Linker struct {
	skillsDir string
}

// NewLinker 创建链接管理器
func NewLinker(projectRoot string) *Linker {
	return &Linker{
		skillsDir: filepath.Join(projectRoot, ".opencode", "skills"),
	}
}

// LinkedDir 返回链接目录路径
func (l *Linker) LinkedDir() string {
	return l.skillsDir
}

// Link 创建软链接
func (l *Linker) Link(skillName string, skillPath string) error {
	// 验证源路径存在
	if _, err := os.Stat(skillPath); err != nil {
		return fmt.Errorf("skill 路径不存在: %s", skillPath)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(l.skillsDir, 0755); err != nil {
		return fmt.Errorf("创建链接目录失败: %w", err)
	}

	linkPath := filepath.Join(l.skillsDir, skillName)

	// 检查是否已链接
	if _, err := os.Lstat(linkPath); err == nil {
		return fmt.Errorf("skill %q 已链接", skillName)
	}

	// 创建相对路径的软链接（跨容器兼容）
	relPath, err := filepath.Rel(l.skillsDir, skillPath)
	if err != nil {
		relPath = skillPath
	}

	if err := os.Symlink(relPath, linkPath); err != nil {
		return fmt.Errorf("创建软链接失败: %w", err)
	}
	return nil
}

// Unlink 删除软链接
func (l *Linker) Unlink(skillName string) error {
	linkPath := filepath.Join(l.skillsDir, skillName)
	if _, err := os.Lstat(linkPath); err != nil {
		return fmt.Errorf("skill %q 未链接", skillName)
	}
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("移除软链接失败: %w", err)
	}
	return nil
}

// ListLinked 返回所有已链接的 skill 名称
func (l *Linker) ListLinked() ([]string, error) {
	entries, err := os.ReadDir(l.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取链接目录失败: %w", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

// IsLinked 检查指定 skill 是否已链接
func (l *Linker) IsLinked(skillName string) bool {
	linkPath := filepath.Join(l.skillsDir, skillName)
	_, err := os.Lstat(linkPath)
	return err == nil
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/linker/ -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add -A && git commit -m "feat: linker with symlink operations"
```

---

### Task 4: CLI 子命令（list/info/add/remove/sync/path/tui）

**Files:**
- Modify: `main.go`
- Create: `internal/cli/root.go` (或直接在 main.go 添加命令)

**Interfaces:**
- Consumes: `config`, `repo.Scanner`, `linker.Linker`
- Produces: 完整的 CLI 子命令树

- [ ] **Step 1: 注册 list 命令**

在 `main.go` 的 `init()` 中添加：

```go
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已链接的 skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		linked, err := l.ListLinked()
		if err != nil {
			return err
		}
		if len(linked) == 0 {
			fmt.Println("当前项目没有链接任何 skill")
			return nil
		}
		fmt.Println("已链接的 skill:")
		for _, name := range linked {
			fmt.Printf("  • %s\n", name)
		}
		return nil
	},
}
```

在 `init()` 中注册：`rootCmd.AddCommand(listCmd)`

- [ ] **Step 2: 注册 info 命令**

```go
var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "查看 skill 详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		skill, err := s.GetSkill(args[0])
		if err != nil {
			return err
		}
		// 读取 SKILL.md
		data, _ := os.ReadFile(filepath.Join(skill.Path, "SKILL.md"))
		fmt.Printf("名称: %s\n分类: %s\n路径: %s\n\n%s\n",
			skill.Name, skill.Category, skill.Path, string(data))
		return nil
	},
}
```

- [ ] **Step 3: 注册 add 命令**

```go
var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "链接指定 skill 到当前项目",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		skill, err := s.GetSkill(args[0])
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Link(skill.Name, skill.Path); err != nil {
			return err
		}
		fmt.Printf("✅ skill %q 已链接到当前项目\n", skill.Name)
		return nil
	},
}
```

- [ ] **Step 4: 注册 remove 命令**

```go
var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "移除已链接的 skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Unlink(args[0]); err != nil {
			return err
		}
		fmt.Printf("✅ skill %q 已移除\n", args[0])
		return nil,
	},
}
```

- [ ] **Step 5: 注册 sync 命令**

```go
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "刷新仓库缓存索引",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		if err := s.BuildCache(); err != nil {
			return fmt.Errorf("刷新缓存失败: %w", err)
		}
		fmt.Println("✅ 缓存已刷新")
		return nil,
	},
}
```

- [ ] **Step 6: 注册 path 命令（含 get/set 子命令）**

```go
var pathCmd = &cobra.Command{
	Use:   "path [get|set <path>]",
	Short: "查看或设置仓库路径",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 默认 get
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println(cfg.RepoPath)
		return nil
	},
}

var pathSetCmd = &cobra.Command{
	Use:   "set <path>",
	Short: "设置仓库路径",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		// 验证路径存在
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("路径不存在: %s", absPath)
		}
		cfg := &config.Config{RepoPath: absPath}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("✅ 仓库路径已设置为: %s\n", absPath)
		return nil
	},
}
```

在 `init()` 中注册：`pathCmd.AddCommand(pathSetCmd)`

- [ ] **Step 7: 注册 tui 命令（骨架，Task 6 填充）**

```go
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "启动 TUI 交互界面浏览/选择 skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Start()
	},
}
```

- [ ] **Step 8: 编译并冒烟测试所有命令**

```bash
go build -o skill-mgr . && ./skill-mgr --help
./skill-mgr list
./skill-mgr path
```

Expected: 各命令正常输出

- [ ] **Step 9: 更新 main.go 中的 import**

确保所有 import 完整：

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"skill-management/internal/config"
	"skill-management/internal/linker"
	"skill-management/internal/repo"
	"skill-management/internal/tui"
)
```

- [ ] **Step 10: 编译验证并提交**

```bash
go build -o skill-mgr . && ./skill-mgr
git add -A && git commit -m "feat: CLI commands (list/info/add/remove/sync/path/tui)"
```

---

### Task 5: Agent JSON 输出模式

**Files:**
- Modify: `main.go`（添加 agent 子命令）

**Interfaces:**
- Consumes: `repo.Scanner`, `linker.Linker`, `config`
- Produces: `agent search <query>`, `agent link <name>`, `agent linked`, `agent remove <name>` 输出 JSON

- [ ] **Step 1: 注册 agent 命令**

```go
var agentCmd = &cobra.Command{
	Use:    "agent",
	Short:  "JSON 输出模式（供 OpenCode agent 内部调用）",
	Hidden: true,
}

var agentSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "搜索 skill 返回 JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return printJSONError(err)
		}
		s := repo.NewScanner(cfg.RepoPath)
		categories, err := s.Scan()
		if err != nil {
			return printJSONError(err)
		}
		query := strings.ToLower(args[0])
		var results []repo.Skill
		for _, cat := range categories {
			for _, sk := range cat.Skills {
				if strings.Contains(strings.ToLower(sk.Name), query) ||
					strings.Contains(strings.ToLower(sk.Description), query) {
					results = append(results, sk)
				}
			}
		}
		if results == nil {
			results = []repo.Skill{}
		}
		return printJSON(map[string]any{"results": results})
	},
}

var agentLinkCmd = &cobra.Command{
	Use:   "link <name>",
	Short: "链接 skill 返回 JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return printJSONError(err)
		}
		cfg, err := config.Load()
		if err != nil {
			return printJSONError(err)
		}
		s := repo.NewScanner(cfg.RepoPath)
		skill, err := s.GetSkill(args[0])
		if err != nil {
			return printJSONError(err)
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Link(skill.Name, skill.Path); err != nil {
			return printJSONError(err)
		}
		return printJSON(map[string]any{"status": "ok", "linked": skill.Name})
	},
}

var agentLinkedCmd = &cobra.Command{
	Use:   "linked",
	Short: "列出已链接 skill 返回 JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return printJSONError(err)
		}
		l := linker.NewLinker(projectRoot)
		linked, err := l.ListLinked()
		if err != nil {
			return printJSONError(err)
		}
		return printJSON(map[string]any{"linked": linked})
	},
}

var agentRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "移除 skill 返回 JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return printJSONError(err)
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Unlink(args[0]); err != nil {
			return printJSONError(err)
		}
		return printJSON(map[string]any{"status": "ok", "removed": args[0]})
	},
}
```

- [ ] **Step 2: 添加 JSON 辅助函数**

```go
func printJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printJSONError(err error) error {
	data, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Println(string(data))
	return nil // 返回 nil 以免 cobra 打印两次错误
}
```

- [ ] **Step 3: 在 init() 中注册**

```go
agentCmd.AddCommand(agentSearchCmd, agentLinkCmd, agentLinkedCmd, agentRemoveCmd)
rootCmd.AddCommand(agentCmd)
```

- [ ] **Step 4: 编译并测试 JSON 输出**

```bash
go build -o skill-mgr .
# 设置一个测试仓库
mkdir -p /tmp/test-skills-repo/frontend/react-component
echo "# React" > /tmp/test-skills-repo/frontend/react-component/SKILL.md
./skill-mgr path set /tmp/test-skills-repo
./skill-mgr agent search react
./skill-mgr agent linked
```

Expected: JSON 格式输出

- [ ] **Step 5: 提交**

```bash
git add -A && git commit -m "feat: agent JSON output mode for OpenCode integration"
```

---

### Task 6: TUI 交互界面（Bubble Tea）

**Files:**
- Create: `internal/tui/model.go`
- Create: `internal/tui/browse.go`
- Create: `internal/tui/preview.go`

**Interfaces:**
- Consumes: `repo.Scanner`, `linker.Linker`, `config`
- Produces: `tui.Start() error` 入口函数

- [ ] **Step 1: 编写 model.go — 主模型 + 入口**

```go
package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"skill-management/internal/config"
	"skill-management/internal/linker"
	"skill-management/internal/repo"
)

type viewType int

const (
	browseView viewType = iota
	previewView
)

type model struct {
	view      viewType
	categories []repo.Category
	skills     []repo.Skill // 扁平列表
	linker     *linker.Linker

	// 浏览视图状态
	cursor     int       // 当前选中项索引
	panelFocus int       // 0=分类面板, 1=skill 面板
	catCursor  int       // 分类面板光标
	selected   map[int]bool // 已选 skill 索引

	// 预览视图状态
	previewSkill *repo.Skill
	previewContent string

	// 窗口尺寸
	width  int
	height int

	// 搜索
	searchQuery string
	searching   bool

	err error
}

func Start() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	s := repo.NewScanner(cfg.RepoPath)
	categories, err := s.Scan()
	if err != nil {
		return fmt.Errorf("扫描仓库失败: %w", err)
	}

	// 扁平化
	var allSkills []repo.Skill
	linkedMap := make(map[string]bool)
	l := linker.NewLinker(projectRoot)
	linked, _ := l.ListLinked()
	for _, name := range linked {
		linkedMap[name] = true
	}
	for _, cat := range categories {
		for _, sk := range cat.Skills {
			sk.Linked = linkedMap[sk.Name]
			allSkills = append(allSkills, sk)
		}
	}

	m := model{
		view:       browseView,
		categories: categories,
		skills:     allSkills,
		linker:     l,
		selected:   make(map[int]bool),
		panelFocus: 1, // 默认聚焦 skill 列表
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
```

- [ ] **Step 2: 编写 model 的 Init/Update/View 方法**

```go
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.view == browseView {
				m.panelFocus = 1 - m.panelFocus
			}

		case "up", "k":
			if m.view == browseView && m.panelFocus == 1 {
				if m.cursor > 0 {
					m.cursor--
				}
			}
			if m.view == browseView && m.panelFocus == 0 {
				if m.catCursor > 0 {
					m.catCursor--
				}
			}

		case "down", "j":
			if m.view == browseView && m.panelFocus == 1 {
				if m.cursor < len(m.skills)-1 {
					m.cursor++
				}
			}
			if m.view == browseView && m.panelFocus == 0 {
				if m.catCursor < len(m.categories)-1 {
					m.catCursor++
				}
			}

		case " ":
			if m.view == browseView && m.panelFocus == 1 {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}

		case "enter":
			if m.view == browseView && m.panelFocus == 1 {
				// 进入预览
				skill := m.skills[m.cursor]
				content, _ := os.ReadFile(filepath.Join(skill.Path, "SKILL.md"))
				m.previewSkill = &skill
				m.previewContent = string(content)
				m.view = previewView
			} else if m.view == previewView {
				// 在预览中按 enter 执行链接
				if m.previewSkill != nil {
					err := m.linker.Link(m.previewSkill.Name, m.previewSkill.Path)
					if err != nil {
						m.err = err
					} else {
						m.skills[m.cursor].Linked = true
						m.selected[m.cursor] = false
					}
				}
				m.view = browseView
			}

		case "l":
			// 直接链接选中的项
			if m.view == browseView {
				for idx := range m.selected {
					skill := m.skills[idx]
					m.linker.Link(skill.Name, skill.Path)
					m.skills[idx].Linked = true
				}
				m.selected = make(map[int]bool)
			}

		case "d":
			// 取消链接光标处的 skill
			if m.view == browseView && m.panelFocus == 1 {
				skill := m.skills[m.cursor]
				m.linker.Unlink(skill.Name)
				m.skills[m.cursor].Linked = false
			}

		case "esc":
			if m.view == previewView {
				m.view = browseView
			}
			if m.searching {
				m.searching = false
				m.searchQuery = ""
			}

		case "/":
			if m.view == browseView {
				m.searching = true
				m.searchQuery = ""
			}

		default:
			if m.searching && len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				// 简单的过滤搜索
				for i, sk := range m.skills {
					if contains(sk.Name, m.searchQuery) {
						m.cursor = i
						break
					}
				}
			}
		}
	}
	return m, nil
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr))
}
```

- [ ] **Step 3: 编写浏览视图渲染**

在 `browse.go` 中：

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED"))
	linkedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

func (m model) renderBrowseView() string {
	// 计算可用高度
	headerHeight := 3
	footerHeight := 2
	panelHeight := m.height - headerHeight - footerHeight
	if panelHeight < 10 {
		panelHeight = 10
	}

	// 标题
	header := titleStyle.Render(" slink — Skill 管理器 ")
	repoPath, _ := config.Load()
	header += mutedStyle.Render(fmt.Sprintf("  仓库: %s", repoPath.RepoPath))

	// 左面板：分类列表
	leftWidth := 20
	var catLines []string
	catLines = append(catLines, mutedStyle.Render("分类"))
	for i, cat := range m.categories {
		line := fmt.Sprintf("  %s", cat.Name)
		if m.panelFocus == 0 && i == m.catCursor {
			line = activeStyle.Render(fmt.Sprintf("  %s", cat.Name))
		} else {
			line = normalStyle.Render(fmt.Sprintf("  %s", cat.Name))
		}
		catLines = append(catLines, line)
	}
	leftPanel := panelStyle.Width(leftWidth).Height(panelHeight).Render(
		strings.Join(catLines, "\n"),
	)

	// 右面板：skill 列表
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 30 {
		rightWidth = 30
	}

	// 过滤：按选中分类
	filteredSkills := m.skills
	if m.panelFocus == 0 && m.catCursor < len(m.categories) {
		catName := m.categories[m.catCursor].Name
		var fs []repo.Skill
		for _, sk := range m.skills {
			if sk.Category == catName {
				fs = append(fs, sk)
			}
		}
		if len(fs) > 0 {
			filteredSkills = fs
		}
	}

	var skillLines []string
	skillLines = append(skillLines, mutedStyle.Render(fmt.Sprintf("Skills (%d)", len(filteredSkills))))
	for i, sk := range filteredSkills {
		prefix := "  "
		if sk.Linked {
			prefix = "★ " + linkedStyle.Render("●")
		}
		if m.selected[i] {
			prefix = "✓ "
		}
		line := fmt.Sprintf("%s %s", prefix, sk.Name)
		if m.panelFocus == 1 && i == m.cursor {
			line = activeStyle.Render(fmt.Sprintf("%s %s", prefix, sk.Name))
		} else if sk.Linked {
			line = linkedStyle.Render(fmt.Sprintf("%s %s", prefix, sk.Name))
		} else {
			line = normalStyle.Render(fmt.Sprintf("%s %s", prefix, sk.Name))
		}
		skillLines = append(skillLines, line)
	}
	rightPanel := panelStyle.Width(rightWidth).Height(panelHeight).Render(
		strings.Join(skillLines, "\n"),
	)

	// 底部帮助
	footer := mutedStyle.Render(
		" [Tab]切换面板  [↑↓]导航  [Space]选择  [Enter]预览  [L]链接选中  [D]取消链接  [/]搜索  [Q]退出",
	)

	return fmt.Sprintf("%s\n%s\n%s",
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
		footer,
	)
}

func (m model) View() string {
	switch m.view {
	case browseView:
		return m.renderBrowseView()
	case previewView:
		return m.renderPreviewView()
	}
	return ""
}
```

- [ ] **Step 4: 编写预览视图渲染**

在 `preview.go` 中：

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	previewTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	previewBodyStyle  = lipgloss.NewStyle().Padding(1, 2)
)

func (m model) renderPreviewView() string {
	if m.previewSkill == nil {
		return ""
	}

	header := previewTitleStyle.Render(fmt.Sprintf(" 📖 %s", m.previewSkill.Name))
	header += fmt.Sprintf("  [%s]", m.previewSkill.Category)

	info := fmt.Sprintf("路径: %s\n已链接: %v\n",
		m.previewSkill.Path, m.previewSkill.Linked)

	body := previewBodyStyle.Render(m.previewContent)

	footer := mutedStyle.Render(
		" [Enter]链接此 skill  [Esc]返回  [Q]退出",
	)

	// 错误显示
	if m.err != nil {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(
			fmt.Sprintf("错误: %v", m.err))
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		header, info, body, footer)
}
```

- [ ] **Step 5: 编译验证**

```bash
go build -o skill-mgr . && ./skill-mgr tui
```

Expected: TUI 正常启动，浏览和预览视图可用

- [ ] **Step 6: 提交**

```bash
git add -A && git commit -m "feat: Bubble Tea TUI with browse and preview views"
```

---

### Task 7: OpenCode Agent 注册 + 集成配置

**Files:**
- Create: `.opencode/agents/slink.json`

**Interfaces:**
- Consumes: skill-mgr 二进制已安装到 PATH

- [ ] **Step 1: 创建 agent 注册文件**

`.opencode/agents/slink.json`：

```json
{
  "name": "slink",
  "description": "管理项目 skill 链接 — 浏览、添加、移除 skill",
  "command": "/slink",
  "handler": {
    "type": "shell",
    "command": "skill-mgr",
    "args": ["agent"]
  },
  "instructions": "用户通过 /slink 指令管理 OpenCode skill 链接。无参数时解析自然语言意图:\n- '添加/链接 X skill' → skill-mgr agent search X → skill-mgr agent link X\n- '查看/列出 skill' → skill-mgr agent linked\n- '移除/删除 X skill' → skill-mgr agent remove X\n- '打开 TUI / 浏览' → skill-mgr tui\n将 JSON 结果格式化后回复用户。",
  "permissions": ["read", "write"]
}
```

- [ ] **Step 2: 创建 .gitignore**

```
skill-mgr
```

- [ ] **Step 3: 全量编译 + 最终验证**

```bash
go build -o skill-mgr .
./skill-mgr --help
./skill-mgr list
./skill-mgr path
```

Expected: 所有命令正常工作

- [ ] **Step 4: 最终提交**

```bash
git add -A && git commit -m "feat: OpenCode agent registration and final integration"
```

---

## 自检清单

- [x] Spec 覆盖：所有 spec 需求都映射到任务（CLI/TUI/Agent JSON/OpenCode 集成）
- [x] 无占位符：所有步骤包含完整代码
- [x] 类型一致性：Config/Skill/Category/Linker 在各任务间签名一致
- [x] 项目根目录检测通过 .opencode/ 向上查找
- [x] 软链接使用相对路径