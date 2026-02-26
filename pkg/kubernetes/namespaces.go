package kubernetes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceInfo 命名空间信息
type NamespaceInfo struct {
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	Age         string            `json:"age"`
}

// ListNamespacesOptions 获取命名空间列表的选项
type ListNamespacesOptions struct {
	// LabelSelector 标签选择器，如 "env=production,team=backend"
	LabelSelector string
}

// ListNamespaces 获取 Kubernetes 集群的命名空间列表
func (c *Client) ListNamespaces(ctx context.Context, opts *ListNamespacesOptions) ([]NamespaceInfo, error) {
	listOpts := metav1.ListOptions{}
	if opts != nil && opts.LabelSelector != "" {
		listOpts.LabelSelector = opts.LabelSelector
	}

	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}

	result := make([]NamespaceInfo, 0, len(nsList.Items))
	now := time.Now()
	for _, ns := range nsList.Items {
		result = append(result, NamespaceInfo{
			Name:        ns.Name,
			Status:      string(ns.Status.Phase),
			Labels:      ns.Labels,
			Annotations: filterSystemAnnotations(ns.Annotations),
			CreatedAt:   ns.CreationTimestamp.Time,
			Age:         formatAge(now.Sub(ns.CreationTimestamp.Time)),
		})
	}

	// 按名称排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// FormatNamespaceList 格式化命名空间列表为可读字符串
func FormatNamespaceList(namespaces []NamespaceInfo) string {
	if len(namespaces) == 0 {
		return "未找到命名空间"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("命名空间列表 (共 %d 个):\n", len(namespaces)))
	sb.WriteString(fmt.Sprintf("%-30s %-10s %-15s\n", "NAME", "STATUS", "AGE"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, ns := range namespaces {
		sb.WriteString(fmt.Sprintf("%-30s %-10s %-15s\n", ns.Name, ns.Status, ns.Age))
	}

	return sb.String()
}

// filterSystemAnnotations 过滤掉系统注解，只保留用户自定义的
func filterSystemAnnotations(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return nil
	}
	filtered := make(map[string]string)
	for k, v := range annotations {
		// 过滤掉 kubernetes.io 和 kubectl.kubernetes.io 的系统注解
		if !strings.Contains(k, "kubernetes.io/") && !strings.Contains(k, "kubectl.kubernetes.io/") {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

// formatAge 将 Duration 格式化为可读字符串（如 5d, 2h30m, 45s）
func formatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
