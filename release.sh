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

# 构建二进制文件
echo "📦 构建二进制文件..."
chmod +x build.sh
./build.sh

# 检查构建结果
if [[ ! -d "build" ]] || [[ -z "$(ls -A build)" ]]; then
    echo "❌ 构建失败，build 目录为空"
    exit 1
fi

echo "✅ 构建完成，生成的文件："
ls -la build/

# 创建并推送标签
echo "🏷️  创建标签 ${VERSION}..."
git tag ${VERSION}
git push origin ${VERSION}

echo "⏳ 等待 GitHub Actions 自动构建和发布..."
echo "📍 你可以在这里查看进度：https://github.com/maxage/h-ui/actions"
echo "📍 发布完成后可以在这里查看：https://github.com/maxage/h-ui/releases"

echo ""
echo "🎉 发布流程已启动！"
echo "📋 安装命令："
echo "   bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh)"