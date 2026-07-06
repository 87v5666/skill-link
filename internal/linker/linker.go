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

	if err := os.Symlink(skillPath, linkPath); err != nil {
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