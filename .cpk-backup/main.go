package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Version 由构建时 -ldflags 注入，开发时显示 "dev"
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		if err := packProject(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "update":
		if err := updateProject(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s\n", err)
			os.Exit(1)
		}
	case "rollback":
		if err := rollbackProject(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s\n", err)
			os.Exit(1)
		}
	case "version", "-v", "--version":
		fmt.Printf("cpk %s\n", Version)
	case "init":
		if err := initCompletion(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`cpk %s — 项目压缩与协作工具

用法:
  cpk              打包当前项目到桌面，弹出 Finder
  cpk update       应用 ~/Downloads/ 中最新的改动包
  cpk rollback     还原上次 update 的改动
  cpk version      显示版本号
  cpk init         安装 Tab 补全
  cpk help         显示此帮助

`, Version)
}

// initCompletion 安装 zsh Tab 补全
func initCompletion() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取 home 目录: %w", err)
	}

	// 创建补全目录
	compDir := filepath.Join(homeDir, ".zsh", "completions")
	if err := os.MkdirAll(compDir, 0755); err != nil {
		return fmt.Errorf("无法创建目录 %s: %w", compDir, err)
	}

	// 写入补全脚本
	compFile := filepath.Join(compDir, "_cpk")
	if err := os.WriteFile(compFile, []byte(zshCompletion), 0644); err != nil {
		return fmt.Errorf("无法写入补全文件: %w", err)
	}

	// 检查 .zshrc 是否已配置
	zshrc := filepath.Join(homeDir, ".zshrc")
	content, _ := os.ReadFile(zshrc)
	contentStr := string(content)

	var lines []string
	if !strings.Contains(contentStr, ".zsh/completions") {
		lines = append(lines, `fpath=(~/.zsh/completions $fpath)`)
	}
	if !strings.Contains(contentStr, "compinit") {
		lines = append(lines, `autoload -Uz compinit && compinit`)
	}

	if len(lines) > 0 {
		f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("无法写入 .zshrc: %w", err)
		}
		defer f.Close()
		f.WriteString("\n# cpk Tab 补全\n")
		for _, line := range lines {
			f.WriteString(line + "\n")
		}
	}

	fmt.Printf("\n✅ Tab 补全已安装\n")
	fmt.Printf("   执行一次: source ~/.zshrc\n")
	fmt.Printf("   之后输入 cpk + Tab 即可补全子命令\n\n")
	return nil
}

const zshCompletion = `#compdef cpk

_cpk() {
    local -a commands
    commands=(
        'update:应用 ~/Downloads/ 中最新的改动包'
        'rollback:还原上次 update 的改动'
        'version:显示版本号'
        'init:安装 Tab 补全'
        'help:显示帮助'
    )
    _describe 'command' commands
}

compdef _cpk cpk
`
