<div align="center">

<a href="https://github.com/jonssonyan/h-ui"><img src="./docs/images/head-cover.png" alt="H UI" width="150" /></a>

<h1 align="center">H UI</h1>

[English](README.md) / 简体中文

仅仅是 Hysteria2 的面板

<p>
<a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://img.shields.io/github/license/jonssonyan/h-ui" alt="License: GPL-3.0"></a>
<a href="https://github.com/jonssonyan/h-ui/stargazers"><img src="https://img.shields.io/github/stars/jonssonyan/h-ui" alt="GitHub stars"></a>
<a href="https://github.com/jonssonyan/h-ui/forks"><img src="https://img.shields.io/github/forks/jonssonyan/h-ui" alt="GitHub forks"></a>
<a href="https://github.com/jonssonyan/h-ui/releases"><img src="https://img.shields.io/github/v/release/jonssonyan/h-ui" alt="GitHub release"></a>
<a href="https://hub.docker.com/r/jonssonyan/h-ui"><img src="https://img.shields.io/docker/pulls/jonssonyan/h-ui" alt="Docker pulls"></a>
</p>

![cover](./docs/images/cover.png)

</div>

## 主要功能

- 轻量级、资源占用低、易于部署
- 监控系统状态和 Hysteria2 状态
- 限制用户流量、用户在线状态、强制用户下线、在线用户数、重设用户流量
- 限制用户同时在线设备数、在线设备数量
- 用户订阅链接、节点URL、导入和导出用户
- 双节点支持：支持配置第二节点通过主节点SOCKS5出站，用户可单独控制节点权限
- 管理 Hysteria2 配置和 Hysteria2 版本、端口跳跃
- 更改 Web 端口、修改 Hysteria2 流量倍数
- Telegram 通知
- 查看、导入和导出系统日志和 Hysteria2 日志
- 多国语言支持: English, 简体中文
- 页面适配、支持夜间模式、自定义页面主题
- 更多功能等待你发现

## 建议系统

系统: CentOS 8+/Ubuntu 20+/Debian 11+

CPU: x86_64/amd64 arm64/aarch64

内存: ≥ 256MB

## 部署

### 快速安装 (推荐)

**双节点支持版本安装（推荐）：**

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh)
```

安装[自定义版本](https://github.com/maxage/h-ui/releases)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/maxage/h-ui/main/install.sh) v0.0.1
```

**原版安装：**

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/jonssonyan/h-ui/main/install.sh)
```

### systemd

下载可执行文件: https://github.com/jonssonyan/h-ui/releases

```bash
mkdir -p /usr/local/h-ui/
curl -fsSL https://github.com/jonssonyan/h-ui/releases/latest/download/h-ui-linux-amd64 -o /usr/local/h-ui/h-ui && chmod +x /usr/local/h-ui/h-ui
curl -fsSL https://raw.githubusercontent.com/jonssonyan/h-ui/main/h-ui.service -o /etc/systemd/system/h-ui.service
# 自定义 Web 端口，默认 8081
# sed -i "s|^ExecStart=.*|ExecStart=/usr/local/h-ui/h-ui -p 8081|" "/etc/systemd/system/h-ui.service"
systemctl daemon-reload
systemctl enable h-ui
systemctl restart h-ui
```

卸载

```bash
systemctl stop h-ui
rm -rf /etc/systemd/system/h-ui.service /usr/local/h-ui/
```

### 容器部署

1. 安装 Docker

   https://docs.docker.com/engine/install/

   ```bash
   bash <(curl -fsSL https://get.docker.com)
   ```

2. 启动容器

   ```bash
   docker pull jonssonyan/h-ui

   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     jonssonyan/h-ui
   ```

   自定义 Web 端口，默认 8081

   ```bash
   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     jonssonyan/h-ui \
     ./h-ui -p 8081
   ```

   设置时区，默认 Asia/Shanghai

   ```bash
   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -e TZ=Asia/Shanghai \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     jonssonyan/h-ui
   ```

卸载

```bash
docker rm -f h-ui
docker rmi jonssonyan/h-ui
rm -rf /h-ui
```

## 默认安装信息

- 面板端口: 8081
- SSH 本地转发端口: 8082
- 登录用户名/密码: 随机6位字符
- 连接密码: {登录用户名}.{登录密码}

## 系统升级

在管理后台将用户、系统配置、Hysteria2 配置导出，重新部署最新版的 h-ui，部署完成之后在管理后台将数据导入

## 双节点配置

### 功能说明

双节点功能允许您配置一个通过主 Hysteria2 节点作为 SOCKS5 出站的第二节点，为用户提供更多连接选择。

### 配置步骤

1. **启用第二节点**
   - 在 Hysteria 管理页面找到"第二节点配置"区域
   - 开启"启用第二节点"开关
   - 设置节点备注名称（可选）

2. **配置 SOCKS5 出站**
   - 填写 SOCKS5 代理地址（格式：host:port）
   - 如需要，填写 SOCKS5 用户名和密码
   - 点击"保存 SOCKS5 配置"

3. **用户权限设置**
   - 在账户管理中编辑用户
   - 在"节点权限"中选择：
     - 单节点：用户只能使用主节点
     - 双节点：用户可以使用主节点和第二节点

### 注意事项

- 第二节点端口自动设置为主节点端口+1
- 第二节点通过主节点的 SOCKS5 代理进行出站连接
- 只有启用第二节点后，用户才能选择双节点权限
- 第二节点故障不会影响主节点正常运行

## 常见问题

[简体中文 > 常见问题](./docs/FAQ_ZH.md)

## 性能优化

- 定时重启服务器

    ```bash
    0 4 * * * /sbin/reboot
    ```

- 安装网络加速
    - [TCP Brutal](https://github.com/apernet/tcp-brutal) (推荐)
    - [teddysun/across#bbrsh](https://github.com/teddysun/across#bbrsh)
    - [Chikage0o0/Linux-NetSpeed](https://github.com/ylx2016/Linux-NetSpeed)
    - [ylx2016/Linux-NetSpeed](https://github.com/ylx2016/Linux-NetSpeed)

## 客户端

https://v2.hysteria.network/zh/docs/getting-started/3rd-party-apps/

## 开发

Go >= 1.20, Node.js >= 18.12.0

- frontend

   ```bash
   cd frontend
   pnpm install
   npm run dev
   ```

- backend

   ```bash
   go run main.go
   ```

## 构建

- frontend

   ```bash
   npm run build:prod
   ```

- backend

  Windows: [build.bat](build.bat)

  Mac/Linux: [build.sh](build.sh)

## 其他

Telegram Channel: https://t.me/jonssonyan_channel

你可以在 YouTube 上订阅我的频道: https://www.youtube.com/@jonssonyan

## 贡献者

在这里感谢所有为此项目做出贡献的人

<a href="https://github.com/jonssonyan/h-ui/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=jonssonyan/h-ui" />
</a>

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=jonssonyan/h-ui&type=Date)](https://star-history.com/#jonssonyan/h-ui&Date)

## 开源协议

[GPL-3.0](LICENSE)