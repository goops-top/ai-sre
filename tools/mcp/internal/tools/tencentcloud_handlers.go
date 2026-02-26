package tools

import (
	"context"
	"encoding/json"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/pkg/logger"
)

var (
	// 全局腾讯云工具实例
	tencentCloudTools *TencentCloudTools
)

// InitTencentCloudTools 初始化腾讯云工具
func InitTencentCloudTools() error {
	var err error
	tencentCloudTools, err = NewTencentCloudTools()
	if err != nil {
		logger.GetLogger().WithError(err).Warn("腾讯云工具初始化失败，相关工具将不可用")
		return err
	}
	
	logger.GetLogger().Info("腾讯云工具初始化成功")
	return nil
}

// DescribeRegionsArgs 查询地域参数
type DescribeRegionsArgs struct {
	Product *string `json:"product,omitempty" jsonschema:"description=产品名称(如tke、cvm、cos等),default=cvm"`
	Format  *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// DescribeRegionsHandler 查询地域处理函数
func DescribeRegionsHandler(arguments DescribeRegionsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "describe_regions",
		"arguments": arguments,
	}).Debug("执行地域查询工具")
	
	// 检查腾讯云工具是否已初始化
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	// 默认产品为 cvm
	product := "cvm"
	if arguments.Product != nil && *arguments.Product != "" {
		product = *arguments.Product
	}
	
	// 调用腾讯云工具
	result, err := tencentCloudTools.DescribeRegions(ctx, DescribeRegionsArgs{
		Product: &product,
		Format:  arguments.Format,
	})
	if err != nil {
		logger.GetLogger().WithError(err).Error("地域查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("地域查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// GetRegionArgs 获取特定地域参数
type GetRegionArgs struct {
	RegionID *string `json:"region_id" jsonschema:"description=地域ID或地域名称,required"`
	Product  *string `json:"product,omitempty" jsonschema:"description=产品名称(如tke、cvm、cos等),default=cvm"`
	Format   *string `json:"format,omitempty" jsonschema:"description=输出格式: json或table,default=table"`
}

// GetRegionHandler 获取特定地域处理函数
func GetRegionHandler(arguments GetRegionArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "get_region",
		"arguments": arguments,
	}).Debug("执行特定地域查询工具")
	
	// 检查腾讯云工具是否已初始化
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	// 验证必需参数
	if arguments.RegionID == nil || *arguments.RegionID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region_id 不能为空")), nil
	}
	
	// 默认产品为 cvm
	product := "cvm"
	if arguments.Product != nil && *arguments.Product != "" {
		product = *arguments.Product
	}
	
	// 调用腾讯云工具
	result, err := tencentCloudTools.GetRegion(ctx, GetRegionArgs{
		RegionID: arguments.RegionID,
		Product:  &product,
		Format:   arguments.Format,
	})
	if err != nil {
		logger.GetLogger().WithError(err).Error("特定地域查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("特定地域查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// TencentCloudValidateArgs 腾讯云连接验证参数
type TencentCloudValidateArgs struct {
	// 暂无参数
}

// TencentCloudValidateHandler 腾讯云连接验证处理函数
func TencentCloudValidateHandler(arguments TencentCloudValidateArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tencentcloud_validate",
		"arguments": arguments,
	}).Debug("执行腾讯云连接验证工具")
	
	// 检查腾讯云工具是否已初始化
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	// 验证连接
	err := tencentCloudTools.ValidateConnection(ctx)
	if err != nil {
		logger.GetLogger().WithError(err).Error("腾讯云连接验证失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("腾讯云连接验证失败: %v", err))), nil
	}
	
	result := map[string]interface{}{
		"status":  "success",
		"message": "腾讯云连接验证成功",
		"services": []string{"TKE"},
	}
	
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("格式化验证结果失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(string(resultJSON))), nil
}

// GetClusterLevelPriceHandler 获取集群等级价格处理函数
func GetClusterLevelPriceHandler(arguments GetClusterLevelPriceArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_get_cluster_level_price",
		"arguments": arguments,
	}).Debug("执行集群等级价格查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterLevel == nil || *arguments.ClusterLevel == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_level 不能为空")), nil
	}
	
	result, err := tencentCloudTools.GetClusterLevelPrice(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群等级价格查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群等级价格查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeAddonHandler 查询集群已安装 addon 列表处理函数
func DescribeAddonHandler(arguments DescribeAddonArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_addon",
		"arguments": arguments,
	}).Debug("执行集群 addon 列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeAddon(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群 addon 列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群 addon 列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// GetTkeAppChartListHandler 获取可安装 addon 列表处理函数
func GetTkeAppChartListHandler(arguments GetTkeAppChartListArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_get_app_chart_list",
		"arguments": arguments,
	}).Debug("执行可安装 addon 列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	
	result, err := tencentCloudTools.GetTkeAppChartList(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("可安装 addon 列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("可安装 addon 列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeImagesHandler 查询 OS 镜像列表处理函数
func DescribeImagesHandler(arguments DescribeImagesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_images",
		"arguments": arguments,
	}).Debug("执行 OS 镜像列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeImages(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("OS 镜像列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("OS 镜像列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeVersionsHandler 查询集群版本列表处理函数
func DescribeVersionsHandler(arguments DescribeVersionsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_versions",
		"arguments": arguments,
	}).Debug("执行集群版本列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeVersions(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群版本列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群版本列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeLogSwitchesHandler 查询集群日志开关处理函数
func DescribeLogSwitchesHandler(arguments DescribeLogSwitchesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_log_switches",
		"arguments": arguments,
	}).Debug("执行集群日志开关查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeLogSwitches(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群日志开关查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群日志开关查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeMasterComponentHandler 查询 master 组件状态处理函数
func DescribeMasterComponentHandler(arguments DescribeMasterComponentArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_master_component",
		"arguments": arguments,
	}).Debug("执行 master 组件状态查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeMasterComponent(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("master 组件状态查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("master 组件状态查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeClusterInstancesHandler 查询集群节点实例列表处理函数
func DescribeClusterInstancesHandler(arguments DescribeClusterInstancesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_cluster_instances",
		"arguments": arguments,
	}).Debug("执行集群节点实例列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeClusterInstances(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群节点实例列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群节点实例列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeClusterVirtualNodeHandler 查询集群超级节点列表处理函数
func DescribeClusterVirtualNodeHandler(arguments DescribeClusterVirtualNodeArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_cluster_virtual_node",
		"arguments": arguments,
	}).Debug("执行集群超级节点列表查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeClusterVirtualNode(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群超级节点列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群超级节点列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeClusterExtraArgsHandler 查询集群自定义参数处理函数
func DescribeClusterExtraArgsHandler(arguments DescribeClusterExtraArgsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_cluster_extra_args",
		"arguments": arguments,
	}).Debug("执行集群自定义参数查询工具")
	
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.ClusterID == nil || *arguments.ClusterID == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 cluster_id 不能为空")), nil
	}
	
	result, err := tencentCloudTools.DescribeClusterExtraArgs(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("集群自定义参数查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("集群自定义参数查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ========== CVM Handlers ==========

// CvmDescribeInstancesHandler 查询 CVM 实例列表处理函数
func CvmDescribeInstancesHandler(arguments CvmDescribeInstancesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cvm_describe_instances",
		"arguments": arguments,
	}).Debug("执行 CVM 实例列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.CvmDescribeInstances(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CVM 实例列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CVM 实例列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// CvmDescribeInstancesStatusHandler 查询 CVM 实例状态处理函数
func CvmDescribeInstancesStatusHandler(arguments CvmDescribeInstancesStatusArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cvm_describe_instances_status",
		"arguments": arguments,
	}).Debug("执行 CVM 实例状态查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.CvmDescribeInstancesStatus(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CVM 实例状态查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CVM 实例状态查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ========== CLB Handlers ==========

// ClbDescribeLoadBalancersHandler 查询 CLB 实例列表处理函数
func ClbDescribeLoadBalancersHandler(arguments ClbDescribeLoadBalancersArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "clb_describe_load_balancers",
		"arguments": arguments,
	}).Debug("执行 CLB 实例列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.ClbDescribeLoadBalancers(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CLB 实例列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CLB 实例列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ClbDescribeListenersHandler 查询 CLB 监听器列表处理函数
func ClbDescribeListenersHandler(arguments ClbDescribeListenersArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "clb_describe_listeners",
		"arguments": arguments,
	}).Debug("执行 CLB 监听器列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.LoadBalancerId == nil || *arguments.LoadBalancerId == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 load_balancer_id 不能为空")), nil
	}

	result, err := tencentCloudTools.ClbDescribeListeners(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CLB 监听器列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CLB 监听器列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ClbDescribeTargetsHandler 查询 CLB 后端服务列表处理函数
func ClbDescribeTargetsHandler(arguments ClbDescribeTargetsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "clb_describe_targets",
		"arguments": arguments,
	}).Debug("执行 CLB 后端服务列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.LoadBalancerId == nil || *arguments.LoadBalancerId == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 load_balancer_id 不能为空")), nil
	}

	result, err := tencentCloudTools.ClbDescribeTargets(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CLB 后端服务列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CLB 后端服务列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ClbDescribeTargetHealthHandler 查询 CLB 后端健康状态处理函数
func ClbDescribeTargetHealthHandler(arguments ClbDescribeTargetHealthArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "clb_describe_target_health",
		"arguments": arguments,
	}).Debug("执行 CLB 后端健康状态查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.LoadBalancerIds == nil || *arguments.LoadBalancerIds == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 load_balancer_ids 不能为空")), nil
	}

	result, err := tencentCloudTools.ClbDescribeTargetHealth(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CLB 后端健康状态查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CLB 后端健康状态查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ========== CDB Handlers ==========

// CdbDescribeDBInstancesHandler 查询 CDB 实例列表处理函数
func CdbDescribeDBInstancesHandler(arguments CdbDescribeDBInstancesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cdb_describe_db_instances",
		"arguments": arguments,
	}).Debug("执行 CDB 实例列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.CdbDescribeDBInstances(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CDB 实例列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CDB 实例列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// CdbDescribeDBInstanceInfoHandler 查询 CDB 实例详细信息处理函数
func CdbDescribeDBInstanceInfoHandler(arguments CdbDescribeDBInstanceInfoArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cdb_describe_db_instance_info",
		"arguments": arguments,
	}).Debug("执行 CDB 实例详细信息查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.InstanceId == nil || *arguments.InstanceId == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 instance_id 不能为空")), nil
	}

	result, err := tencentCloudTools.CdbDescribeDBInstanceInfo(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CDB 实例详细信息查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CDB 实例详细信息查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// CdbDescribeSlowLogsHandler 查询 CDB 慢日志处理函数
func CdbDescribeSlowLogsHandler(arguments CdbDescribeSlowLogsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cdb_describe_slow_logs",
		"arguments": arguments,
	}).Debug("执行 CDB 慢日志查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.InstanceId == nil || *arguments.InstanceId == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 instance_id 不能为空")), nil
	}

	result, err := tencentCloudTools.CdbDescribeSlowLogs(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CDB 慢日志查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CDB 慢日志查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// CdbDescribeErrorLogHandler 查询 CDB 错误日志处理函数
func CdbDescribeErrorLogHandler(arguments CdbDescribeErrorLogArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "cdb_describe_error_log",
		"arguments": arguments,
	}).Debug("执行 CDB 错误日志查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	if arguments.InstanceId == nil || *arguments.InstanceId == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 instance_id 不能为空")), nil
	}

	result, err := tencentCloudTools.CdbDescribeErrorLog(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("CDB 错误日志查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("CDB 错误日志查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// ========== VPC Handlers ==========

// VpcDescribeVpcsHandler 查询 VPC 列表处理函数
func VpcDescribeVpcsHandler(arguments VpcDescribeVpcsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_vpcs",
		"arguments": arguments,
	}).Debug("执行 VPC 列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeVpcs(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("VPC 列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("VPC 列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeSubnetsHandler 查询子网列表处理函数
func VpcDescribeSubnetsHandler(arguments VpcDescribeSubnetsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_subnets",
		"arguments": arguments,
	}).Debug("执行子网列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeSubnets(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("子网列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("子网列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeSecurityGroupsHandler 查询安全组列表处理函数
func VpcDescribeSecurityGroupsHandler(arguments VpcDescribeSecurityGroupsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_security_groups",
		"arguments": arguments,
	}).Debug("执行安全组列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeSecurityGroups(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("安全组列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("安全组列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeNetworkInterfacesHandler 查询弹性网卡列表处理函数
func VpcDescribeNetworkInterfacesHandler(arguments VpcDescribeNetworkInterfacesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_network_interfaces",
		"arguments": arguments,
	}).Debug("执行弹性网卡列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeNetworkInterfaces(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("弹性网卡列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("弹性网卡列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeAddressesHandler 查询弹性公网IP列表处理函数
func VpcDescribeAddressesHandler(arguments VpcDescribeAddressesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_addresses",
		"arguments": arguments,
	}).Debug("执行弹性公网IP列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeAddresses(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("弹性公网IP列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("弹性公网IP列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeBandwidthPackagesHandler 查询带宽包列表处理函数
func VpcDescribeBandwidthPackagesHandler(arguments VpcDescribeBandwidthPackagesArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_bandwidth_packages",
		"arguments": arguments,
	}).Debug("执行带宽包列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeBandwidthPackages(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("带宽包列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("带宽包列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeVpcEndPointHandler 查询终端节点列表处理函数
func VpcDescribeVpcEndPointHandler(arguments VpcDescribeVpcEndPointArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_vpc_endpoint",
		"arguments": arguments,
	}).Debug("执行终端节点列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeVpcEndPoint(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("终端节点列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("终端节点列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeVpcEndPointServiceHandler 查询终端节点服务列表处理函数
func VpcDescribeVpcEndPointServiceHandler(arguments VpcDescribeVpcEndPointServiceArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_vpc_endpoint_service",
		"arguments": arguments,
	}).Debug("执行终端节点服务列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeVpcEndPointService(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("终端节点服务列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("终端节点服务列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// VpcDescribeVpcPeeringConnectionsHandler 查询对等连接列表处理函数
func VpcDescribeVpcPeeringConnectionsHandler(arguments VpcDescribeVpcPeeringConnectionsArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()

	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "vpc_describe_vpc_peering_connections",
		"arguments": arguments,
	}).Debug("执行对等连接列表查询工具")

	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}

	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}

	result, err := tencentCloudTools.VpcDescribeVpcPeeringConnections(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("对等连接列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("对等连接列表查询失败: %v", err))), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}

// DescribeClustersHandler TKE 集群列表查询处理函数
func DescribeClustersHandler(arguments DescribeClustersArgs) (*mcp.ToolResponse, error) {
	ctx := context.Background()
	
	logger.GetLogger().WithFields(logrus.Fields{
		"handler":   "tke_describe_clusters",
		"arguments": arguments,
	}).Debug("执行 TKE 集群列表查询工具")
	
	// 检查腾讯云工具是否已初始化
	if tencentCloudTools == nil {
		return mcp.NewToolResponse(mcp.NewTextContent("腾讯云工具未初始化，请检查配置")), nil
	}
	
	// 验证必需参数
	if arguments.Region == nil || *arguments.Region == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("参数 region 不能为空")), nil
	}
	
	// 调用腾讯云工具
	result, err := tencentCloudTools.DescribeClusters(ctx, arguments)
	if err != nil {
		logger.GetLogger().WithError(err).Error("TKE 集群列表查询失败")
		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("TKE 集群列表查询失败: %v", err))), nil
	}
	
	return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
}