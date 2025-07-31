package bo

// Socks5Config SOCKS5代理配置
type Socks5Config struct {
	Addr     string `json:"addr" yaml:"addr"`         // SOCKS5代理地址，格式: host:port
	Username string `json:"username" yaml:"username"` // SOCKS5用户名（可选）
	Password string `json:"password" yaml:"password"` // SOCKS5密码（可选）
}

// Node2Config 第二节点配置
type Node2Config struct {
	Enable bool         `json:"enable"`         // 是否启用第二节点
	Remark string       `json:"remark"`         // 节点备注名称
	Socks5 Socks5Config `json:"socks5"`         // SOCKS5出站配置
	Port   int          `json:"port,omitempty"` // 第二节点端口（自动计算）
}