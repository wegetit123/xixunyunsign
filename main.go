package main

import (
	"github.com/spf13/cobra"
	"xixuanyunsign/cmd"
	"xixuanyunsign/utils"
)

func main() {
	// 初始化数据库
	err := utils.InitDB()
	if err != nil {
		panic(err)
	}
	// 设置根命令
	var rootCmd = &cobra.Command{Use: "xixun"}
	rootCmd.AddCommand(cmd.LoginCmd)
	rootCmd.AddCommand(cmd.QueryCmd)
	rootCmd.AddCommand(cmd.SignCmd)
	rootCmd.AddCommand(cmd.SchoolSearchIDCmd)

	// 执行根命令
	rootCmd.Execute()
}
