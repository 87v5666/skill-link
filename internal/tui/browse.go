package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"skill-management/internal/config"
	"skill-management/internal/repo"
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