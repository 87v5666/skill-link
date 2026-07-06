package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
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
	view       viewType
	categories []repo.Category
	skills     []repo.Skill // 扁平列表
	linker     *linker.Linker

	// 浏览视图状态
	cursor     int          // 当前选中项索引
	panelFocus int          // 0=分类面板, 1=skill 面板
	catCursor  int          // 分类面板光标
	selected   map[int]bool // 已选 skill 索引

	// 预览视图状态
	previewSkill   *repo.Skill
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