package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"skill-management/internal/config"
	"skill-management/internal/linker"
	"skill-management/internal/repo"
)

var rootCmd = &cobra.Command{
	Use:   "skill-mgr",
	Short: "管理 OpenCode skill 链接的工具",
	Long: `skill-mgr — 从中央仓库浏览、选择、链接 skill 到当前项目的 .opencode/skills/

首次使用前请先配置仓库路径:
  skill-mgr path set ~/my-skills-repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 无参数时打印帮助
		return cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已链接的 skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		linked, err := l.ListLinked()
		if err != nil {
			return err
		}
		if len(linked) == 0 {
			fmt.Println("当前项目没有链接任何 skill")
			return nil
		}
		fmt.Println("已链接的 skill:")
		for _, name := range linked {
			fmt.Printf("  • %s\n", name)
		}
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "查看 skill 详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		skill, err := s.GetSkill(args[0])
		if err != nil {
			return err
		}
		// 读取 SKILL.md
		data, err := os.ReadFile(filepath.Join(skill.Path, "SKILL.md"))
		if err != nil {
			return fmt.Errorf("读取 SKILL.md 失败: %w", err)
		}
		fmt.Printf("名称: %s\n分类: %s\n路径: %s\n\n%s\n",
			skill.Name, skill.Category, skill.Path, string(data))
		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "链接指定 skill 到当前项目",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		skill, err := s.GetSkill(args[0])
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Link(skill.Name, skill.Path); err != nil {
			return err
		}
		fmt.Printf("✅ skill %q 已链接到当前项目\n", skill.Name)
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "移除已链接的 skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}
		l := linker.NewLinker(projectRoot)
		if err := l.Unlink(args[0]); err != nil {
			return err
		}
		fmt.Printf("✅ skill %q 已移除\n", args[0])
		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "刷新仓库缓存索引",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		s := repo.NewScanner(cfg.RepoPath)
		if err := s.BuildCache(); err != nil {
			return fmt.Errorf("刷新缓存失败: %w", err)
		}
		fmt.Println("✅ 缓存已刷新")
		return nil
	},
}

var pathCmd = &cobra.Command{
	Use:   "path [get|set <path>]",
	Short: "查看或设置仓库路径",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 默认 get
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println(cfg.RepoPath)
		return nil
	},
}

var pathSetCmd = &cobra.Command{
	Use:   "set <path>",
	Short: "设置仓库路径",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		// 验证路径存在
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("路径不存在: %s", absPath)
		}
		cfg := &config.Config{RepoPath: absPath}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("✅ 仓库路径已设置为: %s\n", absPath)
		return nil
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "启动 TUI 交互界面浏览/选择 skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("TUI 尚未实现，将在后续版本中提供")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(tuiCmd)
	pathCmd.AddCommand(pathSetCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}