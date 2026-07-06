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
			hasReact := false
			hasCSS := false
			for _, sk := range c.Skills {
				if sk.Name == "react-component" {
					hasReact = true
				}
				if sk.Name == "css-layout" {
					hasCSS = true
				}
			}
			if !hasReact {
				t.Fatal("frontend 下未找到 react-component")
			}
			if !hasCSS {
				t.Fatal("frontend 下未找到 css-layout")
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