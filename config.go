package main

import (
	"fmt"
	"strings"
)

const maxFileSize = 2 * 1024 * 1024 // 2MB

// 第一层：目录黑名单（整个跳过）
var dirBlacklist = map[string]bool{
	".git": true, ".svn": true,
	"node_modules": true, "vendor": true, "bower_components": true,
	"dist": true, "bin": true, "build": true, "target": true,
	".idea": true, ".vscode": true, ".gradle": true,
	"coverage": true, "__pycache__": true, ".next": true,
	"_libs": true, ".claude": true, ".gk": true,
	".cpk-backup": true,
}

// 第二层：扩展名白名单（只保留源代码）
var extWhitelist = map[string]bool{
	// Go
	".go": true, ".mod": true, ".sum": true,
	// JavaScript / TypeScript
	".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
	// Java
	".java": true, ".jsp": true,
	// Web
	".html": true, ".htm": true, ".css": true, ".scss": true, ".less": true,
	// Data / Config
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".xml": true, ".sql": true, ".properties": true,
	".conf": true, ".cfg": true, ".ini": true, ".env": true,
	// Script
	".sh": true, ".bash": true, ".py": true,
	// Template
	".tmpl": true, ".tpl": true,
	// Modern frontend
	".svelte": true, ".vue": true, ".astro": true,
	// Other
	".md": true, ".graphql": true, ".proto": true, ".tf": true, ".hcl": true,
	// EOS
	".bizx": true,
}

// 特殊文件名（没有扩展名但应该包含）
var specialFiles = map[string]bool{
	"Makefile": true, "Dockerfile": true,
	".gitignore": true, ".dockerignore": true,
	".editorconfig": true, ".prettierrc": true,
	".eslintrc": true, ".babelrc": true, ".env.example": true,
}

// 排除的精确文件名（扩展名匹配但不是源代码）
var excludeFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"composer.lock":     true,
}

// 排除的后缀（压缩/打包产物）
var excludeSuffixes = []string{
	".min.js", ".min.css",
	".bundle.js", ".chunk.js",
	".js.map", ".css.map",
}

// shouldInclude 判断一个文件是否应该被打包
func shouldInclude(name string, size int64) bool {
	if size > maxFileSize {
		return false
	}
	if specialFiles[name] {
		return true
	}

	ext := strings.ToLower(extOf(name))
	if !extWhitelist[ext] {
		return false
	}
	if excludeFiles[name] {
		return false
	}

	lower := strings.ToLower(name)
	for _, suffix := range excludeSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return false
		}
	}
	return true
}

// extOf 返回文件扩展名（含点），比 filepath.Ext 更可靠地处理 .env 等
func extOf(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
		if name[i] == '/' {
			break
		}
	}
	return ""
}

// humanSize 把字节数转成人类可读的大小
func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return strings.TrimRight(strings.TrimRight(
			formatFloat(float64(bytes)/float64(1<<30)), "0"), ".") + "GB"
	case bytes >= 1<<20:
		return strings.TrimRight(strings.TrimRight(
			formatFloat(float64(bytes)/float64(1<<20)), "0"), ".") + "MB"
	case bytes >= 1<<10:
		return formatInt(bytes>>10) + "KB"
	default:
		return formatInt(bytes) + "B"
	}
}

func formatFloat(f float64) string {
	s := fmt.Sprintf("%.1f", f)
	return s
}

func formatInt(n int64) string {
	return fmt.Sprintf("%d", n)
}
