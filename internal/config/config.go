package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	RepoPath string `json:"repo_path"`
}

// ConfigDir 返回配置目录路径（设计为可被测试覆盖的变量）
var ConfigDir = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "slink")
}

// DefaultRepoPath 返回默认仓库路径
func DefaultRepoPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode", "skills-repo")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load 从配置文件加载配置，如果文件不存在则返回默认配置
func Load() (*Config, error) {
	cfg := &Config{
		RepoPath: DefaultRepoPath(),
	}
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return cfg, nil
}

// Save 保存配置到文件
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(configPath(), data, 0644); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	return nil
}

// getwd 是 os.Getwd 的可 mock 变量
var getwd = os.Getwd

// FindProjectRoot 从当前目录向上查找包含 .opencode/ 的项目根目录
func FindProjectRoot() (string, error) {
	dir, err := getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 .opencode/ 目录（不在 OpenCode 项目中）")
		}
		dir = parent
	}
}