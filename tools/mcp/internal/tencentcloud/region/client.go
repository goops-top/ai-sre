package region

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"github.com/sirupsen/logrus"
	
	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client 地域管理客户端（基于 CVM 的 DescribeRegions）
type Client struct {
	client  *cvm.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建地域管理客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("cvm")
	
	// 使用默认地域创建 CVM 客户端
	client, err := cvm.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建地域管理客户端失败: %w", err)
	}
	
	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

// GetProductName 获取产品名称
func (c *Client) GetProductName() string {
	return "Region"
}

// GetProductVersion 获取产品版本
func (c *Client) GetProductVersion() string {
	return "2017-03-12"
}

// RegionInfo 地域信息
type RegionInfo struct {
	RegionID    string `json:"region_id"`
	RegionName  string `json:"region_name"`
	RegionState string `json:"region_state"`
}

// DescribeRegions 查询产品支持的地域信息（使用 CVM 的 DescribeRegions）
func (c *Client) DescribeRegions(ctx context.Context, product string) ([]RegionInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"product": product,
	}).Debug("开始查询产品支持的地域信息")
	
	// 创建请求
	request := cvm.NewDescribeRegionsRequest()
	
	// 发送请求
	response, err := c.client.DescribeRegions(request)
	if err != nil {
		// 处理腾讯云 SDK 错误
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"product":    product,
			}).Error("地域管理 API 调用失败")
			return nil, fmt.Errorf("地域管理 API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询产品 %s 地域信息失败: %w", product, err)
	}
	
	// 转换响应数据
	var regions []RegionInfo
	for _, regionData := range response.Response.RegionSet {
		regionInfo := RegionInfo{
			RegionID:    *regionData.Region,
			RegionName:  *regionData.RegionName,
			RegionState: *regionData.RegionState,
		}
		regions = append(regions, regionInfo)
	}
	
	c.logger.WithFields(logrus.Fields{
		"product":      product,
		"region_count": len(regions),
	}).Info("成功查询产品地域信息")
	
	return regions, nil
}

// FormatRegionsAsJSON 将地域信息格式化为 JSON
func (c *Client) FormatRegionsAsJSON(regions []RegionInfo) (string, error) {
	data, err := json.MarshalIndent(regions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化地域信息失败: %w", err)
	}
	return string(data), nil
}

// FormatRegionsAsTable 将地域信息格式化为表格
func (c *Client) FormatRegionsAsTable(regions []RegionInfo, product string) string {
	if len(regions) == 0 {
		return fmt.Sprintf("未找到产品 %s 的任何地域信息", product)
	}
	
	result := fmt.Sprintf("%s 支持的地域信息:\n", product)
	result += "┌─────────────────┬──────────────────────────┬──────────────┐\n"
	result += "│ 地域ID          │ 地域名称                 │ 状态         │\n"
	result += "├─────────────────┼──────────────────────────┼──────────────┤\n"
	
	for _, region := range regions {
		result += fmt.Sprintf("│ %-15s │ %-24s │ %-12s │\n", 
			region.RegionID, region.RegionName, region.RegionState)
	}
	
	result += "└─────────────────┴──────────────────────────┴──────────────┘\n"
	result += fmt.Sprintf("总计: %d 个地域", len(regions))
	
	return result
}

// GetRegionByID 根据地域ID获取地域信息
func (c *Client) GetRegionByID(ctx context.Context, product, regionID string) (*RegionInfo, error) {
	regions, err := c.DescribeRegions(ctx, product)
	if err != nil {
		return nil, err
	}
	
	// 按地域ID匹配
	for _, region := range regions {
		if region.RegionID == regionID {
			return &region, nil
		}
	}
	
	// 按地域名称匹配
	for _, region := range regions {
		if region.RegionName == regionID {
			return &region, nil
		}
	}
	
	return nil, fmt.Errorf("未找到产品 %s 的地域 %s", product, regionID)
}

// ValidatePermissions 验证权限
func (c *Client) ValidatePermissions(ctx context.Context) error {
	// 通过调用 DescribeRegions 来验证权限
	_, err := c.DescribeRegions(ctx, "cvm")
	if err != nil {
		return fmt.Errorf("地域管理权限验证失败: %w", err)
	}
	return nil
}