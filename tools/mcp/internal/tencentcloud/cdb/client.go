package cdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"

	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client CDB 客户端
type Client struct {
	client  *cdb.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建 CDB 客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("cdb")

	client, err := cdb.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}

	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

// GetProductName 获取产品名称
func (c *Client) GetProductName() string {
	return "CDB"
}

// GetProductVersion 获取产品版本
func (c *Client) GetProductVersion() string {
	return "2017-03-20"
}

// ValidatePermissions 验证权限
func (c *Client) ValidatePermissions(ctx context.Context) error {
	_, err := c.DescribeDBInstances(ctx, "ap-beijing")
	if err != nil {
		return fmt.Errorf("CDB 权限验证失败: %w", err)
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

func getUint64Value(u *uint64) uint64 {
	if u != nil {
		return *u
	}
	return 0
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// --- DescribeDBInstances ---

// DBInstanceInfo CDB 实例信息
type DBInstanceInfo struct {
	InstanceId    string `json:"instance_id"`
	InstanceName  string `json:"instance_name"`
	InstanceType  int64  `json:"instance_type"`
	Status        int64  `json:"status"`
	Memory        int64  `json:"memory_mb"`
	Volume        int64  `json:"volume_gb"`
	Cpu           int64  `json:"cpu"`
	Qps           int64  `json:"qps"`
	EngineVersion string `json:"engine_version"`
	EngineType    string `json:"engine_type"`
	Vip           string `json:"vip"`
	Vport         int64  `json:"vport"`
	UniqVpcId     string `json:"uniq_vpc_id"`
	UniqSubnetId  string `json:"uniq_subnet_id"`
	Zone          string `json:"zone"`
	Region        string `json:"region"`
	PayType       int64  `json:"pay_type"`
	CreateTime    string `json:"create_time"`
	DeadlineTime  string `json:"deadline_time"`
	TaskStatus    int64  `json:"task_status"`
	WanStatus     int64  `json:"wan_status"`
	WanDomain     string `json:"wan_domain"`
	WanPort       int64  `json:"wan_port"`
}

// DescribeDBInstancesResult 查询 CDB 实例结果
type DescribeDBInstancesResult struct {
	TotalCount int64            `json:"total_count"`
	Instances  []DBInstanceInfo `json:"instances"`
	Region     string           `json:"region"`
}

// DescribeDBInstances 查询 CDB 实例列表
func (c *Client) DescribeDBInstances(ctx context.Context, region string) (*DescribeDBInstancesResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region": region,
	}).Debug("开始查询 CDB 实例列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cdb")
	client, err := cdb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}

	request := cdb.NewDescribeDBInstancesRequest()
	var limit uint64 = 2000
	request.Limit = &limit

	response, err := client.DescribeDBInstances(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CDB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CDB 实例列表失败: %w", err)
	}

	result := &DescribeDBInstancesResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, inst := range response.Response.Items {
		info := DBInstanceInfo{
			InstanceId:    getStringValue(inst.InstanceId),
			InstanceName:  getStringValue(inst.InstanceName),
			InstanceType:  getInt64Value(inst.InstanceType),
			Status:        getInt64Value(inst.Status),
			Memory:        getInt64Value(inst.Memory),
			Volume:        getInt64Value(inst.Volume),
			Cpu:           getInt64Value(inst.Cpu),
			Qps:           getInt64Value(inst.Qps),
			EngineVersion: getStringValue(inst.EngineVersion),
			EngineType:    getStringValue(inst.EngineType),
			Vip:           getStringValue(inst.Vip),
			Vport:         getInt64Value(inst.Vport),
			UniqVpcId:     getStringValue(inst.UniqVpcId),
			UniqSubnetId:  getStringValue(inst.UniqSubnetId),
			Zone:          getStringValue(inst.Zone),
			Region:        getStringValue(inst.Region),
			PayType:       getInt64Value(inst.PayType),
			CreateTime:    getStringValue(inst.CreateTime),
			DeadlineTime:  getStringValue(inst.DeadlineTime),
			TaskStatus:    getInt64Value(inst.TaskStatus),
			WanStatus:     getInt64Value(inst.WanStatus),
			WanDomain:     getStringValue(inst.WanDomain),
			WanPort:       getInt64Value(inst.WanPort),
		}
		result.Instances = append(result.Instances, info)
	}

	c.logger.WithField("instance_count", len(result.Instances)).Info("成功查询 CDB 实例列表")
	return result, nil
}

// FormatDBInstancesAsJSON 格式化 CDB 实例列表为 JSON
func (c *Client) FormatDBInstancesAsJSON(result *DescribeDBInstancesResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatDBInstancesAsTable 格式化 CDB 实例列表为表格
func (c *Client) FormatDBInstancesAsTable(result *DescribeDBInstancesResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CDB 实例列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 140) + "\n")
	sb.WriteString(fmt.Sprintf("%-20s %-20s %-8s %-6s %-8s %-8s %-5s %-18s %-8s %-12s\n",
		"实例ID", "名称", "版本", "状态", "CPU", "内存(MB)", "类型", "内网IP", "端口", "可用区"))
	sb.WriteString(strings.Repeat("-", 140) + "\n")

	for _, inst := range result.Instances {
		statusStr := fmt.Sprintf("%d", inst.Status)
		switch inst.Status {
		case 0:
			statusStr = "创建中"
		case 1:
			statusStr = "运行中"
		case 4:
			statusStr = "隔离中"
		case 5:
			statusStr = "已隔离"
		}
		typeStr := "主实例"
		switch inst.InstanceType {
		case 2:
			typeStr = "灾备"
		case 3:
			typeStr = "只读"
		}
		sb.WriteString(fmt.Sprintf("%-20s %-20s %-8s %-6s %-8d %-8d %-5s %-18s %-8d %-12s\n",
			inst.InstanceId,
			truncateString(inst.InstanceName, 18),
			inst.EngineVersion,
			statusStr,
			inst.Cpu,
			inst.Memory,
			typeStr,
			inst.Vip,
			inst.Vport,
			inst.Zone))
	}

	return sb.String()
}

// --- DescribeDBInstanceInfo ---

// DBInstanceDetailInfo CDB 实例详细信息
type DBInstanceDetailInfo struct {
	InstanceId       string `json:"instance_id"`
	InstanceName     string `json:"instance_name"`
	Encryption       string `json:"encryption"`
	KeyId            string `json:"key_id"`
	KeyRegion        string `json:"key_region"`
	DefaultKmsRegion string `json:"default_kms_region"`
}

// DescribeDBInstanceInfo 查询 CDB 实例详细信息
func (c *Client) DescribeDBInstanceInfo(ctx context.Context, region string, instanceId string) (*DBInstanceDetailInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"instance_id": instanceId,
	}).Debug("开始查询 CDB 实例详细信息")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cdb")
	client, err := cdb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}

	request := cdb.NewDescribeDBInstanceInfoRequest()
	request.InstanceId = &instanceId

	response, err := client.DescribeDBInstanceInfo(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CDB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CDB 实例详细信息失败: %w", err)
	}

	result := &DBInstanceDetailInfo{
		InstanceId:       getStringValue(response.Response.InstanceId),
		InstanceName:     getStringValue(response.Response.InstanceName),
		Encryption:       getStringValue(response.Response.Encryption),
		KeyId:            getStringValue(response.Response.KeyId),
		KeyRegion:        getStringValue(response.Response.KeyRegion),
		DefaultKmsRegion: getStringValue(response.Response.DefaultKmsRegion),
	}

	c.logger.WithField("instance_id", instanceId).Info("成功查询 CDB 实例详细信息")
	return result, nil
}

// FormatDBInstanceInfoAsJSON 格式化 CDB 实例详细信息为 JSON
func (c *Client) FormatDBInstanceInfoAsJSON(result *DBInstanceDetailInfo) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatDBInstanceInfoAsTable 格式化 CDB 实例详细信息为表格
func (c *Client) FormatDBInstanceInfoAsTable(result *DBInstanceDetailInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CDB 实例详细信息\n"))
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf("实例ID:      %s\n", result.InstanceId))
	sb.WriteString(fmt.Sprintf("实例名称:    %s\n", result.InstanceName))
	sb.WriteString(fmt.Sprintf("加密状态:    %s\n", result.Encryption))
	sb.WriteString(fmt.Sprintf("密钥ID:      %s\n", result.KeyId))
	sb.WriteString(fmt.Sprintf("密钥地域:    %s\n", result.KeyRegion))
	sb.WriteString(fmt.Sprintf("默认KMS地域: %s\n", result.DefaultKmsRegion))
	return sb.String()
}

// --- DescribeSlowLogs ---

// SlowLogInfoItem 慢日志信息
type SlowLogInfoItem struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Date        string `json:"date"`
	IntranetUrl string `json:"intranet_url"`
	InternetUrl string `json:"internet_url"`
	Type        string `json:"type"`
}

// DescribeSlowLogsResult 查询慢日志结果
type DescribeSlowLogsResult struct {
	TotalCount int64             `json:"total_count"`
	Items      []SlowLogInfoItem `json:"items"`
	InstanceId string            `json:"instance_id"`
	Region     string            `json:"region"`
}

// DescribeSlowLogs 查询 CDB 慢日志列表
func (c *Client) DescribeSlowLogs(ctx context.Context, region string, instanceId string) (*DescribeSlowLogsResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"instance_id": instanceId,
	}).Debug("开始查询 CDB 慢日志列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cdb")
	client, err := cdb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}

	request := cdb.NewDescribeSlowLogsRequest()
	request.InstanceId = &instanceId
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeSlowLogs(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CDB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CDB 慢日志失败: %w", err)
	}

	result := &DescribeSlowLogsResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		InstanceId: instanceId,
		Region:     region,
	}

	for _, item := range response.Response.Items {
		info := SlowLogInfoItem{
			Name:        getStringValue(item.Name),
			Size:        getInt64Value(item.Size),
			Date:        getStringValue(item.Date),
			IntranetUrl: getStringValue(item.IntranetUrl),
			InternetUrl: getStringValue(item.InternetUrl),
			Type:        getStringValue(item.Type),
		}
		result.Items = append(result.Items, info)
	}

	c.logger.WithField("log_count", len(result.Items)).Info("成功查询 CDB 慢日志列表")
	return result, nil
}

// FormatSlowLogsAsJSON 格式化慢日志为 JSON
func (c *Client) FormatSlowLogsAsJSON(result *DescribeSlowLogsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatSlowLogsAsTable 格式化慢日志为表格
func (c *Client) FormatSlowLogsAsTable(result *DescribeSlowLogsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CDB 慢日志列表 (实例: %s, 地域: %s, 总数: %d)\n",
		result.InstanceId, result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 100) + "\n")
	sb.WriteString(fmt.Sprintf("%-40s %-15s %-20s %-10s\n",
		"文件名", "大小(Byte)", "日期", "类型"))
	sb.WriteString(strings.Repeat("-", 100) + "\n")

	for _, item := range result.Items {
		sb.WriteString(fmt.Sprintf("%-40s %-15d %-20s %-10s\n",
			truncateString(item.Name, 38),
			item.Size,
			item.Date,
			item.Type))
	}

	return sb.String()
}

// --- DescribeErrorLogData ---

// ErrorLogItem 错误日志条目
type ErrorLogItem struct {
	Timestamp uint64 `json:"timestamp"`
	TimeStr   string `json:"time_str"`
	Content   string `json:"content"`
}

// DescribeErrorLogDataResult 查询错误日志结果
type DescribeErrorLogDataResult struct {
	TotalCount int64          `json:"total_count"`
	Items      []ErrorLogItem `json:"items"`
	InstanceId string         `json:"instance_id"`
	Region     string         `json:"region"`
}

// DescribeErrorLogData 查询 CDB 错误日志
func (c *Client) DescribeErrorLogData(ctx context.Context, region string, instanceId string, startTime, endTime uint64) (*DescribeErrorLogDataResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"instance_id": instanceId,
		"start_time":  startTime,
		"end_time":    endTime,
	}).Debug("开始查询 CDB 错误日志")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("cdb")
	client, err := cdb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}

	request := cdb.NewDescribeErrorLogDataRequest()
	request.InstanceId = &instanceId
	request.StartTime = &startTime
	request.EndTime = &endTime
	var limit int64 = 400
	request.Limit = &limit

	response, err := client.DescribeErrorLogData(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CDB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CDB 错误日志失败: %w", err)
	}

	result := &DescribeErrorLogDataResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		InstanceId: instanceId,
		Region:     region,
	}

	for _, item := range response.Response.Items {
		ts := getUint64Value(item.Timestamp)
		info := ErrorLogItem{
			Timestamp: ts,
			TimeStr:   time.Unix(int64(ts), 0).Format("2006-01-02 15:04:05"),
			Content:   getStringValue(item.Content),
		}
		result.Items = append(result.Items, info)
	}

	c.logger.WithField("log_count", len(result.Items)).Info("成功查询 CDB 错误日志")
	return result, nil
}

// FormatErrorLogDataAsJSON 格式化错误日志为 JSON
func (c *Client) FormatErrorLogDataAsJSON(result *DescribeErrorLogDataResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatErrorLogDataAsTable 格式化错误日志为表格
func (c *Client) FormatErrorLogDataAsTable(result *DescribeErrorLogDataResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CDB 错误日志 (实例: %s, 地域: %s, 总数: %d)\n",
		result.InstanceId, result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 120) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-90s\n", "时间", "内容"))
	sb.WriteString(strings.Repeat("-", 120) + "\n")

	for _, item := range result.Items {
		sb.WriteString(fmt.Sprintf("%-22s %-90s\n",
			item.TimeStr,
			truncateString(item.Content, 88)))
	}

	return sb.String()
}
