package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/entity"
	"strconv"
	"strings"
)

func UpdateConfig(key string, value string) error {
	if key == constant.Hysteria2Enable {
		if value == "1" {
			hysteria2Config, err := GetHysteria2Config()
			if err != nil {
				return err
			}
			if hysteria2Config.Listen == nil || *hysteria2Config.Listen == "" {
				logrus.Errorf("hysteria2 config is empty")
				return errors.New("hysteria2 config is empty")
			}
			// 启动Hysteria2
			if err = StartHysteria2(); err != nil {
				return err
			}
		} else {
			if err := StopHysteria2(); err != nil {
				return err
			}
		}
	}
	return dao.UpdateConfig([]string{key}, map[string]interface{}{"value": value})
}

func GetConfig(key string) (entity.Config, error) {
	return dao.GetConfig("key = ?", key)
}

func ListConfig(keys []string) ([]entity.Config, error) {
	return dao.ListConfig("key in ?", keys)
}

func ListConfigNotIn(keys []string) ([]entity.Config, error) {
	return dao.ListConfig("key not in ?", keys)
}

func GetHysteria2Config() (bo.Hysteria2ServerConfig, error) {
	var serverConfig bo.Hysteria2ServerConfig
	config, err := dao.GetConfig("key = ?", constant.Hysteria2Config)
	if err != nil {
		return serverConfig, err
	}
	if err = yaml.Unmarshal([]byte(*config.Value), &serverConfig); err != nil {
		return serverConfig, err
	}
	return serverConfig, nil
}

func UpdateHysteria2Config(hysteria2ServerConfig bo.Hysteria2ServerConfig) error {
	// 默认值
	config, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.JwtSecret})
	if err != nil {
		return err
	}

	var hUIWebPort string
	var jwtSecret string
	for _, item := range config {
		if *item.Key == constant.HUIWebPort {
			hUIWebPort = *item.Value
		} else if *item.Key == constant.JwtSecret {
			jwtSecret = *item.Value
		}
	}

	if hUIWebPort == "" || jwtSecret == "" {
		logrus.Errorf("hUIWebPort or jwtSecret is nil")
		return errors.New(constant.SysError)
	}

	authHttpUrl, err := GetAuthHttpUrl()
	if err != nil {
		return err
	}

	authType := "http"
	authHttpInsecure := true
	var auth bo.ServerConfigAuth
	auth.Type = &authType
	var http bo.ServerConfigAuthHTTP
	http.URL = &authHttpUrl
	http.Insecure = &authHttpInsecure
	auth.HTTP = &http
	hysteria2ServerConfig.Auth = &auth
	hysteria2ServerConfig.TrafficStats.Secret = &jwtSecret

	yamlConfig, err := yaml.Marshal(&hysteria2ServerConfig)
	if err != nil {
		return err
	}
	return dao.UpdateConfig([]string{constant.Hysteria2Config}, map[string]interface{}{"value": string(yamlConfig)})
}

func SetHysteria2Config(hysteria2ServerConfig bo.Hysteria2ServerConfig) error {
	config, err := yaml.Marshal(&hysteria2ServerConfig)
	if err != nil {
		return err
	}
	return dao.UpdateConfig([]string{constant.Hysteria2Config}, map[string]interface{}{"value": string(config)})
}

func UpsertConfig(configs []entity.Config) error {
	return dao.UpsertConfig(configs)
}

func GetHysteria2ApiPort() (int64, error) {
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return 0, err
	}
	if hysteria2Config.TrafficStats == nil || hysteria2Config.TrafficStats.Listen == nil {
		errMsg := "hysteria2 Traffic Stats API (HTTP) Listen is nil"
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}
	apiPort, err := strconv.ParseInt(strings.Split(*hysteria2Config.TrafficStats.Listen, ":")[1], 10, 64)
	if err != nil {
		errMsg := fmt.Sprintf("apiPort: %s is invalid", *hysteria2Config.TrafficStats.Listen)
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}
	return apiPort, nil
}

func GetPortAndCert() (int64, string, string, error) {
	configs, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.HUICrtPath, constant.HUIKeyPath})
	if err != nil {
		return 0, "", "", err
	}
	port := ""
	crtPath := ""
	keyPath := ""
	for _, config := range configs {
		value := *config.Value
		if *config.Key == constant.HUIWebPort {
			port = value
		} else if *config.Key == constant.HUICrtPath {
			crtPath = value
		} else if *config.Key == constant.HUIKeyPath {
			keyPath = value
		}
	}

	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		logrus.Errorf("port: %s is invalid", port)
		return 0, "", "", errors.New(fmt.Sprintf("port: %s is invalid", port))
	}

	return portInt, crtPath, keyPath, nil
}

func GetAuthHttpUrl() (string, error) {
	port, crtPath, keyPath, err := GetPortAndCert()
	if err != nil {
		return "", err
	}
	protocol := "http"
	if crtPath != "" && keyPath != "" {
		protocol = "https"
	}
	config, err := dao.GetConfig("key = ?", constant.HUIWebContext)
	if err != nil {
		return "", err
	}
	webContext := ""
	if config.Value != nil && *config.Value != "/" && strings.HasPrefix(*config.Value, "/") {
		webContext = *config.Value
	}
	return fmt.Sprintf("%s://127.0.0.1:%d%s/hui/hysteria2/auth", protocol, port, webContext), nil
}
// GetHysteria2Node2Config 获取第二节点配置
func GetHysteria2Node2Config() (bo.Hysteria2ServerConfig, error) {
	var serverConfig bo.Hysteria2ServerConfig
	config, err := dao.GetConfig("key = ?", constant.Hysteria2Node2Config)
	if err != nil {
		return serverConfig, err
	}
	if *config.Value == "" {
		return serverConfig, nil
	}
	if err = yaml.Unmarshal([]byte(*config.Value), &serverConfig); err != nil {
		return serverConfig, err
	}
	return serverConfig, nil
}

// UpdateHysteria2Node2Config 更新第二节点配置
func UpdateHysteria2Node2Config(hysteria2ServerConfig bo.Hysteria2ServerConfig) error {
	// 获取基础配置
	config, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.JwtSecret})
	if err != nil {
		return err
	}

	var hUIWebPort string
	var jwtSecret string
	for _, item := range config {
		if *item.Key == constant.HUIWebPort {
			hUIWebPort = *item.Value
		} else if *item.Key == constant.JwtSecret {
			jwtSecret = *item.Value
		}
	}

	if hUIWebPort == "" || jwtSecret == "" {
		logrus.Errorf("hUIWebPort or jwtSecret is nil")
		return errors.New(constant.SysError)
	}

	// 设置认证URL（使用不同的端口）
	authHttpUrl, err := GetAuthHttpUrl()
	if err != nil {
		return err
	}

	authType := "http"
	authHttpInsecure := true
	var auth bo.ServerConfigAuth
	auth.Type = &authType
	var http bo.ServerConfigAuthHTTP
	http.URL = &authHttpUrl
	http.Insecure = &authHttpInsecure
	auth.HTTP = &http
	hysteria2ServerConfig.Auth = &auth
	hysteria2ServerConfig.TrafficStats.Secret = &jwtSecret

	yamlConfig, err := yaml.Marshal(&hysteria2ServerConfig)
	if err != nil {
		return err
	}
	return dao.UpdateConfig([]string{constant.Hysteria2Node2Config}, map[string]interface{}{"value": string(yamlConfig)})
}

// GetSocks5Config 获取SOCKS5配置
func GetSocks5Config() (bo.Socks5Config, error) {
	var socks5Config bo.Socks5Config
	configs, err := dao.ListConfig("key in ?", []string{
		constant.Hysteria2Socks5Addr,
		constant.Hysteria2Socks5User,
		constant.Hysteria2Socks5Pass,
	})
	if err != nil {
		return socks5Config, err
	}

	for _, config := range configs {
		switch *config.Key {
		case constant.Hysteria2Socks5Addr:
			socks5Config.Addr = *config.Value
		case constant.Hysteria2Socks5User:
			socks5Config.Username = *config.Value
		case constant.Hysteria2Socks5Pass:
			socks5Config.Password = *config.Value
		}
	}
	return socks5Config, nil
}

// UpdateSocks5Config 更新SOCKS5配置
func UpdateSocks5Config(socks5Config bo.Socks5Config) error {
	updates := map[string]interface{}{
		constant.Hysteria2Socks5Addr: socks5Config.Addr,
		constant.Hysteria2Socks5User: socks5Config.Username,
		constant.Hysteria2Socks5Pass: socks5Config.Password,
	}
	
	keys := []string{
		constant.Hysteria2Socks5Addr,
		constant.Hysteria2Socks5User,
		constant.Hysteria2Socks5Pass,
	}
	
	return dao.UpdateConfig(keys, updates)
}

// GenerateNode2ConfigWithSocks5Outbound 生成带SOCKS5出站的第二节点配置
func GenerateNode2ConfigWithSocks5Outbound(baseConfig bo.Hysteria2ServerConfig, socks5Config bo.Socks5Config) (bo.Hysteria2ServerConfig, error) {
	// 复制基础配置
	node2Config := baseConfig
	
	// 修改监听端口（主节点端口+1）
	if baseConfig.Listen != nil && *baseConfig.Listen != "" {
		parts := strings.Split(*baseConfig.Listen, ":")
		if len(parts) == 2 {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return node2Config, fmt.Errorf("invalid port in base config: %s", parts[1])
			}
			newListen := fmt.Sprintf("%s:%d", parts[0], port+1)
			node2Config.Listen = &newListen
		}
	}
	
	// 修改API端口（主节点API端口+1）
	if baseConfig.TrafficStats != nil && baseConfig.TrafficStats.Listen != nil {
		parts := strings.Split(*baseConfig.TrafficStats.Listen, ":")
		if len(parts) == 2 {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return node2Config, fmt.Errorf("invalid API port in base config: %s", parts[1])
			}
			newAPIListen := fmt.Sprintf("%s:%d", parts[0], port+1)
			node2Config.TrafficStats.Listen = &newAPIListen
		}
	}
	
	// 添加SOCKS5出站配置
	if socks5Config.Addr != "" {
		outboundName := "socks5_proxy"
		outboundType := "socks5"
		
		var socks5Outbound bo.ServerConfigOutboundSOCKS5
		socks5Outbound.Addr = &socks5Config.Addr
		if socks5Config.Username != "" {
			socks5Outbound.Username = &socks5Config.Username
		}
		if socks5Config.Password != "" {
			socks5Outbound.Password = &socks5Config.Password
		}
		
		outboundEntry := bo.ServerConfigOutboundEntry{
			Name:   &outboundName,
			Type:   &outboundType,
			SOCKS5: &socks5Outbound,
		}
		
		node2Config.Outbounds = []bo.ServerConfigOutboundEntry{outboundEntry}
	}
	
	return node2Config, nil
}

// IsNode2Enabled 检查第二节点是否启用
func IsNode2Enabled() (bool, error) {
	config, err := dao.GetConfig("key = ?", constant.Hysteria2Node2Enable)
	if err != nil {
		return false, err
	}
	return *config.Value == "1", nil
}

// GetNode2Port 获取第二节点端口
func GetNode2Port() (int, error) {
	// 获取主节点配置
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return 0, err
	}
	
	if hysteria2Config.Listen == nil || *hysteria2Config.Listen == "" {
		return 0, errors.New("main node not configured")
	}
	
	parts := strings.Split(*hysteria2Config.Listen, ":")
	if len(parts) != 2 {
		return 0, errors.New("invalid main node listen format")
	}
	
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", parts[1])
	}
	
	return port + 1, nil
}