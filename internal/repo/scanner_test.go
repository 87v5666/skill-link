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
	s.SetCacheDir(t.TempDir()) // 隔离缓存目录，不写入真实配置目录

	if err := s.BuildCache(); err != nil {
		t.Fatalf("BuildCache() 失败: %v", err)
	}

	skills, err := CachedSkills(s.cacheDir)
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

// createFlatTestRepo 创建扁平结构的测试仓库（每级目录直接是 skill）
func createFlatTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// docx/  — 是 skill，内含 SKILL.md + scripts/ 子目录
	docxDir := filepath.Join(dir, "docx")
	os.MkdirAll(docxDir, 0755)
	os.WriteFile(filepath.Join(docxDir, "SKILL.md"), []byte("# Docx\n\nWord document generation"), 0644)
	os.MkdirAll(filepath.Join(docxDir, "scripts"), 0755) // 子目录，不含 SKILL.md

	// grill-me/ — 是 skill
	os.MkdirAll(filepath.Join(dir, "grill-me"), 0755)
	os.WriteFile(filepath.Join(dir, "grill-me", "SKILL.md"), []byte("# Grill Me\n\nGrilling interviews"), 0644)

	return dir
}

func TestScanner_FlatStructure(t *testing.T) {
	repoDir := createFlatTestRepo(t)
	s := NewScanner(repoDir)
	categories, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() 失败: %v", err)
	}

	// 应该有一个"未分类"组，包含 2 个 skill
	if len(categories) != 1 {
		t.Fatalf("期望 1 个未分类组，实际: %d 个分类", len(categories))
	}
	if categories[0].Name != "未分类" {
		t.Fatalf("期望组名「未分类」，实际: %s", categories[0].Name)
	}
	if len(categories[0].Skills) != 2 {
		t.Fatalf("期望 2 个 skill，实际: %d", len(categories[0].Skills))
	}

	// 检查 scripts/ 没有被误认为 skill
	for _, sk := range categories[0].Skills {
		if sk.Name == "scripts" {
			t.Fatal("scripts/ 不应被识别为 skill")
		}
	}
}

func TestScanner_MixedStructure(t *testing.T) {
	// 混合结构：既有扁平 skill，也有分层的分类
	dir := t.TempDir()

	// 扁平：docx/（含 SKILL.md + scripts/）
	os.MkdirAll(filepath.Join(dir, "docx"), 0755)
	os.WriteFile(filepath.Join(dir, "docx", "SKILL.md"), []byte("# Docx"), 0644)
	os.MkdirAll(filepath.Join(dir, "docx", "scripts"), 0755)

	// 分层：frontend/react-component/ + frontend/css-layout/
	os.MkdirAll(filepath.Join(dir, "frontend", "react-component"), 0755)
	os.WriteFile(filepath.Join(dir, "frontend", "react-component", "SKILL.md"), []byte("# React"), 0644)
	os.MkdirAll(filepath.Join(dir, "frontend", "css-layout"), 0755)
	os.WriteFile(filepath.Join(dir, "frontend", "css-layout", "SKILL.md"), []byte("# CSS"), 0644)

	s := NewScanner(dir)
	categories, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan() 失败: %v", err)
	}

	// 应该有 2 个组：未分类（docx）+ frontend（react, css）
	if len(categories) != 2 {
		t.Fatalf("期望 2 个分组，实际: %d", len(categories))
	}

	// 第一个是未分类（docx）
	if categories[0].Name != "未分类" || len(categories[0].Skills) != 1 {
		t.Fatalf("未分类组异常: name=%s, skills=%d", categories[0].Name, len(categories[0].Skills))
	}

	// 第二个是 frontend（react, css）
	if categories[1].Name != "frontend" || len(categories[1].Skills) != 2 {
		t.Fatalf("frontend 组异常: name=%s, skills=%d", categories[1].Name, len(categories[1].Skills))
	}
}