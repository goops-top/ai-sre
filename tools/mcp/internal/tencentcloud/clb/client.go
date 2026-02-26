package clb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"

	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client CLB 客户端
type Client struct {
	client  *clb.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建 CLB 客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("clb")

	client, err := clb.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}

	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

// GetProductName 获取产品名称
func (c *Client) GetProductName() string {
	return "CLB"
}

// GetProductVersion 获取产品版本
func (c *Client) GetProductVersion() string {
	return "2018-03-17"
}

// ValidatePermissions 验证权限
func (c *Client) ValidatePermissions(ctx context.Context) error {
	_, err := c.DescribeLoadBalancers(ctx, "ap-beijing")
	if err != nil {
		return fmt.Errorf("CLB 权限验证失败: %w", err)
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

func getBoolValue(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
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

// --- DescribeLoadBalancers ---

// LoadBalancerInfo CLB 实例信息
type LoadBalancerInfo struct {
	LoadBalancerId   string   `json:"load_balancer_id"`
	LoadBalancerName string   `json:"load_balancer_name"`
	LoadBalancerType string   `json:"load_balancer_type"`
	Forward          uint64   `json:"forward"`
	Domain           string   `json:"domain"`
	LoadBalancerVips []string `json:"load_balancer_vips"`
	Status           uint64   `json:"status"`
	CreateTime       string   `json:"create_time"`
	VpcId            string   `json:"vpc_id"`
	SubnetId         string   `json:"subnet_id"`
}

// DescribeLoadBalancersResult 查询 CLB 结果
type DescribeLoadBalancersResult struct {
	TotalCount    uint64             `json:"total_count"`
	LoadBalancers []LoadBalancerInfo `json:"load_balancers"`
	Region        string             `json:"region"`
}

// DescribeLoadBalancers 查询 CLB 实例列表
func (c *Client) DescribeLoadBalancers(ctx context.Context, region string) (*DescribeLoadBalancersResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region": region,
	}).Debug("开始查询 CLB 实例列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("clb")
	client, err := clb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}

	request := clb.NewDescribeLoadBalancersRequest()
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeLoadBalancers(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CLB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CLB 实例列表失败: %w", err)
	}

	result := &DescribeLoadBalancersResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, lb := range response.Response.LoadBalancerSet {
		info := LoadBalancerInfo{
			LoadBalancerId:   getStringValue(lb.LoadBalancerId),
			LoadBalancerName: getStringValue(lb.LoadBalancerName),
			LoadBalancerType: getStringValue(lb.LoadBalancerType),
			Forward:          getUint64Value(lb.Forward),
			Domain:           getStringValue(lb.Domain),
			LoadBalancerVips: convertStringPtrSlice(lb.LoadBalancerVips),
			Status:           getUint64Value(lb.Status),
			CreateTime:       getStringValue(lb.CreateTime),
			VpcId:            getStringValue(lb.VpcId),
			SubnetId:         getStringValue(lb.SubnetId),
		}
		result.LoadBalancers = append(result.LoadBalancers, info)
	}

	c.logger.WithField("lb_count", len(result.LoadBalancers)).Info("成功查询 CLB 实例列表")
	return result, nil
}

// FormatLoadBalancersAsJSON 格式化 CLB 列表为 JSON
func (c *Client) FormatLoadBalancersAsJSON(result *DescribeLoadBalancersResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatLoadBalancersAsTable 格式化 CLB 列表为表格
func (c *Client) FormatLoadBalancersAsTable(result *DescribeLoadBalancersResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CLB 实例列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 130) + "\n")
	sb.WriteString(fmt.Sprintf("%-15s %-25s %-10s %-8s %-6s %-18s %-20s\n",
		"实例ID", "名称", "网络类型", "类型", "状态", "VIP", "创建时间"))
	sb.WriteString(strings.Repeat("-", 130) + "\n")

	for _, lb := range result.LoadBalancers {
		vip := "-"
		if len(lb.LoadBalancerVips) > 0 {
			vip = lb.LoadBalancerVips[0]
		}
		lbType := "传统型"
		if lb.Forward == 1 {
			lbType = "负载均衡"
		}
		statusStr := "创建中"
		if lb.Status == 1 {
			statusStr = "正常"
		}
		sb.WriteString(fmt.Sprintf("%-15s %-25s %-10s %-8s %-6s %-18s %-20s\n",
			lb.LoadBalancerId,
			truncateString(lb.LoadBalancerName, 23),
			lb.LoadBalancerType,
			lbType,
			statusStr,
			vip,
			lb.CreateTime))
	}

	return sb.String()
}

// --- DescribeListeners ---

// ListenerInfo 监听器信息
type ListenerInfo struct {
	ListenerId        string `json:"listener_id"`
	ListenerName      string `json:"listener_name"`
	Protocol          string `json:"protocol"`
	Port              int64  `json:"port"`
	Scheduler         string `json:"scheduler"`
	SessionExpireTime int64  `json:"session_expire_time"`
	SniSwitch         int64  `json:"sni_switch"`
	CreateTime        string `json:"create_time"`
}

// DescribeListenersResult 查询监听器结果
type DescribeListenersResult struct {
	TotalCount     uint64         `json:"total_count"`
	Listeners      []ListenerInfo `json:"listeners"`
	LoadBalancerId string         `json:"load_balancer_id"`
	Region         string         `json:"region"`
}

// DescribeListeners 查询 CLB 监听器列表
func (c *Client) DescribeListeners(ctx context.Context, region string, loadBalancerId string) (*DescribeListenersResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region":           region,
		"load_balancer_id": loadBalancerId,
	}).Debug("开始查询 CLB 监听器列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("clb")
	client, err := clb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}

	request := clb.NewDescribeListenersRequest()
	request.LoadBalancerId = &loadBalancerId

	response, err := client.DescribeListeners(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CLB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CLB 监听器列表失败: %w", err)
	}

	result := &DescribeListenersResult{
		TotalCount:     getUint64Value(response.Response.TotalCount),
		LoadBalancerId: loadBalancerId,
		Region:         region,
	}

	for _, listener := range response.Response.Listeners {
		info := ListenerInfo{
			ListenerId:        getStringValue(listener.ListenerId),
			ListenerName:      getStringValue(listener.ListenerName),
			Protocol:          getStringValue(listener.Protocol),
			Port:              getInt64Value(listener.Port),
			Scheduler:         getStringValue(listener.Scheduler),
			SessionExpireTime: getInt64Value(listener.SessionExpireTime),
			SniSwitch:         getInt64Value(listener.SniSwitch),
			CreateTime:        getStringValue(listener.CreateTime),
		}
		result.Listeners = append(result.Listeners, info)
	}

	c.logger.WithField("listener_count", len(result.Listeners)).Info("成功查询 CLB 监听器列表")
	return result, nil
}

// FormatListenersAsJSON 格式化监听器列表为 JSON
func (c *Client) FormatListenersAsJSON(result *DescribeListenersResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatListenersAsTable 格式化监听器列表为表格
func (c *Client) FormatListenersAsTable(result *DescribeListenersResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CLB 监听器列表 (CLB: %s, 地域: %s, 总数: %d)\n",
		result.LoadBalancerId, result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 110) + "\n")
	sb.WriteString(fmt.Sprintf("%-15s %-25s %-10s %-8s %-12s %-20s\n",
		"监听器ID", "名称", "协议", "端口", "调度方式", "创建时间"))
	sb.WriteString(strings.Repeat("-", 110) + "\n")

	for _, l := range result.Listeners {
		sb.WriteString(fmt.Sprintf("%-15s %-25s %-10s %-8d %-12s %-20s\n",
			l.ListenerId,
			truncateString(l.ListenerName, 23),
			l.Protocol,
			l.Port,
			l.Scheduler,
			l.CreateTime))
	}

	return sb.String()
}

// --- DescribeTargets ---

// BackendInfo 后端服务信息
type BackendInfo struct {
	Type               string   `json:"type"`
	InstanceId         string   `json:"instance_id"`
	InstanceName       string   `json:"instance_name"`
	Port               int64    `json:"port"`
	Weight             int64    `json:"weight"`
	PrivateIpAddresses []string `json:"private_ip_addresses"`
	PublicIpAddresses  []string `json:"public_ip_addresses"`
}

// ListenerBackendInfo 监听器后端信息
type ListenerBackendInfo struct {
	ListenerId string        `json:"listener_id"`
	Protocol   string        `json:"protocol"`
	Port       int64         `json:"port"`
	Targets    []BackendInfo `json:"targets"`
}

// DescribeTargetsResult 查询后端服务结果
type DescribeTargetsResult struct {
	Listeners      []ListenerBackendInfo `json:"listeners"`
	LoadBalancerId string                `json:"load_balancer_id"`
	Region         string                `json:"region"`
}

// DescribeTargets 查询 CLB 后端服务列表
func (c *Client) DescribeTargets(ctx context.Context, region string, loadBalancerId string, listenerIds []string) (*DescribeTargetsResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region":           region,
		"load_balancer_id": loadBalancerId,
	}).Debug("开始查询 CLB 后端服务列表")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("clb")
	client, err := clb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}

	request := clb.NewDescribeTargetsRequest()
	request.LoadBalancerId = &loadBalancerId
	if len(listenerIds) > 0 {
		for _, id := range listenerIds {
			idCopy := id
			request.ListenerIds = append(request.ListenerIds, &idCopy)
		}
	}

	response, err := client.DescribeTargets(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CLB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CLB 后端服务失败: %w", err)
	}

	result := &DescribeTargetsResult{
		LoadBalancerId: loadBalancerId,
		Region:         region,
	}

	for _, listener := range response.Response.Listeners {
		lbInfo := ListenerBackendInfo{
			ListenerId: getStringValue(listener.ListenerId),
			Protocol:   getStringValue(listener.Protocol),
			Port:       getInt64Value(listener.Port),
		}
		for _, target := range listener.Targets {
			backend := BackendInfo{
				Type:               getStringValue(target.Type),
				InstanceId:         getStringValue(target.InstanceId),
				InstanceName:       getStringValue(target.InstanceName),
				Port:               getInt64Value(target.Port),
				Weight:             getInt64Value(target.Weight),
				PrivateIpAddresses: convertStringPtrSlice(target.PrivateIpAddresses),
				PublicIpAddresses:  convertStringPtrSlice(target.PublicIpAddresses),
			}
			lbInfo.Targets = append(lbInfo.Targets, backend)
		}
		result.Listeners = append(result.Listeners, lbInfo)
	}

	c.logger.WithField("listener_count", len(result.Listeners)).Info("成功查询 CLB 后端服务列表")
	return result, nil
}

// FormatTargetsAsJSON 格式化后端服务为 JSON
func (c *Client) FormatTargetsAsJSON(result *DescribeTargetsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatTargetsAsTable 格式化后端服务为表格
func (c *Client) FormatTargetsAsTable(result *DescribeTargetsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CLB 后端服务列表 (CLB: %s, 地域: %s)\n",
		result.LoadBalancerId, result.Region))
	sb.WriteString(strings.Repeat("=", 120) + "\n")

	for _, listener := range result.Listeners {
		sb.WriteString(fmt.Sprintf("\n监听器: %s (%s:%d)\n", listener.ListenerId, listener.Protocol, listener.Port))
		sb.WriteString(strings.Repeat("-", 100) + "\n")
		sb.WriteString(fmt.Sprintf("  %-20s %-20s %-8s %-8s %-18s\n",
			"实例ID", "名称", "端口", "权重", "内网IP"))

		for _, t := range listener.Targets {
			privateIP := "-"
			if len(t.PrivateIpAddresses) > 0 {
				privateIP = t.PrivateIpAddresses[0]
			}
			sb.WriteString(fmt.Sprintf("  %-20s %-20s %-8d %-8d %-18s\n",
				t.InstanceId,
				truncateString(t.InstanceName, 18),
				t.Port,
				t.Weight,
				privateIP))
		}
	}

	return sb.String()
}

// --- DescribeTargetHealth ---

// TargetHealthInfo 后端健康状态
type TargetHealthInfo struct {
	IP                 string `json:"ip"`
	Port               int64  `json:"port"`
	HealthStatus       bool   `json:"health_status"`
	TargetId           string `json:"target_id"`
	HealthStatusDetail string `json:"health_status_detail"`
}

// ListenerHealthInfo 监听器健康信息
type ListenerHealthInfo struct {
	ListenerId   string             `json:"listener_id"`
	ListenerName string             `json:"listener_name"`
	Protocol     string             `json:"protocol"`
	Port         int64              `json:"port"`
	Targets      []TargetHealthInfo `json:"targets"`
}

// LBHealthInfo 负载均衡健康信息
type LBHealthInfo struct {
	LoadBalancerId   string               `json:"load_balancer_id"`
	LoadBalancerName string               `json:"load_balancer_name"`
	Listeners        []ListenerHealthInfo `json:"listeners"`
}

// DescribeTargetHealthResult 查询后端健康状态结果
type DescribeTargetHealthResult struct {
	LoadBalancers []LBHealthInfo `json:"load_balancers"`
	Region        string         `json:"region"`
}

// DescribeTargetHealth 查询后端健康状态
func (c *Client) DescribeTargetHealth(ctx context.Context, region string, loadBalancerIds []string) (*DescribeTargetHealthResult, error) {
	c.logger.WithFields(logrus.Fields{
		"region":            region,
		"load_balancer_ids": loadBalancerIds,
	}).Debug("开始查询 CLB 后端健康状态")

	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("clb")
	client, err := clb.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}

	request := clb.NewDescribeTargetHealthRequest()
	for _, id := range loadBalancerIds {
		idCopy := id
		request.LoadBalancerIds = append(request.LoadBalancerIds, &idCopy)
	}

	response, err := client.DescribeTargetHealth(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("CLB API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 CLB 后端健康状态失败: %w", err)
	}

	result := &DescribeTargetHealthResult{
		Region: region,
	}

	for _, lb := range response.Response.LoadBalancers {
		lbHealth := LBHealthInfo{
			LoadBalancerId:   getStringValue(lb.LoadBalancerId),
			LoadBalancerName: getStringValue(lb.LoadBalancerName),
		}
		for _, listener := range lb.Listeners {
			listenerHealth := ListenerHealthInfo{
				ListenerId:   getStringValue(listener.ListenerId),
				ListenerName: getStringValue(listener.ListenerName),
				Protocol:     getStringValue(listener.Protocol),
				Port:         getInt64Value(listener.Port),
			}
			// 收集所有 Rules 下的 Targets
			for _, rule := range listener.Rules {
				for _, target := range rule.Targets {
					th := TargetHealthInfo{
						IP:                 getStringValue(target.IP),
						Port:               getInt64Value(target.Port),
						HealthStatus:       getBoolValue(target.HealthStatus),
						TargetId:           getStringValue(target.TargetId),
						HealthStatusDetail: getStringValue(target.HealthStatusDetail),
					}
					listenerHealth.Targets = append(listenerHealth.Targets, th)
				}
			}
			lbHealth.Listeners = append(lbHealth.Listeners, listenerHealth)
		}
		result.LoadBalancers = append(result.LoadBalancers, lbHealth)
	}

	c.logger.WithField("lb_count", len(result.LoadBalancers)).Info("成功查询 CLB 后端健康状态")
	return result, nil
}

// FormatTargetHealthAsJSON 格式化健康状态为 JSON
func (c *Client) FormatTargetHealthAsJSON(result *DescribeTargetHealthResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

// FormatTargetHealthAsTable 格式化健康状态为表格
func (c *Client) FormatTargetHealthAsTable(result *DescribeTargetHealthResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CLB 后端健康状态 (地域: %s)\n", result.Region))
	sb.WriteString(strings.Repeat("=", 120) + "\n")

	for _, lb := range result.LoadBalancers {
		sb.WriteString(fmt.Sprintf("\nCLB: %s (%s)\n", lb.LoadBalancerId, lb.LoadBalancerName))
		for _, listener := range lb.Listeners {
			sb.WriteString(fmt.Sprintf("  监听器: %s (%s:%d)\n", listener.ListenerId, listener.Protocol, listener.Port))
			sb.WriteString(fmt.Sprintf("  %-18s %-8s %-10s %-20s %-25s\n",
				"IP", "端口", "健康", "实例ID", "详情"))
			sb.WriteString("  " + strings.Repeat("-", 90) + "\n")

			for _, t := range listener.Targets {
				healthStr := "健康"
				if !t.HealthStatus {
					healthStr = "异常"
				}
				sb.WriteString(fmt.Sprintf("  %-18s %-8d %-10s %-20s %-25s\n",
					t.IP,
					t.Port,
					healthStr,
					t.TargetId,
					t.HealthStatusDetail))
			}
		}
	}

	return sb.String()
}
