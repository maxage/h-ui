# 构建和发布指南

## 构建二进制文件

### 1. 本地构建

```bash
# 确保已安装 Go 1.20+
go version

# 构建所有平台的二进制文件
chmod +x build.sh
./build.sh
```

构建完成后，`build/` 目录下会生成所有平台的二进制文件：
- `h-ui-linux-amd64`
- `h-ui-linux-arm64`
- `h-ui-windows-amd64.exe`
- `h-ui-darwin-amd64`
- 等等...

### 2. 使用 GitHub Actions 自动构建

创建 `.github/workflows/build.yml` 文件：

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.20'
    
    - name: Build binaries
      run: |
        chmod +x build.sh
        ./build.sh
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      if: startsWith(github.ref, 'refs/tags/')
      with:
        files: build/*
        generate_release_notes: true
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## 发布 Release

### 方法1: 通过 GitHub Web 界面

1. 进入你的 GitHub 仓库
2. 点击 "Releases" 标签
3. 点击 "Create a new release"
4. 填写标签版本（如 `v1.0.0`）
5. 上传构建的二进制文件
6. 发布 Release

### 方法2: 使用 Git 命令

```bash
# 创建并推送标签
git tag v1.0.0
git push origin v1.0.0

# 如果配置了 GitHub Actions，会自动构建和发布
```

### 方法3: 使用 GitHub CLI

```bash
# 安装 GitHub CLI
# 然后创建 release
gh release create v1.0.0 build/* --title "v1.0.0" --notes "双节点支持版本"
```

## Docker 镜像构建

### 1. 构建 Docker 镜像

```bash
# 构建镜像
docker build -t maxage/h-ui:latest .
docker build -t maxage/h-ui:v1.0.0 .

# 推送到 Docker Hub
docker push maxage/h-ui:latest
docker push maxage/h-ui:v1.0.0
```

### 2. 多架构构建

```bash
# 创建 buildx builder
docker buildx create --name multiarch --use

# 构建并推送多架构镜像
docker buildx build --platform linux/amd64,linux/arm64 \
  -t maxage/h-ui:latest \
  -t maxage/h-ui:v1.0.0 \
  --push .
```

## 快速发布脚本

创建 `release.sh` 脚本：

```bash
#!/bin/bash
set -e

VERSION=${1:-"v1.0.0"}

echo "Building binaries..."
./build.sh

echo "Creating release ${VERSION}..."
gh release create ${VERSION} build/* \
  --title "${VERSION}" \
  --notes "双节点支持版本 - 详见 DUAL_NODE_INSTALL.md"

echo "Building and pushing Docker images..."
docker buildx build --platform linux/amd64,linux/arm64 \
  -t maxage/h-ui:latest \
  -t maxage/h-ui:${VERSION} \
  --push .

echo "Release ${VERSION} completed!"
```

使用方法：
```bash
chmod +x release.sh
./release.sh v1.0.0
```

## 注意事项

1. **确保代码编译通过**：发布前先本地测试
2. **更新版本号**：在代码中更新版本信息
3. **测试安装脚本**：确保新版本的安装脚本正常工作
4. **Docker Hub 权限**：确保有推送 Docker 镜像的权限
5. **GitHub Token**：确保 GitHub Actions 有足够的权限

## 临时解决方案

在你发布正式 Release 之前，当前的安装脚本会自动使用原版的二进制文件，但配置和前端代码使用你的仓库版本。这样可以确保双节点功能正常工作。