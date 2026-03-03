package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 根命令：shorturl
// - 作为命令行入口，聚合具体子命令（如 serve）
var rootCmd = &cobra.Command{
	Use:   "shorturl",
	Short: "短链接服务",
}

// 执行根命令：
// - 解析并运行具体子命令
// - 如运行失败，打印错误并以非零状态退出
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
