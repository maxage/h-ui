package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"h-ui/dao"
	"h-ui/model/constant"
	"h-ui/model/vo"
	"h-ui/proxy"
	"h-ui/util"
	"os"
)

func InitHysteria2() error {
	if !util.Exists(util.GetHysteria2BinPath()) {
		if err := util.DownloadHysteria2(""); err != nil {
			logrus.Errorf("download hysteria2 bin err: %v", err)
			return errors.New("download hysteria2 bin err")
		}
	}

	config, err := dao.GetConfig("key = ?", constant.Hysteria2Enable)
	if err != nil {
		return err
	}

	if *config.Value == "1" {
		if err = StartHysteria2(); err != nil {
			return err
		}
	}

	// 初始化第二节点
	if err := InitHysteria2Node2(); err != nil {
		logrus.Errorf("init hysteria2 node2 err: %v", err)
		// 第二节点初始化失败不影响主节点
	}

	return nil
}

func setHysteria2ConfigYAML() error {
	serverConfig, err := GetHysteria2Config()
	if err != nil {
		return err
	}
	if serverConfig.Listen == nil || *serverConfig.Listen == "" {
		return errors.New("hysteria2 config is empty")
	}

	authHttpUrl, err := GetAuthHttpUrl()
	if err != nil {
		return err
	}

	// update auth http url
	if *serverConfig.Auth.HTTP.URL != authHttpUrl {
		serverConfig.Auth.HTTP.URL = &authHttpUrl
		if err := UpdateHysteria2Config(serverConfig); err != nil {
			return err
		}
	}

	hysteria2Config, err := yaml.Marshal(&serverConfig)
	if err != nil {
		logrus.Errorf("marshal hysteria2 config err: %v", err)
		return errors.New("marshal hysteria2 config err")
	}
	file, err := os.OpenFile(constant.Hysteria2ConfigPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("create hysteria2 server config file err: %v", err)
		return errors.New("create hysteria2 server config file err")
	}
	_, err = file.WriteString(string(hysteria2Config))
	if err != nil {
		logrus.Errorf("write hysteria2 config.json file err: %v", err)
		return errors.New("hysteria2 config.json file write err")
	}
	return nil
}

func Hysteria2IsRunning() bool {
	return proxy.NewHysteria2Instance().IsRunning()
}

func StartHysteria2() error {
	if err := setHysteria2ConfigYAML(); err != nil {
		return err
	}
	if err := proxy.NewHysteria2Instance().StartHysteria2(); err != nil {
		return err
	}
	return nil
}

func StopHysteria2() error {
	return proxy.NewHysteria2Instance().StopHysteria2()
}

func RestartHysteria2() error {
	if err := StopHysteria2(); err != nil {
		return err
	}
	if err := StartHysteria2(); err != nil {
		return err
	}
	return nil
}

func ReleaseHysteria2() error {
	return proxy.NewHysteria2Instance().Release()
}

func Hysteria2AcmePath() (vo.Hysteria2AcmePathVo, error) {
	hysteria2AcmePathVo := vo.Hysteria2AcmePathVo{}
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return hysteria2AcmePathVo, err
	}
	if hysteria2Config.TLS != nil &&
		hysteria2Config.TLS.Cert != nil && *hysteria2Config.TLS.Cert != "" &&
		hysteria2Config.TLS.Key != nil && *hysteria2Config.TLS.Key != "" {
		if util.Exists(*hysteria2Config.TLS.Cert) && util.Exists(*hysteria2Config.TLS.Key) {
			hysteria2AcmePathVo.CrtPath = *hysteria2Config.TLS.Cert
			hysteria2AcmePathVo.KeyPath = *hysteria2Config.TLS.Key
			return hysteria2AcmePathVo, nil
		}
		return hysteria2AcmePathVo, errors.New("cert not found")
	} else if hysteria2Config.ACME != nil &&
		hysteria2Config.ACME.Domains != nil &&
		len(hysteria2Config.ACME.Domains) > 0 &&
		hysteria2Config.ACME.CA != nil &&
		*hysteria2Config.ACME.CA != "" &&
		hysteria2Config.ACME.Dir != nil &&
		*hysteria2Config.ACME.Dir != "" {
		acmeDir := *hysteria2Config.ACME.Dir
		for _, domain := range hysteria2Config.ACME.Domains {
			crtPath, err := util.FindFile(acmeDir, fmt.Sprintf("%s.crt", domain))
			if err != nil {
				continue
			}
			keyPath, err := util.FindFile(acmeDir, fmt.Sprintf("%s.key", domain))
			if err != nil {
				continue
			}
			hysteria2AcmePathVo.CrtPath = crtPath
			hysteria2AcmePathVo.KeyPath = keyPath
			return hysteria2AcmePathVo, nil
		}
	}
	return vo.Hysteria2AcmePathVo{}, errors.New("cert not found")
}
// 第二节点相关函数

// InitHysteria2Node2 初始化第二节点
func InitHysteria2Node2() error {
	enabled, err := IsNode2Enabled()
	if err != nil {
		logrus.Errorf("check node2 enabled status failed: %v", err)
		return err
	}

	if enabled {
		logrus.Info("Node2 is enabled, attempting to start...")
		if err = StartHysteria2Node2(); err != nil {
			logrus.Errorf("start hysteria2 node2 failed: %v", err)
			// 第二节点启动失败不影响主节点
			logrus.Warn("Node2 startup failed, but main node will continue to operate")
			return nil
		}
		logrus.Info("Node2 started successfully")
	} else {
		logrus.Info("Node2 is disabled, skipping initialization")
	}
	return nil
}

// setHysteria2Node2ConfigYAML 设置第二节点配置文件
func setHysteria2Node2ConfigYAML() error {
	// 获取主节点配置
	baseConfig, err := GetHysteria2Config()
	if err != nil {
		return err
	}
	if baseConfig.Listen == nil || *baseConfig.Listen == "" {
		return errors.New("main hysteria2 config is empty")
	}

	// 获取SOCKS5配置
	socks5Config, err := GetSocks5Config()
	if err != nil {
		return err
	}
	if socks5Config.Addr == "" {
		return errors.New("socks5 config is empty")
	}

	// 生成第二节点配置
	node2Config, err := GenerateNode2ConfigWithSocks5Outbound(baseConfig, socks5Config)
	if err != nil {
		return err
	}

	// 更新认证URL
	authHttpUrl, err := GetAuthHttpUrl()
	if err != nil {
		return err
	}

	if node2Config.Auth != nil && node2Config.Auth.HTTP != nil {
		node2Config.Auth.HTTP.URL = &authHttpUrl
		if err := UpdateHysteria2Node2Config(node2Config); err != nil {
			return err
		}
	}

	// 生成YAML配置
	hysteria2Config, err := yaml.Marshal(&node2Config)
	if err != nil {
		logrus.Errorf("marshal hysteria2 node2 config err: %v", err)
		return errors.New("marshal hysteria2 node2 config err")
	}

	// 写入配置文件
	file, err := os.OpenFile(constant.Hysteria2Node2ConfigPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("create hysteria2 node2 config file err: %v", err)
		return errors.New("create hysteria2 node2 config file err")
	}
	defer file.Close()

	_, err = file.WriteString(string(hysteria2Config))
	if err != nil {
		logrus.Errorf("write hysteria2 node2 config file err: %v", err)
		return errors.New("hysteria2 node2 config file write err")
	}
	return nil
}

// Hysteria2Node2IsRunning 检查第二节点是否运行
func Hysteria2Node2IsRunning() bool {
	return proxy.NewHysteria2Node2Instance().IsRunning()
}

// StartHysteria2Node2 启动第二节点
func StartHysteria2Node2() error {
	logrus.Info("Starting Hysteria2 Node2...")
	
	if err := setHysteria2Node2ConfigYAML(); err != nil {
		logrus.Errorf("failed to set node2 config: %v", err)
		return err
	}
	
	if err := proxy.NewHysteria2Node2Instance().StartHysteria2(); err != nil {
		logrus.Errorf("failed to start node2 process: %v", err)
		return err
	}
	
	logrus.Info("Hysteria2 Node2 started successfully")
	return nil
}

// StopHysteria2Node2 停止第二节点
func StopHysteria2Node2() error {
	logrus.Info("Stopping Hysteria2 Node2...")
	
	if err := proxy.NewHysteria2Node2Instance().StopHysteria2(); err != nil {
		logrus.Errorf("failed to stop node2: %v", err)
		return err
	}
	
	logrus.Info("Hysteria2 Node2 stopped successfully")
	return nil
}

// RestartHysteria2Node2 重启第二节点
func RestartHysteria2Node2() error {
	if err := StopHysteria2Node2(); err != nil {
		return err
	}
	if err := StartHysteria2Node2(); err != nil {
		return err
	}
	return nil
}

// ReleaseHysteria2Node2 释放第二节点资源
func ReleaseHysteria2Node2() error {
	return proxy.NewHysteria2Node2Instance().Release()
}

// 双节点统一管理函数

// StartAllNodes 启动所有节点
func StartAllNodes() error {
	// 启动主节点
	if err := StartHysteria2(); err != nil {
		return err
	}

	// 检查是否启用第二节点
	enabled, err := IsNode2Enabled()
	if err != nil {
		logrus.Errorf("check node2 enabled err: %v", err)
		return nil // 不影响主节点
	}

	if enabled {
		if err := StartHysteria2Node2(); err != nil {
			logrus.Errorf("start node2 err: %v", err)
			// 第二节点启动失败不影响主节点
		}
	}

	return nil
}

// StopAllNodes 停止所有节点
func StopAllNodes() error {
	var lastErr error

	// 停止第二节点
	if Hysteria2Node2IsRunning() {
		if err := StopHysteria2Node2(); err != nil {
			logrus.Errorf("stop node2 err: %v", err)
			lastErr = err
		}
	}

	// 停止主节点
	if err := StopHysteria2(); err != nil {
		logrus.Errorf("stop main node err: %v", err)
		lastErr = err
	}

	return lastErr
}

// GetNodesStatus 获取所有节点状态
func GetNodesStatus() map[string]bool {
	return map[string]bool{
		"node1": Hysteria2IsRunning(),
		"node2": Hysteria2Node2IsRunning(),
	}
}

// ReleaseAllNodes 释放所有节点资源
func ReleaseAllNodes() error {
	var lastErr error

	if err := ReleaseHysteria2Node2(); err != nil {
		logrus.Errorf("release node2 err: %v", err)
		lastErr = err
	}

	if err := ReleaseHysteria2(); err != nil {
		logrus.Errorf("release main node err: %v", err)
		lastErr = err
	}

	return lastErr
}