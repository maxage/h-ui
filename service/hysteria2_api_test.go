package service

import (
	"h-ui/model/bo"
	"testing"
)

func TestGenerateNodeConfig(t *testing.T) {
	config := bo.Hysteria2ServerConfig{
		Listen: stringPtr(":443"),
		Bandwidth: &bo.ServerConfigBandwidth{
			Up:   stringPtr("100 Mbps"),
			Down: stringPtr("100 Mbps"),
		},
		Obfs: &bo.ServerConfigObfs{
			Type: stringPtr("salamander"),
			Salamander: &bo.ServerConfigObfsSalamander{
				Password: stringPtr("testpass"),
			},
		},
		ACME: &bo.ServerConfigACME{
			Domains: []string{"example.com"},
		},
	}

	nodeConfig := generateNodeConfig(config, "TestNode", "testuser.testpass", "example.com", "8080,9090")

	// 验证基本配置
	if nodeConfig.Name != "TestNode" {
		t.Errorf("Expected name TestNode, got: %s", nodeConfig.Name)
	}

	if nodeConfig.Type != "hysteria2" {
		t.Errorf("Expected type hysteria2, got: %s", nodeConfig.Type)
	}

	if nodeConfig.Server != "example.com" {
		t.Errorf("Expected server example.com, got: %s", nodeConfig.Server)
	}

	if nodeConfig.Port != "443" {
		t.Errorf("Expected port 443, got: %s", nodeConfig.Port)
	}

	if nodeConfig.Password != "testuser.testpass" {
		t.Errorf("Expected password testuser.testpass, got: %s", nodeConfig.Password)
	}

	// 验证带宽配置
	if nodeConfig.Up != "100 Mbps" {
		t.Errorf("Expected up 100 Mbps, got: %s", nodeConfig.Up)
	}

	if nodeConfig.Down != "100 Mbps" {
		t.Errorf("Expected down 100 Mbps, got: %s", nodeConfig.Down)
	}

	// 验证混淆配置
	if nodeConfig.Obfs != "testpass" {
		t.Errorf("Expected obfs testpass, got: %s", nodeConfig.Obfs)
	}

	// 验证SNI配置
	if nodeConfig.Sni != "example.com" {
		t.Errorf("Expected sni example.com, got: %s", nodeConfig.Sni)
	}

	// 验证端口跳跃
	if nodeConfig.Ports != "8080,9090" {
		t.Errorf("Expected ports 8080,9090, got: %s", nodeConfig.Ports)
	}
}

func TestGenerateHysteria2Url(t *testing.T) {
	nodeConfig := NodeConfig{
		Name:     "TestNode",
		Type:     "hysteria2",
		Server:   "example.com",
		Port:     "443",
		Password: "testuser.testpass",
		Obfs:     "testobfs",
		Sni:      "example.com",
		Down:     "100 Mbps",
		Ports:    "8080,9090",
	}

	url, err := generateHysteria2Url(nodeConfig, "example.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证URL格式
	expectedPrefix := "hysteria2://testuser.testpass@example.com:443"
	if !contains(url, expectedPrefix) {
		t.Errorf("Expected URL to contain %s, got: %s", expectedPrefix, url)
	}

	// 验证参数
	if !contains(url, "obfs=salamander") {
		t.Error("Expected URL to contain obfs=salamander")
	}

	if !contains(url, "sni=example.com") {
		t.Error("Expected URL to contain sni=example.com")
	}

	if !contains(url, "#TestNode") {
		t.Error("Expected URL to contain #TestNode")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}