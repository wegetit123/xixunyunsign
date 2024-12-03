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

	var rootCmd = &cobra.Command{Use: "xixun"}
	rootCmd.AddCommand(cmd.LoginCmd)
	rootCmd.AddCommand(cmd.QueryCmd)
	rootCmd.AddCommand(cmd.SignCmd)
	rootCmd.Execute()
}
