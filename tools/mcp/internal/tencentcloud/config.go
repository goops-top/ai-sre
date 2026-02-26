package tencentcloud

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LoadConfigFromEnv 从环境变量加载腾讯云配置
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		SecretID:    os.Getenv("TENCENTCLOUD_SECRET_ID"),
		SecretKey:   os.Getenv("TENCENTCLOUD_SECRET_KEY"),
		Region:      os.Getenv("TENCENTCLOUD_REGION"),
		Endpoint:    os.Getenv("TENCENTCLOUD_ENDPOINT"),
		UseInternal: strings.EqualFold(os.Getenv("TENCENTCLOUD_USE_INTERNAL"), "true"),
	}

	// 如果没有设置地域，默认使用北京
	if config.Region == "" {
		config.Region = "ap-beijing"
	}

	return config, nil
}

// LoadConfigFromFile 从配置文件加载腾讯云配置
func LoadConfigFromFile(configPath string) (*Config, error) {
	// 这里可以实现从 JSON/YAML 文件加载配置的逻辑
	// 暂时返回空配置，后续可以扩展
	return &Config{}, nil
}

// GetConfigFromMultipleSources 从多个来源获取配置
func GetConfigFromMultipleSources(logger *logrus.Logger) (*Config, error) {
	// 优先级：环境变量 > 配置文件 > 默认值
	
	// 1. 尝试从环境变量加载
	config, err := LoadConfigFromEnv()
	if err != nil {
		logger.WithError(err).Warn("Failed to load config from environment variables")
	}
	
	// 2. 验证必要的配置项
	if config.SecretID == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("腾讯云认证信息不完整，请设置环境变量 TENCENTCLOUD_SECRET_ID 和 TENCENTCLOUD_SECRET_KEY")
	}
	
	logger.WithFields(logrus.Fields{
		"region":       config.Region,
		"endpoint":     config.Endpoint,
		"use_internal": config.UseInternal,
	}).Info("腾讯云配置加载成功")
	
	return config, nil
}

// MaskSensitiveInfo 屏蔽敏感信息用于日志输出
func (c *Config) MaskSensitiveInfo() map[string]string {
	masked := make(map[string]string)
	
	if c.SecretID != "" {
		if len(c.SecretID) > 8 {
			masked["secret_id"] = c.SecretID[:4] + "****" + c.SecretID[len(c.SecretID)-4:]
		} else {
			masked["secret_id"] = "****"
		}
	}
	
	if c.SecretKey != "" {
		masked["secret_key"] = "****"
	}
	
	masked["region"] = c.Region
	masked["endpoint"] = c.Endpoint
	if c.UseInternal {
		masked["use_internal"] = "true"
	} else {
		masked["use_internal"] = "false"
	}
	
	return masked
}

// GetRegionDisplayName 获取地域的显示名称
func GetRegionDisplayName(region string) string {
	regionMap := map[string]string{
		"ap-beijing":     "华北地区(北京)",
		"ap-shanghai":    "华东地区(上海)",
		"ap-guangzhou":   "华南地区(广州)",
		"ap-shenzhen":    "华南地区(深圳)",
		"ap-chengdu":     "西南地区(成都)",
		"ap-chongqing":   "西南地区(重庆)",
		"ap-hongkong":    "港澳台地区(香港)",
		"ap-singapore":   "亚太地区(新加坡)",
		"ap-tokyo":       "亚太地区(东京)",
		"ap-seoul":       "亚太地区(首尔)",
		"ap-mumbai":      "亚太地区(孟买)",
		"ap-bangkok":     "亚太地区(曼谷)",
		"na-ashburn":     "美国东部(弗吉尼亚)",
		"na-siliconvalley": "美国西部(硅谷)",
		"na-toronto":     "北美地区(多伦多)",
		"eu-frankfurt":   "欧洲地区(法兰克福)",
		"eu-moscow":      "欧洲地区(莫斯科)",
	}
	
	if displayName, exists := regionMap[region]; exists {
		return displayName
	}
	return region
}

// ValidateRegion 验证地域是否有效
func ValidateRegion(region string) bool {
	validRegions := []string{
		"ap-beijing", "ap-shanghai", "ap-guangzhou", "ap-shenzhen",
		"ap-chengdu", "ap-chongqing", "ap-hongkong", "ap-singapore",
		"ap-tokyo", "ap-seoul", "ap-mumbai", "ap-bangkok",
		"na-ashburn", "na-siliconvalley", "na-toronto",
		"eu-frankfurt", "eu-moscow",
	}
	
	for _, validRegion := range validRegions {
		if strings.EqualFold(region, validRegion) {
			return true
		}
	}
	return false
}