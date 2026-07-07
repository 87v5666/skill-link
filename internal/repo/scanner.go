package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skill-management/internal/config"
)

// Scanner 扫描技能仓库目录
type Scanner struct {
	repoPath   string
	cacheDir   string     // 可选；覆盖默认缓存目录（用于测试隔离）
	categories []Category // 内存缓存
	scanned    bool       // 是否已扫描
}

// NewScanner 创建 Scanner 实例
func NewScanner(repoPath string) *Scanner {
	return &Scanner{repoPath: repoPath}
}

// SetCacheDir 覆盖默认缓存目录（用于测试隔离）
func (s *Scanner) SetCacheDir(dir string) {
	s.cacheDir = dir
}

// cachePath 返回缓存文件路径
func (s *Scanner) cachePath() string {
	if s.cacheDir != "" {
		return filepath.Join(s.cacheDir, "cache.json")
	}
	return filepath.Join(config.ConfigDir(), "cache.json")
}

// hasSKILLMD 检查目录下是否包含 SKILL.md 文件
func hasSKILLMD(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "SKILL.md"))
	return err == nil
}

// scanDisk 扫描仓库目录，返回分类树（实际的磁盘 I/O）
// 支持两种目录结构：
//   扁平结构：每级目录直接是 skill（含 SKILL.md）
//   分层结构：分类 → skill 子目录 → SKILL.md
func (s *Scanner) scanDisk() ([]Category, error) {
	entries, err := os.ReadDir(s.repoPath)
	if err != nil {
		return nil, fmt.Errorf("读取仓库目录失败: %w", err)
	}

	var uncategorized []Skill
	var categories []Category
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		catPath := filepath.Join(s.repoPath, entry.Name())

		if hasSKILLMD(catPath) {
			// 扁平模式：此目录本身就是一个 skill
			uncategorized = append(uncategorized, Skill{
				Name:        entry.Name(),
				Category:    "",
				Path:        catPath,
				Description: readDescription(catPath),
			})
			continue
		}

		// 分层模式：按分类 → skill 子目录扫描
		skillDirs, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}
		cat := Category{Name: entry.Name()}
		for _, sd := range skillDirs {
			if !sd.IsDir() || strings.HasPrefix(sd.Name(), ".") {
				continue
			}
			skillPath := filepath.Join(catPath, sd.Name())
			if !hasSKILLMD(skillPath) {
				continue
			}
			cat.Skills = append(cat.Skills, Skill{
				Name:        sd.Name(),
				Category:    entry.Name(),
				Path:        skillPath,
				Description: readDescription(skillPath),
			})
		}
		if len(cat.Skills) > 0 {
			categories = append(categories, cat)
		}
	}
	// 把扁平结构的 skill 放到"未分类"组
	if len(uncategorized) > 0 {
		categories = append([]Category{{
			Name:   "未分类",
			Skills: uncategorized,
		}}, categories...)
	}
	return categories, nil
}

// getCategories 返回分类树（带内存缓存，避免重复磁盘 I/O）
func (s *Scanner) getCategories() ([]Category, error) {
	if s.scanned {
		// 返回防御性副本，防止调用方修改内部缓存
		result := make([]Category, len(s.categories))
		copy(result, s.categories)
		return result, nil
	}
	cats, err := s.scanDisk()
	if err != nil {
		return nil, err
	}
	s.categories = cats
	s.scanned = true
	// 返回防御性副本
	result := make([]Category, len(s.categories))
	copy(result, s.categories)
	return result, nil
}

// Scan 扫描仓库目录，返回分类树
func (s *Scanner) Scan() ([]Category, error) {
	return s.getCategories()
}

// GetSkill 按名称查找 skill（不区分分类）
func (s *Scanner) GetSkill(name string) (*Skill, error) {
	categories, err := s.getCategories()
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
	cachePath := s.cachePath()
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}
	return os.WriteFile(cachePath, data, 0644)
}

// CachedSkills 从指定缓存目录读取所有 skill（扁平列表）
func CachedSkills(cacheDir string) ([]Skill, error) {
	cachePath := filepath.Join(cacheDir, "cache.json")
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