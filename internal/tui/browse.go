package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"skill-management/internal/config"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED"))
	linkedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	panelStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
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

	// 左面板：自定义分类列表
	leftWidth := 24
	var catLines []string
	catLines = append(catLines, mutedStyle.Render("分类"))

	// "全部" - always first
	allLabel := "  📋 全部"
	if m.panelFocus == 0 && m.catCursor == 0 {
		catLines = append(catLines, activeStyle.Render(allLabel))
	} else {
		catLines = append(catLines, normalStyle.Render(allLabel))
	}

	// Custom categories
	for i, catName := range m.customCatNames {
		idx := i + 1
		if m.panelFocus == 0 && m.catCursor == idx {
			catLines = append(catLines, activeStyle.Render(fmt.Sprintf("  %s", catName)))
		} else {
			catLines = append(catLines, normalStyle.Render(fmt.Sprintf("  %s", catName)))
		}
	}

	// "新建分类" action
	newLabel := "  ＋ 新建分类"
	if m.panelFocus == 0 && m.catCursor == len(m.customCatNames)+1 {
		catLines = append(catLines, activeStyle.Render(newLabel))
	} else {
		catLines = append(catLines, mutedStyle.Render(newLabel))
	}
	leftPanel := panelStyle.Width(leftWidth).Height(panelHeight).Render(
		strings.Join(catLines, "\n"),
	)

	// 右面板：skill 列表
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 30 {
		rightWidth = 30
	}

	// 过滤：按自定义分类
	filteredSkills := m.filteredSkills()

	var skillLines []string
	skillLines = append(skillLines, mutedStyle.Render(fmt.Sprintf("Skills (%d)", len(filteredSkills))))
	for _, sk := range filteredSkills {
		prefix := "  "
		if sk.Linked {
			prefix = "★ " + linkedStyle.Render("●")
		}
		if m.selected[sk.Name] {
			prefix = "✓ "
		}

		// Build line with note
		note := m.dataStore.GetNote(sk.Name)
		noteStr := ""
		if note != "" {
			noteStr = mutedStyle.Render(fmt.Sprintf(" — %s", note))
		}

		line := fmt.Sprintf("%s %s%s", prefix, sk.Name, noteStr)
		if m.panelFocus == 1 && sk.Name == m.cursorName {
			line = activeStyle.Render(fmt.Sprintf("%s %s%s", prefix, sk.Name, noteStr))
		} else if sk.Linked {
			line = linkedStyle.Render(fmt.Sprintf("%s %s%s", prefix, sk.Name, noteStr))
		} else {
			line = normalStyle.Render(fmt.Sprintf("%s %s%s", prefix, sk.Name, noteStr))
		}
		skillLines = append(skillLines, line)
	}
	rightPanel := panelStyle.Width(rightWidth).Height(panelHeight).Render(
		strings.Join(skillLines, "\n"),
	)

	// Footer — either normal shortcuts or input prompt
	var footer string
	if m.inputMode != noInput {
		// Input mode: show prompt instead of shortcuts
		var inputContent string
		switch m.inputMode {
		case addingCategory, editingNote, editingCategoryName:
			inputContent = fmt.Sprintf(" %s%s", m.inputPrompt, m.inputBuffer)
		case addToCategory:
			catNames := m.dataStore.SortedCategoryNames()
			var listLines []string
			listLines = append(listLines, " 选择分类（Enter确认）:")
			for i, n := range catNames {
				inCat := false
				for _, sn := range m.dataStore.Categories[n] {
					if sn == m.cursorName {
						inCat = true
						break
					}
				}
				status := "[+]"
				if inCat {
					status = "[-]"
				}
				var marker string
				if i == m.catSelectCursor {
					marker = fmt.Sprintf("▸ %s %s", status, n)
				} else {
					marker = fmt.Sprintf("  %s %s", status, n)
				}
				listLines = append(listLines, marker)
			}
			inputContent = strings.Join(listLines, "\n")
		}
		footer = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Padding(0, 1).
			Render(inputContent)
	} else {
		// Normal shortcuts
		footer = mutedStyle.Render(
			" [Tab]切换  [↑↓]导航  [Space]选  [Enter]预览  [L]链接  [D]取消  [/]搜索  [N]分类  [X]删  [R]改名  [E]备注  [A]分组  [Q]退出",
		)
		if m.version != "dev" {
			footer += mutedStyle.Render(fmt.Sprintf("  %s", m.version))
		}
	}

	// 错误显示 (always shown below footer)
	if m.err != nil {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(
			fmt.Sprintf("错误: %v", m.err))
	}

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