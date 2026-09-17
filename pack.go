package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type fileEntry struct {
	Path string // 相对于项目根目录的路径
	Size int64
}

func packProject() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("无法获取当前目录: %w", err)
	}
	projectName := filepath.Base(projectDir)

	fmt.Printf("\n⏳ 正在扫描 %s ...\n", projectName)

	// 扫描并收集命中的文件
	var files []fileEntry
	var totalSize int64

	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		relPath, _ := filepath.Rel(projectDir, path)
		if relPath == "." {
			return nil
		}

		// 目录黑名单
		if info.IsDir() {
			if dirBlacklist[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// 白名单 + 排除规则 + 大小限制
		if shouldInclude(info.Name(), info.Size()) {
			files = append(files, fileEntry{Path: relPath, Size: info.Size()})
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("扫描失败: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	// 创建 zip
	homeDir, _ := os.UserHomeDir()
	zipPath := filepath.Join(homeDir, "Desktop", fmt.Sprintf("cpk-%s.zip", projectName))

	if err := writeZip(zipPath, projectDir, files, projectName, totalSize); err != nil {
		return fmt.Errorf("打包失败: %w", err)
	}

	zipInfo, _ := os.Stat(zipPath)

	fmt.Printf("\n✅ 打包完成\n")
	fmt.Printf("   文件数:   %d\n", len(files))
	fmt.Printf("   原始大小: %s\n", humanSize(totalSize))
	fmt.Printf("   ZIP 大小: %s\n", humanSize(zipInfo.Size()))
	fmt.Printf("   路径:     %s\n\n", zipPath)

	// macOS: 弹出 Finder 并选中文件
	exec.Command("open", "-R", zipPath).Run()

	return nil
}

func writeZip(zipPath, projectDir string, files []fileEntry, projectName string, totalSize int64) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// 写入源代码文件
	for _, entry := range files {
		fullPath := filepath.Join(projectDir, entry.Path)
		if err := addToZip(w, fullPath, entry.Path); err != nil {
			continue // 跳过读不了的文件
		}
	}

	// 写入 PROJECT_MAP.md
	mapContent := buildProjectMap(projectName, projectDir, files, totalSize)
	if mw, err := w.Create("PROJECT_MAP.md"); err == nil {
		mw.Write([]byte(mapContent))
	}

	// 写入 CPACK_FOR_CLAUDE.md
	if cw, err := w.Create("CPACK_FOR_CLAUDE.md"); err == nil {
		cw.Write([]byte(claudeReadme))
	}

	return nil
}

func addToZip(w *zip.Writer, srcPath, zipPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := w.Create(zipPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

// ========== PROJECT_MAP.md 生成 ==========

func buildProjectMap(project, projectDir string, files []fileEntry, totalSize int64) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", project))
	b.WriteString(fmt.Sprintf("扫描: %s | 文件数: %d | 大小: %s\n\n",
		time.Now().Format("2006-01-02 15:04"),
		len(files), humanSize(totalSize)))

	// 一级目录大小
	b.WriteString("## 一级目录大小\n\n```\n")
	type dirEntry struct {
		Name string
		Size int64
	}
	var topDirs []dirEntry

	entries, _ := os.ReadDir(projectDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		size := calcDirSize(filepath.Join(projectDir, e.Name()))
		topDirs = append(topDirs, dirEntry{e.Name(), size})
	}
	sort.Slice(topDirs, func(i, j int) bool { return topDirs[i].Size > topDirs[j].Size })
	for _, d := range topDirs {
		b.WriteString(fmt.Sprintf("  %8s  %s/\n", humanSize(d.Size), d.Name))
	}
	b.WriteString("```\n\n")

	// 目录结构（3层）
	b.WriteString("## 目录结构\n\n```\n")
	dirs := collectDirs(projectDir, 3)
	for _, d := range dirs {
		b.WriteString(d + "\n")
	}
	b.WriteString("```\n\n")

	// 文件清单
	b.WriteString("## 文件清单\n\n```\n")
	for _, f := range files {
		b.WriteString(fmt.Sprintf("  %8s  %s\n", humanSize(f.Size), f.Path))
	}
	b.WriteString("```\n")

	return b.String()
}

func calcDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})
	return size
}

func collectDirs(root string, maxDepth int) []string {
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			dirs = append(dirs, ".")
			return nil
		}
		depth := strings.Count(rel, string(os.PathSeparator)) + 1
		if depth > maxDepth {
			return filepath.SkipDir
		}
		if dirBlacklist[info.Name()] {
			return filepath.SkipDir
		}
		dirs = append(dirs, "./"+rel)
		return nil
	})
	sort.Strings(dirs)
	return dirs
}

// ========== CPACK_FOR_CLAUDE.md ==========

const claudeReadme = `# cpk 协作说明

这是通过 cpk 工具打包的项目源代码，只包含源代码文件。
完整的项目结构和文件清单见 PROJECT_MAP.md。

## 给 Claude 的约定

当你完成代码修改后，请将改动的文件打成 zip 包给用户下载。

要求：
1. 只放改动过的文件，不需要全量
2. zip 内的目录结构必须与项目一致
   例如改了 pkg/server/config.go → zip 里就是 pkg/server/config.go
3. 不要在 zip 外面套额外的文件夹

用户会在项目目录执行 cpk update，自动应用你的改动。
`
