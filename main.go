package main

import "shorturl/cmd"

// 程序入口：执行 Cobra 根命令，触发子命令（如 serve）启动服务
func main() {
	cmd.Execute()
}

