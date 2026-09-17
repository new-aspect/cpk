# cpk

跟 Claude 网页版协作写代码的命令行工具。打包项目、应用改动，各一条命令。

## 安装

```bash
brew tap new-aspect/tap
brew install cpk
cpk init && source ~/.zshrc
```

没装 Homebrew？用 curl：

```bash
# Apple Silicon (M1/M2/M3)
sudo curl -L https://github.com/new-aspect/cpk/releases/latest/download/cpk-darwin-arm64 -o /usr/local/bin/cpk
sudo chmod +x /usr/local/bin/cpk
cpk init && source ~/.zshrc
```

<details><summary>Intel Mac</summary>

```bash
sudo curl -L https://github.com/new-aspect/cpk/releases/latest/download/cpk-darwin-amd64 -o /usr/local/bin/cpk
sudo chmod +x /usr/local/bin/cpk
cpk init && source ~/.zshrc
```
</details>

## 使用

```bash
cd your-project

cpk              # 打包 → 桌面出现 zip → Finder 弹出 → 拖给 Claude
cpk update       # Claude 给的 zip 下载后 → 自动覆盖到项目
cpk rollback     # 覆盖炸了 → 一键还原
```

## 为什么做这个

无法用 Claude Code，所有协作靠网页版上传压缩包。核心决策：用扩展名白名单（而非目录黑名单）过滤源代码——项目类型差异大，黑名单追不完，白名单一套规则通吃 Go/Java/JS。全量打包而非按需打包，因为操作成本比 token 成本贵。
