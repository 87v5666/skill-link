package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Store 持久化数据：自定义分类和备注
type Store struct {
	Categories map[string][]string `json:"categories"` // 分类名 -> skill 名称列表
	Notes      map[string]string   `json:"notes"`       // skill 名称 -> 备注
}

var storePath = func() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "slink")
	return filepath.Join(dir, "data.json")
}

func Load() (*Store, error) {
	s := &Store{
		Categories: make(map[string][]string),
		Notes:      make(map[string]string),
	}
	data, err := os.ReadFile(storePath())
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("解析数据失败: %w", err)
	}
	return s, nil
}

func Save(s *Store) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "slink")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storePath(), data, 0644)
}

func (s *Store) CreateCategory(name string) error {
	if name == "" {
		return fmt.Errorf("分类名不能为空")
	}
	if _, ok := s.Categories[name]; ok {
		return fmt.Errorf("分类 %q 已存在", name)
	}
	s.Categories[name] = []string{}
	return Save(s)
}

func (s *Store) DeleteCategory(name string) error {
	delete(s.Categories, name)
	return Save(s)
}

func (s *Store) RenameCategory(oldName, newName string) error {
	if newName == "" {
		return fmt.Errorf("分类名不能为空")
	}
	if _, ok := s.Categories[newName]; ok {
		return fmt.Errorf("分类 %q 已存在", newName)
	}
	skills := s.Categories[oldName]
	if skills == nil {
		return fmt.Errorf("分类 %q 不存在", oldName)
	}
	delete(s.Categories, oldName)
	s.Categories[newName] = skills
	return Save(s)
}

func (s *Store) AddSkillToCategory(category, skillName string) error {
	list := s.Categories[category]
	for _, n := range list {
		if n == skillName {
			return nil // 已在其中
		}
	}
	s.Categories[category] = append(list, skillName)
	return Save(s)
}

func (s *Store) RemoveSkillFromCategory(category, skillName string) {
	list := s.Categories[category]
	var newList []string
	for _, n := range list {
		if n != skillName {
			newList = append(newList, n)
		}
	}
	s.Categories[category] = newList
	Save(s) // ignore error
}

func (s *Store) SetNote(skillName, note string) error {
	if note == "" {
		delete(s.Notes, skillName)
	} else {
		s.Notes[skillName] = note
	}
	return Save(s)
}

func (s *Store) GetNote(skillName string) string {
	return s.Notes[skillName]
}

// SortedCategoryNames returns sorted category names
func (s *Store) SortedCategoryNames() []string {
	var names []string
	for n := range s.Categories {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}