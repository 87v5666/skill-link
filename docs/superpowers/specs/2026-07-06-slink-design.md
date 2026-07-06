# slink — OpenCode Skill 链接管理工具

## 概述

`slink`（skill-link 的缩写）是一个用于管理 OpenCode skill 软链接的 CLI 工具 + OpenCode 内置指令。用户可以从中央 skill 仓库浏览、选择、链接 skill 到当前项目的 `.opencode/skills/` 目录下。

## 用户需求

- Skill 统一存放于中央仓库文件夹
- 用户可在项目文件夹下通过 TUI 浏览并选择需要的 skill，自动创建软链接
- 用户可在 OpenCode 启动后，通过 `/slink` 指令管理 skill
- 支持自然语言交互：`/slink` 默认无参数时，智能体解析用户意图并执行

## 技术选型

| 层 | 选型 | 理由 |
|----|------|------|
| 语言 | Go | 单二进制分发，跨平台，TUI 生态成熟 |
| TUI 框架 | Bubble Tea | Go 生态最流行的 TUI 框架 |
| 存储 | JSON 文件 | 轻量，无需数据库 |

## 架构

```
用户交互层
├── 终端下: skill-mgr (独立 TUI)
└── OpenCode 内: /slink 指令 (自定义 agent)

       ▼
skill-mgr CLI 二进制
├── TUI 模式 (browse/select)
├── CLI 模式 (list/add/remove/info)
└── JSON 输出模式 (agent 内部调用)

       ▼
核心操作层
├── 读取中央仓库目录结构
├── 创建/删除软链接到 .opencode/skills/
└── 刷新配置缓存
```

## Skill 仓库结构

中央仓库位置：`~/.config/opencode/skills-repo/`（可配置）

```
~/.config/opencode/skills-repo/
├── frontend/
│   ├── react-component/
│   │   └── SKILL.md
│   ├── css-layout/
│   │   └── SKILL.md
│   └── form-validation/
│       └── SKILL.md
├── backend/
│   ├── api-design/
│   │   └── SKILL.md
│   ├── database-schema/
│   │   └── SKILL.md
│   └── auth-flow/
│       └── SKILL.md
├── testing/
│   ├── unit-test/
│   │   └── SKILL.md
│   └── e2e-test/
│       └── SKILL.md
└── meta/
    ├── category-list.json
    └── skill-registry.json
```

每个 skill 文件夹是一个可选单元，软链接到项目的 `.opencode/skills/` 下。

## 配置文件

```
~/.config/slink/
├── config.json       # 仓库路径等配置
├── cache.json        # 仓库 skill 索引缓存
└── links.db.json     # 各项目链接记录（可选审计）
```

## CLI 子命令

```
skill-mgr
├── tui               # 启动 TUI 浏览/选择
├── list              # 列出已链接 skill
├── add <path>        # 添加指定 skill 链接
├── remove <name>     # 移除 skill
├── sync              # 刷新仓库缓存索引
├── info <name>       # 查看 skill 详情
├── path              # 查看/设置仓库路径
└── agent             # 供 OpenCode agent 调用的 JSON 模式
```

## OpenCode 指令集 (`/slink`)

| 指令 | 功能 |
|------|------|
| `/slink` | 默认无参数 — 自然语言模式，智能体理解意图执行 |
| `/slink tui` | 打开 TUI 交互界面浏览/选择 |
| `/slink list` | 列出已链接的 skill |
| `/slink add <name>` | 按名称添加 skill |
| `/slink remove <name>` | 移除指定 skill |
| `/slink info <name>` | 预览 skill 详情 |
| `/slink path` | 配置中央仓库位置 |

### 自然语言模式

`/slink` 无参数时，用户的自然语言由 **OpenCode 的 slink agent（LLM）** 解析意图，再调用 skill-mgr 的 JSON 模式或 CLI 模式执行。skill-mgr 二进制本身不做 NLP。

解析示例：

| 用户输入 | 解析结果 | 执行命令 |
|----------|---------|---------|
| `/slink 我想把 react 和 tailwind 的 skill 加进来` | 搜索 react、tailwind，创建链接 | `skill-mgr agent link react-component` `skill-mgr agent link css-layout` |
| `/slink 看看我现在有哪些 skill` | 列出已链接 | `skill-mgr agent linked` |
| `/slink 帮我把 api-design 那个 skill 链接上` | 搜索 api-design，创建链接 | `skill-mgr agent search api-design` → `skill-mgr agent link api-design` |
| `/slink 不需要 vue 的 skill 了，删掉吧` | 搜索 vue 相关，移除链接 | `skill-mgr agent search vue` → `skill-mgr agent remove vue-components` |

## OpenCode 集成方案

在项目的 `.opencode/agents/` 目录注册自定义 agent `slink`，或通过 opencode.json 配置声明。

工作流程：
1. 用户输入 `/slink ...`
2. 匹配到 slink agent
3. 解析指令或自然语言意图
4. 简单操作直接执行（list/info）
5. 复杂操作调用 skill-mgr 的 JSON 输出模式或启动 TUI
6. 返回结果给用户

### agent JSON 输出模式

供 slink agent 内部调用，输出结构化数据：

```
skill-mgr agent search "react"
→ {"results": [{"name": "react-component", "path": "...", "description": "..."}]}

skill-mgr agent link "react-component"
→ {"status": "ok", "linked": "react-component"}

skill-mgr agent linked
→ {"linked": ["react-component", "css-layout"]}
```

## TUI 界面设计

### 浏览视图

左右分栏布局：左侧分类树，右侧 skill 列表。
- 支持 Tab 切换面板
- 方向键导航
- Space 多选
- Enter 确认
- `/` 搜索
- `q` 退出
- 已链接 skill 显示标记（★ 已链接）

### 预览视图

查看选中 skill 的 SKILL.md 内容，支持 Enter 链接或返回。

## 软链接目标

所有选中的 skill 软链接到项目根目录下的 `.opencode/skills/`。

## 未来可扩展方向

- 多仓库源支持（本地 + 远程 Git 仓库）
- skill 版本管理
- 依赖解析（skill 之间的依赖关系）
- 社区仓库索引