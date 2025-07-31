#!/bin/bash
set -e

echo "🚀 开始完整构建流程..."

# 检查 Node.js 和 npm
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ npm 未安装，请先安装 npm"
    exit 1
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

echo "📦 构建前端..."
cd frontend

# 安装依赖
if [[ ! -d "node_modules" ]]; then
    echo "📥 安装前端依赖..."
    npm install
fi

# 构建前端
echo "🔨 构建前端代码..."
npm run build:prod

# 检查构建结果
if [[ ! -d "dist" ]]; then
    echo "❌ 前端构建失败，dist 目录不存在"
    exit 1
fi

echo "✅ 前端构建完成"

# 返回根目录
cd ..

echo "🔨 构建后端..."
chmod +x build.sh
./build.sh

echo "✅ 构建完成！"
echo "📋 生成的文件："
ls -la build/