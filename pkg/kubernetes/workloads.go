package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// WorkloadType 工作负载类型
type WorkloadType string

const (
	WorkloadDeployment      WorkloadType = "Deployment"
	WorkloadStatefulSet     WorkloadType = "StatefulSet"
	WorkloadCronJob         WorkloadType = "CronJob"
	WorkloadJob             WorkloadType = "Job"
	WorkloadDaemonSet       WorkloadType = "DaemonSet"
	WorkloadStatefulSetPlus WorkloadType = "StatefulSetPlus"
)

// AllWorkloadTypes 所有支持的原生工作负载类型
var AllWorkloadTypes = []WorkloadType{
	WorkloadDeployment,
	WorkloadStatefulSet,
	WorkloadCronJob,
	WorkloadJob,
	WorkloadDaemonSet,
}

// WorkloadInfo 工作负载基本信息
type WorkloadInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      WorkloadType      `json:"type"`
	Replicas  *ReplicaStatus    `json:"replicas,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Age       string            `json:"age"`
	// CronJob 专属字段
	Schedule     string `json:"schedule,omitempty"`
	LastSchedule string `json:"last_schedule,omitempty"`
	ActiveJobs   int    `json:"active_jobs,omitempty"`
	Suspend      *bool  `json:"suspend,omitempty"`
	// Job 专属字段
	Completions *int32 `json:"completions,omitempty"`
	Succeeded   int32  `json:"succeeded,omitempty"`
	Failed      int32  `json:"failed,omitempty"`
	JobStatus   string `json:"job_status,omitempty"`
}

// ReplicaStatus 副本状态
type ReplicaStatus struct {
	Desired   int32 `json:"desired"`
	Ready     int32 `json:"ready"`
	Available int32 `json:"available"`
	Updated   int32 `json:"updated"`
}

// WorkloadDetail 工作负载详细信息
type WorkloadDetail struct {
	WorkloadInfo
	// 详细规格信息
	Strategy       string              `json:"strategy,omitempty"`
	Selector       map[string]string   `json:"selector,omitempty"`
	Annotations    map[string]string   `json:"annotations,omitempty"`
	Containers     []ContainerSpec     `json:"containers"`
	InitContainers []ContainerSpec     `json:"init_containers,omitempty"`
	Volumes        []string            `json:"volumes,omitempty"`
	NodeSelector   map[string]string   `json:"node_selector,omitempty"`
	ServiceAccount string              `json:"service_account,omitempty"`
	Conditions     []WorkloadCondition `json:"conditions,omitempty"`
	// 原始 JSON（完整信息备查）
	RawJSON string `json:"raw_json,omitempty"`
}

// ContainerSpec 容器规格
type ContainerSpec struct {
	Name            string           `json:"name"`
	Image           string           `json:"image"`
	Ports           []ContainerPort  `json:"ports,omitempty"`
	Env             []EnvVar         `json:"env,omitempty"`
	Resources       *ResourceRequire `json:"resources,omitempty"`
	VolumeMounts    []string         `json:"volume_mounts,omitempty"`
	LivenessProbe   string           `json:"liveness_probe,omitempty"`
	ReadinessProbe  string           `json:"readiness_probe,omitempty"`
	ImagePullPolicy string           `json:"image_pull_policy,omitempty"`
}

// ContainerPort 容器端口
type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"container_port"`
	Protocol      string `json:"protocol,omitempty"`
}

// EnvVar 环境变量（只记录名称，不暴露 Secret 值）
type EnvVar struct {
	Name      string `json:"name"`
	ValueFrom string `json:"value_from,omitempty"`
}

// ResourceRequire 资源需求
type ResourceRequire struct {
	RequestsCPU    string `json:"requests_cpu,omitempty"`
	RequestsMemory string `json:"requests_memory,omitempty"`
	LimitsCPU      string `json:"limits_cpu,omitempty"`
	LimitsMemory   string `json:"limits_memory,omitempty"`
}

// WorkloadCondition 工作负载状态条件
type WorkloadCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ListWorkloadsOptions 获取工作负载列表的选项
type ListWorkloadsOptions struct {
	// WorkloadTypes 要查询的工作负载类型，为空则查询所有类型
	WorkloadTypes []WorkloadType
	// LabelSelector 标签选择器
	LabelSelector string
	// IncludeStatefulSetPlus 是否包含 StatefulSetPlus (TKE CRD)
	IncludeStatefulSetPlus bool
}

// GetWorkloadDetailOptions 获取工作负载详情的选项
type GetWorkloadDetailOptions struct {
	// IncludeRawJSON 是否包含原始 JSON
	IncludeRawJSON bool
}

// ListWorkloads 获取指定命名空间下的工作负载列表
func (c *Client) ListWorkloads(ctx context.Context, namespace string, opts *ListWorkloadsOptions) ([]WorkloadInfo, error) {
	if namespace == "" {
		return nil, fmt.Errorf("命名空间不能为空")
	}

	workloadTypes := AllWorkloadTypes
	includeSSP := false
	labelSelector := ""

	if opts != nil {
		if len(opts.WorkloadTypes) > 0 {
			workloadTypes = opts.WorkloadTypes
		}
		includeSSP = opts.IncludeStatefulSetPlus
		labelSelector = opts.LabelSelector
	}

	listOpts := metav1.ListOptions{}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}

	var allWorkloads []WorkloadInfo
	now := time.Now()

	for _, wt := range workloadTypes {
		switch wt {
		case WorkloadDeployment:
			items, err := c.listDeployments(ctx, namespace, listOpts, now)
			if err != nil {
				c.logger.WithError(err).Warnf("获取 Deployment 列表失败")
				continue
			}
			allWorkloads = append(allWorkloads, items...)

		case WorkloadStatefulSet:
			items, err := c.listStatefulSets(ctx, namespace, listOpts, now)
			if err != nil {
				c.logger.WithError(err).Warnf("获取 StatefulSet 列表失败")
				continue
			}
			allWorkloads = append(allWorkloads, items...)

		case WorkloadDaemonSet:
			items, err := c.listDaemonSets(ctx, namespace, listOpts, now)
			if err != nil {
				c.logger.WithError(err).Warnf("获取 DaemonSet 列表失败")
				continue
			}
			allWorkloads = append(allWorkloads, items...)

		case WorkloadCronJob:
			items, err := c.listCronJobs(ctx, namespace, listOpts, now)
			if err != nil {
				c.logger.WithError(err).Warnf("获取 CronJob 列表失败")
				continue
			}
			allWorkloads = append(allWorkloads, items...)

		case WorkloadJob:
			items, err := c.listJobs(ctx, namespace, listOpts, now)
			if err != nil {
				c.logger.WithError(err).Warnf("获取 Job 列表失败")
				continue
			}
			allWorkloads = append(allWorkloads, items...)
		}
	}

	// StatefulSetPlus (TKE CRD) 需要单独查询
	if includeSSP {
		items, err := c.listStatefulSetPlus(ctx, namespace, listOpts, now)
		if err != nil {
			c.logger.WithError(err).Warnf("获取 StatefulSetPlus 列表失败（可能集群不支持该 CRD）")
		} else {
			allWorkloads = append(allWorkloads, items...)
		}
	}

	// 按类型 + 名称排序
	sort.Slice(allWorkloads, func(i, j int) bool {
		if allWorkloads[i].Type != allWorkloads[j].Type {
			return allWorkloads[i].Type < allWorkloads[j].Type
		}
		return allWorkloads[i].Name < allWorkloads[j].Name
	})

	return allWorkloads, nil
}

// GetWorkloadDetail 获取指定工作负载的详细信息
func (c *Client) GetWorkloadDetail(ctx context.Context, namespace, name string, workloadType WorkloadType, opts *GetWorkloadDetailOptions) (*WorkloadDetail, error) {
	if namespace == "" {
		return nil, fmt.Errorf("命名空间不能为空")
	}
	if name == "" {
		return nil, fmt.Errorf("工作负载名称不能为空")
	}

	includeRaw := false
	if opts != nil {
		includeRaw = opts.IncludeRawJSON
	}

	now := time.Now()

	switch workloadType {
	case WorkloadDeployment:
		return c.getDeploymentDetail(ctx, namespace, name, now, includeRaw)
	case WorkloadStatefulSet:
		return c.getStatefulSetDetail(ctx, namespace, name, now, includeRaw)
	case WorkloadDaemonSet:
		return c.getDaemonSetDetail(ctx, namespace, name, now, includeRaw)
	case WorkloadCronJob:
		return c.getCronJobDetail(ctx, namespace, name, now, includeRaw)
	case WorkloadJob:
		return c.getJobDetail(ctx, namespace, name, now, includeRaw)
	case WorkloadStatefulSetPlus:
		return c.getStatefulSetPlusDetail(ctx, namespace, name, now, includeRaw)
	default:
		return nil, fmt.Errorf("不支持的工作负载类型: %s", workloadType)
	}
}

// ========== Deployment ==========

func (c *Client) listDeployments(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	deployList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	result := make([]WorkloadInfo, 0, len(deployList.Items))
	for _, d := range deployList.Items {
		var desired int32
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		result = append(result, WorkloadInfo{
			Name:      d.Name,
			Namespace: d.Namespace,
			Type:      WorkloadDeployment,
			Replicas: &ReplicaStatus{
				Desired:   desired,
				Ready:     d.Status.ReadyReplicas,
				Available: d.Status.AvailableReplicas,
				Updated:   d.Status.UpdatedReplicas,
			},
			Labels:    d.Labels,
			CreatedAt: d.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(d.CreationTimestamp.Time)),
		})
	}
	return result, nil
}

func (c *Client) getDeploymentDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Deployment %s/%s 详情失败: %w", namespace, name, err)
	}

	var desired int32
	if deploy.Spec.Replicas != nil {
		desired = *deploy.Spec.Replicas
	}

	detail := &WorkloadDetail{
		WorkloadInfo: WorkloadInfo{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Type:      WorkloadDeployment,
			Replicas: &ReplicaStatus{
				Desired:   desired,
				Ready:     deploy.Status.ReadyReplicas,
				Available: deploy.Status.AvailableReplicas,
				Updated:   deploy.Status.UpdatedReplicas,
			},
			Labels:    deploy.Labels,
			CreatedAt: deploy.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(deploy.CreationTimestamp.Time)),
		},
		Strategy:       string(deploy.Spec.Strategy.Type),
		Annotations:    filterSystemAnnotations(deploy.Annotations),
		ServiceAccount: deploy.Spec.Template.Spec.ServiceAccountName,
		NodeSelector:   deploy.Spec.Template.Spec.NodeSelector,
	}

	if deploy.Spec.Selector != nil {
		detail.Selector = deploy.Spec.Selector.MatchLabels
	}

	detail.Containers = extractContainerSpecs(deploy.Spec.Template.Spec.Containers)
	detail.InitContainers = extractContainerSpecs(deploy.Spec.Template.Spec.InitContainers)

	for _, v := range deploy.Spec.Template.Spec.Volumes {
		detail.Volumes = append(detail.Volumes, v.Name)
	}

	for _, cond := range deploy.Status.Conditions {
		detail.Conditions = append(detail.Conditions, WorkloadCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	if includeRaw {
		raw, _ := json.MarshalIndent(deploy, "", "  ")
		detail.RawJSON = string(raw)
	}

	return detail, nil
}

// ========== StatefulSet ==========

func (c *Client) listStatefulSets(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	stsList, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	result := make([]WorkloadInfo, 0, len(stsList.Items))
	for _, s := range stsList.Items {
		var desired int32
		if s.Spec.Replicas != nil {
			desired = *s.Spec.Replicas
		}
		result = append(result, WorkloadInfo{
			Name:      s.Name,
			Namespace: s.Namespace,
			Type:      WorkloadStatefulSet,
			Replicas: &ReplicaStatus{
				Desired:   desired,
				Ready:     s.Status.ReadyReplicas,
				Available: s.Status.AvailableReplicas,
				Updated:   s.Status.UpdatedReplicas,
			},
			Labels:    s.Labels,
			CreatedAt: s.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(s.CreationTimestamp.Time)),
		})
	}
	return result, nil
}

func (c *Client) getStatefulSetDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	sts, err := c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StatefulSet %s/%s 详情失败: %w", namespace, name, err)
	}

	var desired int32
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}

	detail := &WorkloadDetail{
		WorkloadInfo: WorkloadInfo{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Type:      WorkloadStatefulSet,
			Replicas: &ReplicaStatus{
				Desired:   desired,
				Ready:     sts.Status.ReadyReplicas,
				Available: sts.Status.AvailableReplicas,
				Updated:   sts.Status.UpdatedReplicas,
			},
			Labels:    sts.Labels,
			CreatedAt: sts.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(sts.CreationTimestamp.Time)),
		},
		Strategy:       string(sts.Spec.UpdateStrategy.Type),
		Annotations:    filterSystemAnnotations(sts.Annotations),
		ServiceAccount: sts.Spec.Template.Spec.ServiceAccountName,
		NodeSelector:   sts.Spec.Template.Spec.NodeSelector,
	}

	if sts.Spec.Selector != nil {
		detail.Selector = sts.Spec.Selector.MatchLabels
	}
	detail.Containers = extractContainerSpecs(sts.Spec.Template.Spec.Containers)
	detail.InitContainers = extractContainerSpecs(sts.Spec.Template.Spec.InitContainers)
	for _, v := range sts.Spec.Template.Spec.Volumes {
		detail.Volumes = append(detail.Volumes, v.Name)
	}
	for _, cond := range sts.Status.Conditions {
		detail.Conditions = append(detail.Conditions, WorkloadCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	if includeRaw {
		raw, _ := json.MarshalIndent(sts, "", "  ")
		detail.RawJSON = string(raw)
	}

	return detail, nil
}

// ========== DaemonSet ==========

func (c *Client) listDaemonSets(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	dsList, err := c.clientset.AppsV1().DaemonSets(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	result := make([]WorkloadInfo, 0, len(dsList.Items))
	for _, d := range dsList.Items {
		result = append(result, WorkloadInfo{
			Name:      d.Name,
			Namespace: d.Namespace,
			Type:      WorkloadDaemonSet,
			Replicas: &ReplicaStatus{
				Desired:   d.Status.DesiredNumberScheduled,
				Ready:     d.Status.NumberReady,
				Available: d.Status.NumberAvailable,
				Updated:   d.Status.UpdatedNumberScheduled,
			},
			Labels:    d.Labels,
			CreatedAt: d.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(d.CreationTimestamp.Time)),
		})
	}
	return result, nil
}

func (c *Client) getDaemonSetDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	ds, err := c.clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 DaemonSet %s/%s 详情失败: %w", namespace, name, err)
	}

	detail := &WorkloadDetail{
		WorkloadInfo: WorkloadInfo{
			Name:      ds.Name,
			Namespace: ds.Namespace,
			Type:      WorkloadDaemonSet,
			Replicas: &ReplicaStatus{
				Desired:   ds.Status.DesiredNumberScheduled,
				Ready:     ds.Status.NumberReady,
				Available: ds.Status.NumberAvailable,
				Updated:   ds.Status.UpdatedNumberScheduled,
			},
			Labels:    ds.Labels,
			CreatedAt: ds.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(ds.CreationTimestamp.Time)),
		},
		Strategy:       string(ds.Spec.UpdateStrategy.Type),
		Annotations:    filterSystemAnnotations(ds.Annotations),
		ServiceAccount: ds.Spec.Template.Spec.ServiceAccountName,
		NodeSelector:   ds.Spec.Template.Spec.NodeSelector,
	}

	if ds.Spec.Selector != nil {
		detail.Selector = ds.Spec.Selector.MatchLabels
	}
	detail.Containers = extractContainerSpecs(ds.Spec.Template.Spec.Containers)
	detail.InitContainers = extractContainerSpecs(ds.Spec.Template.Spec.InitContainers)
	for _, v := range ds.Spec.Template.Spec.Volumes {
		detail.Volumes = append(detail.Volumes, v.Name)
	}
	for _, cond := range ds.Status.Conditions {
		detail.Conditions = append(detail.Conditions, WorkloadCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	if includeRaw {
		raw, _ := json.MarshalIndent(ds, "", "  ")
		detail.RawJSON = string(raw)
	}

	return detail, nil
}

// ========== CronJob ==========

func (c *Client) listCronJobs(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	cronJobList, err := c.clientset.BatchV1().CronJobs(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	result := make([]WorkloadInfo, 0, len(cronJobList.Items))
	for _, cj := range cronJobList.Items {
		info := WorkloadInfo{
			Name:      cj.Name,
			Namespace: cj.Namespace,
			Type:      WorkloadCronJob,
			Labels:    cj.Labels,
			CreatedAt: cj.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(cj.CreationTimestamp.Time)),
			Schedule:  cj.Spec.Schedule,
			Suspend:   cj.Spec.Suspend,
		}
		if cj.Status.LastScheduleTime != nil {
			info.LastSchedule = formatAge(now.Sub(cj.Status.LastScheduleTime.Time)) + " ago"
		}
		info.ActiveJobs = len(cj.Status.Active)
		result = append(result, info)
	}
	return result, nil
}

func (c *Client) getCronJobDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	cj, err := c.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 CronJob %s/%s 详情失败: %w", namespace, name, err)
	}

	detail := &WorkloadDetail{
		WorkloadInfo: WorkloadInfo{
			Name:      cj.Name,
			Namespace: cj.Namespace,
			Type:      WorkloadCronJob,
			Labels:    cj.Labels,
			CreatedAt: cj.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(cj.CreationTimestamp.Time)),
			Schedule:  cj.Spec.Schedule,
			Suspend:   cj.Spec.Suspend,
		},
		Annotations: filterSystemAnnotations(cj.Annotations),
	}

	if cj.Status.LastScheduleTime != nil {
		detail.LastSchedule = formatAge(now.Sub(cj.Status.LastScheduleTime.Time)) + " ago"
	}
	detail.ActiveJobs = len(cj.Status.Active)

	detail.Containers = extractContainerSpecs(cj.Spec.JobTemplate.Spec.Template.Spec.Containers)
	detail.InitContainers = extractContainerSpecs(cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
	detail.ServiceAccount = cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName
	detail.NodeSelector = cj.Spec.JobTemplate.Spec.Template.Spec.NodeSelector

	for _, v := range cj.Spec.JobTemplate.Spec.Template.Spec.Volumes {
		detail.Volumes = append(detail.Volumes, v.Name)
	}

	if includeRaw {
		raw, _ := json.MarshalIndent(cj, "", "  ")
		detail.RawJSON = string(raw)
	}

	return detail, nil
}

// ========== Job ==========

func (c *Client) listJobs(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	jobList, err := c.clientset.BatchV1().Jobs(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	result := make([]WorkloadInfo, 0, len(jobList.Items))
	for _, j := range jobList.Items {
		info := WorkloadInfo{
			Name:      j.Name,
			Namespace: j.Namespace,
			Type:      WorkloadJob,
			Labels:    j.Labels,
			CreatedAt: j.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(j.CreationTimestamp.Time)),
			Succeeded: j.Status.Succeeded,
			Failed:    j.Status.Failed,
		}
		if j.Spec.Completions != nil {
			info.Completions = j.Spec.Completions
		}
		info.JobStatus = getJobStatus(j.Status.Conditions)
		result = append(result, info)
	}
	return result, nil
}

func (c *Client) getJobDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	job, err := c.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Job %s/%s 详情失败: %w", namespace, name, err)
	}

	detail := &WorkloadDetail{
		WorkloadInfo: WorkloadInfo{
			Name:      job.Name,
			Namespace: job.Namespace,
			Type:      WorkloadJob,
			Labels:    job.Labels,
			CreatedAt: job.CreationTimestamp.Time,
			Age:       formatAge(now.Sub(job.CreationTimestamp.Time)),
			Succeeded: job.Status.Succeeded,
			Failed:    job.Status.Failed,
		},
		Annotations:    filterSystemAnnotations(job.Annotations),
		ServiceAccount: job.Spec.Template.Spec.ServiceAccountName,
		NodeSelector:   job.Spec.Template.Spec.NodeSelector,
	}

	if job.Spec.Completions != nil {
		detail.Completions = job.Spec.Completions
	}
	detail.JobStatus = getJobStatus(job.Status.Conditions)

	detail.Containers = extractContainerSpecs(job.Spec.Template.Spec.Containers)
	detail.InitContainers = extractContainerSpecs(job.Spec.Template.Spec.InitContainers)

	for _, v := range job.Spec.Template.Spec.Volumes {
		detail.Volumes = append(detail.Volumes, v.Name)
	}

	for _, cond := range job.Status.Conditions {
		detail.Conditions = append(detail.Conditions, WorkloadCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	if includeRaw {
		raw, _ := json.MarshalIndent(job, "", "  ")
		detail.RawJSON = string(raw)
	}

	return detail, nil
}

// ========== StatefulSetPlus (TKE CRD) ==========
// StatefulSetPlus 是腾讯云 TKE 提供的增强型 StatefulSet
// GVR: apps.kruise.io/v1beta1/statefulsets 或 platform.tkestack.io/v1/statefulsetplus

var statefulSetPlusGVRs = []schema.GroupVersionResource{
	{Group: "apps.kruise.io", Version: "v1beta1", Resource: "statefulsets"},
	{Group: "platform.tkestack.io", Version: "v1", Resource: "statefulsetplus"},
}

func (c *Client) listStatefulSetPlus(ctx context.Context, namespace string, listOpts metav1.ListOptions, now time.Time) ([]WorkloadInfo, error) {
	for _, gvr := range statefulSetPlusGVRs {
		list, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
		if err != nil {
			continue
		}

		result := make([]WorkloadInfo, 0, len(list.Items))
		for _, item := range list.Items {
			result = append(result, parseUnstructuredWorkload(item, WorkloadStatefulSetPlus, now))
		}
		return result, nil
	}
	return nil, fmt.Errorf("集群不支持 StatefulSetPlus CRD")
}

func (c *Client) getStatefulSetPlusDetail(ctx context.Context, namespace, name string, now time.Time, includeRaw bool) (*WorkloadDetail, error) {
	for _, gvr := range statefulSetPlusGVRs {
		item, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		info := parseUnstructuredWorkload(*item, WorkloadStatefulSetPlus, now)
		detail := &WorkloadDetail{
			WorkloadInfo: info,
			Annotations:  filterSystemAnnotations(item.GetAnnotations()),
		}

		containers, _, _ := unstructured.NestedSlice(item.Object, "spec", "template", "spec", "containers")
		for _, c := range containers {
			if cm, ok := c.(map[string]interface{}); ok {
				cs := ContainerSpec{
					Name:  getNestedString(cm, "name"),
					Image: getNestedString(cm, "image"),
				}
				detail.Containers = append(detail.Containers, cs)
			}
		}

		if includeRaw {
			raw, _ := json.MarshalIndent(item.Object, "", "  ")
			detail.RawJSON = string(raw)
		}

		return detail, nil
	}
	return nil, fmt.Errorf("获取 StatefulSetPlus %s/%s 详情失败：集群不支持该 CRD", namespace, name)
}

// ========== 辅助函数 ==========

func parseUnstructuredWorkload(item unstructured.Unstructured, wt WorkloadType, now time.Time) WorkloadInfo {
	info := WorkloadInfo{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
		Type:      wt,
		Labels:    item.GetLabels(),
		CreatedAt: item.GetCreationTimestamp().Time,
		Age:       formatAge(now.Sub(item.GetCreationTimestamp().Time)),
	}

	desired, _, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas")
	info.Replicas = &ReplicaStatus{
		Desired: int32(desired),
		Ready:   int32(readyReplicas),
	}

	return info
}

func extractContainerSpecs(containers []corev1.Container) []ContainerSpec {
	if len(containers) == 0 {
		return nil
	}

	result := make([]ContainerSpec, 0, len(containers))
	for _, c := range containers {
		cs := ContainerSpec{
			Name:            c.Name,
			Image:           c.Image,
			ImagePullPolicy: string(c.ImagePullPolicy),
		}

		for _, p := range c.Ports {
			cs.Ports = append(cs.Ports, ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				Protocol:      string(p.Protocol),
			})
		}

		for _, e := range c.Env {
			ev := EnvVar{Name: e.Name}
			if e.ValueFrom != nil {
				if e.ValueFrom.SecretKeyRef != nil {
					ev.ValueFrom = fmt.Sprintf("secret:%s/%s", e.ValueFrom.SecretKeyRef.Name, e.ValueFrom.SecretKeyRef.Key)
				} else if e.ValueFrom.ConfigMapKeyRef != nil {
					ev.ValueFrom = fmt.Sprintf("configmap:%s/%s", e.ValueFrom.ConfigMapKeyRef.Name, e.ValueFrom.ConfigMapKeyRef.Key)
				} else if e.ValueFrom.FieldRef != nil {
					ev.ValueFrom = fmt.Sprintf("fieldRef:%s", e.ValueFrom.FieldRef.FieldPath)
				}
			}
			cs.Env = append(cs.Env, ev)
		}

		if c.Resources.Requests != nil || c.Resources.Limits != nil {
			cs.Resources = &ResourceRequire{}
			if cpu := c.Resources.Requests.Cpu(); cpu != nil {
				cs.Resources.RequestsCPU = cpu.String()
			}
			if mem := c.Resources.Requests.Memory(); mem != nil {
				cs.Resources.RequestsMemory = mem.String()
			}
			if cpu := c.Resources.Limits.Cpu(); cpu != nil {
				cs.Resources.LimitsCPU = cpu.String()
			}
			if mem := c.Resources.Limits.Memory(); mem != nil {
				cs.Resources.LimitsMemory = mem.String()
			}
		}

		for _, vm := range c.VolumeMounts {
			cs.VolumeMounts = append(cs.VolumeMounts, fmt.Sprintf("%s -> %s", vm.Name, vm.MountPath))
		}

		if c.LivenessProbe != nil {
			cs.LivenessProbe = formatProbe(c.LivenessProbe)
		}
		if c.ReadinessProbe != nil {
			cs.ReadinessProbe = formatProbe(c.ReadinessProbe)
		}

		result = append(result, cs)
	}
	return result
}

func formatProbe(probe *corev1.Probe) string {
	if probe == nil {
		return ""
	}
	if probe.HTTPGet != nil {
		return fmt.Sprintf("HTTP GET %s:%d%s", probe.HTTPGet.Host, probe.HTTPGet.Port.IntValue(), probe.HTTPGet.Path)
	}
	if probe.TCPSocket != nil {
		return fmt.Sprintf("TCP %d", probe.TCPSocket.Port.IntValue())
	}
	if probe.Exec != nil {
		return fmt.Sprintf("exec: %s", strings.Join(probe.Exec.Command, " "))
	}
	if probe.GRPC != nil {
		return fmt.Sprintf("gRPC port=%d", probe.GRPC.Port)
	}
	return "unknown"
}

func getJobStatus(conditions []batchv1.JobCondition) string {
	for _, c := range conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return "Complete"
		}
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return "Failed"
		}
	}
	return "Running"
}

func getNestedString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// FormatWorkloadList 格式化工作负载列表为可读字符串
func FormatWorkloadList(workloads []WorkloadInfo) string {
	if len(workloads) == 0 {
		return "未找到工作负载"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("工作负载列表 (共 %d 个):\n", len(workloads)))
	sb.WriteString(fmt.Sprintf("%-20s %-35s %-15s %-15s\n", "TYPE", "NAME", "READY", "AGE"))
	sb.WriteString(strings.Repeat("-", 90) + "\n")

	for _, w := range workloads {
		ready := "-"
		if w.Replicas != nil {
			ready = fmt.Sprintf("%d/%d", w.Replicas.Ready, w.Replicas.Desired)
		}
		if w.Type == WorkloadCronJob {
			suspend := "Active"
			if w.Suspend != nil && *w.Suspend {
				suspend = "Suspended"
			}
			ready = fmt.Sprintf("%s (%s)", w.Schedule, suspend)
		}
		if w.Type == WorkloadJob {
			ready = fmt.Sprintf("%d/%d (%s)", w.Succeeded, safeInt32(w.Completions), w.JobStatus)
		}
		sb.WriteString(fmt.Sprintf("%-20s %-35s %-15s %-15s\n", w.Type, w.Name, ready, w.Age))
	}

	return sb.String()
}

func safeInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
