package vo

type ConfigVo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}
// Node2ConfigVo 第二节点配置VO
type Node2ConfigVo struct {
	Enable bool   `json:"enable"`
	Remark string `json:"remark"`
	Port   int    `json:"port"`
	Status bool   `json:"status"` // 运行状态
}

// Socks5ConfigVo SOCKS5配置VO
type Socks5ConfigVo struct {
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // 密码在返回时可能需要隐藏
}