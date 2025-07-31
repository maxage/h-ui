package service

import (
	"errors"
	"fmt"
	"github.com/skip2/go-qrcode"
	"gopkg.in/yaml.v3"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/entity"
	"h-ui/model/vo"
	"h-ui/proxy"
	"net/url"
	"strings"
	"time"
)

func Hysteria2Auth(conPass string) (int64, string, error) {
	if !Hysteria2IsRunning() {
		return 0, "", errors.New("hysteria2 is not running")
	}

	now := time.Now().UnixMilli()
	account, err := dao.GetAccount("con_pass = ? and deleted = 0 and (quota < 0 or quota > download + upload) and ? < expire_time and ? > kick_util_time", conPass, now, now)
	if err != nil {
		return 0, "", err
	}

	// 限制设备数
	onlineUsers, err := Hysteria2Online()
	if err != nil {
		return 0, "", err
	}
	device, exist := onlineUsers[*account.Username]
	if exist && *account.DeviceNo <= device {
		return 0, "", errors.New("device limited")
	}

	return *account.Id, *account.Username, nil
}

func Hysteria2Online() (map[string]int64, error) {
	if !Hysteria2IsRunning() {
		return map[string]int64{}, nil
	}
	apiPort, err := GetHysteria2ApiPort()
	if err != nil {
		return nil, errors.New("get hysteria2 apiPort err")
	}
	jwtSecretConfig, err := dao.GetConfig("key = ?", constant.JwtSecret)
	if err != nil {
		return nil, err
	}
	onlineUsers, err := proxy.NewHysteria2Api(apiPort).OnlineUsers(*jwtSecretConfig.Value)
	if err != nil {
		return nil, err
	}
	return onlineUsers, nil
}

func Hysteria2Kick(ids []int64, kickUtilTime int64) error {
	if !Hysteria2IsRunning() {
		return errors.New("hysteria2 is not running")
	}
	if err := dao.UpdateAccount(ids, map[string]interface{}{"kick_util_time": kickUtilTime}); err != nil {
		return err
	}

	accounts, err := dao.ListAccount("id in ?", ids)
	if err != nil {
		return err
	}
	var keys []string
	for _, item := range accounts {
		keys = append(keys, *item.Username)
	}
	apiPort, err := GetHysteria2ApiPort()
	if err != nil {
		return errors.New("get hysteria2 apiPort err")
	}
	jwtSecretConfig, err := dao.GetConfig("key = ?", constant.JwtSecret)
	if err != nil {
		return err
	}
	if err = proxy.NewHysteria2Api(apiPort).KickUsers(keys, *jwtSecretConfig.Value); err != nil {
		return err
	}
	return nil
}

func Hysteria2SubscribeUrl(accountId int64, protocol string, host string) (string, error) {
	account, err := dao.GetAccount("id = ?", accountId)
	if err != nil {
		return "", err
	}
	config, err := dao.GetConfig("key = ?", constant.HUIWebContext)
	if err != nil {
		return "", err
	}
	webContext := ""
	if config.Value != nil && *config.Value != "/" && strings.HasPrefix(*config.Value, "/") {
		webContext = *config.Value
	}
	return fmt.Sprintf("%s//%s%s/hui/%s", protocol, host, webContext, url.QueryEscape(*account.ConPass)), nil
}

func Hysteria2Subscribe(conPass string, clientType string, host string) (string, string, error) {
	// 获取用户信息
	account, err := dao.GetAccount("con_pass = ?", conPass)
	if err != nil {
		return "", "", err
	}

	// 根据用户权限生成订阅
	return Hysteria2SubscribeWithNodeAccess(conPass, clientType, host, *account.NodeAccess)
}

func Hysteria2SubscribeWithNodeAccess(conPass string, clientType string, host string, nodeAccess int64) (string, string, error) {
	account, err := dao.GetAccount("con_pass = ?", conPass)
	if err != nil {
		return "", "", err
	}

	// 生成多节点配置
	nodeConfigs, err := generateMultiNodeConfig(*account, clientType, host, nodeAccess)
	if err != nil {
		return "", "", err
	}

	userInfo := ""
	configStr := ""
	
	if clientType == constant.Shadowrocket || clientType == constant.Clash {
		userInfo = fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d",
			*account.Upload,
			*account.Download,
			*account.Quota,
			*account.ExpireTime/1000)

		// 转换为bo.Hysteria2格式
		var proxies []interface{}
		var proxyNames []string
		
		for _, nodeConfig := range nodeConfigs {
			hysteria2 := bo.Hysteria2{
				Name:     nodeConfig.Name,
				Type:     nodeConfig.Type,
				Server:   nodeConfig.Server,
				Port:     nodeConfig.Port,
				Ports:    nodeConfig.Ports,
				Password: nodeConfig.Password,
				Up:       nodeConfig.Up,
				Down:     nodeConfig.Down,
				Sni:      nodeConfig.Sni,
				SkipCertVerify: nodeConfig.SkipCertVerify,
			}
			
			if nodeConfig.Obfs != "" {
				if clientType == constant.Shadowrocket {
					hysteria2.Obfs = nodeConfig.Obfs
				} else {
					hysteria2.Obfs = "salamander"
					hysteria2.ObfsPassword = nodeConfig.Obfs
				}
			}
			
			proxies = append(proxies, hysteria2)
			proxyNames = append(proxyNames, nodeConfig.Name)
		}

		proxyGroup := bo.ProxyGroup{
			Name:    "PROXY",
			Type:    "select",
			Proxies: proxyNames,
		}

		clashConfig := bo.ClashConfig{
			ProxyGroups: []bo.ProxyGroup{
				proxyGroup,
			},
			Proxies: proxies,
		}
		clashConfigYaml, err := yaml.Marshal(&clashConfig)
		if err != nil {
			return "", "", err
		}
		configStr = string(clashConfigYaml)
		if clientType == constant.Clash {
			clashExtension, err := GetConfig(constant.ClashExtension)
			if err != nil {
				return "", "", err
			}
			if clashExtension.Value != nil && *clashExtension.Value != "" {
				configStr = fmt.Sprintf("%s%s", configStr, *clashExtension.Value)
			}
		}
	} else if clientType == constant.V2rayN {
		// V2rayN支持多个URL，用换行分隔
		var urls []string
		for _, nodeConfig := range nodeConfigs {
			url, err := generateHysteria2Url(nodeConfig, strings.Split(host, ":")[0])
			if err != nil {
				return "", "", err
			}
			urls = append(urls, url)
		}
		configStr = strings.Join(urls, "\n")
	}

	return userInfo, configStr, nil
}

func Hysteria2Url(accountId int64, hostname string) (string, error) {
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return "", err
	}
	if hysteria2Config.Listen == nil || *hysteria2Config.Listen == "" {
		return "", errors.New("hysteria2 config is empty")
	}

	account, err := dao.GetAccount("id = ?", accountId)
	if err != nil {
		return "", err
	}

	urlConfig := ""
	if hysteria2Config.Obfs != nil &&
		hysteria2Config.Obfs.Type != nil &&
		*hysteria2Config.Obfs.Type == "salamander" &&
		hysteria2Config.Obfs.Salamander != nil &&
		hysteria2Config.Obfs.Salamander.Password != nil &&
		*hysteria2Config.Obfs.Salamander.Password != "" {
		urlConfig += fmt.Sprintf("&obfs=salamander&obfs-password=%s", *hysteria2Config.Obfs.Salamander.Password)
	}

	if hysteria2Config.ACME != nil &&
		hysteria2Config.ACME.Domains != nil &&
		len(hysteria2Config.ACME.Domains) > 0 {
		urlConfig += fmt.Sprintf("&sni=%s", hysteria2Config.ACME.Domains[0])
		// shadowrocket
		urlConfig += fmt.Sprintf("&peer=%s", hysteria2Config.ACME.Domains[0])
	}

	urlConfig += "&insecure=0"

	if hysteria2Config.Bandwidth != nil &&
		hysteria2Config.Bandwidth.Down != nil &&
		*hysteria2Config.Bandwidth.Down != "" {
		// shadowrocket
		urlConfig += fmt.Sprintf("&downmbps=%s", url.PathEscape(*hysteria2Config.Bandwidth.Down))
	}

	hysteria2ConfigPortHopping, err := dao.GetConfig("key = ?", constant.Hysteria2ConfigPortHopping)
	if err != nil {
		return "", err
	}
	if *hysteria2ConfigPortHopping.Value != "" {
		// shadowrocket
		urlConfig += fmt.Sprintf("&mport=%s", *hysteria2ConfigPortHopping.Value)
	}

	hysteria2ConfigRemark, err := dao.GetConfig("key = ?", constant.Hysteria2ConfigRemark)
	if err != nil {
		return "", err
	}
	if *hysteria2ConfigRemark.Value != "" {
		urlConfig += fmt.Sprintf("#%s", *hysteria2ConfigRemark.Value)
	}
	if urlConfig != "" {
		urlConfig = "/?" + strings.TrimPrefix(urlConfig, "&")
	}
	return fmt.Sprintf("hysteria2://%s@%s%s", *account.ConPass, hostname, *hysteria2Config.Listen) + urlConfig, nil
}
// generateMultiNodeConfig 生成多节点配置
func generateMultiNodeConfig(account entity.Account, clientType string, host string, nodeAccess int64) ([]NodeConfig, error) {
	var nodeConfigs []NodeConfig

	// 获取主节点配置
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return nil, err
	}
	if hysteria2Config.Listen == nil || *hysteria2Config.Listen == "" {
		return nil, errors.New("hysteria2 config is empty")
	}

	// 获取主节点备注
	hysteria2ConfigRemark, err := dao.GetConfig("key = ?", constant.Hysteria2ConfigRemark)
	if err != nil {
		return nil, err
	}
	mainNodeName := "hysteria2"
	if *hysteria2ConfigRemark.Value != "" {
		mainNodeName = *hysteria2ConfigRemark.Value
	}

	// 获取端口跳跃配置
	hysteria2ConfigPortHopping, err := dao.GetConfig("key = ?", constant.Hysteria2ConfigPortHopping)
	if err != nil {
		return nil, err
	}

	// 生成主节点配置
	mainNode := generateNodeConfig(hysteria2Config, mainNodeName, *account.ConPass, host, *hysteria2ConfigPortHopping.Value)
	nodeConfigs = append(nodeConfigs, mainNode)

	// 如果用户有双节点权限且第二节点启用，添加第二节点
	if nodeAccess == 2 {
		enabled, err := IsNode2Enabled()
		if err != nil {
			return nil, err
		}
		if enabled {
			// 获取第二节点配置
			node2Config, err := GetHysteria2Node2Config()
			if err != nil {
				return nil, err
			}
			if node2Config.Listen != nil && *node2Config.Listen != "" {
				// 获取第二节点备注
				node2Remark, err := dao.GetConfig("key = ?", constant.Hysteria2Node2Remark)
				if err != nil {
					return nil, err
				}
				node2Name := "Node2"
				if *node2Remark.Value != "" {
					node2Name = *node2Remark.Value
				}

				// 生成第二节点配置
				node2 := generateNodeConfig(node2Config, node2Name, *account.ConPass, host, *hysteria2ConfigPortHopping.Value)
				nodeConfigs = append(nodeConfigs, node2)
			}
		}
	}

	return nodeConfigs, nil
}

// generateNodeConfig 生成单个节点配置
func generateNodeConfig(config bo.Hysteria2ServerConfig, nodeName string, conPass string, host string, portHopping string) NodeConfig {
	nodeConfig := NodeConfig{
		Name:     nodeName,
		Type:     "hysteria2",
		Server:   strings.Split(host, ":")[0],
		Port:     strings.Split(*config.Listen, ":")[1],
		Ports:    portHopping,
		Password: conPass,
	}

	// 设置带宽
	if config.Bandwidth != nil {
		if config.Bandwidth.Up != nil && *config.Bandwidth.Up != "" {
			nodeConfig.Up = *config.Bandwidth.Up
		}
		if config.Bandwidth.Down != nil && *config.Bandwidth.Down != "" {
			nodeConfig.Down = *config.Bandwidth.Down
		}
	}

	// 设置混淆
	if config.Obfs != nil &&
		config.Obfs.Type != nil &&
		*config.Obfs.Type == "salamander" &&
		config.Obfs.Salamander != nil &&
		config.Obfs.Salamander.Password != nil &&
		*config.Obfs.Salamander.Password != "" {
		nodeConfig.Obfs = *config.Obfs.Salamander.Password
	}

	// 设置SNI
	if config.ACME != nil &&
		config.ACME.Domains != nil &&
		len(config.ACME.Domains) > 0 {
		nodeConfig.Sni = config.ACME.Domains[0]
	}

	nodeConfig.SkipCertVerify = false

	return nodeConfig
}

// NodeConfig 节点配置结构
type NodeConfig struct {
	Name           string `json:"name" yaml:"name"`
	Type           string `json:"type" yaml:"type"`
	Server         string `json:"server" yaml:"server"`
	Port           string `json:"port" yaml:"port"`
	Ports          string `json:"ports,omitempty" yaml:"ports,omitempty"`
	Password       string `json:"password" yaml:"password"`
	Up             string `json:"up,omitempty" yaml:"up,omitempty"`
	Down           string `json:"down,omitempty" yaml:"down,omitempty"`
	Obfs           string `json:"obfs,omitempty" yaml:"obfs,omitempty"`
	ObfsPassword   string `json:"obfs-password,omitempty" yaml:"obfs-password,omitempty"`
	Sni            string `json:"sni,omitempty" yaml:"sni,omitempty"`
	SkipCertVerify bool   `json:"skip-cert-verify" yaml:"skip-cert-verify"`
}//
 generateHysteria2Url 生成Hysteria2 URL
func generateHysteria2Url(nodeConfig NodeConfig, hostname string) (string, error) {
	urlConfig := ""
	
	if nodeConfig.Obfs != "" {
		urlConfig += fmt.Sprintf("&obfs=salamander&obfs-password=%s", nodeConfig.Obfs)
	}

	if nodeConfig.Sni != "" {
		urlConfig += fmt.Sprintf("&sni=%s", nodeConfig.Sni)
		// shadowrocket
		urlConfig += fmt.Sprintf("&peer=%s", nodeConfig.Sni)
	}

	urlConfig += "&insecure=0"

	if nodeConfig.Down != "" {
		// shadowrocket
		urlConfig += fmt.Sprintf("&downmbps=%s", url.PathEscape(nodeConfig.Down))
	}

	if nodeConfig.Ports != "" {
		// shadowrocket
		urlConfig += fmt.Sprintf("&mport=%s", nodeConfig.Ports)
	}

	// 添加节点名称作为备注
	urlConfig += fmt.Sprintf("#%s", nodeConfig.Name)

	if urlConfig != "" {
		urlConfig = "/?" + strings.TrimPrefix(urlConfig, "&")
	}
	
	return fmt.Sprintf("hysteria2://%s@%s:%s%s", nodeConfig.Password, hostname, nodeConfig.Port, urlConfig), nil
}/
/ Hysteria2MultiNodeUrl 生成多节点URL
func Hysteria2MultiNodeUrl(accountId int64, hostname string) ([]vo.Hysteria2NodeUrlVo, error) {
	account, err := dao.GetAccount("id = ?", accountId)
	if err != nil {
		return nil, err
	}

	// 生成多节点配置
	nodeConfigs, err := generateMultiNodeConfig(*account, constant.V2rayN, hostname, *account.NodeAccess)
	if err != nil {
		return nil, err
	}

	var nodeUrls []vo.Hysteria2NodeUrlVo
	for _, nodeConfig := range nodeConfigs {
		url, err := generateHysteria2Url(nodeConfig, hostname)
		if err != nil {
			return nil, err
		}

		// 生成二维码
		qrCode, err := qrcode.Encode(url, qrcode.Medium, 300)
		if err != nil {
			return nil, err
		}

		nodeUrl := vo.Hysteria2NodeUrlVo{
			NodeName: nodeConfig.Name,
			Url:      url,
			QrCode:   qrCode,
		}
		nodeUrls = append(nodeUrls, nodeUrl)
	}

	return nodeUrls, nil
}