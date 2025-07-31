# 双节点支持功能设计文档

## 概述

本设计文档描述了如何在现有 H UI 架构基础上，以最小化修改的方式实现双节点支持功能。设计遵循现有的代码结构和UI模式，确保功能的无缝集成。

## 架构设计

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

### 数据模型扩展

#### 1. 配置表新增字段
```go
// 新增配置常量
const (
    Hysteria2Node2Enable = "HYSTERIA2_NODE2_ENABLE"     // 第二节点开关
    Hysteria2Node2Config = "HYSTERIA2_NODE2_CONFIG"     // 第二节点配置
    Hysteria2Node2Remark = "HYSTERIA2_NODE2_REMARK"    // 第二节点备注
    Hysteria2Socks5Addr = "HYSTERIA2_SOCKS5_ADDR"      // SOCKS5地址
    Hysteria2Socks5User = "HYSTERIA2_SOCKS5_USER"      // SOCKS5用户名
    Hysteria2Socks5Pass = "HYSTERIA2_SOCKS5_PASS"      // SOCKS5密码
)
```

#### 2. 用户表扩展
```go
// Account 实体添加字段
type Account struct {
    // ... 现有字段
    NodeAccess *int64 `gorm:"column:node_access;default:1" json:"nodeAccess"` // 1=单节点, 2=双节点
}
```

## 组件设计

### 1. 配置管理组件

#### 扩展现有 service/config.go
```go
// 获取第二节点配置
func GetHysteria2Node2Config() (bo.Hysteria2ServerConfig, error)

// 更新第二节点配置  
func UpdateHysteria2Node2Config(config bo.Hysteria2ServerConfig) error

// 生成第二节点配置（包含SOCKS5出站）
func generateNode2ConfigWithSocks5Outbound(baseConfig bo.Hysteria2ServerConfig, socks5Config Socks5Config) bo.Hysteria2ServerConfig
```

#### SOCKS5配置结构
```go
type Socks5Config struct {
    Addr     string `json:"addr"`
    Username string `json:"username"`  
    Password string `json:"password"`
}
```

### 2. 实例管理组件

#### 扩展现有 service/hysteria2.go
```go
// 第二节点实例管理
func StartHysteria2Node2() error
func StopHysteria2Node2() error  
func RestartHysteria2Node2() error
func Hysteria2Node2IsRunning() bool

// 双节点统一管理
func StartAllNodes() error
func StopAllNodes() error
func GetNodesStatus() map[string]bool
```

#### 实例隔离策略
- 第二节点使用独立的配置文件: `bin/hysteria2-node2.yaml`
- 第二节点使用独立的端口: 主节点端口 + 1
- 第二节点使用独立的API端口: 主节点API端口 + 1

### 3. 订阅服务组件

#### 扩展现有 service/hysteria2_api.go
```go
// 根据用户权限生成订阅
func Hysteria2SubscribeWithNodeAccess(conPass string, clientType string, host string, nodeAccess int64) (string, string, error)

// 生成多节点配置
func generateMultiNodeConfig(account entity.Account, clientType string, host string) ([]NodeConfig, error)

// 节点配置结构
type NodeConfig struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Server   string `json:"server"`
    Port     string `json:"port"`
    Password string `json:"password"`
    // ... 其他配置
}
```

### 4. 控制器扩展

#### 扩展现有 controller/config.go
```go
// 获取第二节点配置
func GetHysteria2Node2Config(c *gin.Context)

// 更新第二节点配置
func UpdateHysteria2Node2Config(c *gin.Context)

// 获取SOCKS5配置
func GetSocks5Config(c *gin.Context)

// 更新SOCKS5配置  
func UpdateSocks5Config(c *gin.Context)
```

#### 扩展现有 controller/account.go
```go
// 在现有的 SaveAccount, UpdateAccount 中添加 nodeAccess 字段处理
// 无需新增接口，只需扩展现有接口
```

## 数据库迁移

### 1. 配置表初始化
```sql
INSERT INTO config (key, value, remark) VALUES 
('HYSTERIA2_NODE2_ENABLE', '0', '第二节点开关'),
('HYSTERIA2_NODE2_REMARK', 'Node2', '第二节点备注'),
('HYSTERIA2_SOCKS5_ADDR', '', 'SOCKS5代理地址'),
('HYSTERIA2_SOCKS5_USER', '', 'SOCKS5用户名'),
('HYSTERIA2_SOCKS5_PASS', '', 'SOCKS5密码');
```

### 2. 用户表迁移
```sql
ALTER TABLE account ADD COLUMN node_access INTEGER DEFAULT 1;
```

## 前端设计

### 1. 配置页面扩展 (frontend/src/views/config/hysteria2.vue)

#### 在现有配置表单中添加：
```vue
<!-- 第二节点配置区域 -->
<el-card class="box-card" style="margin-top: 20px;">
  <template #header>
    <div class="card-header">
      <span>第二节点配置</span>
      <el-switch v-model="node2Enable" @change="handleNode2EnableChange" />
    </div>
  </template>
  
  <div v-show="node2Enable">
    <!-- SOCKS5出站配置 -->
    <el-form-item label="SOCKS5地址">
      <el-input v-model="socks5Config.addr" placeholder="127.0.0.1:1080" />
    </el-form-item>
    <!-- 用户名密码等 -->
  </div>
</el-card>
```

### 2. 用户管理页面扩展 (frontend/src/views/account/index.vue)

#### 在用户编辑对话框中添加：
```vue
<el-form-item label="节点权限">
  <el-radio-group v-model="accountForm.nodeAccess">
    <el-radio :label="1">单节点</el-radio>
    <el-radio :label="2" :disabled="!node2Enabled">双节点</el-radio>
  </el-radio-group>
</el-form-item>
```

#### 在用户列表中添加状态显示：
```vue
<el-table-column prop="nodeAccess" label="节点权限" width="100">
  <template #default="{ row }">
    <el-tag :type="row.nodeAccess === 2 ? 'success' : 'info'">
      {{ row.nodeAccess === 2 ? '双节点' : '单节点' }}
    </el-tag>
  </template>
</el-table-column>
```

### 3. API接口扩展

#### 扩展现有API (frontend/src/api/config/index.ts)
```typescript
// 第二节点配置相关
export function getNode2ConfigApi(): AxiosPromise<Node2Config>
export function updateNode2ConfigApi(data: Node2Config): AxiosPromise
export function getSocks5ConfigApi(): AxiosPromise<Socks5Config>  
export function updateSocks5ConfigApi(data: Socks5Config): AxiosPromise
```

## 错误处理

### 1. 配置验证
- SOCKS5地址格式验证
- 端口冲突检测
- 配置完整性验证

### 2. 实例管理
- 第二节点启动失败时不影响主节点
- 资源清理和错误恢复
- 状态同步和监控

### 3. 用户体验
- 配置错误时的友好提示
- 节点状态的实时显示
- 权限变更的即时生效

## 测试策略

### 1. 单元测试
- 配置生成逻辑测试
- SOCKS5出站配置测试
- 用户权限控制测试

### 2. 集成测试
- 双节点启动停止测试
- 订阅生成测试
- 权限控制端到端测试

### 3. 兼容性测试
- 现有功能回归测试
- 数据库迁移测试
- 配置导入导出测试

## 部署和维护

### 1. 升级策略
- 数据库自动迁移
- 配置向下兼容
- 渐进式功能启用

### 2. 监控和日志
- 第二节点状态监控
- SOCKS5连接状态日志
- 用户权限变更审计

### 3. 性能考虑
- 双节点资源占用
- 配置缓存优化
- 订阅生成性能