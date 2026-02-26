package kubernetes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodStatusInfo Pod 状态信息
type PodStatusInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Phase          string            `json:"phase"`
	Status         string            `json:"status"`
	Ready          string            `json:"ready"`
	Restarts       int32             `json:"restarts"`
	Node           string            `json:"node"`
	IP             string            `json:"ip"`
	Labels         map[string]string `json:"labels,omitempty"`
	Containers     []ContainerStatus `json:"containers"`
	InitContainers []ContainerStatus `json:"init_containers,omitempty"`
	Conditions     []PodCondition    `json:"conditions,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	Age            string            `json:"age"`
	// QoS 等级
	QOSClass string `json:"qos_class,omitempty"`
	// 控制器信息
	OwnerKind string `json:"owner_kind,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
}

// ContainerStatus 容器状态
type ContainerStatus struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
	StateDetail  string `json:"state_detail,omitempty"`
	// 最后一次状态变更原因
	LastTerminationReason   string `json:"last_termination_reason,omitempty"`
	LastTerminationMessage  string `json:"last_termination_message,omitempty"`
	LastTerminationExitCode *int32 `json:"last_termination_exit_code,omitempty"`
}

// PodCondition Pod 条件
type PodCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ListPodsOptions 获取 Pod 列表的选项
type ListPodsOptions struct {
	// LabelSelector 标签选择器
	LabelSelector string
	// FieldSelector 字段选择器，如 "status.phase=Running"
	FieldSelector string
	// OnlyUnhealthy 只返回不健康的 Pod
	OnlyUnhealthy bool
}

// ListPodStatus 获取指定命名空间下的 Pod 状态列表
func (c *Client) ListPodStatus(ctx context.Context, namespace string, opts *ListPodsOptions) ([]PodStatusInfo, error) {
	if namespace == "" {
		return nil, fmt.Errorf("命名空间不能为空")
	}

	listOpts := metav1.ListOptions{}
	onlyUnhealthy := false

	if opts != nil {
		if opts.LabelSelector != "" {
			listOpts.LabelSelector = opts.LabelSelector
		}
		if opts.FieldSelector != "" {
			listOpts.FieldSelector = opts.FieldSelector
		}
		onlyUnhealthy = opts.OnlyUnhealthy
	}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	now := time.Now()
	result := make([]PodStatusInfo, 0, len(podList.Items))

	for _, pod := range podList.Items {
		info := buildPodStatusInfo(pod, now)

		if onlyUnhealthy && isPodHealthy(pod) {
			continue
		}

		result = append(result, info)
	}

	// 按状态排序：异常 Pod 排前面
	sort.Slice(result, func(i, j int) bool {
		iPriority := statusPriority(result[i].Status)
		jPriority := statusPriority(result[j].Status)
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// buildPodStatusInfo 从 Pod 对象构建状态信息
func buildPodStatusInfo(pod corev1.Pod, now time.Time) PodStatusInfo {
	info := PodStatusInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		IP:        pod.Status.PodIP,
		Labels:    pod.Labels,
		CreatedAt: pod.CreationTimestamp.Time,
		Age:       formatAge(now.Sub(pod.CreationTimestamp.Time)),
		QOSClass:  string(pod.Status.QOSClass),
	}

	if len(pod.OwnerReferences) > 0 {
		info.OwnerKind = pod.OwnerReferences[0].Kind
		info.OwnerName = pod.OwnerReferences[0].Name
	}

	var totalContainers int
	var readyContainers int
	var totalRestarts int32

	for _, cs := range pod.Status.InitContainerStatuses {
		info.InitContainers = append(info.InitContainers, buildContainerStatus(cs))
	}

	for _, cs := range pod.Status.ContainerStatuses {
		cStatus := buildContainerStatus(cs)
		info.Containers = append(info.Containers, cStatus)
		totalContainers++
		if cs.Ready {
			readyContainers++
		}
		totalRestarts += cs.RestartCount
	}

	info.Ready = fmt.Sprintf("%d/%d", readyContainers, totalContainers)
	info.Restarts = totalRestarts
	info.Status = determinePodStatus(pod)

	for _, cond := range pod.Status.Conditions {
		info.Conditions = append(info.Conditions, PodCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return info
}

// buildContainerStatus 构建容器状态
func buildContainerStatus(cs corev1.ContainerStatus) ContainerStatus {
	status := ContainerStatus{
		Name:         cs.Name,
		Image:        cs.Image,
		Ready:        cs.Ready,
		RestartCount: cs.RestartCount,
	}

	if cs.State.Running != nil {
		status.State = "Running"
		status.StateDetail = fmt.Sprintf("Started at %s", cs.State.Running.StartedAt.Format(time.RFC3339))
	} else if cs.State.Waiting != nil {
		status.State = "Waiting"
		status.StateDetail = cs.State.Waiting.Reason
		if cs.State.Waiting.Message != "" {
			status.StateDetail += ": " + cs.State.Waiting.Message
		}
	} else if cs.State.Terminated != nil {
		status.State = "Terminated"
		status.StateDetail = cs.State.Terminated.Reason
		if cs.State.Terminated.Message != "" {
			status.StateDetail += ": " + cs.State.Terminated.Message
		}
	}

	if cs.LastTerminationState.Terminated != nil {
		t := cs.LastTerminationState.Terminated
		status.LastTerminationReason = t.Reason
		status.LastTerminationMessage = t.Message
		status.LastTerminationExitCode = &t.ExitCode
	}

	return status
}

// determinePodStatus 模拟 kubectl 的 Pod 状态显示逻辑
func determinePodStatus(pod corev1.Pod) string {
	reason := string(pod.Status.Phase)

	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
			reason = "Init:Error"
		}
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			reason = fmt.Sprintf("Init:%s", cs.State.Waiting.Reason)
		}
	}

	hasRunning := false
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			reason = cs.State.Waiting.Reason
		} else if cs.State.Terminated != nil {
			if cs.State.Terminated.Reason != "" {
				reason = cs.State.Terminated.Reason
			} else if cs.State.Terminated.Signal != 0 {
				reason = fmt.Sprintf("Signal:%d", cs.State.Terminated.Signal)
			} else {
				reason = fmt.Sprintf("ExitCode:%d", cs.State.Terminated.ExitCode)
			}
		} else if cs.State.Running != nil {
			hasRunning = true
		}
	}

	if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	if hasRunning && pod.Status.Phase == corev1.PodRunning && reason == string(corev1.PodRunning) {
		reason = "Running"
	}

	return reason
}

// isPodHealthy 判断 Pod 是否健康
func isPodHealthy(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
		return false
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready || cs.RestartCount > 5 {
			return false
		}
		if cs.State.Waiting != nil {
			return false
		}
	}
	return true
}

// statusPriority 状态优先级（数字越小越优先显示 = 越异常）
func statusPriority(status string) int {
	switch status {
	case "CrashLoopBackOff":
		return 0
	case "Error", "OOMKilled":
		return 1
	case "ImagePullBackOff", "ErrImagePull":
		return 2
	case "Init:Error", "Init:CrashLoopBackOff":
		return 3
	case "Pending":
		return 4
	case "Terminating":
		return 5
	case "Running":
		return 8
	case "Completed", "Succeeded":
		return 9
	default:
		return 6
	}
}

// FormatPodStatusList 格式化 Pod 状态列表为可读字符串
func FormatPodStatusList(pods []PodStatusInfo) string {
	if len(pods) == 0 {
		return "未找到 Pod"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pod 状态列表 (共 %d 个):\n", len(pods)))
	sb.WriteString(fmt.Sprintf("%-45s %-20s %-8s %-10s %-15s %-15s\n", "NAME", "STATUS", "READY", "RESTARTS", "NODE", "AGE"))
	sb.WriteString(strings.Repeat("-", 120) + "\n")

	for _, p := range pods {
		sb.WriteString(fmt.Sprintf("%-45s %-20s %-8s %-10d %-15s %-15s\n",
			truncateString(p.Name, 44),
			p.Status,
			p.Ready,
			p.Restarts,
			truncateString(p.Node, 14),
			p.Age,
		))
	}

	statusCount := make(map[string]int)
	for _, p := range pods {
		statusCount[p.Status]++
	}
	sb.WriteString("\n状态摘要: ")
	summaryParts := make([]string, 0)
	for status, count := range statusCount {
		summaryParts = append(summaryParts, fmt.Sprintf("%s=%d", status, count))
	}
	sort.Strings(summaryParts)
	sb.WriteString(strings.Join(summaryParts, ", "))
	sb.WriteString("\n")

	return sb.String()
}

// truncateString 截断过长字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
