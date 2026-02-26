package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client Kubernetes 客户端封装
type Client struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	logger        *logrus.Logger
	kubeconfig    string
}

// NewClient 通过 kubeconfig 文件路径创建 Kubernetes 客户端
// kubeconfigPath 为空时，按以下顺序尝试：
// 1. KUBECONFIG 环境变量
// 2. ~/.kube/config 默认路径
// 3. in-cluster 配置（Pod 内运行时）
func NewClient(kubeconfigPath string, logger *logrus.Logger) (*Client, error) {
	if logger == nil {
		logger = logrus.New()
	}

	config, resolvedPath, err := buildConfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("构建 Kubernetes 配置失败: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes clientset 失败: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes dynamic client 失败: %w", err)
	}

	logger.WithField("kubeconfig", resolvedPath).Info("Kubernetes 客户端初始化成功")

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		logger:        logger,
		kubeconfig:    resolvedPath,
	}, nil
}

// NewClientFromRESTConfig 通过已有的 rest.Config 创建客户端（用于测试或自定义场景）
func NewClientFromRESTConfig(config *rest.Config, logger *logrus.Logger) (*Client, error) {
	if logger == nil {
		logger = logrus.New()
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes clientset 失败: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes dynamic client 失败: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		logger:        logger,
		kubeconfig:    "rest-config",
	}, nil
}

// GetClientset 获取原始 kubernetes.Interface，供外部直接使用
func (c *Client) GetClientset() kubernetes.Interface {
	return c.clientset
}

// GetDynamicClient 获取 dynamic.Interface，供查询 CRD 使用
func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// buildConfig 按优先级构建 rest.Config
func buildConfig(kubeconfigPath string) (*rest.Config, string, error) {
	// 1. 如果明确指定了 kubeconfig 路径
	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, "", fmt.Errorf("加载 kubeconfig 文件 %s 失败: %w", kubeconfigPath, err)
		}
		return config, kubeconfigPath, nil
	}

	// 2. 从 KUBECONFIG 环境变量
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", envPath)
		if err != nil {
			return nil, "", fmt.Errorf("加载 KUBECONFIG 环境变量指定的文件 %s 失败: %w", envPath, err)
		}
		return config, envPath, nil
	}

	// 3. 默认 ~/.kube/config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		defaultPath := filepath.Join(homeDir, ".kube", "config")
		if _, statErr := os.Stat(defaultPath); statErr == nil {
			config, err := clientcmd.BuildConfigFromFlags("", defaultPath)
			if err == nil {
				return config, defaultPath, nil
			}
		}
	}

	// 4. in-cluster 配置
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, "", fmt.Errorf("无法找到有效的 Kubernetes 配置：未指定 kubeconfig 路径，KUBECONFIG 环境变量未设置，~/.kube/config 不存在，且不在集群内运行")
	}
	return config, "in-cluster", nil
}
