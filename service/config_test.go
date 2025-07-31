package service

import (
	"h-ui/model/bo"
	"testing"
)

func TestGenerateNode2ConfigWithSocks5Outbound(t *testing.T) {
	// 测试第二节点配置生成
	baseConfig := bo.Hysteria2ServerConfig{
		Listen: stringPtr(":443"),
		TrafficStats: &bo.ServerConfigTrafficStats{
			Listen: stringPtr(":9999"),
		},
	}

	socks5Config := bo.Socks5Config{
		Addr:     "127.0.0.1:1080",
		Username: "testuser",
		Password: "testpass",
	}

	node2Config, err := GenerateNode2ConfigWithSocks5Outbound(baseConfig, socks5Config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证端口是否正确设置为主端口+1
	if node2Config.Listen == nil || *node2Config.Listen != ":444" {
		t.Errorf("Expected listen port :444, got: %v", node2Config.Listen)
	}

	// 验证API端口是否正确设置
	if node2Config.TrafficStats == nil || node2Config.TrafficStats.Listen == nil || *node2Config.TrafficStats.Listen != ":10000" {
		t.Errorf("Expected API port :10000, got: %v", node2Config.TrafficStats.Listen)
	}

	// 验证SOCKS5出站配置
	if len(node2Config.Outbounds) != 1 {
		t.Errorf("Expected 1 outbound, got: %d", len(node2Config.Outbounds))
	}

	if node2Config.Outbounds[0].SOCKS5 == nil {
		t.Error("Expected SOCKS5 outbound configuration")
	} else {
		if *node2Config.Outbounds[0].SOCKS5.Addr != socks5Config.Addr {
			t.Errorf("Expected SOCKS5 addr %s, got: %s", socks5Config.Addr, *node2Config.Outbounds[0].SOCKS5.Addr)
		}
	}
}

func TestGetNode2Port(t *testing.T) {
	// 这里需要mock数据库，暂时跳过实际测试
	t.Skip("Requires database mocking")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}