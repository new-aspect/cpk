# 改动说明

全量重命名 cpack → cpk，涉及 7 个文件：

- main.go — 命令输出、帮助文字、Tab 补全脚本
- pack.go — 输出文件名 cpk-项目名.zip、CPACK_FOR_CLAUDE.md 内容
- update.go — 备份目录名 .cpk-backup、提示文字
- config.go — 备份目录名加入黑名单
- README.md — 安装命令、使用示例
- .gitignore — 二进制文件名
- .github/workflows/release.yml — 构建产物名、安装命令
