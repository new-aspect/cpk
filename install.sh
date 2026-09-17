#!/bin/bash
set -e

REPO="new-aspect/cpk"
BINARY=""

# 检测芯片
ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64) BINARY="cpk-darwin-arm64" ;;
  x86_64)        BINARY="cpk-darwin-amd64" ;;
  *)             echo "❌ 不支持的架构: $ARCH"; exit 1 ;;
esac

GITHUB_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
MIRROR_URL="https://gh-proxy.com/${GITHUB_URL}"

echo ""
echo "⏳ 正在安装 cpk ..."
echo "   架构: ${ARCH}"

# 试 GitHub 直连（3秒超时），失败走镜像
TMP=$(mktemp)
if curl -fsSL --connect-timeout 3 -o "$TMP" "$GITHUB_URL" 2>/dev/null; then
  echo "   来源: GitHub"
else
  echo "   GitHub 超时，切换镜像 ..."
  if ! curl -fsSL --connect-timeout 10 -o "$TMP" "$MIRROR_URL" 2>/dev/null; then
    rm -f "$TMP"
    echo "❌ 下载失败，请检查网络"
    exit 1
  fi
  echo "   来源: gh-proxy.com"
fi

# 检查下载内容是否有效（至少 1MB）
SIZE=$(wc -c < "$TMP" | tr -d ' ')
if [ "$SIZE" -lt 1000000 ]; then
  rm -f "$TMP"
  echo "❌ 下载的文件不完整（${SIZE} 字节），请重试"
  exit 1
fi

# 安装
sudo mv "$TMP" /usr/local/bin/cpk
sudo chmod +x /usr/local/bin/cpk

# Tab 补全
cpk init 2>/dev/null || true

VER=$(cpk version 2>/dev/null || echo "cpk")
echo ""
echo "✅ ${VER} 已安装"
echo "   执行一次: source ~/.zshrc"
echo ""
