package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"skill-management/internal/config"
	"skill-management/internal/data"
	"skill-management/internal/linker"
	"skill-management/internal/repo"
)

type viewType int

const (
	browseView viewType = iota
	previewView
)

type textInputMode int

const (
	noInput textInputMode = iota
	addingCategory
	editingNote
	addToCategory
	editingCategoryName
)

type model struct {
	view       viewType
	categories []repo.Category
	skills     []repo.Skill // 扁平列表
	linker     *linker.Linker

	// 浏览视图状态
	cursorName string          // 当前选中 skill 名称
	panelFocus int             // 0=分类面板, 1=skill 面板
	catCursor  int             // 分类面板光标
	selectedCategoryIdx int    // 0=全部, 1+=自定义分类索引（独立于 panelFocus）
	selected   map[string]bool // 已选 skill 名称

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

	// 数据持久化
	dataStore     *data.Store
	version       string
	customCatNames []string // sorted custom category names for display

	// 文本输入模式
	inputMode    textInputMode
	inputBuffer  string
	inputPrompt  string
	inputTitle   string

	// 选择分类模式
	catSelectCursor int

	renamingCategory string // 正在重命名的分类名
}

func Start(version string) error {
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

	ds, err := data.Load()
	if err != nil {
		return fmt.Errorf("加载数据失败: %w", err)
	}

	m := model{
		view:           browseView,
		categories:     categories,
		skills:         allSkills,
		linker:         l,
		selected:       make(map[string]bool),
		panelFocus:     1, // 默认聚焦 skill 列表
		dataStore:      ds,
		version:        version,
		customCatNames: ds.SortedCategoryNames(),
	}
	if len(allSkills) > 0 {
		m.cursorName = allSkills[0].Name
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
		// 文本输入模式优先处理
		if m.inputMode != noInput {
			switch msg.String() {
			case "enter":
				if m.inputMode == addingCategory {
					if m.inputBuffer != "" {
						m.dataStore.CreateCategory(m.inputBuffer)
						m.customCatNames = m.dataStore.SortedCategoryNames()
					}
					m.inputMode = noInput
				} else if m.inputMode == editingNote {
					m.dataStore.SetNote(m.cursorName, m.inputBuffer)
					m.inputMode = noInput
				} else if m.inputMode == addToCategory {
					catNames := m.dataStore.SortedCategoryNames()
					if m.catSelectCursor >= 0 && m.catSelectCursor < len(catNames) {
						catName := catNames[m.catSelectCursor]
						// Toggle: if already in category, remove; else add
						found := false
						for _, n := range m.dataStore.Categories[catName] {
							if n == m.cursorName {
								found = true
								break
							}
						}
						if found {
							m.dataStore.RemoveSkillFromCategory(catName, m.cursorName)
						} else {
							m.dataStore.AddSkillToCategory(catName, m.cursorName)
						}
					}
					m.inputMode = noInput
				} else if m.inputMode == editingCategoryName {
					if m.inputBuffer != "" && m.inputBuffer != m.renamingCategory {
						m.dataStore.RenameCategory(m.renamingCategory, m.inputBuffer)
						m.customCatNames = m.dataStore.SortedCategoryNames()
					}
					m.renamingCategory = ""
					m.inputMode = noInput
				}
				return m, nil
				m.inputMode = noInput
				m.inputBuffer = ""
				return m, nil

			case "backspace":
				if len(m.inputBuffer) > 0 {
					m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
				}
				return m, nil

			default:
				if m.inputMode == addToCategory {
					switch msg.String() {
					case "up", "k":
						if m.catSelectCursor > 0 {
							m.catSelectCursor--
						}
					case "down", "j":
						catNames := m.dataStore.SortedCategoryNames()
						if m.catSelectCursor < len(catNames)-1 {
							m.catSelectCursor++
						}
					}
					return m, nil
				}
				// Accept Chinese and other multi-byte characters
				if msg.Type == tea.KeyRunes {
					m.inputBuffer += string(msg.Runes)
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.view == browseView {
				m.panelFocus = 1 - m.panelFocus
			}

		case "up", "k":
			if m.view == browseView && m.panelFocus == 1 {
				filtered := m.filteredSkills()
				for i, sk := range filtered {
					if sk.Name == m.cursorName && i > 0 {
						m.cursorName = filtered[i-1].Name
						break
					}
				}
			}
			if m.view == browseView && m.panelFocus == 0 {
				if m.catCursor > 0 {
					m.catCursor--
				}
				if m.catCursor <= len(m.customCatNames) {
					m.selectedCategoryIdx = m.catCursor
				}
			}

		case "down", "j":
			if m.view == browseView && m.panelFocus == 1 {
				filtered := m.filteredSkills()
				for i, sk := range filtered {
					if sk.Name == m.cursorName && i < len(filtered)-1 {
						m.cursorName = filtered[i+1].Name
						break
					}
				}
			}
			if m.view == browseView && m.panelFocus == 0 {
				if m.catCursor < len(m.customCatNames)+1 {
					m.catCursor++
				}
				if m.catCursor <= len(m.customCatNames) {
					m.selectedCategoryIdx = m.catCursor
				}
			}

		case " ":
			if m.view == browseView && m.panelFocus == 1 {
				m.selected[m.cursorName] = !m.selected[m.cursorName]
			}

		case "enter":
			if m.view == browseView && m.panelFocus == 0 && m.catCursor == len(m.customCatNames)+1 {
				m.inputMode = addingCategory
				m.inputBuffer = ""
				m.inputPrompt = "输入分类名称: "
				m.inputTitle = "新建分类"
				return m, nil
			}
			if m.view == browseView && m.panelFocus == 1 {
				// 进入预览
				var skill repo.Skill
				for _, sk := range m.skills {
					if sk.Name == m.cursorName {
						skill = sk
						break
					}
				}
				content, err := os.ReadFile(filepath.Join(skill.Path, "SKILL.md"))
				if err != nil {
					m.err = err
					m.previewContent = ""
				} else {
					m.previewContent = string(content)
				}
				m.previewSkill = &skill
				m.view = previewView
			} else if m.view == previewView {
				// 在预览中按 enter 执行链接
				if m.previewSkill != nil {
					err := m.linker.Link(m.previewSkill.Name, m.previewSkill.Path)
					if err != nil {
						m.err = err
					} else {
						for i, sk := range m.skills {
							if sk.Name == m.previewSkill.Name {
								m.skills[i].Linked = true
								break
							}
						}
						m.selected[m.previewSkill.Name] = false
					}
				}
				m.view = browseView
			}

		case "l":
			// 直接链接选中的项
			if m.view == browseView {
				for name := range m.selected {
					for i, sk := range m.skills {
						if sk.Name == name {
							if err := m.linker.Link(sk.Name, sk.Path); err != nil {
								m.err = err
							} else {
								m.skills[i].Linked = true
							}
							break
						}
					}
				}
				m.selected = make(map[string]bool)
			}

		case "d":
			// 取消链接光标处的 skill
			if m.view == browseView && m.panelFocus == 1 {
				for i, sk := range m.skills {
					if sk.Name == m.cursorName {
						if err := m.linker.Unlink(sk.Name); err != nil {
							m.err = err
						} else {
							m.skills[i].Linked = false
						}
						break
					}
				}
			}

		case "n":
			// New category (only when not in input mode)
			if m.view == browseView {
				m.inputMode = addingCategory
				m.inputBuffer = ""
				m.inputPrompt = "输入分类名称: "
				m.inputTitle = "新建分类"
			}

		case "e":
			// Edit note for current skill
			if m.view == browseView && m.panelFocus == 1 {
				m.inputMode = editingNote
				m.inputBuffer = m.dataStore.GetNote(m.cursorName)
				m.inputPrompt = fmt.Sprintf("备注 [%s]: ", m.cursorName)
				m.inputTitle = "编辑备注"
			}

		case "a":
			// Toggle add/remove current skill from category
			if m.view == browseView && m.panelFocus == 1 {
				m.inputMode = addToCategory
				m.catSelectCursor = 0
				m.inputTitle = "选择分类"
			}

		case "x":
			// Delete category (when on category panel, on a custom category)
			if m.view == browseView && m.panelFocus == 0 && m.catCursor > 0 && m.catCursor <= len(m.customCatNames) {
				catName := m.customCatNames[m.catCursor-1]
				m.dataStore.DeleteCategory(catName)
				m.customCatNames = m.dataStore.SortedCategoryNames()
				if m.catCursor > len(m.customCatNames) {
					m.catCursor = len(m.customCatNames)
				}
				m.selectedCategoryIdx = m.catCursor
			}

		case "r":
			// Rename category (when on category panel, on a custom category)
			if m.view == browseView && m.panelFocus == 0 && m.catCursor > 0 && m.catCursor <= len(m.customCatNames) {
				m.renamingCategory = m.customCatNames[m.catCursor-1]
				m.inputMode = editingCategoryName
				m.inputBuffer = m.renamingCategory
				m.inputPrompt = fmt.Sprintf("重命名分类 [%s] → ", m.renamingCategory)
				m.inputTitle = "重命名分类"
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
				for _, sk := range m.skills {
					if strings.Contains(strings.ToLower(sk.Name), strings.ToLower(m.searchQuery)) {
						m.cursorName = sk.Name
						break
					}
				}
			}
		}
	}
	return m, nil
}

func (m model) filteredSkills() []repo.Skill {
	// 使用 selectedCategoryIdx 追踪选中的分类，不受 panelFocus 影响
	if m.selectedCategoryIdx <= 0 {
		return m.skills
	}
	idx := m.selectedCategoryIdx - 1
	if idx >= 0 && idx < len(m.customCatNames) {
		catName := m.customCatNames[idx]
		skillNames := m.dataStore.Categories[catName]
		var fs []repo.Skill
		for _, sk := range m.skills {
			for _, n := range skillNames {
				if sk.Name == n {
					fs = append(fs, sk)
					break
				}
			}
		}
		return fs
	}

	return m.skills
}