package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const backupDir = ".cpack-backup"

func updateProject() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取 home 目录: %w", err)
	}

	// 在 ~/Downloads/ 找最新的 zip
	zipPath, err := findLatestZip(filepath.Join(homeDir, "Downloads"))
	if err != nil {
		return err
	}

	fmt.Printf("\n📦 找到: %s\n", filepath.Base(zipPath))

	// 打开 zip，列出文件
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("无法打开 zip: %w", err)
	}
	defer r.Close()

	// 检测并去掉可能的公共前缀目录（比如 Claude 打包时套了一层文件夹）
	prefix := detectCommonPrefix(r.File)

	var targets []zipTarget
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.TrimPrefix(f.Name, prefix)
		if name == "" {
			continue
		}
		targets = append(targets, zipTarget{ZipEntry: f, DestPath: name})
	}

	if len(targets) == 0 {
		return fmt.Errorf("zip 中没有文件")
	}

	// 显示文件清单
	fmt.Printf("   包含 %d 个文件:\n\n", len(targets))
	for _, t := range targets {
		marker := "  ✚ 新增"
		if _, err := os.Stat(t.DestPath); err == nil {
			marker = "  ✎ 覆盖"
		}
		fmt.Printf("   %s  %s\n", marker, t.DestPath)
	}

	// 等待确认
	fmt.Printf("\n   按回车确认，Ctrl+C 取消: ")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	// 清理旧备份，创建新备份目录
	os.RemoveAll(backupDir)
	os.MkdirAll(backupDir, 0755)

	// 备份已有文件 → 解压覆盖
	var overwritten, added int
	for _, t := range targets {
		// 如果目标文件已存在，先备份
		if _, err := os.Stat(t.DestPath); err == nil {
			backupPath := filepath.Join(backupDir, t.DestPath)
			os.MkdirAll(filepath.Dir(backupPath), 0755)
			if err := copyFile(t.DestPath, backupPath); err != nil {
				fmt.Printf("   ⚠️  备份失败 %s: %s\n", t.DestPath, err)
				continue
			}
			overwritten++
		} else {
			added++
		}

		// 解压到目标位置
		os.MkdirAll(filepath.Dir(t.DestPath), 0755)
		if err := extractZipEntry(t.ZipEntry, t.DestPath); err != nil {
			fmt.Printf("   ⚠️  写入失败 %s: %s\n", t.DestPath, err)
		}
	}

	fmt.Printf("\n✅ 更新完成\n")
	fmt.Printf("   覆盖: %d 个文件\n", overwritten)
	fmt.Printf("   新增: %d 个文件\n", added)
	fmt.Printf("   备份: %s/\n", backupDir)
	fmt.Printf("\n   如需还原: cpack rollback\n\n")

	return nil
}

func rollbackProject() error {
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("没有找到备份（%s/ 不存在）", backupDir)
	}

	var restored int
	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(backupDir, path)
		os.MkdirAll(filepath.Dir(relPath), 0755)
		if err := copyFile(path, relPath); err != nil {
			fmt.Printf("   ⚠️  还原失败 %s: %s\n", relPath, err)
			return nil
		}
		restored++
		return nil
	})
	if err != nil {
		return fmt.Errorf("还原失败: %w", err)
	}

	os.RemoveAll(backupDir)

	fmt.Printf("\n✅ 已还原 %d 个文件\n\n", restored)
	return nil
}

// ========== 工具函数 ==========

type zipTarget struct {
	ZipEntry *zip.File
	DestPath string // 去掉公共前缀后的相对路径
}

// findLatestZip 在指定目录中找到最新的 .zip 文件
func findLatestZip(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("无法读取 %s: %w", dir, err)
	}

	type zipFile struct {
		path    string
		modTime int64
	}
	var zips []zipFile

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".zip") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		zips = append(zips, zipFile{
			path:    filepath.Join(dir, e.Name()),
			modTime: info.ModTime().UnixNano(),
		})
	}

	if len(zips) == 0 {
		return "", fmt.Errorf("~/Downloads/ 中没有找到 zip 文件")
	}

	sort.Slice(zips, func(i, j int) bool { return zips[i].modTime > zips[j].modTime })
	return zips[0].path, nil
}

// detectCommonPrefix 检测 zip 里所有文件是否共享一个顶层目录前缀
// 如果是（比如 "changes-xxx/"），返回该前缀以便去掉
func detectCommonPrefix(files []*zip.File) string {
	prefix := ""
	for _, f := range files {
		if f.FileInfo().IsDir() {
			continue
		}
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) < 2 {
			return "" // 有文件直接在根层，没有公共前缀
		}
		dir := parts[0] + "/"
		if prefix == "" {
			prefix = dir
		} else if dir != prefix {
			return "" // 不同的顶层目录
		}
	}
	return prefix
}

func extractZipEntry(f *zip.File, destPath string) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
