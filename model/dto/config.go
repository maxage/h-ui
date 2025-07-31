package dto

type ConfigDto struct {
	Key *string `json:"key" form:"key" validate:"required,min=1,max=128"`
}

type ConfigsDto struct {
	Keys []string `json:"keys" form:"keys" validate:"required"`
}

type ConfigUpdateDto struct {
	Key   *string `json:"key" form:"key" validate:"required,min=1,max=128"`
	Value *string `json:"value" form:"value" validate:"required,min=0,max=128"`
}

type ConfigsUpdateDto struct {
	ConfigUpdateDtos []ConfigUpdateDto `json:"configUpdateDtos" form:"configUpdateDtos" validate:"required"`
}
// Node2ConfigDto 第二节点配置DTO
type Node2ConfigDto struct {
	Enable bool   `json:"enable" form:"enable"`
	Remark string `json:"remark" form:"remark" validate:"omitempty,max=50"`
}

// Socks5ConfigDto SOCKS5配置DTO
type Socks5ConfigDto struct {
	Addr     string `json:"addr" form:"addr" validate:"required_if=Enable true,omitempty,hostname_port"`
	Username string `json:"username" form:"username" validate:"omitempty,max=50"`
	Password string `json:"password" form:"password" validate:"omitempty,max=100"`
}