package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	
	"ai-sre/tools/mcp/internal/tencentcloud"
	"ai-sre/tools/mcp/internal/tencentcloud/cdb"
	"ai-sre/tools/mcp/internal/tencentcloud/clb"
	"ai-sre/tools/mcp/internal/tencentcloud/cvm"
	"ai-sre/tools/mcp/internal/tencentcloud/region"
	"ai-sre/tools/mcp/internal/tencentcloud/tke"
	"ai-sre/tools/mcp/internal/tencentcloud/vpc"
	"ai-sre/tools/mcp/pkg/logger"
)

// GetClusterLevelPriceArgs 获取集群等级价格参数
type GetClusterLevelPriceArgs struct {
	Region       *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterLevel *string `json:"cluster_level" jsonschema:"description=集群等级: L20、L50、L100、L200、L500、L1000、L3000、L5000,enum=L20,enum=L50,enum=L100,enum=L200,enum=L500,enum=L1000,enum=L3000,enum=L5000,required"`
	Format       *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeAddonArgs 查询集群已安装的 addon 列表参数
type DescribeAddonArgs struct {
	Region    *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	AddonName *string `json:"addon_name,omitempty" jsonschema:"description=addon名称(不传时返回集群下全部addon)"`
	Format    *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// GetTkeAppChartListArgs 获取可安装的 addon 列表参数
type GetTkeAppChartListArgs struct {
	Region      *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Kind        *string `json:"kind,omitempty" jsonschema:"description=app类型: log、scheduler、network、storage、monitor、dns、image、other、invisible"`
	Arch        *string `json:"arch,omitempty" jsonschema:"description=支持的操作系统架构: arm32、arm64、amd64"`
	ClusterType *string `json:"cluster_type,omitempty" jsonschema:"description=集群类型: tke、eks"`
	Format      *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeImagesArgs 查询 OS 镜像列表参数
type DescribeImagesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeVersionsArgs 查询集群版本列表参数
type DescribeVersionsArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeLogSwitchesArgs 查询集群日志开关参数
type DescribeLogSwitchesArgs struct {
	Region    *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	Format    *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeMasterComponentArgs 查询 master 组件状态参数
type DescribeMasterComponentArgs struct {
	Region    *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	Component *string `json:"component,omitempty" jsonschema:"description=master组件名称,enum=kube-apiserver,enum=kube-scheduler,enum=kube-controller-manager,default=kube-apiserver"`
	Format    *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeClusterInstancesArgs 查询集群节点实例列表参数
type DescribeClusterInstancesArgs struct {
	Region       *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID    *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	InstanceRole *string `json:"instance_role,omitempty" jsonschema:"description=节点角色: WORKER、MASTER、ETCD、MASTER_ETCD、ALL,enum=WORKER,enum=MASTER,enum=ETCD,enum=MASTER_ETCD,enum=ALL,default=WORKER"`
	Format       *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeClusterVirtualNodeArgs 查询集群超级节点列表参数
type DescribeClusterVirtualNodeArgs struct {
	Region     *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID  *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	NodePoolId *string `json:"node_pool_id,omitempty" jsonschema:"description=节点池ID(不传时返回集群下全部超级节点)"`
	Format     *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeClusterExtraArgsArgs 查询集群自定义参数
type DescribeClusterExtraArgsArgs struct {
	Region    *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterID *string `json:"cluster_id" jsonschema:"description=集群ID,required"`
	Format    *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// === CVM Args ===

// CvmDescribeInstancesArgs 查询 CVM 实例列表参数
type CvmDescribeInstancesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// CvmDescribeInstancesStatusArgs 查询 CVM 实例状态列表参数
type CvmDescribeInstancesStatusArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// === CLB Args ===

// ClbDescribeLoadBalancersArgs 查询 CLB 实例列表参数
type ClbDescribeLoadBalancersArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// ClbDescribeListenersArgs 查询 CLB 监听器列表参数
type ClbDescribeListenersArgs struct {
	Region         *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	LoadBalancerId *string `json:"load_balancer_id" jsonschema:"description=负载均衡实例ID,required"`
	Format         *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// ClbDescribeTargetsArgs 查询 CLB 后端服务列表参数
type ClbDescribeTargetsArgs struct {
	Region         *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	LoadBalancerId *string `json:"load_balancer_id" jsonschema:"description=负载均衡实例ID,required"`
	Format         *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// ClbDescribeTargetHealthArgs 查询 CLB 后端健康状态参数
type ClbDescribeTargetHealthArgs struct {
	Region          *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	LoadBalancerIds *string `json:"load_balancer_ids" jsonschema:"description=负载均衡实例ID(多个用逗号分隔),required"`
	Format          *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// === CDB Args ===

// CdbDescribeDBInstancesArgs 查询 CDB 实例列表参数
type CdbDescribeDBInstancesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// CdbDescribeDBInstanceInfoArgs 查询 CDB 实例详细信息参数
type CdbDescribeDBInstanceInfoArgs struct {
	Region     *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	InstanceId *string `json:"instance_id" jsonschema:"description=CDB实例ID,required"`
	Format     *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// CdbDescribeSlowLogsArgs 查询 CDB 慢日志参数
type CdbDescribeSlowLogsArgs struct {
	Region     *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	InstanceId *string `json:"instance_id" jsonschema:"description=CDB实例ID,required"`
	Format     *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// CdbDescribeErrorLogArgs 查询 CDB 错误日志参数
type CdbDescribeErrorLogArgs struct {
	Region     *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	InstanceId *string `json:"instance_id" jsonschema:"description=CDB实例ID,required"`
	StartTime  *string `json:"start_time,omitempty" jsonschema:"description=开始时间(Unix时间戳秒级 或 不传则默认最近1小时)"`
	EndTime    *string `json:"end_time,omitempty" jsonschema:"description=结束时间(Unix时间戳秒级 或 不传则默认当前时间)"`
	Format     *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// === VPC Args ===

// VpcDescribeVpcsArgs 查询 VPC 列表参数
type VpcDescribeVpcsArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeSubnetsArgs 查询子网列表参数
type VpcDescribeSubnetsArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	VpcId  *string `json:"vpc_id,omitempty" jsonschema:"description=VPC实例ID(不传则查询所有子网)"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeSecurityGroupsArgs 查询安全组列表参数
type VpcDescribeSecurityGroupsArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeNetworkInterfacesArgs 查询弹性网卡列表参数
type VpcDescribeNetworkInterfacesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	VpcId  *string `json:"vpc_id,omitempty" jsonschema:"description=VPC实例ID(不传则查询所有网卡)"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeAddressesArgs 查询弹性公网IP列表参数
type VpcDescribeAddressesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeBandwidthPackagesArgs 查询带宽包列表参数
type VpcDescribeBandwidthPackagesArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeVpcEndPointArgs 查询终端节点列表参数
type VpcDescribeVpcEndPointArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeVpcEndPointServiceArgs 查询终端节点服务列表参数
type VpcDescribeVpcEndPointServiceArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// VpcDescribeVpcPeeringConnectionsArgs 查询对等连接列表参数
type VpcDescribeVpcPeeringConnectionsArgs struct {
	Region *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	Format *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeClustersArgs 查询 TKE 集群列表参数
type DescribeClustersArgs struct {
	Region      *string `json:"region" jsonschema:"description=地域ID(如ap-beijing、ap-shanghai等),required"`
	ClusterType *string `json:"cluster_type,omitempty" jsonschema:"description=集群类型: all(全部集群)、tke(普通集群)、serverless(弹性集群),enum=all,enum=tke,enum=serverless,default=all"`
	Format      *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// TencentCloudTools 腾讯云工具集
type TencentCloudTools struct {
	clientManager *tencentcloud.ClientManager
	regionClient  *region.Client
	tkeClient     *tke.Client
	cvmClient     *cvm.Client
	clbClient     *clb.Client
	cdbClient     *cdb.Client
	vpcClient     *vpc.Client
	logger        *logrus.Logger
}

// NewTencentCloudTools 创建腾讯云工具集
func NewTencentCloudTools() (*TencentCloudTools, error) {
	// 加载腾讯云配置
	config, err := tencentcloud.GetConfigFromMultipleSources(logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("加载腾讯云配置失败: %w", err)
	}
	
	// 创建客户端管理器
	clientManager := tencentcloud.NewClientManager(config, logger.GetLogger())
	
	// 验证配置
	if err := clientManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("腾讯云配置验证失败: %w", err)
	}
	
	// 创建地域管理客户端
	regionClient, err := region.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建地域管理客户端失败: %w", err)
	}
	
	// 创建 TKE 客户端
	tkeClient, err := tke.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建 TKE 客户端失败: %w", err)
	}
	
	// 创建 CVM 客户端
	cvmClient, err := cvm.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建 CVM 客户端失败: %w", err)
	}
	
	// 创建 CLB 客户端
	clbClient, err := clb.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建 CLB 客户端失败: %w", err)
	}
	
	// 创建 CDB 客户端
	cdbClient, err := cdb.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建 CDB 客户端失败: %w", err)
	}
	
	// 创建 VPC 客户端
	vpcClient, err := vpc.NewClient(clientManager, logger.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}
	
	logger.GetLogger().WithFields(logrus.Fields{
		"config": config.MaskSensitiveInfo(),
	}).Info("腾讯云工具集初始化成功")
	
	return &TencentCloudTools{
		clientManager: clientManager,
		regionClient:  regionClient,
		tkeClient:     tkeClient,
		cvmClient:     cvmClient,
		clbClient:     clbClient,
		cdbClient:     cdbClient,
		vpcClient:     vpcClient,
		logger:        logger.GetLogger(),
	}, nil
}

// DescribeRegions 查询产品支持的地域信息
func (t *TencentCloudTools) DescribeRegions(ctx context.Context, args DescribeRegionsArgs) (string, error) {
	product := "cvm"
	if args.Product != nil && *args.Product != "" {
		product = *args.Product
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":    "describe_regions",
		"product": product,
		"format": func() string {
			if args.Format != nil {
				return *args.Format
			}
			return ""
		}(),
	}).Info("开始执行地域查询")
	
	// 使用地域管理系统查询产品支持的地域信息
	regions, err := t.regionClient.DescribeRegions(ctx, product)
	if err != nil {
		t.logger.WithError(err).Error("地域查询失败")
		return "", fmt.Errorf("查询产品 %s 地域信息失败: %w", product, err)
	}
	
	// 根据格式返回结果
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.regionClient.FormatRegionsAsJSON(regions)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		// 默认使用表格格式
		return t.regionClient.FormatRegionsAsTable(regions, strings.ToUpper(product)), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// GetRegion 获取特定地域信息
func (t *TencentCloudTools) GetRegion(ctx context.Context, args GetRegionArgs) (string, error) {
	product := "cvm"
	if args.Product != nil && *args.Product != "" {
		product = *args.Product
	}
	
	regionID := ""
	if args.RegionID != nil {
		regionID = *args.RegionID
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":      "get_region",
		"product":   product,
		"region_id": regionID,
		"format": func() string {
			if args.Format != nil {
				return *args.Format
			}
			return ""
		}(),
	}).Info("开始执行特定地域查询")
	
	if regionID == "" {
		return "", fmt.Errorf("地域ID不能为空")
	}
	
	// 使用地域管理系统查询特定地域信息
	region, err := t.regionClient.GetRegionByID(ctx, product, regionID)
	if err != nil {
		t.logger.WithError(err).Error("特定地域查询失败")
		return "", fmt.Errorf("查询产品 %s 地域 %s 信息失败: %w", product, regionID, err)
	}
	
	// 根据格式返回结果
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(region, "", "  ")
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return string(data), nil
	case "table", "":
		// 默认使用表格格式
		result := fmt.Sprintf("%s 地域信息:\n", strings.ToUpper(product))
		result += fmt.Sprintf("地域ID: %s\n", region.RegionID)
		result += fmt.Sprintf("地域名称: %s\n", region.RegionName)
		result += fmt.Sprintf("状态: %s\n", region.RegionState)
		return result, nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ValidateConnection 验证腾讯云连接
func (t *TencentCloudTools) ValidateConnection(ctx context.Context) error {
	t.logger.Info("开始验证腾讯云连接")
	
	// 验证地域管理权限
	if err := t.regionClient.ValidatePermissions(ctx); err != nil {
		t.logger.WithError(err).Error("地域管理权限验证失败")
		return fmt.Errorf("地域管理权限验证失败: %w", err)
	}
	
	t.logger.Info("腾讯云连接验证成功")
	return nil
}

// DescribeClusters 查询 TKE 集群列表（支持按集群类型过滤）
func (t *TencentCloudTools) DescribeClusters(ctx context.Context, args DescribeClustersArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterType := "all"
	if args.ClusterType != nil && *args.ClusterType != "" {
		clusterType = strings.ToLower(*args.ClusterType)
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":         "tke_describe_clusters",
		"region":       region,
		"cluster_type": clusterType,
		"format": func() string {
			if args.Format != nil {
				return *args.Format
			}
			return ""
		}(),
	}).Info("开始执行 TKE 集群列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	
	// 验证 cluster_type 参数
	switch clusterType {
	case "all", "tke", "serverless":
		// 合法值
	default:
		return "", fmt.Errorf("不支持的集群类型: %s，支持的类型: all, tke, serverless", clusterType)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	var resultParts []string
	
	// 查询普通集群
	if clusterType == "all" || clusterType == "tke" {
		clusters, err := t.tkeClient.DescribeClusters(ctx, region)
		if err != nil {
			t.logger.WithError(err).Error("TKE 普通集群列表查询失败")
			return "", fmt.Errorf("查询地域 %s 的 TKE 普通集群列表失败: %w", region, err)
		}
		
		switch strings.ToLower(format) {
		case "json":
			result, err := t.tkeClient.FormatClustersAsJSON(clusters)
			if err != nil {
				return "", fmt.Errorf("格式化普通集群 JSON 结果失败: %w", err)
			}
			resultParts = append(resultParts, fmt.Sprintf("=== TKE 普通集群 ===\n%s", result))
		case "table", "":
			resultParts = append(resultParts, t.tkeClient.FormatClustersAsTable(clusters, region))
		default:
			return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
		}
	}
	
	// 查询 Serverless 集群
	if clusterType == "all" || clusterType == "serverless" {
		eksClusters, err := t.tkeClient.DescribeEKSClusters(ctx, region)
		if err != nil {
			t.logger.WithError(err).Error("EKS Serverless 集群列表查询失败")
			return "", fmt.Errorf("查询地域 %s 的 EKS Serverless 集群列表失败: %w", region, err)
		}
		
		switch strings.ToLower(format) {
		case "json":
			result, err := t.tkeClient.FormatEKSClustersAsJSON(eksClusters)
			if err != nil {
				return "", fmt.Errorf("格式化 Serverless 集群 JSON 结果失败: %w", err)
			}
			resultParts = append(resultParts, fmt.Sprintf("=== EKS Serverless 集群 ===\n%s", result))
		case "table", "":
			resultParts = append(resultParts, t.tkeClient.FormatEKSClustersAsTable(eksClusters, region))
		default:
			return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
		}
	}
	
	return strings.Join(resultParts, "\n\n"), nil
}

// GetClusterLevelPrice 获取集群等级价格
func (t *TencentCloudTools) GetClusterLevelPrice(ctx context.Context, args GetClusterLevelPriceArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterLevel := ""
	if args.ClusterLevel != nil {
		clusterLevel = *args.ClusterLevel
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":          "tke_get_cluster_level_price",
		"region":        region,
		"cluster_level": clusterLevel,
	}).Info("开始执行集群等级价格查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterLevel == "" {
		return "", fmt.Errorf("集群等级参数不能为空")
	}
	
	info, err := t.tkeClient.GetClusterLevelPrice(ctx, region, clusterLevel)
	if err != nil {
		t.logger.WithError(err).Error("集群等级价格查询失败")
		return "", fmt.Errorf("查询集群等级 %s 的价格失败: %w", clusterLevel, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatClusterLevelPriceAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatClusterLevelPriceAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeAddon 查询集群已安装的 addon 列表
func (t *TencentCloudTools) DescribeAddon(ctx context.Context, args DescribeAddonArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	addonName := ""
	if args.AddonName != nil {
		addonName = *args.AddonName
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":       "tke_describe_addon",
		"region":     region,
		"cluster_id": clusterID,
		"addon_name": addonName,
	}).Info("开始执行集群 addon 列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeAddon(ctx, region, clusterID, addonName)
	if err != nil {
		t.logger.WithError(err).Error("集群 addon 列表查询失败")
		return "", fmt.Errorf("查询集群 %s 的 addon 列表失败: %w", clusterID, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatAddonListAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatAddonListAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// GetTkeAppChartList 获取可安装的 addon 列表
func (t *TencentCloudTools) GetTkeAppChartList(ctx context.Context, args GetTkeAppChartListArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	kind := ""
	if args.Kind != nil {
		kind = *args.Kind
	}
	
	arch := ""
	if args.Arch != nil {
		arch = *args.Arch
	}
	
	clusterType := ""
	if args.ClusterType != nil {
		clusterType = *args.ClusterType
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":         "tke_get_app_chart_list",
		"region":       region,
		"kind":         kind,
		"arch":         arch,
		"cluster_type": clusterType,
	}).Info("开始执行可安装 addon 列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	
	info, err := t.tkeClient.GetTkeAppChartList(ctx, region, kind, arch, clusterType)
	if err != nil {
		t.logger.WithError(err).Error("可安装 addon 列表查询失败")
		return "", fmt.Errorf("查询可安装 addon 列表失败: %w", err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatAppChartListAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatAppChartListAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeImages 查询指定地域支持的 OS 镜像列表
func (t *TencentCloudTools) DescribeImages(ctx context.Context, args DescribeImagesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":   "tke_describe_images",
		"region": region,
	}).Info("开始执行 OS 镜像列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeImages(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("OS 镜像列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的 OS 镜像列表失败: %w", region, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatImagesAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatImagesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeVersions 查询指定地域支持的集群版本列表
func (t *TencentCloudTools) DescribeVersions(ctx context.Context, args DescribeVersionsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":   "tke_describe_versions",
		"region": region,
	}).Info("开始执行集群版本列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeVersions(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("集群版本列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的集群版本列表失败: %w", region, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatVersionsAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatVersionsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeLogSwitches 查询集群日志开关信息
func (t *TencentCloudTools) DescribeLogSwitches(ctx context.Context, args DescribeLogSwitchesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":       "tke_describe_log_switches",
		"region":     region,
		"cluster_id": clusterID,
	}).Info("开始执行集群日志开关查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeLogSwitches(ctx, region, clusterID)
	if err != nil {
		t.logger.WithError(err).Error("集群日志开关查询失败")
		return "", fmt.Errorf("查询集群 %s 的日志开关失败: %w", clusterID, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatLogSwitchesAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatLogSwitchesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeMasterComponent 查询 master 组件状态
func (t *TencentCloudTools) DescribeMasterComponent(ctx context.Context, args DescribeMasterComponentArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	component := "kube-apiserver"
	if args.Component != nil && *args.Component != "" {
		component = *args.Component
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":       "tke_describe_master_component",
		"region":     region,
		"cluster_id": clusterID,
		"component":  component,
	}).Info("开始执行 master 组件状态查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeMasterComponent(ctx, region, clusterID, component)
	if err != nil {
		t.logger.WithError(err).Error("master 组件状态查询失败")
		return "", fmt.Errorf("查询集群 %s 的 master 组件 %s 状态失败: %w", clusterID, component, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatMasterComponentAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatMasterComponentAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeClusterInstances 查询集群节点实例列表
func (t *TencentCloudTools) DescribeClusterInstances(ctx context.Context, args DescribeClusterInstancesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	instanceRole := ""
	if args.InstanceRole != nil {
		instanceRole = *args.InstanceRole
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":          "tke_describe_cluster_instances",
		"region":        region,
		"cluster_id":    clusterID,
		"instance_role": instanceRole,
	}).Info("开始执行集群节点实例列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeClusterInstances(ctx, region, clusterID, instanceRole)
	if err != nil {
		t.logger.WithError(err).Error("集群节点实例列表查询失败")
		return "", fmt.Errorf("查询集群 %s 的节点实例列表失败: %w", clusterID, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatClusterInstancesAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatClusterInstancesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeClusterVirtualNode 查询集群超级节点列表
func (t *TencentCloudTools) DescribeClusterVirtualNode(ctx context.Context, args DescribeClusterVirtualNodeArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	nodePoolId := ""
	if args.NodePoolId != nil {
		nodePoolId = *args.NodePoolId
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":         "tke_describe_cluster_virtual_node",
		"region":       region,
		"cluster_id":   clusterID,
		"node_pool_id": nodePoolId,
	}).Info("开始执行集群超级节点列表查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeClusterVirtualNode(ctx, region, clusterID, nodePoolId)
	if err != nil {
		t.logger.WithError(err).Error("集群超级节点列表查询失败")
		return "", fmt.Errorf("查询集群 %s 的超级节点列表失败: %w", clusterID, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatVirtualNodesAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatVirtualNodesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// DescribeClusterExtraArgs 查询集群自定义参数
func (t *TencentCloudTools) DescribeClusterExtraArgs(ctx context.Context, args DescribeClusterExtraArgsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}
	
	clusterID := ""
	if args.ClusterID != nil {
		clusterID = *args.ClusterID
	}
	
	t.logger.WithFields(logrus.Fields{
		"tool":       "tke_describe_cluster_extra_args",
		"region":     region,
		"cluster_id": clusterID,
	}).Info("开始执行集群自定义参数查询")
	
	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if clusterID == "" {
		return "", fmt.Errorf("集群ID参数不能为空")
	}
	
	info, err := t.tkeClient.DescribeClusterExtraArgs(ctx, region, clusterID)
	if err != nil {
		t.logger.WithError(err).Error("集群自定义参数查询失败")
		return "", fmt.Errorf("查询集群 %s 的自定义参数失败: %w", clusterID, err)
	}
	
	format := ""
	if args.Format != nil {
		format = *args.Format
	}
	
	switch strings.ToLower(format) {
	case "json":
		result, err := t.tkeClient.FormatClusterExtraArgsAsJSON(info)
		if err != nil {
			return "", fmt.Errorf("格式化 JSON 结果失败: %w", err)
		}
		return result, nil
	case "table", "":
		return t.tkeClient.FormatClusterExtraArgsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ========== CVM 工具方法 ==========

// CvmDescribeInstances 查询 CVM 实例列表
func (t *TencentCloudTools) CvmDescribeInstances(ctx context.Context, args CvmDescribeInstancesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "cvm_describe_instances",
		"region": region,
	}).Info("开始执行 CVM 实例列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.cvmClient.DescribeInstances(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("CVM 实例列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的 CVM 实例列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cvmClient.FormatInstancesAsJSON(info)
	case "table", "":
		return t.cvmClient.FormatInstancesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// CvmDescribeInstancesStatus 查询 CVM 实例状态列表
func (t *TencentCloudTools) CvmDescribeInstancesStatus(ctx context.Context, args CvmDescribeInstancesStatusArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "cvm_describe_instances_status",
		"region": region,
	}).Info("开始执行 CVM 实例状态查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.cvmClient.DescribeInstancesStatus(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("CVM 实例状态查询失败")
		return "", fmt.Errorf("查询地域 %s 的 CVM 实例状态失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cvmClient.FormatInstancesStatusAsJSON(info)
	case "table", "":
		return t.cvmClient.FormatInstancesStatusAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ========== CLB 工具方法 ==========

// ClbDescribeLoadBalancers 查询 CLB 实例列表
func (t *TencentCloudTools) ClbDescribeLoadBalancers(ctx context.Context, args ClbDescribeLoadBalancersArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "clb_describe_load_balancers",
		"region": region,
	}).Info("开始执行 CLB 实例列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.clbClient.DescribeLoadBalancers(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("CLB 实例列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的 CLB 实例列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.clbClient.FormatLoadBalancersAsJSON(info)
	case "table", "":
		return t.clbClient.FormatLoadBalancersAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ClbDescribeListeners 查询 CLB 监听器列表
func (t *TencentCloudTools) ClbDescribeListeners(ctx context.Context, args ClbDescribeListenersArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	loadBalancerId := ""
	if args.LoadBalancerId != nil {
		loadBalancerId = *args.LoadBalancerId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":             "clb_describe_listeners",
		"region":           region,
		"load_balancer_id": loadBalancerId,
	}).Info("开始执行 CLB 监听器列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if loadBalancerId == "" {
		return "", fmt.Errorf("负载均衡实例ID参数不能为空")
	}

	info, err := t.clbClient.DescribeListeners(ctx, region, loadBalancerId)
	if err != nil {
		t.logger.WithError(err).Error("CLB 监听器列表查询失败")
		return "", fmt.Errorf("查询 CLB %s 的监听器列表失败: %w", loadBalancerId, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.clbClient.FormatListenersAsJSON(info)
	case "table", "":
		return t.clbClient.FormatListenersAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ClbDescribeTargets 查询 CLB 后端服务列表
func (t *TencentCloudTools) ClbDescribeTargets(ctx context.Context, args ClbDescribeTargetsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	loadBalancerId := ""
	if args.LoadBalancerId != nil {
		loadBalancerId = *args.LoadBalancerId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":             "clb_describe_targets",
		"region":           region,
		"load_balancer_id": loadBalancerId,
	}).Info("开始执行 CLB 后端服务列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if loadBalancerId == "" {
		return "", fmt.Errorf("负载均衡实例ID参数不能为空")
	}

	info, err := t.clbClient.DescribeTargets(ctx, region, loadBalancerId, nil)
	if err != nil {
		t.logger.WithError(err).Error("CLB 后端服务列表查询失败")
		return "", fmt.Errorf("查询 CLB %s 的后端服务列表失败: %w", loadBalancerId, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.clbClient.FormatTargetsAsJSON(info)
	case "table", "":
		return t.clbClient.FormatTargetsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ClbDescribeTargetHealth 查询 CLB 后端健康状态
func (t *TencentCloudTools) ClbDescribeTargetHealth(ctx context.Context, args ClbDescribeTargetHealthArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	lbIdsStr := ""
	if args.LoadBalancerIds != nil {
		lbIdsStr = *args.LoadBalancerIds
	}

	t.logger.WithFields(logrus.Fields{
		"tool":              "clb_describe_target_health",
		"region":            region,
		"load_balancer_ids": lbIdsStr,
	}).Info("开始执行 CLB 后端健康状态查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if lbIdsStr == "" {
		return "", fmt.Errorf("负载均衡实例ID参数不能为空")
	}

	lbIds := strings.Split(lbIdsStr, ",")
	for i := range lbIds {
		lbIds[i] = strings.TrimSpace(lbIds[i])
	}

	info, err := t.clbClient.DescribeTargetHealth(ctx, region, lbIds)
	if err != nil {
		t.logger.WithError(err).Error("CLB 后端健康状态查询失败")
		return "", fmt.Errorf("查询 CLB 后端健康状态失败: %w", err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.clbClient.FormatTargetHealthAsJSON(info)
	case "table", "":
		return t.clbClient.FormatTargetHealthAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ========== CDB 工具方法 ==========

// CdbDescribeDBInstances 查询 CDB 实例列表
func (t *TencentCloudTools) CdbDescribeDBInstances(ctx context.Context, args CdbDescribeDBInstancesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "cdb_describe_db_instances",
		"region": region,
	}).Info("开始执行 CDB 实例列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.cdbClient.DescribeDBInstances(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("CDB 实例列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的 CDB 实例列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cdbClient.FormatDBInstancesAsJSON(info)
	case "table", "":
		return t.cdbClient.FormatDBInstancesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// CdbDescribeDBInstanceInfo 查询 CDB 实例详细信息
func (t *TencentCloudTools) CdbDescribeDBInstanceInfo(ctx context.Context, args CdbDescribeDBInstanceInfoArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	instanceId := ""
	if args.InstanceId != nil {
		instanceId = *args.InstanceId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":        "cdb_describe_db_instance_info",
		"region":      region,
		"instance_id": instanceId,
	}).Info("开始执行 CDB 实例详细信息查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if instanceId == "" {
		return "", fmt.Errorf("实例ID参数不能为空")
	}

	info, err := t.cdbClient.DescribeDBInstanceInfo(ctx, region, instanceId)
	if err != nil {
		t.logger.WithError(err).Error("CDB 实例详细信息查询失败")
		return "", fmt.Errorf("查询 CDB 实例 %s 的详细信息失败: %w", instanceId, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cdbClient.FormatDBInstanceInfoAsJSON(info)
	case "table", "":
		return t.cdbClient.FormatDBInstanceInfoAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// CdbDescribeSlowLogs 查询 CDB 慢日志
func (t *TencentCloudTools) CdbDescribeSlowLogs(ctx context.Context, args CdbDescribeSlowLogsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	instanceId := ""
	if args.InstanceId != nil {
		instanceId = *args.InstanceId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":        "cdb_describe_slow_logs",
		"region":      region,
		"instance_id": instanceId,
	}).Info("开始执行 CDB 慢日志查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if instanceId == "" {
		return "", fmt.Errorf("实例ID参数不能为空")
	}

	info, err := t.cdbClient.DescribeSlowLogs(ctx, region, instanceId)
	if err != nil {
		t.logger.WithError(err).Error("CDB 慢日志查询失败")
		return "", fmt.Errorf("查询 CDB 实例 %s 的慢日志失败: %w", instanceId, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cdbClient.FormatSlowLogsAsJSON(info)
	case "table", "":
		return t.cdbClient.FormatSlowLogsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// CdbDescribeErrorLog 查询 CDB 错误日志
func (t *TencentCloudTools) CdbDescribeErrorLog(ctx context.Context, args CdbDescribeErrorLogArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	instanceId := ""
	if args.InstanceId != nil {
		instanceId = *args.InstanceId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":        "cdb_describe_error_log",
		"region":      region,
		"instance_id": instanceId,
	}).Info("开始执行 CDB 错误日志查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}
	if instanceId == "" {
		return "", fmt.Errorf("实例ID参数不能为空")
	}

	// 默认查询最近1小时
	now := time.Now()
	endTime := uint64(now.Unix())
	startTime := uint64(now.Add(-1 * time.Hour).Unix())

	if args.StartTime != nil && *args.StartTime != "" {
		var st uint64
		if _, err := fmt.Sscanf(*args.StartTime, "%d", &st); err == nil {
			startTime = st
		}
	}
	if args.EndTime != nil && *args.EndTime != "" {
		var et uint64
		if _, err := fmt.Sscanf(*args.EndTime, "%d", &et); err == nil {
			endTime = et
		}
	}

	info, err := t.cdbClient.DescribeErrorLogData(ctx, region, instanceId, startTime, endTime)
	if err != nil {
		t.logger.WithError(err).Error("CDB 错误日志查询失败")
		return "", fmt.Errorf("查询 CDB 实例 %s 的错误日志失败: %w", instanceId, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.cdbClient.FormatErrorLogDataAsJSON(info)
	case "table", "":
		return t.cdbClient.FormatErrorLogDataAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// ========== VPC 工具方法 ==========

// VpcDescribeVpcs 查询 VPC 列表
func (t *TencentCloudTools) VpcDescribeVpcs(ctx context.Context, args VpcDescribeVpcsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_vpcs",
		"region": region,
	}).Info("开始执行 VPC 列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeVpcs(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("VPC 列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的 VPC 列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatVpcsAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatVpcsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeSubnets 查询子网列表
func (t *TencentCloudTools) VpcDescribeSubnets(ctx context.Context, args VpcDescribeSubnetsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	vpcId := ""
	if args.VpcId != nil {
		vpcId = *args.VpcId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_subnets",
		"region": region,
		"vpc_id": vpcId,
	}).Info("开始执行子网列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeSubnets(ctx, region, vpcId)
	if err != nil {
		t.logger.WithError(err).Error("子网列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的子网列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatSubnetsAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatSubnetsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeSecurityGroups 查询安全组列表
func (t *TencentCloudTools) VpcDescribeSecurityGroups(ctx context.Context, args VpcDescribeSecurityGroupsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_security_groups",
		"region": region,
	}).Info("开始执行安全组列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeSecurityGroups(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("安全组列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的安全组列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatSecurityGroupsAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatSecurityGroupsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeNetworkInterfaces 查询弹性网卡列表
func (t *TencentCloudTools) VpcDescribeNetworkInterfaces(ctx context.Context, args VpcDescribeNetworkInterfacesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	vpcId := ""
	if args.VpcId != nil {
		vpcId = *args.VpcId
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_network_interfaces",
		"region": region,
		"vpc_id": vpcId,
	}).Info("开始执行弹性网卡列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeNetworkInterfaces(ctx, region, vpcId)
	if err != nil {
		t.logger.WithError(err).Error("弹性网卡列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的弹性网卡列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatNetworkInterfacesAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatNetworkInterfacesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeAddresses 查询弹性公网IP列表
func (t *TencentCloudTools) VpcDescribeAddresses(ctx context.Context, args VpcDescribeAddressesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_addresses",
		"region": region,
	}).Info("开始执行弹性公网IP列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeAddresses(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("弹性公网IP列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的弹性公网IP列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatAddressesAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatAddressesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeBandwidthPackages 查询带宽包列表
func (t *TencentCloudTools) VpcDescribeBandwidthPackages(ctx context.Context, args VpcDescribeBandwidthPackagesArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_bandwidth_packages",
		"region": region,
	}).Info("开始执行带宽包列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeBandwidthPackages(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("带宽包列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的带宽包列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatBandwidthPackagesAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatBandwidthPackagesAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeVpcEndPoint 查询终端节点列表
func (t *TencentCloudTools) VpcDescribeVpcEndPoint(ctx context.Context, args VpcDescribeVpcEndPointArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_vpc_endpoint",
		"region": region,
	}).Info("开始执行终端节点列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeVpcEndPoint(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("终端节点列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的终端节点列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatVpcEndPointAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatVpcEndPointAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeVpcEndPointService 查询终端节点服务列表
func (t *TencentCloudTools) VpcDescribeVpcEndPointService(ctx context.Context, args VpcDescribeVpcEndPointServiceArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_vpc_endpoint_service",
		"region": region,
	}).Info("开始执行终端节点服务列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeVpcEndPointService(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("终端节点服务列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的终端节点服务列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatVpcEndPointServiceAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatVpcEndPointServiceAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}

// VpcDescribeVpcPeeringConnections 查询对等连接列表
func (t *TencentCloudTools) VpcDescribeVpcPeeringConnections(ctx context.Context, args VpcDescribeVpcPeeringConnectionsArgs) (string, error) {
	region := ""
	if args.Region != nil {
		region = *args.Region
	}

	t.logger.WithFields(logrus.Fields{
		"tool":   "vpc_describe_vpc_peering_connections",
		"region": region,
	}).Info("开始执行对等连接列表查询")

	if region == "" {
		return "", fmt.Errorf("地域参数不能为空")
	}

	info, err := t.vpcClient.DescribeVpcPeeringConnections(ctx, region)
	if err != nil {
		t.logger.WithError(err).Error("对等连接列表查询失败")
		return "", fmt.Errorf("查询地域 %s 的对等连接列表失败: %w", region, err)
	}

	format := ""
	if args.Format != nil {
		format = *args.Format
	}

	switch strings.ToLower(format) {
	case "json":
		return t.vpcClient.FormatVpcPeeringConnectionsAsJSON(info)
	case "table", "":
		return t.vpcClient.FormatVpcPeeringConnectionsAsTable(info), nil
	default:
		return "", fmt.Errorf("不支持的输出格式: %s，支持的格式: json, table", format)
	}
}