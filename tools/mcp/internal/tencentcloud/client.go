package tencentcloud

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	"github.com/sirupsen/logrus"
)

// Config 腾讯云配置
type Config struct {
	SecretID    string `json:"secret_id" yaml:"secret_id"`
	SecretKey   string `json:"secret_key" yaml:"secret_key"`
	Region      string `json:"region" yaml:"region"`
	Endpoint    string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	UseInternal bool   `json:"use_internal,omitempty" yaml:"use_internal,omitempty"` // 是否使用内网域名访问云 API
}

// GetEndpointForProduct 根据产品名获取 API 域名
// 优先级: Endpoint(手动指定) > UseInternal(内网域名) > SDK默认(公网域名)
// 产品域名规则:
//   公网: {product}.tencentcloudapi.com
//   内网: {product}.internal.tencentcloudapi.com
func (c *Config) GetEndpointForProduct(product string) string {
	// 手动指定 endpoint 优先级最高
	if c.Endpoint != "" {
		return c.Endpoint
	}
	// 内网模式: 自动拼接内网域名
	if c.UseInternal {
		return fmt.Sprintf("%s.internal.tencentcloudapi.com", product)
	}
	// 返回空字符串，由 SDK 自动使用默认公网域名
	return ""
}

// ClientManager 腾讯云客户端管理器
type ClientManager struct {
	config *Config
	logger *logrus.Logger
}

// NewClientManager 创建腾讯云客户端管理器
func NewClientManager(config *Config, logger *logrus.Logger) *ClientManager {
	return &ClientManager{
		config: config,
		logger: logger,
	}
}

// GetCredential 获取腾讯云凭证
func (cm *ClientManager) GetCredential() *common.Credential {
	return common.NewCredential(cm.config.SecretID, cm.config.SecretKey)
}

// GetClientProfile 获取客户端配置
// product 为产品标识（如 tke、cvm、clb、cdb、vpc），用于生成对应的 API 域名
func (cm *ClientManager) GetClientProfile(product string) *profile.ClientProfile {
	cpf := profile.NewClientProfile()
	
	// 根据产品名获取对应的 endpoint
	endpoint := cm.config.GetEndpointForProduct(product)
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}
	
	// 设置请求方法和协议
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 30
	cpf.SignMethod = "HmacSHA256"
	
	return cpf
}

// ValidateConfig 验证配置
func (cm *ClientManager) ValidateConfig() error {
	if cm.config.SecretID == "" {
		return fmt.Errorf("腾讯云 SecretID 不能为空")
	}
	if cm.config.SecretKey == "" {
		return fmt.Errorf("腾讯云 SecretKey 不能为空")
	}
	return nil
}

// GetAvailableRegions 获取可用地域列表
func GetAvailableRegions() []RegionInfo {
	return []RegionInfo{
		{Region: regions.Beijing, Name: "华北地区(北京)", EnglishName: "Beijing"},
		{Region: regions.Shanghai, Name: "华东地区(上海)", EnglishName: "Shanghai"},
		{Region: regions.Guangzhou, Name: "华南地区(广州)", EnglishName: "Guangzhou"},
		{Region: "ap-shenzhen", Name: "华南地区(深圳)", EnglishName: "Shenzhen"},
		{Region: regions.Chengdu, Name: "西南地区(成都)", EnglishName: "Chengdu"},
		{Region: regions.Chongqing, Name: "西南地区(重庆)", EnglishName: "Chongqing"},
		{Region: regions.HongKong, Name: "港澳台地区(香港)", EnglishName: "Hong Kong"},
		{Region: regions.Singapore, Name: "亚太地区(新加坡)", EnglishName: "Singapore"},
		{Region: regions.Tokyo, Name: "亚太地区(东京)", EnglishName: "Tokyo"},
		{Region: regions.Seoul, Name: "亚太地区(首尔)", EnglishName: "Seoul"},
		{Region: regions.Mumbai, Name: "亚太地区(孟买)", EnglishName: "Mumbai"},
		{Region: regions.Bangkok, Name: "亚太地区(曼谷)", EnglishName: "Bangkok"},
		{Region: "na-ashburn", Name: "美国东部(弗吉尼亚)", EnglishName: "Virginia"},
		{Region: regions.SiliconValley, Name: "美国西部(硅谷)", EnglishName: "Silicon Valley"},
		{Region: regions.Toronto, Name: "北美地区(多伦多)", EnglishName: "Toronto"},
		{Region: regions.Frankfurt, Name: "欧洲地区(法兰克福)", EnglishName: "Frankfurt"},
		{Region: regions.Moscow, Name: "欧洲地区(莫斯科)", EnglishName: "Moscow"},
	}
}

// RegionInfo 地域信息
type RegionInfo struct {
	Region      string `json:"region"`
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
}

// ProductClient 产品客户端接口
type ProductClient interface {
	GetProductName() string
	GetProductVersion() string
	ValidatePermissions(ctx context.Context) error
}