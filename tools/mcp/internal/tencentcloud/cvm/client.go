package cvm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client CVM 客户端
type Client struct {
	client  *cvm.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建 CVM 客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("cvm")

	client, err := cvm.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CVM 客户端失败: %w", err)
	}

	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

// GetProductName 获取产品名称
func (c *Client) GetProductName() string {
	return "CVM"
}

// GetProductVersion 获取产品版本
func (c *Client) GetProductVersion() string {
	return "2017-03-12"
}

// ValidatePermissions 验证权限
func (c *Client) ValidatePermissions(ctx context.Context) error {
	_, err := c.DescribeInstances(ctx, "ap-beijing")
	if err != nil {
		return fmt.Errorf("CVM 权限验证失败: %w", err)
	}
	return nil
}

// --- Helper functions ---

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func getInt64Value(i *int64) int64 {
	if i != nil {
		return *i
	}
	return 0
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func convertStringPtrSlice(ptrs []*string) []string {
	result := make([]string, 0, len(ptrs))
	for _, p := range ptrs {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}

// --- DescribeInstances ---

// InstanceInfo CVM 实例信息
type InstanceInfo struct {
	InstanceId         string   `json:"instance_id"`
	InstanceName       string   `json:"instance_name"`
	InstanceType       string   `json:"instance_type"`
	InstanceState      string   `json:"instance_state"`
	InstanceChargeType string   `json:"instance_charge_type"`
	CPU                int64    `json:"cpu"`
	Memory             int64    `json:"memory"`
	OsName             string   `json:"os_name"`
	PrivateIpAddresses []string `json:"private_ip_addresses"`
	PublicIpAddresses  []string `json:"public_ip_addresses"`
	VpcId              string   `json:"vpc_id"`
	SubnetId           string   `json:"subnet_id"`
	CreatedTime        string   `json:"created_time"`
	ExpiredTime        string   `json:"expired_time"`
	Zone               string   `json:"zone"`
	ImageId            string   `json:"image_id"`
}

// DescribeInstancesResult 查询实例结果
type DescribeInstancesResult struct {
	TotalCount int64          `json:"total_count"`
	Instances  []InstanceInfo `json:"instances"`
	Region     string         `json:"region"`
}

// DescribeInstances 查询 CVM 实例列表
func (c *Client) DescribeInstances(ctx context.Context, region string) (*DescribeInstancesResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region": region,
	}).Debug("开始查询 CVM 实例列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cvm")
	client, err := cvm.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CVM 客户端失败: %w", err)
	}

	request := cvm.NewDescribeInstancesRequest()
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeInstances(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CVM API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CVM 实例列表失败: %w", err)
	}

	result := &DescribeInstancesResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, inst := range response.Response.InstanceSet {
		info := InstanceInfo{
			InstanceId:         getStringValue(inst.InstanceId),
			InstanceName:       getStringValue(inst.InstanceName),
			InstanceType:       getStringValue(inst.InstanceType),
			InstanceState:      getStringValue(inst.InstanceState),
			InstanceChargeType: getStringValue(inst.InstanceChargeType),
			CPU:                getInt64Value(inst.CPU),
			Memory:             getInt64Value(inst.Memory),
			OsName:             getStringValue(inst.OsName),
			PrivateIpAddresses: convertStringPtrSlice(inst.PrivateIpAddresses),
			PublicIpAddresses:  convertStringPtrSlice(inst.PublicIpAddresses),
			CreatedTime:        getStringValue(inst.CreatedTime),
			ExpiredTime:        getStringValue(inst.ExpiredTime),
			ImageId:            getStringValue(inst.ImageId),
		}
		if inst.Placement != nil {
			info.Zone = getStringValue(inst.Placement.Zone)
		}
		if inst.VirtualPrivateCloud != nil {
			info.VpcId = getStringValue(inst.VirtualPrivateCloud.VpcId)
			info.SubnetId = getStringValue(inst.VirtualPrivateCloud.SubnetId)
		}
		result.Instances = append(result.Instances, info)
	}

	c.logger.WithField("instance_count", len(result.Instances)).Info("成功查询 CVM 实例列表")
	return result, nil
}

// FormatInstancesAsJSON 格式化实例列表为 JSON
func (c *Client) FormatInstancesAsJSON(result *DescribeInstancesResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatInstancesAsTable 格式化实例列表为表格
func (c *Client) FormatInstancesAsTable(result *DescribeInstancesResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CVM 实例列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 140) + "\n")
	sb.WriteString(fmt.Sprintf("%-20s %-20s %-15s %-12s %-6s %-8s %-18s %-18s %-15s\n",
		"实例ID", "名称", "机型", "状态", "CPU", "内存(GB)", "内网IP", "公网IP", "可用区"))
	sb.WriteString(strings.Repeat("-", 140) + "\n")

	for _, inst := range result.Instances {
		privateIP := "-"
		if len(inst.PrivateIpAddresses) > 0 {
			privateIP = inst.PrivateIpAddresses[0]
		}
		publicIP := "-"
		if len(inst.PublicIpAddresses) > 0 {
			publicIP = inst.PublicIpAddresses[0]
		}
		sb.WriteString(fmt.Sprintf("%-20s %-20s %-15s %-12s %-6d %-8d %-18s %-18s %-15s\n",
			inst.InstanceId,
			truncateString(inst.InstanceName, 18),
			truncateString(inst.InstanceType, 13),
			inst.InstanceState,
			inst.CPU,
			inst.Memory,
			privateIP,
			publicIP,
			inst.Zone))
	}

	return sb.String()
}

// --- DescribeInstancesStatus ---

// InstanceStatusInfo CVM 实例状态信息
type InstanceStatusInfo struct {
	InstanceId    string `json:"instance_id"`
	InstanceState string `json:"instance_state"`
}

// DescribeInstancesStatusResult 查询实例状态结果
type DescribeInstancesStatusResult struct {
	TotalCount int64                `json:"total_count"`
	Instances  []InstanceStatusInfo `json:"instances"`
	Region     string               `json:"region"`
}

// DescribeInstancesStatus 查询 CVM 实例状态列表
func (c *Client) DescribeInstancesStatus(ctx context.Context, region string) (*DescribeInstancesStatusResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region": region,
	}).Debug("开始查询 CVM 实例状态列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cvm")
	client, err := cvm.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CVM 客户端失败: %w", err)
	}

	request := cvm.NewDescribeInstancesStatusRequest()
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeInstancesStatus(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CVM API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CVM 实例状态失败: %w", err)
	}

	result := &DescribeInstancesStatusResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, status := range response.Response.InstanceStatusSet {
		result.Instances = append(result.Instances, InstanceStatusInfo{
			InstanceId:    getStringValue(status.InstanceId),
			InstanceState: getStringValue(status.InstanceState),
		})
	}

	c.logger.WithField("instance_count", len(result.Instances)).Info("成功查询 CVM 实例状态列表")
	return result, nil
}

// FormatInstancesStatusAsJSON 格式化实例状态为 JSON
func (c *Client) FormatInstancesStatusAsJSON(result *DescribeInstancesStatusResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatInstancesStatusAsTable 格式化实例状态为表格
func (c *Client) FormatInstancesStatusAsTable(result *DescribeInstancesStatusResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CVM 实例状态列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-25s %-20s\n", "实例ID", "状态"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, inst := range result.Instances {
		sb.WriteString(fmt.Sprintf("%-25s %-20s\n", inst.InstanceId, inst.InstanceState))
	}

	return sb.String()
}
