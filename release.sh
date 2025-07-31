#!/bin/bash
set -e

# 检查参数
VERSION=${1:-"v1.0.0"}

echo "🚀 开始发布 H UI ${VERSION}..."

# 检查是否有未提交的更改
if [[ -n $(git status --porcelain) ]]; then
    echo "❌ 有未提交的更改，请先提交所有更改"
    git status
    exit 1
fi

# 创建并推送标签（GitHub Actions 会自动构建）
echo "🏷️  创建标签 ${VERSION}..."
git tag ${VERSION}
git push origin ${VERSION}

echo "⏳ GitHub Actions 正在自动构建和发布..."
echo "📍 查看构建进度：https://github.com/maxage/h-ui/actions"
echo "📍 发布完成后查看：https://github.com/maxage/h-ui/releases"

echo ""
echo "🎉 发布流程已启动！"
echo "📋 安装命令："
echo "   bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh)"
echo ""
echo "⏰ 预计 5-10 分钟后完成构建和发布"