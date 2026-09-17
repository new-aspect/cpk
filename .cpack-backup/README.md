# cpack

项目压缩与协作工具。给 Claude 网页版发代码、应用改动，各一条命令。

## 安装

**Apple Silicon (M1/M2/M3):**

```bash
sudo curl -L https://github.com/new-aspect/cpack/releases/latest/download/cpack-darwin-arm64 -o /usr/local/bin/cpack
sudo chmod +x /usr/local/bin/cpack
cpack init
source ~/.zshrc
```

**Intel Mac:**

```bash
sudo curl -L https://github.com/new-aspect/cpack/releases/latest/download/cpack-darwin-amd64 -o /usr/local/bin/cpack
sudo chmod +x /usr/local/bin/cpack
cpack init
source ~/.zshrc
```

## 使用

```bash
cd your-project

cpack              # 打包 → 桌面出现 zip → Finder 弹出 → 拖给 Claude
cpack update       # Claude 给的 zip 下载后 → 自动覆盖到项目
cpack rollback     # 覆盖炸了 → 一键还原
cpack version      # 版本号
cpack init         # 安装 Tab 补全（只需跑一次）
```

## 协作流程

```
你                              Claude
│                                │
├── cpack ──────────────────────→│ 收到完整源码 + PROJECT_MAP.md
│                                │ 改代码
│← 下载 zip ←───────────────────┤ 给你改动包
├── cpack update                 │
│   (自动备份 → 覆盖)            │
│                                │
├── 跑不起来了？                  │
│   贴报错给 Claude ────────────→│ 不用重新发包
│                                │
├── cpack rollback               │
│   (一键还原)                   │
```

## 打包策略

三层过滤，只保留源代码：

1. **目录黑名单** — .git, node_modules, vendor, dist, bin, .idea 等
2. **扩展名白名单** — .go .js .ts .java .html .css .json .yaml .sql .py .md 等
3. **2MB 上限** — 超大文件通常是生成物，跳过
