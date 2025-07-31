package router

import (
	"github.com/gin-gonic/gin"
	"h-ui/controller"
)

func initConfigRouter(configApi *gin.RouterGroup) {
	config := configApi.Group("/config")
	{
		config.POST("/updateConfigs", controller.UpdateConfigs)
		config.GET("/getConfig", controller.GetConfig)
		config.POST("/listConfig", controller.ListConfig)
		config.GET("/getHysteria2Config", controller.GetHysteria2Config)
		config.POST("/updateHysteria2Config", controller.UpdateHysteria2Config)
		config.POST("/exportHysteria2Config", controller.ExportHysteria2Config)
		config.POST("/importHysteria2Config", controller.ImportHysteria2Config)
		config.GET("/getHysteria2Node2Config", controller.GetHysteria2Node2Config)
		config.POST("/updateHysteria2Node2Config", controller.UpdateHysteria2Node2Config)
		config.GET("/getSocks5Config", controller.GetSocks5Config)
		config.POST("/updateSocks5Config", controller.UpdateSocks5Config)
		config.GET("/getNode2Status", controller.GetNode2Status)
		config.POST("/toggleNode2", controller.ToggleNode2)
		config.GET("/getAllNodesStatus", controller.GetAllNodesStatus)
		config.POST("/exportNode2Config", controller.ExportNode2Config)
		config.POST("/importNode2Config", controller.ImportNode2Config)
		config.POST("/exportFullConfig", controller.ExportFullConfig)
		config.POST("/exportConfig", controller.ExportConfig)
		config.POST("/importConfig", controller.ImportConfig)
		config.GET("/hysteria2AcmePath", controller.Hysteria2AcmePath)
		config.POST("/restartServer", controller.RestartServer)
		config.POST("/uploadCertFile", controller.UploadCertFile)
	}
}
