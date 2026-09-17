#!/bin/bash
set -e

VERSION="$1"
MESSAGE="$2"

if [ -z "$VERSION" ]; then
  echo "用法: ./release.sh 0.3.0 \"提交信息\" "
  exit 1
fi

TAG="v${VERSION}"

# 检查 tag 是否已存在
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "❌ ${TAG} 已存在"
  exit 1
fi

# 检查工作区状态
if [ -n "$(git status --porcelain)" ]; then
  # 有改动
  if [ -z "$MESSAGE" ]; then
    echo "❌ 有未提交的改动，请提供提交信息:"
    echo "   ./release.sh ${VERSION} \"你的提交信息\""
    exit 1
  fi
  git add -A
  git commit -m "$MESSAGE"
fi

git tag "$TAG"
git push
git push --tags

echo ""
echo "✅ ${TAG} 已发布，等待 Actions 构建 ..."
echo ""
