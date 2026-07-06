package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkAndUnlink(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	// 模拟中央仓库中的 skill 目录
	repoSkillDir := filepath.Join(tmpDir, "repo", "react-component")
	os.MkdirAll(repoSkillDir, 0755)
	os.WriteFile(filepath.Join(repoSkillDir, "SKILL.md"), []byte("# test"), 0644)

	l := NewLinker(tmpDir)

	// 链接
	if err := l.Link("react-component", repoSkillDir); err != nil {
		t.Fatalf("Link() 失败: %v", err)
	}

	// 检查软链接是否存在
	linkPath := filepath.Join(skillsDir, "react-component")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("读取软链接失败: %v", err)
	}
	if target != repoSkillDir {
		t.Fatalf("期望链接到 %s，实际: %s", repoSkillDir, target)
	}

	// IsLinked
	if !l.IsLinked("react-component") {
		t.Fatal("IsLinked() 应返回 true")
	}

	// ListLinked
	linked, err := l.ListLinked()
	if err != nil {
		t.Fatalf("ListLinked() 失败: %v", err)
	}
	if len(linked) != 1 || linked[0] != "react-component" {
		t.Fatalf("期望 [react-component]，实际: %v", linked)
	}

	// 取消链接
	if err := l.Unlink("react-component"); err != nil {
		t.Fatalf("Unlink() 失败: %v", err)
	}
	if l.IsLinked("react-component") {
		t.Fatal("取消链接后 IsLinked() 应返回 false")
	}
}

func TestLink_AlreadyLinked(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	repoSkillDir := filepath.Join(tmpDir, "repo", "react-component")
	os.MkdirAll(repoSkillDir, 0755)

	l := NewLinker(tmpDir)
	l.Link("react-component", repoSkillDir)

	// 重复链接应返回错误
	err := l.Link("react-component", repoSkillDir)
	if err == nil {
		t.Fatal("重复链接应返回错误")
	}
}

func TestLink_InvalidSkillPath(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".opencode", "skills"), 0755)

	l := NewLinker(tmpDir)
	err := l.Link("invalid", "/nonexistent/path")
	if err == nil {
		t.Fatal("无效路径应返回错误")
	}
}

func TestUnlink_NotLinked(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".opencode", "skills"), 0755)

	l := NewLinker(tmpDir)
	err := l.Unlink("nonexistent")
	if err == nil {
		t.Fatal("取消未链接的 skill 应返回错误")
	}
}

func TestLink_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, ".opencode", "skills")
	os.MkdirAll(skillsDir, 0755)

	names := []string{"react", "vue", "angular"}
	for _, n := range names {
		d := filepath.Join(tmpDir, "repo", n)
		os.MkdirAll(d, 0755)
	}

	l := NewLinker(tmpDir)
	for _, n := range names {
		if err := l.Link(n, filepath.Join(tmpDir, "repo", n)); err != nil {
			t.Fatalf("Link(%s) 失败: %v", n, err)
		}
	}

	linked, _ := l.ListLinked()
	if len(linked) != 3 {
		t.Fatalf("期望 3 个，实际: %d", len(linked))
	}
}