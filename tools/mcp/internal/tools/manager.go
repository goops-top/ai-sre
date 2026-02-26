package tools

import (
	"fmt"
	"sync"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/sirupsen/logrus"
	"ai-sre/tools/mcp/pkg/logger"
)

// MCPServerInterface 定义MCP服务器接口，避免循环依赖
type MCPServerInterface interface {
	RegisterTool(name, description string, handler interface{}) error
}

// ToolManager 工具管理器
type ToolManager struct {
	server MCPServerInterface
	tools  map[string]interface{}
	mutex  sync.RWMutex
}

// NewToolManager 创建新的工具管理器
func NewToolManager(server MCPServerInterface) *ToolManager {
	return &ToolManager{
		server: server,
		tools:  make(map[string]interface{}),
	}
}

// RegisterDefaultTools 注册默认工具
func (tm *ToolManager) RegisterDefaultTools() error {
	logger.Info("Registering default tools")
	
	// 注册ping工具
	if err := tm.server.RegisterTool(
		"ping",
		"简单的ping工具，用于测试MCP服务器连接和响应。返回指定的消息或默认的'pong'响应。",
		PingHandler,
	); err != nil {
		return fmt.Errorf("failed to register ping tool: %w", err)
	}
	
	// 同时注册到全局注册表（HTTP模式使用）
	GetGlobalRegistry().RegisterHandler("ping", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(PingArguments); ok {
			return PingHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args PingArguments
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return PingHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册echo工具
	if err := tm.server.RegisterTool(
		"echo",
		"高级文本处理和格式化工具，支持大小写转换、前缀后缀添加、文本重复等功能。",
		EchoHandler,
	); err != nil {
		return fmt.Errorf("failed to register echo tool: %w", err)
	}
	
	// 同时注册到全局注册表（HTTP模式使用）
	GetGlobalRegistry().RegisterHandler("echo", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(EchoArguments); ok {
			return EchoHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args EchoArguments
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return EchoHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册system_info工具
	if err := tm.server.RegisterTool(
		"system_info",
		"获取系统运行时信息，包括Go运行时、内存使用、环境变量、进程信息等。",
		SystemInfoHandler,
	); err != nil {
		return fmt.Errorf("failed to register system_info tool: %w", err)
	}
	
	// 同时注册到全局注册表（HTTP模式使用）
	GetGlobalRegistry().RegisterHandler("system_info", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(SystemInfoArguments); ok {
			return SystemInfoHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args SystemInfoArguments
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return SystemInfoHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 尝试初始化和注册腾讯云工具
	if err := tm.RegisterTencentCloudTools(); err != nil {
		logger.WithError(err).Warn("腾讯云工具注册失败，相关工具将不可用")
		// 不返回错误，继续运行其他工具
	}
	
	logger.WithFields(logrus.Fields{
		"tool_count": 3,
		"tools":      []string{"ping", "echo", "system_info"},
	}).Info("Default tools registered successfully")
	
	return nil
}

// RegisterTencentCloudTools 注册腾讯云工具
func (tm *ToolManager) RegisterTencentCloudTools() error {
	logger.Info("Registering Tencent Cloud tools")
	
	// 初始化腾讯云工具
	if err := InitTencentCloudTools(); err != nil {
		return fmt.Errorf("腾讯云工具初始化失败: %w", err)
	}
	
	// 注册地域查询工具
	if err := tm.server.RegisterTool(
		"describe_regions",
		"查询腾讯云产品支持的地域信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
		DescribeRegionsHandler,
	); err != nil {
		return fmt.Errorf("failed to register describe_regions tool: %w", err)
	}
	
	// 同时注册到全局注册表
	GetGlobalRegistry().RegisterHandler("describe_regions", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeRegionsArgs); ok {
			return DescribeRegionsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeRegionsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeRegionsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册特定地域查询工具
	if err := tm.server.RegisterTool(
		"get_region",
		"根据地域ID查询腾讯云产品特定地域的详细信息。支持多种产品(如tke、cvm、cos等)，支持 JSON 和表格两种输出格式。",
		GetRegionHandler,
	); err != nil {
		return fmt.Errorf("failed to register get_region tool: %w", err)
	}
	
	// 同时注册到全局注册表
	GetGlobalRegistry().RegisterHandler("get_region", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(GetRegionArgs); ok {
			return GetRegionHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args GetRegionArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return GetRegionHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册腾讯云连接验证工具
	if err := tm.server.RegisterTool(
		"tencentcloud_validate",
		"验证腾讯云 API 连接和权限配置。检查 SecretID、SecretKey 是否正确以及相关服务权限。",
		TencentCloudValidateHandler,
	); err != nil {
		return fmt.Errorf("failed to register tencentcloud_validate tool: %w", err)
	}
	
	// 同时注册到全局注册表
	GetGlobalRegistry().RegisterHandler("tencentcloud_validate", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(TencentCloudValidateArgs); ok {
			return TencentCloudValidateHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args TencentCloudValidateArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return TencentCloudValidateHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群列表查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_clusters",
		"查询指定地域的 TKE 集群列表。支持按集群类型过滤：all(全部)、tke(普通集群)、serverless(弹性集群)。默认查询全部集群。",
		DescribeClustersHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_clusters tool: %w", err)
	}
	
	// 同时注册到全局注册表
	GetGlobalRegistry().RegisterHandler("tke_describe_clusters", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeClustersArgs); ok {
			return DescribeClustersHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeClustersArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeClustersHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群自定义参数查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_cluster_extra_args",
		"查询指定地域下指定 TKE 集群的自定义参数(Etcd、KubeAPIServer、KubeControllerManager、KubeScheduler)。",
		DescribeClusterExtraArgsHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_cluster_extra_args tool: %w", err)
	}
	
	// 同时注册到全局注册表
	GetGlobalRegistry().RegisterHandler("tke_describe_cluster_extra_args", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeClusterExtraArgsArgs); ok {
			return DescribeClusterExtraArgsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeClusterExtraArgsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeClusterExtraArgsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群等级价格查询工具
	if err := tm.server.RegisterTool(
		"tke_get_cluster_level_price",
		"获取指定地域下指定集群等级的价格信息。集群等级可选：L20、L50、L100、L200、L500、L1000、L3000、L5000。",
		GetClusterLevelPriceHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_get_cluster_level_price tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_get_cluster_level_price", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(GetClusterLevelPriceArgs); ok {
			return GetClusterLevelPriceHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args GetClusterLevelPriceArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return GetClusterLevelPriceHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群 addon 查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_addon",
		"查询指定地域下指定 TKE 集群已安装的 addon 列表。可选指定 addon 名称查询特定 addon。",
		DescribeAddonHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_addon tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_addon", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeAddonArgs); ok {
			return DescribeAddonHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeAddonArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeAddonHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 可安装 addon 列表查询工具
	if err := tm.server.RegisterTool(
		"tke_get_app_chart_list",
		"获取指定地域可安装的 TKE addon 列表。支持按类型(kind)、架构(arch)、集群类型(cluster_type)过滤。",
		GetTkeAppChartListHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_get_app_chart_list tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_get_app_chart_list", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(GetTkeAppChartListArgs); ok {
			return GetTkeAppChartListHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args GetTkeAppChartListArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return GetTkeAppChartListHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE OS 镜像列表查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_images",
		"获取指定地域支持的 TKE 节点 OS 镜像列表。",
		DescribeImagesHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_images tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_images", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeImagesArgs); ok {
			return DescribeImagesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeImagesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeImagesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群版本列表查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_versions",
		"获取指定地域支持的 TKE 集群 Kubernetes 版本列表。",
		DescribeVersionsHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_versions tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_versions", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeVersionsArgs); ok {
			return DescribeVersionsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeVersionsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeVersionsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群日志开关查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_log_switches",
		"查询指定地域下指定 TKE 集群的日志采集开关状态，包括审计日志、事件日志、普通日志和 Master 日志。",
		DescribeLogSwitchesHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_log_switches tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_log_switches", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeLogSwitchesArgs); ok {
			return DescribeLogSwitchesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeLogSwitchesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeLogSwitchesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE master 组件状态查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_master_component",
		"查询指定地域下指定 TKE 集群的 master 组件运行状态。支持 kube-apiserver、kube-scheduler、kube-controller-manager，默认查询 kube-apiserver。",
		DescribeMasterComponentHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_master_component tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_master_component", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeMasterComponentArgs); ok {
			return DescribeMasterComponentHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeMasterComponentArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeMasterComponentHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群节点实例列表查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_cluster_instances",
		"查询指定地域下指定 TKE 集群的节点实例列表，包含节点IP、角色、状态、封锁状态、节点池等信息。支持按节点角色过滤。",
		DescribeClusterInstancesHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_cluster_instances tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_cluster_instances", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeClusterInstancesArgs); ok {
			return DescribeClusterInstancesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeClusterInstancesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeClusterInstancesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 TKE 集群超级节点列表查询工具
	if err := tm.server.RegisterTool(
		"tke_describe_cluster_virtual_node",
		"查询指定地域下指定 TKE 集群的超级节点列表。可选指定节点池ID过滤。",
		DescribeClusterVirtualNodeHandler,
	); err != nil {
		return fmt.Errorf("failed to register tke_describe_cluster_virtual_node tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("tke_describe_cluster_virtual_node", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(DescribeClusterVirtualNodeArgs); ok {
			return DescribeClusterVirtualNodeHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args DescribeClusterVirtualNodeArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return DescribeClusterVirtualNodeHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// ========== CVM 工具注册 ==========
	
	// 注册 CVM 实例列表查询工具
	if err := tm.server.RegisterTool(
		"cvm_describe_instances",
		"查询指定地域的 CVM 实例列表。支持按实例ID、实例名称、可用区、项目ID等过滤。返回实例的基本信息、网络配置、磁盘信息等。",
		CvmDescribeInstancesHandler,
	); err != nil {
		return fmt.Errorf("failed to register cvm_describe_instances tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cvm_describe_instances", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CvmDescribeInstancesArgs); ok {
			return CvmDescribeInstancesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CvmDescribeInstancesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CvmDescribeInstancesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CVM 实例状态查询工具
	if err := tm.server.RegisterTool(
		"cvm_describe_instances_status",
		"查询指定地域的 CVM 实例状态列表。返回实例ID和对应的运行状态(RUNNING/STOPPED/PENDING等)。",
		CvmDescribeInstancesStatusHandler,
	); err != nil {
		return fmt.Errorf("failed to register cvm_describe_instances_status tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cvm_describe_instances_status", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CvmDescribeInstancesStatusArgs); ok {
			return CvmDescribeInstancesStatusHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CvmDescribeInstancesStatusArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CvmDescribeInstancesStatusHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// ========== CLB 工具注册 ==========
	
	// 注册 CLB 负载均衡实例列表查询工具
	if err := tm.server.RegisterTool(
		"clb_describe_load_balancers",
		"查询指定地域的 CLB 负载均衡实例列表。支持按实例ID、名称、类型(OPEN/INTERNAL)、VIP等过滤。",
		ClbDescribeLoadBalancersHandler,
	); err != nil {
		return fmt.Errorf("failed to register clb_describe_load_balancers tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("clb_describe_load_balancers", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(ClbDescribeLoadBalancersArgs); ok {
			return ClbDescribeLoadBalancersHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args ClbDescribeLoadBalancersArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return ClbDescribeLoadBalancersHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CLB 监听器列表查询工具
	if err := tm.server.RegisterTool(
		"clb_describe_listeners",
		"查询指定地域下指定 CLB 实例的监听器列表。返回监听器的协议、端口、健康检查配置等信息。",
		ClbDescribeListenersHandler,
	); err != nil {
		return fmt.Errorf("failed to register clb_describe_listeners tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("clb_describe_listeners", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(ClbDescribeListenersArgs); ok {
			return ClbDescribeListenersHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args ClbDescribeListenersArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return ClbDescribeListenersHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CLB 后端目标列表查询工具
	if err := tm.server.RegisterTool(
		"clb_describe_targets",
		"查询指定地域下指定 CLB 实例绑定的后端目标(RS)列表。可选指定监听器ID过滤。",
		ClbDescribeTargetsHandler,
	); err != nil {
		return fmt.Errorf("failed to register clb_describe_targets tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("clb_describe_targets", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(ClbDescribeTargetsArgs); ok {
			return ClbDescribeTargetsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args ClbDescribeTargetsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return ClbDescribeTargetsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CLB 后端目标健康状态查询工具
	if err := tm.server.RegisterTool(
		"clb_describe_target_health",
		"查询指定地域下指定 CLB 实例后端目标的健康检查状态。支持查询多个 CLB 实例(逗号分隔)。",
		ClbDescribeTargetHealthHandler,
	); err != nil {
		return fmt.Errorf("failed to register clb_describe_target_health tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("clb_describe_target_health", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(ClbDescribeTargetHealthArgs); ok {
			return ClbDescribeTargetHealthHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args ClbDescribeTargetHealthArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return ClbDescribeTargetHealthHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// ========== CDB 工具注册 ==========
	
	// 注册 CDB 实例列表查询工具
	if err := tm.server.RegisterTool(
		"cdb_describe_db_instances",
		"查询指定地域的 CDB (MySQL) 实例列表。支持按实例ID、实例名称、状态等过滤。返回实例基本信息、配置、网络等。",
		CdbDescribeDBInstancesHandler,
	); err != nil {
		return fmt.Errorf("failed to register cdb_describe_db_instances tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cdb_describe_db_instances", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CdbDescribeDBInstancesArgs); ok {
			return CdbDescribeDBInstancesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CdbDescribeDBInstancesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CdbDescribeDBInstancesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CDB 实例详情查询工具
	if err := tm.server.RegisterTool(
		"cdb_describe_db_instance_info",
		"查询指定地域下指定 CDB (MySQL) 实例的详细信息，包括实例配置、网络信息、参数等。",
		CdbDescribeDBInstanceInfoHandler,
	); err != nil {
		return fmt.Errorf("failed to register cdb_describe_db_instance_info tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cdb_describe_db_instance_info", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CdbDescribeDBInstanceInfoArgs); ok {
			return CdbDescribeDBInstanceInfoHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CdbDescribeDBInstanceInfoArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CdbDescribeDBInstanceInfoHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CDB 慢查询日志查询工具
	if err := tm.server.RegisterTool(
		"cdb_describe_slow_logs",
		"查询指定地域下指定 CDB (MySQL) 实例的慢查询日志文件列表。返回慢日志文件名、大小、时间等信息。",
		CdbDescribeSlowLogsHandler,
	); err != nil {
		return fmt.Errorf("failed to register cdb_describe_slow_logs tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cdb_describe_slow_logs", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CdbDescribeSlowLogsArgs); ok {
			return CdbDescribeSlowLogsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CdbDescribeSlowLogsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CdbDescribeSlowLogsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册 CDB 错误日志查询工具
	if err := tm.server.RegisterTool(
		"cdb_describe_error_log",
		"查询指定地域下指定 CDB (MySQL) 实例的错误日志数据。支持按时间范围和关键字过滤。默认查询最近1小时。",
		CdbDescribeErrorLogHandler,
	); err != nil {
		return fmt.Errorf("failed to register cdb_describe_error_log tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("cdb_describe_error_log", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(CdbDescribeErrorLogArgs); ok {
			return CdbDescribeErrorLogHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args CdbDescribeErrorLogArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return CdbDescribeErrorLogHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// ========== VPC 工具注册 ==========
	
	// 注册 VPC 列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_vpcs",
		"查询指定地域的 VPC 列表。返回 VPC ID、名称、CIDR、是否默认、DHCP、DNS 等信息。",
		VpcDescribeVpcsHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_vpcs tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_vpcs", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeVpcsArgs); ok {
			return VpcDescribeVpcsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeVpcsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeVpcsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册子网列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_subnets",
		"查询指定地域的子网列表。支持按 VPC ID 过滤。返回子网ID、CIDR、可用区、可用IP数等信息。",
		VpcDescribeSubnetsHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_subnets tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_subnets", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeSubnetsArgs); ok {
			return VpcDescribeSubnetsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeSubnetsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeSubnetsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册安全组列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_security_groups",
		"查询指定地域的安全组列表。返回安全组ID、名称、描述、是否默认等信息。",
		VpcDescribeSecurityGroupsHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_security_groups tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_security_groups", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeSecurityGroupsArgs); ok {
			return VpcDescribeSecurityGroupsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeSecurityGroupsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeSecurityGroupsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册弹性网卡列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_network_interfaces",
		"查询指定地域的弹性网卡(ENI)列表。支持按 VPC ID 过滤。返回网卡ID、MAC、状态、内网IP等信息。",
		VpcDescribeNetworkInterfacesHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_network_interfaces tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_network_interfaces", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeNetworkInterfacesArgs); ok {
			return VpcDescribeNetworkInterfacesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeNetworkInterfacesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeNetworkInterfacesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册弹性公网IP列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_addresses",
		"查询指定地域的弹性公网IP(EIP)列表。返回 EIP ID、公网IP、状态、绑定实例、带宽等信息。",
		VpcDescribeAddressesHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_addresses tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_addresses", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeAddressesArgs); ok {
			return VpcDescribeAddressesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeAddressesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeAddressesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册带宽包列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_bandwidth_packages",
		"查询指定地域的带宽包列表。返回带宽包ID、名称、网络类型、计费类型、带宽、状态等信息。",
		VpcDescribeBandwidthPackagesHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_bandwidth_packages tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_bandwidth_packages", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeBandwidthPackagesArgs); ok {
			return VpcDescribeBandwidthPackagesHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeBandwidthPackagesArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeBandwidthPackagesHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册终端节点列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_vpc_endpoint",
		"查询指定地域的终端节点列表。返回终端节点ID、名称、VPC、VIP、服务ID、状态等信息。",
		VpcDescribeVpcEndPointHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_vpc_endpoint tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_vpc_endpoint", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeVpcEndPointArgs); ok {
			return VpcDescribeVpcEndPointHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeVpcEndPointArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeVpcEndPointHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册终端节点服务列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_vpc_endpoint_service",
		"查询指定地域的终端节点服务列表。返回服务ID、名称、VPC、VIP、服务类型、终端节点数等信息。",
		VpcDescribeVpcEndPointServiceHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_vpc_endpoint_service tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_vpc_endpoint_service", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeVpcEndPointServiceArgs); ok {
			return VpcDescribeVpcEndPointServiceHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeVpcEndPointServiceArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeVpcEndPointServiceHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	// 注册对等连接列表查询工具
	if err := tm.server.RegisterTool(
		"vpc_describe_vpc_peering_connections",
		"查询指定地域的对等连接列表。返回对等连接ID、名称、本端/对端VPC、地域、状态、带宽等信息。",
		VpcDescribeVpcPeeringConnectionsHandler,
	); err != nil {
		return fmt.Errorf("failed to register vpc_describe_vpc_peering_connections tool: %w", err)
	}
	
	GetGlobalRegistry().RegisterHandler("vpc_describe_vpc_peering_connections", func(arguments interface{}) (*mcp.ToolResponse, error) {
		if args, ok := arguments.(VpcDescribeVpcPeeringConnectionsArgs); ok {
			return VpcDescribeVpcPeeringConnectionsHandler(args)
		} else if argsMap, ok := arguments.(map[string]interface{}); ok {
			var args VpcDescribeVpcPeeringConnectionsArgs
			if err := ConvertArgumentsToStruct(argsMap, &args); err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("参数转换失败: %v", err))), nil
			}
			return VpcDescribeVpcPeeringConnectionsHandler(args)
		}
		return mcp.NewToolResponse(mcp.NewTextContent("无效的参数类型")), nil
	})
	
	logger.WithFields(logrus.Fields{
		"tool_count": 33,
		"tools":      []string{"describe_regions", "get_region", "tencentcloud_validate", "tke_describe_clusters", "tke_describe_cluster_extra_args", "tke_get_cluster_level_price", "tke_describe_addon", "tke_get_app_chart_list", "tke_describe_images", "tke_describe_versions", "tke_describe_log_switches", "tke_describe_master_component", "tke_describe_cluster_instances", "tke_describe_cluster_virtual_node", "cvm_describe_instances", "cvm_describe_instances_status", "clb_describe_load_balancers", "clb_describe_listeners", "clb_describe_targets", "clb_describe_target_health", "cdb_describe_db_instances", "cdb_describe_db_instance_info", "cdb_describe_slow_logs", "cdb_describe_error_log", "vpc_describe_vpcs", "vpc_describe_subnets", "vpc_describe_security_groups", "vpc_describe_network_interfaces", "vpc_describe_addresses", "vpc_describe_bandwidth_packages", "vpc_describe_vpc_endpoint", "vpc_describe_vpc_endpoint_service", "vpc_describe_vpc_peering_connections"},
	}).Info("Tencent Cloud tools registered successfully")
	
	return nil
}

// GetRegisteredTools 获取已注册的工具列表
func (tm *ToolManager) GetRegisteredTools() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	tools := make([]string, 0, len(tm.tools))
	for name := range tm.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetToolCount 获取工具数量
func (tm *ToolManager) GetToolCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.tools)
}