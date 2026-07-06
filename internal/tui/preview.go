package tui

import (
	"fmt"

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