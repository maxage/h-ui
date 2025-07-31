# H UI 双节点支持版本安装指南

## 概述

这是 H UI 的增强版本，添加了双节点支持功能。允许管理员配置第二个节点通过主节点的 SOCKS5 代理出站，并为用户提供灵活的节点权限控制。

## 主要新功能

- **双节点支持**：配置第二节点通过主节点 SOCKS5 出站
- **用户权限控制**：为每个用户单独设置单节点或双节点权限
- **智能订阅生成**：根据用户权限动态生成订阅内容
- **配置管理**：完整的第二节点配置导入导出功能

## 快速安装

### 一键安装脚本

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh)
```

### 指定版本安装

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh) v0.0.1
```

### 环境变量自定义

如果你想从其他 GitHub 仓库安装，可以设置环境变量：

```bash
# 从自定义仓库安装
GITHUB_USER=your-username GITHUB_REPO=your-repo bash <(curl -fsSL https://raw.githubusercontent.com/your-username/your-repo/main/install.sh)
```

## 安装后配置

### 1. 访问管理面板

默认访问地址：`http://your-server-ip:8081`
默认账号：`sysadmin`
默认密码：`sysadmin`

### 2. 配置第二节点

1. 进入 **Hysteria2 配置** 页面
2. 找到 **第二节点配置** 区域
3. 启用第二节点开关
4. 配置 SOCKS5 出站设置：
   - **地址**：主节点的 SOCKS5 代理地址（如：127.0.0.1:1080）
   - **用户名**：SOCKS5 代理用户名（可选）
   - **密码**：SOCKS5 代理密码（可选）

### 3. 设置用户权限

1. 进入 **用户管理** 页面
2. 编辑用户信息
3. 在 **节点权限** 选项中选择：
   - **单节点**：用户只能访问主节点
   - **双节点**：用户可以访问主节点和第二节点

## 功能说明

### 节点架构

```
主节点 (Node1): Hysteria2 直连
├── 端口: 原有配置端口
├── 配置: 现有 HYSTERIA2_CONFIG
└── 管理: 现有 Hysteria2 实例

第二节点 (Node2): Hysteria2 + SOCKS5 出站
├── 端口: 主节点端口 + 1
├── 配置: 新增 HYSTERIA2_NODE2_CONFIG  
├── 出站: 通过主节点的 SOCKS5 代理
└── 管理: 新增 Hysteria2 实例
```

### 用户权限控制

- **单节点权限**：用户订阅只包含主节点配置
- **双节点权限**：用户订阅包含主节点和第二节点配置
- **动态权限**：第二节点禁用时，所有用户自动回退到单节点

### 配置管理

- **独立配置**：第二节点使用独立的配置文件和端口
- **错误隔离**：第二节点问题不影响主节点运行
- **配置同步**：支持配置的导入导出和备份

## 故障排除

### 第二节点无法启动

1. 检查 SOCKS5 代理配置是否正确
2. 确认端口没有被占用（主节点端口+1）
3. 查看系统日志：`journalctl -u h-ui -f`

### 用户无法选择双节点权限

1. 确认第二节点已启用
2. 检查第二节点状态是否正常
3. 验证 SOCKS5 配置是否完整

### 订阅内容不正确

1. 检查用户权限设置
2. 确认第二节点配置正确
3. 重新生成用户订阅链接

## 升级说明

从原版 H UI 升级到双节点版本：

1. 数据库会自动迁移，添加必要的字段和配置
2. 现有用户默认设置为单节点权限
3. 所有原有功能保持不变

## 技术支持

- GitHub Issues: https://github.com/maxage/h-ui/issues
- 原版项目: https://github.com/jonssonyan/h-ui

## 许可证

本项目基于原版 H UI 开发，遵循相同的 GPL-3.0 许可证。