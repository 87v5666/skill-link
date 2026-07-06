package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

func init() {
	// 子命令将在后续任务中注册
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}