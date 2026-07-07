package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// 确保没有配置文件干扰
	t.Setenv("HOME", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 不应返回错误，got: %v", err)
	}
	if cfg.RepoPath == "" {
		t.Fatal("默认 RepoPath 不应为空")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigDir := ConfigDir
	ConfigDir = func() string { return tmpDir }
	defer func() { ConfigDir = origConfigDir }()

	cfg := &Config{RepoPath: "/tmp/test-repo"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() 失败: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}
	if loaded.RepoPath != "/tmp/test-repo" {
		t.Fatalf("期望 RepoPath=/tmp/test-repo, 实际: %s", loaded.RepoPath)
	}
}

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	subDir := filepath.Join(projectDir, "src", "components")

	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	origGetwd := getwd
	getwd = func() (string, error) { return subDir, nil }
	defer func() { getwd = origGetwd }()

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() 失败: %v", err)
	}
	if root != projectDir {
		t.Fatalf("期望 root=%s, 实际: %s", projectDir, root)
	}
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	origGetwd := getwd
	getwd = func() (string, error) { return tmpDir, nil }
	defer func() { getwd = origGetwd }()

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() 不应返回错误，got: %v", err)
	}
	if root != tmpDir {
		t.Fatalf("期望 root=%s, 实际: %s", tmpDir, root)
	}
}