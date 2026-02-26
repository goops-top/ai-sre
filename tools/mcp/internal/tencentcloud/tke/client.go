package tke

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	"github.com/sirupsen/logrus"
	
	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client TKE 客户端
type Client struct {
	client  *tke.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建 TKE 客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("tke")
	
	// 使用默认地域创建客户端
	client, err := tke.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

// GetProductName 获取产品名称
func (c *Client) GetProductName() string {
	return "TKE"
}

// GetProductVersion 获取产品版本
func (c *Client) GetProductVersion() string {
	return "2018-05-25"
}

// ValidatePermissions 验证权限
func (c *Client) ValidatePermissions(ctx context.Context) error {
	// 通过调用 DescribeRegions 来验证权限
	_, err := c.DescribeRegions(ctx)
	if err != nil {
		return fmt.Errorf("TKE 权限验证失败: %w", err)
	}
	return nil
}

// RegionInfo TKE 地域信息
type RegionInfo struct {
	RegionID   int64  `json:"region_id"`
	RegionName string `json:"region_name"`
	Status     string `json:"status"`
}

// DescribeRegions 查询 TKE 支持的地域信息
func (c *Client) DescribeRegions(ctx context.Context) ([]RegionInfo, error) {
	c.logger.Debug("开始查询 TKE 支持的地域信息")
	
	// 创建请求
	request := tke.NewDescribeRegionsRequest()
	
	// 发送请求
	response, err := c.client.DescribeRegions(request)
	if err != nil {
		// 处理腾讯云 SDK 错误
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":    sdkError.Code,
				"message": sdkError.Message,
				"request_id": sdkError.RequestId,
			}).Error("TKE API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 TKE 地域信息失败: %w", err)
	}
	
	// 转换响应数据
	var regions []RegionInfo
	for _, region := range response.Response.RegionInstanceSet {
		regionInfo := RegionInfo{
			RegionID:   *region.RegionId,
			RegionName: *region.RegionName,
			Status:     *region.Status,
		}
		regions = append(regions, regionInfo)
	}
	
	c.logger.WithField("region_count", len(regions)).Info("成功查询 TKE 地域信息")
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
func (c *Client) FormatRegionsAsTable(regions []RegionInfo) string {
	if len(regions) == 0 {
		return "未找到任何地域信息"
	}
	
	result := "TKE 支持的地域信息:\n"
	result += "┌─────────────────┬──────────────────────────┬──────────────┐\n"
	result += "│ 地域ID          │ 地域名称                 │ 状态         │\n"
	result += "├─────────────────┼──────────────────────────┼──────────────┤\n"
	
	for _, region := range regions {
		result += fmt.Sprintf("│ %-15d │ %-24s │ %-12s │\n", 
			region.RegionID, region.RegionName, region.Status)
	}
	
	result += "└─────────────────┴──────────────────────────┴──────────────┘\n"
	result += fmt.Sprintf("总计: %d 个地域", len(regions))
	
	return result
}

// GetRegionByID 根据地域ID获取地域信息
func (c *Client) GetRegionByID(ctx context.Context, regionID string) (*RegionInfo, error) {
	regions, err := c.DescribeRegions(ctx)
	if err != nil {
		return nil, err
	}
	
	// 尝试按地域ID（数字）匹配
	for _, region := range regions {
		if fmt.Sprintf("%d", region.RegionID) == regionID {
			return &region, nil
		}
	}
	
	// 尝试按地域名称匹配
	for _, region := range regions {
		if region.RegionName == regionID {
			return &region, nil
		}
	}
	
	return nil, fmt.Errorf("未找到地域 %s", regionID)
}

// ClusterInfo TKE 普通集群信息
type ClusterInfo struct {
	ClusterID          string `json:"cluster_id"`
	ClusterName        string `json:"cluster_name"`
	ClusterDescription string `json:"cluster_description"`
	ClusterVersion     string `json:"cluster_version"`
	ClusterOs          string `json:"cluster_os"`
	ClusterType        string `json:"cluster_type"`
	ClusterKind        string `json:"cluster_kind"` // tke 或 serverless，用于标识集群种类
	Region             string `json:"region"`
	VpcID              string `json:"vpc_id"`
	ProjectID          int64  `json:"project_id"`
	Status             string `json:"status"`
	CreatedTime        string `json:"created_time"`
	NodeNum            int64  `json:"node_num"`
	EnableExternalNode bool   `json:"enable_external_node"`
}

// EKSClusterInfo EKS Serverless 集群信息
type EKSClusterInfo struct {
	ClusterID    string   `json:"cluster_id"`
	ClusterName  string   `json:"cluster_name"`
	ClusterDesc  string   `json:"cluster_desc"`
	ClusterKind  string   `json:"cluster_kind"` // 固定为 serverless
	K8SVersion   string   `json:"k8s_version"`
	Region       string   `json:"region"`
	VpcID        string   `json:"vpc_id"`
	SubnetIDs    []string `json:"subnet_ids"`
	Status       string   `json:"status"`
	CreatedTime  string   `json:"created_time"`
}

// DescribeClusters 查询 TKE 集群列表
func (c *Client) DescribeClusters(ctx context.Context, region string) ([]ClusterInfo, error) {
	c.logger.WithField("region", region).Debug("开始查询 TKE 集群列表")
	
	// 创建指定地域的客户端
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	// 创建请求
	request := tke.NewDescribeClustersRequest()
	
	// 发送请求
	response, err := client.DescribeClusters(request)
	if err != nil {
		// 处理腾讯云 SDK 错误
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
			}).Error("TKE API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 TKE 集群列表失败: %w", err)
	}
	
	// 转换响应数据
	var clusters []ClusterInfo
	for _, cluster := range response.Response.Clusters {
		vpcID := ""
		if cluster.ClusterNetworkSettings != nil {
			vpcID = getStringValue(cluster.ClusterNetworkSettings.VpcId)
		}
		clusterInfo := ClusterInfo{
			ClusterID:          getStringValue(cluster.ClusterId),
			ClusterName:        getStringValue(cluster.ClusterName),
			ClusterDescription: getStringValue(cluster.ClusterDescription),
			ClusterVersion:     getStringValue(cluster.ClusterVersion),
			ClusterOs:          getStringValue(cluster.ClusterOs),
			ClusterType:        getStringValue(cluster.ClusterType),
			ClusterKind:        "tke",
			Region:             region,
			VpcID:              vpcID,
			ProjectID:          getUint64AsInt64Value(cluster.ProjectId),
			Status:             getStringValue(cluster.ClusterStatus),
			CreatedTime:        getStringValue(cluster.CreatedTime),
			NodeNum:            getUint64AsInt64Value(cluster.ClusterNodeNum),
			EnableExternalNode: getBoolValue(cluster.EnableExternalNode),
		}
		clusters = append(clusters, clusterInfo)
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"cluster_count": len(clusters),
	}).Info("成功查询 TKE 集群列表")
	
	return clusters, nil
}

// FormatClustersAsJSON 将普通集群信息格式化为 JSON
func (c *Client) FormatClustersAsJSON(clusters []ClusterInfo) (string, error) {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化集群信息失败: %w", err)
	}
	return string(data), nil
}

// FormatClustersAsTable 将普通集群信息格式化为表格
func (c *Client) FormatClustersAsTable(clusters []ClusterInfo, region string) string {
	if len(clusters) == 0 {
		return fmt.Sprintf("地域 %s 未找到任何 TKE 普通集群", region)
	}
	
	result := fmt.Sprintf("地域 %s 的 TKE 普通集群列表:\n", region)
	result += "┌─────────────────────────┬──────────────────────────┬─────────────┬──────────────┬──────────────┬──────────────┐\n"
	result += "│ 集群ID                  │ 集群名称                 │ 版本        │ 状态         │ 节点数       │ 创建时间     │\n"
	result += "├─────────────────────────┼──────────────────────────┼─────────────┼──────────────┼──────────────┼──────────────┤\n"
	
	for _, cluster := range clusters {
		// 截断过长的字段以适应表格显示
		clusterID := truncateString(cluster.ClusterID, 23)
		clusterName := truncateString(cluster.ClusterName, 24)
		version := truncateString(cluster.ClusterVersion, 11)
		status := truncateString(cluster.Status, 12)
		nodeNum := fmt.Sprintf("%d", cluster.NodeNum)
		createdTime := truncateString(cluster.CreatedTime, 12)
		
		result += fmt.Sprintf("│ %-23s │ %-24s │ %-11s │ %-12s │ %-12s │ %-12s │\n",
			clusterID, clusterName, version, status, nodeNum, createdTime)
	}
	
	result += "└─────────────────────────┴──────────────────────────┴─────────────┴──────────────┴──────────────┴──────────────┘\n"
	result += fmt.Sprintf("总计: %d 个普通集群", len(clusters))
	
	return result
}

// DescribeEKSClusters 查询 EKS Serverless 集群列表
func (c *Client) DescribeEKSClusters(ctx context.Context, region string) ([]EKSClusterInfo, error) {
	c.logger.WithField("region", region).Debug("开始查询 EKS Serverless 集群列表")
	
	// 创建指定地域的客户端
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	// 创建请求
	request := tke.NewDescribeEKSClustersRequest()
	
	// 发送请求
	response, err := client.DescribeEKSClusters(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
			}).Error("EKS API 调用失败")
			return nil, fmt.Errorf("EKS API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 EKS Serverless 集群列表失败: %w", err)
	}
	
	// 转换响应数据
	var clusters []EKSClusterInfo
	if response.Response != nil && response.Response.Clusters != nil {
		for _, cluster := range response.Response.Clusters {
			subnetIDs := make([]string, 0)
			if cluster.SubnetIds != nil {
				for _, sid := range cluster.SubnetIds {
					if sid != nil {
						subnetIDs = append(subnetIDs, *sid)
					}
				}
			}
			
			clusterInfo := EKSClusterInfo{
				ClusterID:   getStringValue(cluster.ClusterId),
				ClusterName: getStringValue(cluster.ClusterName),
				ClusterDesc: getStringValue(cluster.ClusterDesc),
				ClusterKind: "serverless",
				K8SVersion:  getStringValue(cluster.K8SVersion),
				Region:      region,
				VpcID:       getStringValue(cluster.VpcId),
				SubnetIDs:   subnetIDs,
				Status:      getStringValue(cluster.Status),
				CreatedTime: getStringValue(cluster.CreatedTime),
			}
			clusters = append(clusters, clusterInfo)
		}
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"cluster_count": len(clusters),
	}).Info("成功查询 EKS Serverless 集群列表")
	
	return clusters, nil
}

// FormatEKSClustersAsJSON 将 EKS 集群信息格式化为 JSON
func (c *Client) FormatEKSClustersAsJSON(clusters []EKSClusterInfo) (string, error) {
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 EKS 集群信息失败: %w", err)
	}
	return string(data), nil
}

// FormatEKSClustersAsTable 将 EKS 集群信息格式化为表格
func (c *Client) FormatEKSClustersAsTable(clusters []EKSClusterInfo, region string) string {
	if len(clusters) == 0 {
		return fmt.Sprintf("地域 %s 未找到任何 EKS Serverless 集群", region)
	}
	
	result := fmt.Sprintf("地域 %s 的 EKS Serverless 集群列表:\n", region)
	result += "┌─────────────────────────┬──────────────────────────┬─────────────┬──────────────┬──────────────┐\n"
	result += "│ 集群ID                  │ 集群名称                 │ K8S版本     │ 状态         │ 创建时间     │\n"
	result += "├─────────────────────────┼──────────────────────────┼─────────────┼──────────────┼──────────────┤\n"
	
	for _, cluster := range clusters {
		clusterID := truncateString(cluster.ClusterID, 23)
		clusterName := truncateString(cluster.ClusterName, 24)
		version := truncateString(cluster.K8SVersion, 11)
		status := truncateString(cluster.Status, 12)
		createdTime := truncateString(cluster.CreatedTime, 12)
		
		result += fmt.Sprintf("│ %-23s │ %-24s │ %-11s │ %-12s │ %-12s │\n",
			clusterID, clusterName, version, status, createdTime)
	}
	
	result += "└─────────────────────────┴──────────────────────────┴─────────────┴──────────────┴──────────────┘\n"
	result += fmt.Sprintf("总计: %d 个 Serverless 集群", len(clusters))
	
	return result
}

// ClusterExtraArgsInfo 集群自定义参数信息
type ClusterExtraArgsInfo struct {
	ClusterID             string   `json:"cluster_id"`
	Region                string   `json:"region"`
	HasExtraArgs          bool     `json:"has_extra_args"`
	Etcd                  []string `json:"etcd,omitempty"`
	KubeAPIServer         []string `json:"kube_apiserver,omitempty"`
	KubeControllerManager []string `json:"kube_controller_manager,omitempty"`
	KubeScheduler         []string `json:"kube_scheduler,omitempty"`
}

// DescribeClusterExtraArgs 查询集群自定义参数
func (c *Client) DescribeClusterExtraArgs(ctx context.Context, region string, clusterID string) (*ClusterExtraArgsInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
	}).Debug("开始查询集群自定义参数")
	
	// 创建指定地域的客户端
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	// 创建请求
	request := tke.NewDescribeClusterExtraArgsRequest()
	request.ClusterId = &clusterID
	
	// 发送请求
	response, err := client.DescribeClusterExtraArgs(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
				"cluster_id": clusterID,
			}).Error("查询集群自定义参数 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询集群自定义参数失败: %w", err)
	}
	
	// 转换响应数据
	info := &ClusterExtraArgsInfo{
		ClusterID: clusterID,
		Region:    region,
	}
	
	if response.Response != nil && response.Response.ClusterExtraArgs != nil {
		extraArgs := response.Response.ClusterExtraArgs
		info.Etcd = convertStringPtrSlice(extraArgs.Etcd)
		info.KubeAPIServer = convertStringPtrSlice(extraArgs.KubeAPIServer)
		info.KubeControllerManager = convertStringPtrSlice(extraArgs.KubeControllerManager)
		info.KubeScheduler = convertStringPtrSlice(extraArgs.KubeScheduler)
		info.HasExtraArgs = len(info.Etcd) > 0 || len(info.KubeAPIServer) > 0 || len(info.KubeControllerManager) > 0 || len(info.KubeScheduler) > 0
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":         region,
		"cluster_id":     clusterID,
		"has_extra_args": info.HasExtraArgs,
	}).Info("成功查询集群自定义参数")
	
	return info, nil
}

// FormatClusterExtraArgsAsJSON 将集群自定义参数格式化为 JSON
func (c *Client) FormatClusterExtraArgsAsJSON(info *ClusterExtraArgsInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化集群自定义参数失败: %w", err)
	}
	return string(data), nil
}

// FormatClusterExtraArgsAsTable 将集群自定义参数格式化为表格
func (c *Client) FormatClusterExtraArgsAsTable(info *ClusterExtraArgsInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) 自定义参数:\n", info.ClusterID, info.Region)
	
	if !info.HasExtraArgs {
		result += "\n该集群未配置任何自定义参数。\n"
		return result
	}
	
	result += "\n"
	
	if len(info.Etcd) > 0 {
		result += "【Etcd 自定义参数】\n"
		for _, arg := range info.Etcd {
			result += fmt.Sprintf("  - %s\n", arg)
		}
		result += "\n"
	}
	
	if len(info.KubeAPIServer) > 0 {
		result += "【KubeAPIServer 自定义参数】\n"
		for _, arg := range info.KubeAPIServer {
			result += fmt.Sprintf("  - %s\n", arg)
		}
		result += "\n"
	}
	
	if len(info.KubeControllerManager) > 0 {
		result += "【KubeControllerManager 自定义参数】\n"
		for _, arg := range info.KubeControllerManager {
			result += fmt.Sprintf("  - %s\n", arg)
		}
		result += "\n"
	}
	
	if len(info.KubeScheduler) > 0 {
		result += "【KubeScheduler 自定义参数】\n"
		for _, arg := range info.KubeScheduler {
			result += fmt.Sprintf("  - %s\n", arg)
		}
	}
	
	return result
}

// ClusterLevelPriceInfo 集群等级价格信息
type ClusterLevelPriceInfo struct {
	Region       string  `json:"region"`
	ClusterLevel string  `json:"cluster_level"`
	Cost         int64   `json:"cost"`
	TotalCost    int64   `json:"total_cost"`
	Policy       float64 `json:"policy"`
	CostYuan     string  `json:"cost_yuan"`
	TotalCostYuan string `json:"total_cost_yuan"`
}

// GetClusterLevelPrice 获取集群等级价格
func (c *Client) GetClusterLevelPrice(ctx context.Context, region string, clusterLevel string) (*ClusterLevelPriceInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"cluster_level": clusterLevel,
	}).Debug("开始查询集群等级价格")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewGetClusterLevelPriceRequest()
	request.ClusterLevel = &clusterLevel
	
	response, err := client.GetClusterLevelPrice(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":          sdkError.Code,
				"message":       sdkError.Message,
				"request_id":    sdkError.RequestId,
				"region":        region,
				"cluster_level": clusterLevel,
			}).Error("获取集群等级价格 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("获取集群等级价格失败: %w", err)
	}
	
	info := &ClusterLevelPriceInfo{
		Region:       region,
		ClusterLevel: clusterLevel,
	}
	
	if response.Response != nil {
		if response.Response.Cost != nil {
			info.Cost = int64(*response.Response.Cost)
			info.CostYuan = fmt.Sprintf("%.2f", float64(*response.Response.Cost)/100.0)
		}
		if response.Response.TotalCost != nil {
			info.TotalCost = int64(*response.Response.TotalCost)
			info.TotalCostYuan = fmt.Sprintf("%.2f", float64(*response.Response.TotalCost)/100.0)
		}
		if response.Response.Policy != nil {
			info.Policy = *response.Response.Policy
		}
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"cluster_level": clusterLevel,
		"cost":          info.Cost,
		"total_cost":    info.TotalCost,
	}).Info("成功获取集群等级价格")
	
	return info, nil
}

// FormatClusterLevelPriceAsJSON 将集群等级价格格式化为 JSON
func (c *Client) FormatClusterLevelPriceAsJSON(info *ClusterLevelPriceInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化集群等级价格失败: %w", err)
	}
	return string(data), nil
}

// FormatClusterLevelPriceAsTable 将集群等级价格格式化为表格
func (c *Client) FormatClusterLevelPriceAsTable(info *ClusterLevelPriceInfo) string {
	result := fmt.Sprintf("集群等级 %s (地域: %s) 价格信息:\n\n", info.ClusterLevel, info.Region)
	result += fmt.Sprintf("折后价格: %s 元/月 (%d 分)\n", info.CostYuan, info.Cost)
	result += fmt.Sprintf("原价:     %s 元/月 (%d 分)\n", info.TotalCostYuan, info.TotalCost)
	result += fmt.Sprintf("折扣:     %.0f%% (100=不打折)\n", info.Policy)
	return result
}

// AddonInfo 组件信息
type AddonInfo struct {
	AddonName    string `json:"addon_name"`
	AddonVersion string `json:"addon_version"`
	Phase        string `json:"phase"`
	Reason       string `json:"reason,omitempty"`
}

// ClusterAddonListInfo 集群组件列表信息
type ClusterAddonListInfo struct {
	ClusterID string      `json:"cluster_id"`
	Region    string      `json:"region"`
	AddonCount int        `json:"addon_count"`
	Addons    []AddonInfo `json:"addons"`
}

// DescribeAddon 查询集群已安装的 addon 列表
func (c *Client) DescribeAddon(ctx context.Context, region string, clusterID string, addonName string) (*ClusterAddonListInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
		"addon_name": addonName,
	}).Debug("开始查询集群 addon 列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeAddonRequest()
	request.ClusterId = &clusterID
	if addonName != "" {
		request.AddonName = &addonName
	}
	
	response, err := client.DescribeAddon(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
				"cluster_id": clusterID,
			}).Error("查询集群 addon 列表 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询集群 addon 列表失败: %w", err)
	}
	
	info := &ClusterAddonListInfo{
		ClusterID: clusterID,
		Region:    region,
	}
	
	if response.Response != nil && response.Response.Addons != nil {
		for _, addon := range response.Response.Addons {
			addonInfo := AddonInfo{
				AddonName:    getStringValue(addon.AddonName),
				AddonVersion: getStringValue(addon.AddonVersion),
				Phase:        getStringValue(addon.Phase),
				Reason:       getStringValue(addon.Reason),
			}
			info.Addons = append(info.Addons, addonInfo)
		}
	}
	info.AddonCount = len(info.Addons)
	
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"cluster_id":  clusterID,
		"addon_count": info.AddonCount,
	}).Info("成功查询集群 addon 列表")
	
	return info, nil
}

// FormatAddonListAsJSON 将 addon 列表格式化为 JSON
func (c *Client) FormatAddonListAsJSON(info *ClusterAddonListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 addon 列表失败: %w", err)
	}
	return string(data), nil
}

// FormatAddonListAsTable 将 addon 列表格式化为表格
func (c *Client) FormatAddonListAsTable(info *ClusterAddonListInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) 已安装的 Addon 列表:\n", info.ClusterID, info.Region)
	
	if info.AddonCount == 0 {
		result += "\n该集群未安装任何 Addon。\n"
		return result
	}
	
	result += "┌──────────────────────────────┬──────────────────┬──────────────────┬──────────────────────────┐\n"
	result += "│ Addon 名称                   │ 版本             │ 状态             │ 原因                     │\n"
	result += "├──────────────────────────────┼──────────────────┼──────────────────┼──────────────────────────┤\n"
	
	for _, addon := range info.Addons {
		name := truncateString(addon.AddonName, 28)
		version := truncateString(addon.AddonVersion, 16)
		phase := truncateString(addon.Phase, 16)
		reason := truncateString(addon.Reason, 24)
		
		result += fmt.Sprintf("│ %-28s │ %-16s │ %-16s │ %-24s │\n",
			name, version, phase, reason)
	}
	
	result += "└──────────────────────────────┴──────────────────┴──────────────────┴──────────────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个 Addon", info.AddonCount)
	
	return result
}

// AppChartInfo App Chart 信息
type AppChartInfo struct {
	Name          string `json:"name"`
	Label         string `json:"label"`
	LatestVersion string `json:"latest_version"`
}

// AppChartListInfo App Chart 列表信息
type AppChartListInfo struct {
	Region     string         `json:"region"`
	Kind       string         `json:"kind,omitempty"`
	Arch       string         `json:"arch,omitempty"`
	ClusterType string        `json:"cluster_type,omitempty"`
	ChartCount int            `json:"chart_count"`
	Charts     []AppChartInfo `json:"charts"`
}

// GetTkeAppChartList 获取可安装的 addon 列表
func (c *Client) GetTkeAppChartList(ctx context.Context, region string, kind string, arch string, clusterType string) (*AppChartListInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":       region,
		"kind":         kind,
		"arch":         arch,
		"cluster_type": clusterType,
	}).Debug("开始查询可安装的 addon 列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewGetTkeAppChartListRequest()
	if kind != "" {
		request.Kind = &kind
	}
	if arch != "" {
		request.Arch = &arch
	}
	if clusterType != "" {
		request.ClusterType = &clusterType
	}
	
	response, err := client.GetTkeAppChartList(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
			}).Error("查询可安装 addon 列表 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询可安装 addon 列表失败: %w", err)
	}
	
	info := &AppChartListInfo{
		Region:      region,
		Kind:        kind,
		Arch:        arch,
		ClusterType: clusterType,
	}
	
	if response.Response != nil && response.Response.AppCharts != nil {
		for _, chart := range response.Response.AppCharts {
			chartInfo := AppChartInfo{
				Name:          getStringValue(chart.Name),
				Label:         getStringValue(chart.Label),
				LatestVersion: getStringValue(chart.LatestVersion),
			}
			info.Charts = append(info.Charts, chartInfo)
		}
	}
	info.ChartCount = len(info.Charts)
	
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"chart_count": info.ChartCount,
	}).Info("成功查询可安装 addon 列表")
	
	return info, nil
}

// FormatAppChartListAsJSON 将 App Chart 列表格式化为 JSON
func (c *Client) FormatAppChartListAsJSON(info *AppChartListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 App Chart 列表失败: %w", err)
	}
	return string(data), nil
}

// FormatAppChartListAsTable 将 App Chart 列表格式化为表格
func (c *Client) FormatAppChartListAsTable(info *AppChartListInfo) string {
	filterInfo := fmt.Sprintf("地域: %s", info.Region)
	if info.Kind != "" {
		filterInfo += fmt.Sprintf(", 类型: %s", info.Kind)
	}
	if info.Arch != "" {
		filterInfo += fmt.Sprintf(", 架构: %s", info.Arch)
	}
	if info.ClusterType != "" {
		filterInfo += fmt.Sprintf(", 集群类型: %s", info.ClusterType)
	}
	
	result := fmt.Sprintf("可安装的 Addon 列表 (%s):\n", filterInfo)
	
	if info.ChartCount == 0 {
		result += "\n未找到任何可安装的 Addon。\n"
		return result
	}
	
	result += "┌──────────────────────────────┬──────────────────────────────┬──────────────────┐\n"
	result += "│ 名称                         │ 标签                         │ 最新版本         │\n"
	result += "├──────────────────────────────┼──────────────────────────────┼──────────────────┤\n"
	
	for _, chart := range info.Charts {
		name := truncateString(chart.Name, 28)
		label := truncateString(chart.Label, 28)
		version := truncateString(chart.LatestVersion, 16)
		
		result += fmt.Sprintf("│ %-28s │ %-28s │ %-16s │\n",
			name, label, version)
	}
	
	result += "└──────────────────────────────┴──────────────────────────────┴──────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个可安装 Addon", info.ChartCount)
	
	return result
}

// ImageInfo 镜像信息
type ImageInfo struct {
	Alias           string `json:"alias"`
	OsName          string `json:"os_name"`
	ImageId         string `json:"image_id"`
	OsCustomizeType string `json:"os_customize_type"`
}

// ImageListInfo 镜像列表信息
type ImageListInfo struct {
	Region     string      `json:"region"`
	ImageCount int         `json:"image_count"`
	Images     []ImageInfo `json:"images"`
}

// DescribeImages 获取指定地域支持的 OS 镜像列表
func (c *Client) DescribeImages(ctx context.Context, region string) (*ImageListInfo, error) {
	c.logger.WithField("region", region).Debug("开始查询 OS 镜像列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeImagesRequest()
	
	response, err := client.DescribeImages(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
			}).Error("查询 OS 镜像列表 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 OS 镜像列表失败: %w", err)
	}
	
	info := &ImageListInfo{Region: region}
	
	if response.Response != nil && response.Response.ImageInstanceSet != nil {
		for _, img := range response.Response.ImageInstanceSet {
			info.Images = append(info.Images, ImageInfo{
				Alias:           getStringValue(img.Alias),
				OsName:          getStringValue(img.OsName),
				ImageId:         getStringValue(img.ImageId),
				OsCustomizeType: getStringValue(img.OsCustomizeType),
			})
		}
	}
	info.ImageCount = len(info.Images)
	
	c.logger.WithFields(logrus.Fields{
		"region":      region,
		"image_count": info.ImageCount,
	}).Info("成功查询 OS 镜像列表")
	
	return info, nil
}

// FormatImagesAsJSON 将镜像列表格式化为 JSON
func (c *Client) FormatImagesAsJSON(info *ImageListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化镜像列表失败: %w", err)
	}
	return string(data), nil
}

// FormatImagesAsTable 将镜像列表格式化为表格
func (c *Client) FormatImagesAsTable(info *ImageListInfo) string {
	result := fmt.Sprintf("地域 %s 支持的 OS 镜像列表:\n", info.Region)
	
	if info.ImageCount == 0 {
		result += "\n未找到任何 OS 镜像。\n"
		return result
	}
	
	result += "┌──────────────────────────────┬──────────────────────────────┬──────────────────┬──────────────────┐\n"
	result += "│ 别名                         │ 操作系统名称                 │ 镜像ID           │ 定制类型         │\n"
	result += "├──────────────────────────────┼──────────────────────────────┼──────────────────┼──────────────────┤\n"
	
	for _, img := range info.Images {
		alias := truncateString(img.Alias, 28)
		osName := truncateString(img.OsName, 28)
		imageId := truncateString(img.ImageId, 16)
		customizeType := truncateString(img.OsCustomizeType, 16)
		
		result += fmt.Sprintf("│ %-28s │ %-28s │ %-16s │ %-16s │\n",
			alias, osName, imageId, customizeType)
	}
	
	result += "└──────────────────────────────┴──────────────────────────────┴──────────────────┴──────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个 OS 镜像", info.ImageCount)
	
	return result
}

// VersionInfo 版本信息
type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Remark  string `json:"remark,omitempty"`
}

// VersionListInfo 版本列表信息
type VersionListInfo struct {
	Region       string        `json:"region"`
	VersionCount int           `json:"version_count"`
	Versions     []VersionInfo `json:"versions"`
}

// DescribeVersions 获取指定地域支持的集群版本列表
func (c *Client) DescribeVersions(ctx context.Context, region string) (*VersionListInfo, error) {
	c.logger.WithField("region", region).Debug("开始查询集群版本列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeVersionsRequest()
	
	response, err := client.DescribeVersions(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
			}).Error("查询集群版本列表 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询集群版本列表失败: %w", err)
	}
	
	info := &VersionListInfo{Region: region}
	
	if response.Response != nil && response.Response.VersionInstanceSet != nil {
		for _, v := range response.Response.VersionInstanceSet {
			info.Versions = append(info.Versions, VersionInfo{
				Name:    getStringValue(v.Name),
				Version: getStringValue(v.Version),
				Remark:  getStringValue(v.Remark),
			})
		}
	}
	info.VersionCount = len(info.Versions)
	
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"version_count": info.VersionCount,
	}).Info("成功查询集群版本列表")
	
	return info, nil
}

// FormatVersionsAsJSON 将版本列表格式化为 JSON
func (c *Client) FormatVersionsAsJSON(info *VersionListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化版本列表失败: %w", err)
	}
	return string(data), nil
}

// FormatVersionsAsTable 将版本列表格式化为表格
func (c *Client) FormatVersionsAsTable(info *VersionListInfo) string {
	result := fmt.Sprintf("地域 %s 支持的集群版本列表:\n", info.Region)
	
	if info.VersionCount == 0 {
		result += "\n未找到任何集群版本。\n"
		return result
	}
	
	result += "┌──────────────────────────────┬──────────────────┬──────────────────────────────────────────┐\n"
	result += "│ 名称                         │ 版本             │ 备注                                     │\n"
	result += "├──────────────────────────────┼──────────────────┼──────────────────────────────────────────┤\n"
	
	for _, v := range info.Versions {
		name := truncateString(v.Name, 28)
		version := truncateString(v.Version, 16)
		remark := truncateString(v.Remark, 38)
		
		result += fmt.Sprintf("│ %-28s │ %-16s │ %-38s │\n",
			name, version, remark)
	}
	
	result += "└──────────────────────────────┴──────────────────┴──────────────────────────────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个版本", info.VersionCount)
	
	return result
}

// LogSwitchDetailInfo 日志开关详情
type LogSwitchDetailInfo struct {
	Enable   bool   `json:"enable"`
	ErrorMsg string `json:"error_msg,omitempty"`
	LogsetId string `json:"logset_id,omitempty"`
	Status   string `json:"status,omitempty"`
	TopicId  string `json:"topic_id,omitempty"`
}

// ClusterLogSwitchInfo 集群日志开关信息
type ClusterLogSwitchInfo struct {
	ClusterID string               `json:"cluster_id"`
	Region    string               `json:"region"`
	Audit     *LogSwitchDetailInfo `json:"audit,omitempty"`
	Event     *LogSwitchDetailInfo `json:"event,omitempty"`
	Log       *LogSwitchDetailInfo `json:"log,omitempty"`
	MasterLog *LogSwitchDetailInfo `json:"master_log,omitempty"`
}

// DescribeLogSwitches 查询集群日志开关信息
func (c *Client) DescribeLogSwitches(ctx context.Context, region string, clusterID string) (*ClusterLogSwitchInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
	}).Debug("开始查询集群日志开关信息")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeLogSwitchesRequest()
	request.ClusterIds = []*string{&clusterID}
	
	response, err := client.DescribeLogSwitches(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
				"cluster_id": clusterID,
			}).Error("查询集群日志开关 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询集群日志开关失败: %w", err)
	}
	
	info := &ClusterLogSwitchInfo{
		ClusterID: clusterID,
		Region:    region,
	}
	
	if response.Response != nil && response.Response.SwitchSet != nil {
		for _, sw := range response.Response.SwitchSet {
			if getStringValue(sw.ClusterId) == clusterID {
				if sw.Audit != nil {
					info.Audit = &LogSwitchDetailInfo{
						Enable:   getBoolValue(sw.Audit.Enable),
						ErrorMsg: getStringValue(sw.Audit.ErrorMsg),
						LogsetId: getStringValue(sw.Audit.LogsetId),
						Status:   getStringValue(sw.Audit.Status),
						TopicId:  getStringValue(sw.Audit.TopicId),
					}
				}
				if sw.Event != nil {
					info.Event = &LogSwitchDetailInfo{
						Enable:   getBoolValue(sw.Event.Enable),
						ErrorMsg: getStringValue(sw.Event.ErrorMsg),
						LogsetId: getStringValue(sw.Event.LogsetId),
						Status:   getStringValue(sw.Event.Status),
						TopicId:  getStringValue(sw.Event.TopicId),
					}
				}
				if sw.Log != nil {
					info.Log = &LogSwitchDetailInfo{
						Enable:   getBoolValue(sw.Log.Enable),
						ErrorMsg: getStringValue(sw.Log.ErrorMsg),
						LogsetId: getStringValue(sw.Log.LogsetId),
						Status:   getStringValue(sw.Log.Status),
						TopicId:  getStringValue(sw.Log.TopicId),
					}
				}
				if sw.MasterLog != nil {
					info.MasterLog = &LogSwitchDetailInfo{
						Enable:   getBoolValue(sw.MasterLog.Enable),
						ErrorMsg: getStringValue(sw.MasterLog.ErrorMsg),
						LogsetId: getStringValue(sw.MasterLog.LogsetId),
						Status:   getStringValue(sw.MasterLog.Status),
						TopicId:  getStringValue(sw.MasterLog.TopicId),
					}
				}
				break
			}
		}
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
	}).Info("成功查询集群日志开关信息")
	
	return info, nil
}

// FormatLogSwitchesAsJSON 将日志开关信息格式化为 JSON
func (c *Client) FormatLogSwitchesAsJSON(info *ClusterLogSwitchInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化日志开关信息失败: %w", err)
	}
	return string(data), nil
}

// formatSwitchDetail 格式化单个开关详情
func formatSwitchDetail(name string, detail *LogSwitchDetailInfo) string {
	if detail == nil {
		return fmt.Sprintf("【%s】未配置\n", name)
	}
	enableStr := "关闭"
	if detail.Enable {
		enableStr = "开启"
	}
	result := fmt.Sprintf("【%s】%s\n", name, enableStr)
	if detail.Status != "" {
		result += fmt.Sprintf("  状态: %s\n", detail.Status)
	}
	if detail.LogsetId != "" {
		result += fmt.Sprintf("  日志集ID: %s\n", detail.LogsetId)
	}
	if detail.TopicId != "" {
		result += fmt.Sprintf("  日志主题ID: %s\n", detail.TopicId)
	}
	if detail.ErrorMsg != "" {
		result += fmt.Sprintf("  错误信息: %s\n", detail.ErrorMsg)
	}
	return result
}

// FormatLogSwitchesAsTable 将日志开关信息格式化为表格
func (c *Client) FormatLogSwitchesAsTable(info *ClusterLogSwitchInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) 日志开关信息:\n\n", info.ClusterID, info.Region)
	result += formatSwitchDetail("审计日志 (Audit)", info.Audit)
	result += "\n"
	result += formatSwitchDetail("事件日志 (Event)", info.Event)
	result += "\n"
	result += formatSwitchDetail("普通日志 (Log)", info.Log)
	result += "\n"
	result += formatSwitchDetail("Master日志 (MasterLog)", info.MasterLog)
	return result
}

// MasterComponentInfo Master 组件状态信息
type MasterComponentInfo struct {
	ClusterID string `json:"cluster_id"`
	Region    string `json:"region"`
	Component string `json:"component"`
	Status    string `json:"status"`
}

// DescribeMasterComponent 查询 master 组件状态
func (c *Client) DescribeMasterComponent(ctx context.Context, region string, clusterID string, component string) (*MasterComponentInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
		"component":  component,
	}).Debug("开始查询 master 组件状态")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeMasterComponentRequest()
	request.ClusterId = &clusterID
	request.Component = &component
	
	response, err := client.DescribeMasterComponent(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
				"cluster_id": clusterID,
				"component":  component,
			}).Error("查询 master 组件状态 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 master 组件状态失败: %w", err)
	}
	
	info := &MasterComponentInfo{
		ClusterID: clusterID,
		Region:    region,
		Component: component,
	}
	
	if response.Response != nil {
		info.Component = getStringValue(response.Response.Component)
		info.Status = getStringValue(response.Response.Status)
	}
	
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
		"component":  info.Component,
		"status":     info.Status,
	}).Info("成功查询 master 组件状态")
	
	return info, nil
}

// FormatMasterComponentAsJSON 将 master 组件状态格式化为 JSON
func (c *Client) FormatMasterComponentAsJSON(info *MasterComponentInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 master 组件状态失败: %w", err)
	}
	return string(data), nil
}

// FormatMasterComponentAsTable 将 master 组件状态格式化为表格
func (c *Client) FormatMasterComponentAsTable(info *MasterComponentInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) Master 组件状态:\n\n", info.ClusterID, info.Region)
	result += fmt.Sprintf("组件: %s\n", info.Component)
	result += fmt.Sprintf("状态: %s\n", info.Status)
	return result
}

// ClusterInstanceInfo 集群节点实例信息
type ClusterInstanceInfo struct {
	InstanceId         string `json:"instance_id"`
	InstanceRole       string `json:"instance_role"`
	InstanceState      string `json:"instance_state"`
	FailedReason       string `json:"failed_reason,omitempty"`
	DrainStatus        string `json:"drain_status,omitempty"`
	LanIP              string `json:"lan_ip,omitempty"`
	NodePoolId         string `json:"node_pool_id,omitempty"`
	AutoscalingGroupId string `json:"autoscaling_group_id,omitempty"`
	CreatedTime        string `json:"created_time,omitempty"`
}

// ClusterInstanceListInfo 集群节点实例列表信息
type ClusterInstanceListInfo struct {
	ClusterID     string                `json:"cluster_id"`
	Region        string                `json:"region"`
	TotalCount    int64                 `json:"total_count"`
	InstanceCount int                   `json:"instance_count"`
	Instances     []ClusterInstanceInfo `json:"instances"`
}

// DescribeClusterInstances 查询集群节点实例列表
func (c *Client) DescribeClusterInstances(ctx context.Context, region string, clusterID string, instanceRole string) (*ClusterInstanceListInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":        region,
		"cluster_id":    clusterID,
		"instance_role": instanceRole,
	}).Debug("开始查询集群节点实例列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	info := &ClusterInstanceListInfo{
		ClusterID: clusterID,
		Region:    region,
	}
	
	var offset int64 = 0
	var limit int64 = 100
	
	for {
		request := tke.NewDescribeClusterInstancesRequest()
		request.ClusterId = &clusterID
		request.Offset = &offset
		request.Limit = &limit
		if instanceRole != "" {
			request.InstanceRole = &instanceRole
		}
		
		response, err := client.DescribeClusterInstances(request)
		if err != nil {
			if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
				c.logger.WithFields(logrus.Fields{
					"code":       sdkError.Code,
					"message":    sdkError.Message,
					"request_id": sdkError.RequestId,
					"region":     region,
					"cluster_id": clusterID,
				}).Error("查询集群节点实例 API 调用失败")
				return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
			}
			return nil, fmt.Errorf("查询集群节点实例失败: %w", err)
		}
		
		if response.Response != nil {
			if response.Response.TotalCount != nil {
				info.TotalCount = int64(*response.Response.TotalCount)
			}
			for _, inst := range response.Response.InstanceSet {
				info.Instances = append(info.Instances, ClusterInstanceInfo{
					InstanceId:         getStringValue(inst.InstanceId),
					InstanceRole:       getStringValue(inst.InstanceRole),
					InstanceState:      getStringValue(inst.InstanceState),
					FailedReason:       getStringValue(inst.FailedReason),
					DrainStatus:        getStringValue(inst.DrainStatus),
					LanIP:              getStringValue(inst.LanIP),
					NodePoolId:         getStringValue(inst.NodePoolId),
					AutoscalingGroupId: getStringValue(inst.AutoscalingGroupId),
					CreatedTime:        getStringValue(inst.CreatedTime),
				})
			}
		}
		
		if int64(len(info.Instances)) >= info.TotalCount {
			break
		}
		offset += limit
	}
	info.InstanceCount = len(info.Instances)
	
	c.logger.WithFields(logrus.Fields{
		"region":         region,
		"cluster_id":     clusterID,
		"instance_count": info.InstanceCount,
	}).Info("成功查询集群节点实例列表")
	
	return info, nil
}

// FormatClusterInstancesAsJSON 将集群节点实例列表格式化为 JSON
func (c *Client) FormatClusterInstancesAsJSON(info *ClusterInstanceListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化集群节点实例列表失败: %w", err)
	}
	return string(data), nil
}

// FormatClusterInstancesAsTable 将集群节点实例列表格式化为表格
func (c *Client) FormatClusterInstancesAsTable(info *ClusterInstanceListInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) 节点实例列表:\n", info.ClusterID, info.Region)
	
	if info.InstanceCount == 0 {
		result += "\n该集群未找到任何节点实例。\n"
		return result
	}
	
	result += "┌──────────────────────┬──────────────┬──────────────┬──────────────────┬──────────────┬──────────────────────┐\n"
	result += "│ 实例ID               │ 角色         │ 状态         │ 内网IP           │ 封锁状态     │ 节点池ID             │\n"
	result += "├──────────────────────┼──────────────┼──────────────┼──────────────────┼──────────────┼──────────────────────┤\n"
	
	for _, inst := range info.Instances {
		instanceId := truncateString(inst.InstanceId, 20)
		role := truncateString(inst.InstanceRole, 12)
		state := truncateString(inst.InstanceState, 12)
		lanIP := truncateString(inst.LanIP, 16)
		drain := truncateString(inst.DrainStatus, 12)
		poolId := truncateString(inst.NodePoolId, 20)
		
		result += fmt.Sprintf("│ %-20s │ %-12s │ %-12s │ %-16s │ %-12s │ %-20s │\n",
			instanceId, role, state, lanIP, drain, poolId)
	}
	
	result += "└──────────────────────┴──────────────┴──────────────┴──────────────────┴──────────────┴──────────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个节点实例", info.InstanceCount)
	
	return result
}

// VirtualNodeInfo 超级节点信息
type VirtualNodeInfo struct {
	Name        string `json:"name"`
	SubnetId    string `json:"subnet_id"`
	Phase       string `json:"phase"`
	CreatedTime string `json:"created_time,omitempty"`
}

// VirtualNodeListInfo 超级节点列表信息
type VirtualNodeListInfo struct {
	ClusterID  string            `json:"cluster_id"`
	Region     string            `json:"region"`
	NodePoolId string            `json:"node_pool_id,omitempty"`
	TotalCount int64             `json:"total_count"`
	NodeCount  int               `json:"node_count"`
	Nodes      []VirtualNodeInfo `json:"nodes"`
}

// DescribeClusterVirtualNode 查询集群超级节点列表
func (c *Client) DescribeClusterVirtualNode(ctx context.Context, region string, clusterID string, nodePoolId string) (*VirtualNodeListInfo, error) {
	c.logger.WithFields(logrus.Fields{
		"region":       region,
		"cluster_id":   clusterID,
		"node_pool_id": nodePoolId,
	}).Debug("开始查询集群超级节点列表")
	
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("tke")
	
	client, err := tke.NewClient(credential, region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	request := tke.NewDescribeClusterVirtualNodeRequest()
	request.ClusterId = &clusterID
	if nodePoolId != "" {
		request.NodePoolId = &nodePoolId
	}
	
	response, err := client.DescribeClusterVirtualNode(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			c.logger.WithFields(logrus.Fields{
				"code":       sdkError.Code,
				"message":    sdkError.Message,
				"request_id": sdkError.RequestId,
				"region":     region,
				"cluster_id": clusterID,
			}).Error("查询集群超级节点 API 调用失败")
			return nil, fmt.Errorf("TKE API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询集群超级节点失败: %w", err)
	}
	
	info := &VirtualNodeListInfo{
		ClusterID:  clusterID,
		Region:     region,
		NodePoolId: nodePoolId,
	}
	
	if response.Response != nil {
		if response.Response.TotalCount != nil {
			info.TotalCount = int64(*response.Response.TotalCount)
		}
		if response.Response.Nodes != nil {
			for _, node := range response.Response.Nodes {
				info.Nodes = append(info.Nodes, VirtualNodeInfo{
					Name:        getStringValue(node.Name),
					SubnetId:    getStringValue(node.SubnetId),
					Phase:       getStringValue(node.Phase),
					CreatedTime: getStringValue(node.CreatedTime),
				})
			}
		}
	}
	info.NodeCount = len(info.Nodes)
	
	c.logger.WithFields(logrus.Fields{
		"region":     region,
		"cluster_id": clusterID,
		"node_count": info.NodeCount,
	}).Info("成功查询集群超级节点列表")
	
	return info, nil
}

// FormatVirtualNodesAsJSON 将超级节点列表格式化为 JSON
func (c *Client) FormatVirtualNodesAsJSON(info *VirtualNodeListInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化超级节点列表失败: %w", err)
	}
	return string(data), nil
}

// FormatVirtualNodesAsTable 将超级节点列表格式化为表格
func (c *Client) FormatVirtualNodesAsTable(info *VirtualNodeListInfo) string {
	result := fmt.Sprintf("集群 %s (地域: %s) 超级节点列表:\n", info.ClusterID, info.Region)
	
	if info.NodeCount == 0 {
		result += "\n该集群未找到任何超级节点。\n"
		return result
	}
	
	result += "┌──────────────────────────────┬──────────────────────────┬──────────────────┬──────────────────────────┐\n"
	result += "│ 节点名称                     │ 子网ID                   │ 状态             │ 创建时间                 │\n"
	result += "├──────────────────────────────┼──────────────────────────┼──────────────────┼──────────────────────────┤\n"
	
	for _, node := range info.Nodes {
		name := truncateString(node.Name, 28)
		subnetId := truncateString(node.SubnetId, 24)
		phase := truncateString(node.Phase, 16)
		createdTime := truncateString(node.CreatedTime, 24)
		
		result += fmt.Sprintf("│ %-28s │ %-24s │ %-16s │ %-24s │\n",
			name, subnetId, phase, createdTime)
	}
	
	result += "└──────────────────────────────┴──────────────────────────┴──────────────────┴──────────────────────────┘\n"
	result += fmt.Sprintf("总计: %d 个超级节点", info.NodeCount)
	
	return result
}

// convertStringPtrSlice 将 []*string 转换为 []string
func convertStringPtrSlice(ptrs []*string) []string {
	if ptrs == nil {
		return nil
	}
	result := make([]string, 0, len(ptrs))
	for _, ptr := range ptrs {
		if ptr != nil {
			result = append(result, *ptr)
		}
	}
	return result
}

// 辅助函数：安全获取字符串指针的值
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// 辅助函数：安全获取int64指针的值
func getInt64Value(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// 辅助函数：安全获取bool指针的值
func getBoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// 辅助函数：安全获取uint64指针的值并转换为int64
func getUint64AsInt64Value(ptr *uint64) int64 {
	if ptr == nil {
		return 0
	}
	return int64(*ptr)
}

// 辅助函数：截断字符串以适应表格显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}