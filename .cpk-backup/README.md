# cpk

项目压缩与协作工具。给 Claude 网页版发代码、应用改动，各一条命令。

## 安装

**Apple Silicon (M1/M2/M3):**

```bash
sudo curl -L https://github.com/new-aspect/cpack/releases/latest/download/cpk-darwin-arm64 -o /usr/local/bin/cpk
sudo chmod +x /usr/local/bin/cpk
cpk init
source ~/.zshrc
```

**Intel Mac:**

```bash
sudo curl -L https://github.com/new-aspect/cpack/releases/latest/download/cpk-darwin-amd64 -o /usr/local/bin/cpk
sudo chmod +x /usr/local/bin/cpk
cpk init
source ~/.zshrc
```

## 使用

```bash
cd your-project

cpk              # 打包 → 桌面出现 zip → Finder 弹出 → 拖给 Claude
cpk update       # Claude 给的 zip 下载后 → 自动覆盖到项目
cpk rollback     # 覆盖炸了 → 一键还原
cpk version      # 版本号
cpk init         # 安装 Tab 补全（只需跑一次）
```

## 协作流程

```
你                              Claude
│                                │
├── cpack ──────────────────────→│ 收到完整源码 + PROJECT_MAP.md
│                                │ 改代码
│← 下载 zip ←───────────────────┤ 给你改动包
├── cpk update                 │
│   (自动备份 → 覆盖)            │
│                                │
├── 跑不起来了？                  │
│   贴报错给 Claude ────────────→│ 不用重新发包
│                                │
├── cpk rollback               │
│   (一键还原)                   │
```

## 打包策略

三层过滤，只保留源代码：

1. **目录黑名单** — .git, node_modules, vendor, dist, bin, .idea 等
2. **扩展名白名单** — .go .js .ts .java .html .css .json .yaml .sql .py .md 等
3. **2MB 上限** — 超大文件通常是生成物，跳过

## 为什么这样做

无法用 Claude Code，所有协作靠网页版上传压缩包。核心决策：用扩展名白名单（而非目录黑名单）过滤源代码——因为项目类型差异大，黑名单追不完，白名单一套规则通吃 Go/Java/JS。全量打包而非按需打包，因为操作成本比 token 成本贵。