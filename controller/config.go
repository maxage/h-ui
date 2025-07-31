package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/dto"
	"h-ui/model/entity"
	"h-ui/model/vo"
	"h-ui/service"
	"h-ui/util"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func UpdateConfigs(c *gin.Context) {
	configsUpdateDto, err := validateField(c, dto.ConfigsUpdateDto{})
	if err != nil {
		return
	}

	port, crtPath, keyPath, err := service.GetPortAndCert()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	needResetPortHopping := false
	needRestart := false

	for _, item := range configsUpdateDto.ConfigUpdateDtos {
		key := *item.Key
		value := *item.Value

		if key == constant.HUIWebPort && strconv.FormatInt(port, 10) != value {
			port, err := strconv.Atoi(value)
			if err != nil {
				vo.Fail(fmt.Sprintf("port: %s is invalid", value), c)
				return
			}
			if !util.IsPortAvailable(uint(port), "tcp") {
				vo.Fail(fmt.Sprintf("port: %s is used", value), c)
				return
			}
			needRestart = true
		}
		if key == constant.HUICrtPath && crtPath != value {
			if value != "" && !util.Exists(value) {
				vo.Fail(fmt.Sprintf("crt path: %s is not exist", value), c)
				return
			}
			needRestart = true
		}
		if key == constant.HUIKeyPath && keyPath != value {
			if value != "" && !util.Exists(value) {
				vo.Fail(fmt.Sprintf("key path: %s is not exist", value), c)
				return
			}
			needRestart = true
		}

		if key == constant.HUIWebContext {
			huiWebContext, err := service.GetConfig(constant.HUIWebContext)
			if err != nil {
				vo.Fail(err.Error(), c)
				return
			}
			if *huiWebContext.Value != value {
				needRestart = true
			}
		}

		if key == constant.Hysteria2ConfigPortHopping {
			re := regexp.MustCompile("^\\d+(?:-\\d+)?(?:,\\d+(?:-\\d+)?)*$")
			if value != "" && !re.MatchString(value) {
				vo.Fail(fmt.Sprintf("port hopping: %s is invalid", value), c)
				return
			}
			hysteria2ConfigPortHopping, err := service.GetConfig(constant.Hysteria2ConfigPortHopping)
			if err != nil {
				vo.Fail(err.Error(), c)
				return
			}
			if *hysteria2ConfigPortHopping.Value != value {
				needResetPortHopping = true
			}
		}

		if key == constant.ResetTrafficCron {
			resetTrafficCron, err := service.GetConfig(constant.ResetTrafficCron)
			if err != nil {
				vo.Fail(err.Error(), c)
				return
			}
			if *resetTrafficCron.Value != value {
				needRestart = true
			}
		}

		if key == constant.TelegramEnable {
			telegramEnable, err := service.GetConfig(constant.TelegramEnable)
			if err != nil {
				vo.Fail(err.Error(), c)
				return
			}
			if *telegramEnable.Value != value {
				needRestart = true
			}
		}

		if err = service.UpdateConfig(key, value); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	if needResetPortHopping {
		if err := service.InitPortHopping(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	if needRestart {
		go func() {
			_ = service.StopServer()
		}()
	}

	vo.Success(nil, c)
}

func GetConfig(c *gin.Context) {
	configDto, err := validateField(c, dto.ConfigDto{})
	if err != nil {
		return
	}
	config, err := service.GetConfig(*configDto.Key)
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	configVo := vo.ConfigVo{
		Key:   *config.Key,
		Value: *config.Value,
	}

	running := service.Hysteria2IsRunning()

	if (*config.Value == "1") != running {
		enable := "0"
		if running {
			enable = "1"
		}
		if err := service.UpdateConfig(constant.Hysteria2Enable, enable); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
		configVo.Value = enable
	}

	vo.Success(configVo, c)
}

func ListConfig(c *gin.Context) {
	configsDto, err := validateField(c, dto.ConfigsDto{})
	if err != nil {
		return
	}
	configs, err := service.ListConfig(configsDto.Keys)
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	var configVos []vo.ConfigVo
	for _, item := range configs {
		configVo := vo.ConfigVo{
			Key:   *item.Key,
			Value: *item.Value,
		}
		configVos = append(configVos, configVo)
	}
	vo.Success(configVos, c)
}

func GetHysteria2Config(c *gin.Context) {
	config, err := service.GetHysteria2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	vo.Success(config, c)
}

func UpdateHysteria2Config(c *gin.Context) {
	hysteria2ServerConfig, err := validateField(c, bo.Hysteria2ServerConfig{})
	if err != nil {
		return
	}

	hysteria2Config, err := service.GetHysteria2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	needResetPortHopping := false
	if hysteria2Config.Listen != nil &&
		*hysteria2Config.Listen != "" &&
		hysteria2ServerConfig.Listen != nil &&
		*hysteria2ServerConfig.Listen != "" &&
		*hysteria2ServerConfig.Listen != *hysteria2Config.Listen {
		needResetPortHopping = true
	}

	if err = service.UpdateHysteria2Config(hysteria2ServerConfig); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	if needResetPortHopping {
		if err := service.InitPortHopping(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	running := service.Hysteria2IsRunning()
	if running {
		if err = service.RestartHysteria2(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}
	vo.Success(nil, c)
}

func ExportHysteria2Config(c *gin.Context) {
	hysteria2ServerConfig, err := service.GetHysteria2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 默认值
	config, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.JwtSecret})
	if err != nil {
		vo.Fail(err.Error(), c)
		return
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
		vo.Fail(constant.SysError, c)
		return
	}

	authHttpUrl, err := service.GetAuthHttpUrl()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
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

	fileName := fmt.Sprintf("Hysteria2Config-%s.yaml", time.Now().Format("20060102150405"))
	filePath := constant.ExportPathDir + fileName

	if err = util.ExportFile(filePath, hysteria2ServerConfig, 1); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	if !util.Exists(filePath) {
		vo.Fail("file not exist", c)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.File(filePath)
}

func ImportHysteria2Config(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		vo.Fail(constant.SysError, c)
		return
	}
	if header.Size > 1024*1024*2 {
		vo.Fail("the file is too big", c)
		return
	}
	if !strings.HasSuffix(header.Filename, ".yaml") {
		vo.Fail("file format not supported", c)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		vo.Fail("yaml file read err", c)
		return
	}
	var hysteria2ServerConfig bo.Hysteria2ServerConfig
	if err = yaml.Unmarshal(content, &hysteria2ServerConfig); err != nil {
		vo.Fail("content Unmarshal err", c)
		return
	}

	// 默认值
	config, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.JwtSecret})
	if err != nil {
		vo.Fail(err.Error(), c)
		return
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
		vo.Fail(constant.SysError, c)
		return
	}

	authHttpUrl, err := service.GetAuthHttpUrl()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
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

	if err = service.SetHysteria2Config(hysteria2ServerConfig); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	running := service.Hysteria2IsRunning()
	if running {
		if err = service.RestartHysteria2(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	vo.Success(nil, c)
}

func ExportConfig(c *gin.Context) {
	configs, err := service.ListConfigNotIn([]string{constant.Hysteria2Config})
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	fileName := fmt.Sprintf("SystemConfig-%s.json", time.Now().Format("20060102150405"))
	filePath := constant.ExportPathDir + fileName

	if err = util.ExportFile(filePath, configs, 0); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	if !util.Exists(filePath) {
		vo.Fail("file not exist", c)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.File(filePath)
}

func ImportConfig(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		vo.Fail(constant.SysError, c)
		return
	}
	if header.Size > 1024*1024*2 {
		vo.Fail("the file is too big", c)
		return
	}
	if !strings.HasSuffix(header.Filename, ".json") {
		vo.Fail("file format not supported", c)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		vo.Fail("json file read err", c)
		return
	}
	var configs []entity.Config
	if err = json.Unmarshal(content, &configs); err != nil {
		vo.Fail("content Unmarshal err", c)
		return
	}
	if err = service.UpsertConfig(configs); err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	go func() {
		_ = service.StopServer()
	}()
	vo.Success(nil, c)
}

func Hysteria2AcmePath(c *gin.Context) {
	hysteria2AcmePathVo, err := service.Hysteria2AcmePath()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	vo.Success(hysteria2AcmePathVo, c)
}

func RestartServer(c *gin.Context) {
	go func() {
		_ = service.StopServer()
	}()
	vo.Success(nil, c)
}

func UploadCertFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		vo.Fail(constant.SysError, c)
		return
	}
	ext := filepath.Ext(file.Filename)
	if ext != ".crt" && ext != ".key" {
		vo.Fail("file format not supported", c)
		return
	}
	if file.Size > 1024*1024 {
		vo.Fail("the file is too big", c)
		return
	}
	err = filepath.WalkDir(constant.BinDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fileExt := filepath.Ext(path)
		if !d.IsDir() && fileExt == ext {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to delete file: %s, error: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("error during file deletion: %v", err)
		vo.Fail("delete file failed", c)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		vo.Fail(constant.SysError, c)
		return
	}
	safeFilename := filepath.Base(file.Filename)
	certPath := filepath.Join(wd, constant.BinDir, safeFilename)

	if err := c.SaveUploadedFile(file, certPath); err != nil {
		vo.Fail("file upload failed", c)
		return
	}
	vo.Success(certPath, c)
}
// GetHysteria2Node2Config 获取第二节点配置
func GetHysteria2Node2Config(c *gin.Context) {
	config, err := service.GetHysteria2Node2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	vo.Success(config, c)
}

// UpdateHysteria2Node2Config 更新第二节点配置
func UpdateHysteria2Node2Config(c *gin.Context) {
	hysteria2ServerConfig, err := validateField(c, bo.Hysteria2ServerConfig{})
	if err != nil {
		return
	}

	if err = service.UpdateHysteria2Node2Config(hysteria2ServerConfig); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 如果第二节点正在运行，重启它
	if service.Hysteria2Node2IsRunning() {
		if err = service.RestartHysteria2Node2(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}
	vo.Success(nil, c)
}

// GetSocks5Config 获取SOCKS5配置
func GetSocks5Config(c *gin.Context) {
	config, err := service.GetSocks5Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	
	// 隐藏密码
	configVo := vo.Socks5ConfigVo{
		Addr:     config.Addr,
		Username: config.Username,
		// Password 不返回，保护隐私
	}
	vo.Success(configVo, c)
}

// UpdateSocks5Config 更新SOCKS5配置
func UpdateSocks5Config(c *gin.Context) {
	socks5ConfigDto, err := validateField(c, dto.Socks5ConfigDto{})
	if err != nil {
		return
	}

	socks5Config := bo.Socks5Config{
		Addr:     socks5ConfigDto.Addr,
		Username: socks5ConfigDto.Username,
		Password: socks5ConfigDto.Password,
	}

	if err = service.UpdateSocks5Config(socks5Config); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 如果第二节点正在运行，重启它以应用新的SOCKS5配置
	if service.Hysteria2Node2IsRunning() {
		if err = service.RestartHysteria2Node2(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	vo.Success(nil, c)
}

// GetNode2Status 获取第二节点状态
func GetNode2Status(c *gin.Context) {
	enabled, err := service.IsNode2Enabled()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	port := 0
	if enabled {
		nodePort, err := service.GetNode2Port()
		if err == nil {
			port = nodePort
		}
	}

	node2Remark, err := service.GetConfig(constant.Hysteria2Node2Remark)
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	statusVo := vo.Node2ConfigVo{
		Enable: enabled,
		Remark: *node2Remark.Value,
		Port:   port,
		Status: service.Hysteria2Node2IsRunning(),
	}

	vo.Success(statusVo, c)
}

// ToggleNode2 切换第二节点开关
func ToggleNode2(c *gin.Context) {
	node2ConfigDto, err := validateField(c, dto.Node2ConfigDto{})
	if err != nil {
		return
	}

	// 更新开关状态
	enableValue := "0"
	if node2ConfigDto.Enable {
		enableValue = "1"
	}

	if err = service.UpdateConfig(constant.Hysteria2Node2Enable, enableValue); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 更新备注
	if node2ConfigDto.Remark != "" {
		if err = service.UpdateConfig(constant.Hysteria2Node2Remark, node2ConfigDto.Remark); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	}

	// 根据开关状态启动或停止第二节点
	if node2ConfigDto.Enable {
		// 检查SOCKS5配置是否完整
		socks5Config, err := service.GetSocks5Config()
		if err != nil {
			vo.Fail(err.Error(), c)
			return
		}
		if socks5Config.Addr == "" {
			vo.Fail("SOCKS5 address is required", c)
			return
		}

		// 启动第二节点
		if err = service.StartHysteria2Node2(); err != nil {
			vo.Fail(err.Error(), c)
			return
		}
	} else {
		// 停止第二节点
		if service.Hysteria2Node2IsRunning() {
			if err = service.StopHysteria2Node2(); err != nil {
				vo.Fail(err.Error(), c)
				return
			}
		}
		
		// 将所有双节点用户降级为单节点
		if err = service.DowngradeUsersToSingleNode(); err != nil {
			logrus.Errorf("failed to downgrade users to single node: %v", err)
			// 不返回错误，因为节点已经成功禁用，用户降级失败不应该影响主要操作
		}
	}

	vo.Success(nil, c)
}

// GetAllNodesStatus 获取所有节点状态
func GetAllNodesStatus(c *gin.Context) {
	status := service.GetNodesStatus()
	vo.Success(status, c)
}// 
ExportNode2Config 导出第二节点配置
func ExportNode2Config(c *gin.Context) {
	// 检查第二节点是否启用
	enabled, err := service.IsNode2Enabled()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}
	if !enabled {
		vo.Fail("node2 is not enabled", c)
		return
	}

	// 获取第二节点配置
	node2Config, err := service.GetHysteria2Node2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 获取SOCKS5配置
	socks5Config, err := service.GetSocks5Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 组合导出数据
	exportData := map[string]interface{}{
		"node2Config":  node2Config,
		"socks5Config": socks5Config,
		"exportTime":   time.Now().Format("2006-01-02 15:04:05"),
		"version":      "1.0",
	}

	fileName := fmt.Sprintf("Node2Config-%s.json", time.Now().Format("20060102150405"))
	filePath := constant.ExportPathDir + fileName

	if err = util.ExportFile(filePath, exportData, 0); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	if !util.Exists(filePath) {
		vo.Fail("file not exist", c)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.File(filePath)
}

// ImportNode2Config 导入第二节点配置
func ImportNode2Config(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		vo.Fail(constant.SysError, c)
		return
	}
	if header.Size > 1024*1024*2 {
		vo.Fail("the file is too big", c)
		return
	}
	if !strings.HasSuffix(header.Filename, ".json") {
		vo.Fail("file format not supported", c)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		vo.Fail("json file read err", c)
		return
	}

	var importData map[string]interface{}
	if err = json.Unmarshal(content, &importData); err != nil {
		vo.Fail("content Unmarshal err", c)
		return
	}

	// 验证导入数据格式
	if _, exists := importData["node2Config"]; !exists {
		vo.Fail("invalid node2 config file format", c)
		return
	}
	if _, exists := importData["socks5Config"]; !exists {
		vo.Fail("invalid socks5 config file format", c)
		return
	}

	// 解析第二节点配置
	node2ConfigData, _ := json.Marshal(importData["node2Config"])
	var node2Config bo.Hysteria2ServerConfig
	if err = json.Unmarshal(node2ConfigData, &node2Config); err != nil {
		vo.Fail("invalid node2 config format", c)
		return
	}

	// 解析SOCKS5配置
	socks5ConfigData, _ := json.Marshal(importData["socks5Config"])
	var socks5Config bo.Socks5Config
	if err = json.Unmarshal(socks5ConfigData, &socks5Config); err != nil {
		vo.Fail("invalid socks5 config format", c)
		return
	}

	// 导入第二节点配置
	if err = service.UpdateHysteria2Node2Config(node2Config); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 导入SOCKS5配置
	if err = service.UpdateSocks5Config(socks5Config); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	vo.Success(nil, c)
}

// ExportFullConfig 导出完整配置（包含第二节点）
func ExportFullConfig(c *gin.Context) {
	// 获取系统配置
	systemConfigs, err := service.ListConfigNotIn([]string{constant.Hysteria2Config, constant.Hysteria2Node2Config})
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	// 获取主节点配置
	hysteria2Config, err := service.GetHysteria2Config()
	if err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	exportData := map[string]interface{}{
		"systemConfigs":   systemConfigs,
		"hysteria2Config": hysteria2Config,
		"exportTime":      time.Now().Format("2006-01-02 15:04:05"),
		"version":         "1.0",
	}

	// 如果第二节点启用，包含第二节点配置
	enabled, err := service.IsNode2Enabled()
	if err == nil && enabled {
		node2Config, err := service.GetHysteria2Node2Config()
		if err == nil {
			exportData["node2Config"] = node2Config
		}

		socks5Config, err := service.GetSocks5Config()
		if err == nil {
			exportData["socks5Config"] = socks5Config
		}
	}

	fileName := fmt.Sprintf("FullConfig-%s.json", time.Now().Format("20060102150405"))
	filePath := constant.ExportPathDir + fileName

	if err = util.ExportFile(filePath, exportData, 0); err != nil {
		vo.Fail(err.Error(), c)
		return
	}

	if !util.Exists(filePath) {
		vo.Fail("file not exist", c)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.File(filePath)
}