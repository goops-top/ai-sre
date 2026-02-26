package vpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"ai-sre/tools/mcp/internal/tencentcloud"
)

// Client VPC 客户端
type Client struct {
	client  *vpc.Client
	manager *tencentcloud.ClientManager
	logger  *logrus.Logger
}

// NewClient 创建 VPC 客户端
func NewClient(manager *tencentcloud.ClientManager, logger *logrus.Logger) (*Client, error) {
	credential := manager.GetCredential()
	clientProfile := manager.GetClientProfile("vpc")

	client, err := vpc.NewClient(credential, "ap-beijing", clientProfile)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	return &Client{
		client:  client,
		manager: manager,
		logger:  logger,
	}, nil
}

func (c *Client) GetProductName() string    { return "VPC" }
func (c *Client) GetProductVersion() string { return "2017-03-12" }
func (c *Client) ValidatePermissions(ctx context.Context) error {
	_, err := c.DescribeVpcs(ctx, "ap-beijing")
	if err != nil {
		return fmt.Errorf("VPC 权限验证失败: %w", err)
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

func getUint64Value(u *uint64) uint64 {
	if u != nil {
		return *u
	}
	return 0
}

func getInt64Value(i *int64) int64 {
	if i != nil {
		return *i
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

// newRegionClient 创建指定地域的 VPC 客户端
func (c *Client) newRegionClient(region string) (*vpc.Client, error) {
	credential := c.manager.GetCredential()
	clientProfile := c.manager.GetClientProfile("vpc")
	return vpc.NewClient(credential, region, clientProfile)
}

// ===================== DescribeVpcs =====================

// VpcInfo VPC 信息
type VpcInfo struct {
	VpcId       string   `json:"vpc_id"`
	VpcName     string   `json:"vpc_name"`
	CidrBlock   string   `json:"cidr_block"`
	IsDefault   bool     `json:"is_default"`
	CreatedTime string   `json:"created_time"`
	DnsServers  []string `json:"dns_servers,omitempty"`
	DomainName  string   `json:"domain_name,omitempty"`
	Ipv6Cidr    string   `json:"ipv6_cidr,omitempty"`
	EnableDhcp  bool     `json:"enable_dhcp"`
}

// DescribeVpcsResult 查询 VPC 结果
type DescribeVpcsResult struct {
	TotalCount uint64    `json:"total_count"`
	Vpcs       []VpcInfo `json:"vpcs"`
	Region     string    `json:"region"`
}

// DescribeVpcs 查询 VPC 列表
func (c *Client) DescribeVpcs(ctx context.Context, region string) (*DescribeVpcsResult, error) {
	c.logger.WithField("region", region).Debug("开始查询 VPC 列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeVpcsRequest()
	limit := "100"
	request.Limit = &limit

	response, err := client.DescribeVpcs(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询 VPC 列表失败: %w", err)
	}

	result := &DescribeVpcsResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, v := range response.Response.VpcSet {
		info := VpcInfo{
			VpcId:       getStringValue(v.VpcId),
			VpcName:     getStringValue(v.VpcName),
			CidrBlock:   getStringValue(v.CidrBlock),
			IsDefault:   getBoolValue(v.IsDefault),
			CreatedTime: getStringValue(v.CreatedTime),
			DnsServers:  convertStringPtrSlice(v.DnsServerSet),
			DomainName:  getStringValue(v.DomainName),
			Ipv6Cidr:    getStringValue(v.Ipv6CidrBlock),
			EnableDhcp:  getBoolValue(v.EnableDhcp),
		}
		result.Vpcs = append(result.Vpcs, info)
	}

	c.logger.WithField("vpc_count", len(result.Vpcs)).Info("成功查询 VPC 列表")
	return result, nil
}

func (c *Client) FormatVpcsAsJSON(result *DescribeVpcsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatVpcsAsTable(result *DescribeVpcsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("VPC 列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 120) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-25s %-18s %-10s %-8s %-26s\n",
		"VPC ID", "名称", "CIDR", "默认VPC", "DHCP", "创建时间"))
	sb.WriteString(strings.Repeat("-", 120) + "\n")

	for _, v := range result.Vpcs {
		isDefault := "否"
		if v.IsDefault {
			isDefault = "是"
		}
		dhcp := "关"
		if v.EnableDhcp {
			dhcp = "开"
		}
		sb.WriteString(fmt.Sprintf("%-22s %-25s %-18s %-10s %-8s %-26s\n",
			v.VpcId,
			truncateString(v.VpcName, 23),
			v.CidrBlock,
			isDefault,
			dhcp,
			v.CreatedTime))
	}

	return sb.String()
}

// ===================== DescribeSubnets =====================

// SubnetInfo 子网信息
type SubnetInfo struct {
	SubnetId              string `json:"subnet_id"`
	SubnetName            string `json:"subnet_name"`
	VpcId                 string `json:"vpc_id"`
	CidrBlock             string `json:"cidr_block"`
	Zone                  string `json:"zone"`
	IsDefault             bool   `json:"is_default"`
	AvailableIpCount      uint64 `json:"available_ip_count"`
	TotalIpCount          uint64 `json:"total_ip_count"`
	RouteTableId          string `json:"route_table_id"`
	NetworkAclId          string `json:"network_acl_id,omitempty"`
	CreatedTime           string `json:"created_time"`
}

// DescribeSubnetsResult 查询子网结果
type DescribeSubnetsResult struct {
	TotalCount uint64       `json:"total_count"`
	Subnets    []SubnetInfo `json:"subnets"`
	Region     string       `json:"region"`
}

// DescribeSubnets 查询子网列表
func (c *Client) DescribeSubnets(ctx context.Context, region string, vpcId string) (*DescribeSubnetsResult, error) {
	c.logger.WithFields(logrus.Fields{"region": region, "vpc_id": vpcId}).Debug("开始查询子网列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeSubnetsRequest()
	limit := "100"
	request.Limit = &limit

	if vpcId != "" {
		request.Filters = []*vpc.Filter{
			{
				Name:   common.StringPtr("vpc-id"),
				Values: common.StringPtrs([]string{vpcId}),
			},
		}
	}

	response, err := client.DescribeSubnets(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询子网列表失败: %w", err)
	}

	result := &DescribeSubnetsResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, s := range response.Response.SubnetSet {
		info := SubnetInfo{
			SubnetId:         getStringValue(s.SubnetId),
			SubnetName:       getStringValue(s.SubnetName),
			VpcId:            getStringValue(s.VpcId),
			CidrBlock:        getStringValue(s.CidrBlock),
			Zone:             getStringValue(s.Zone),
			IsDefault:        getBoolValue(s.IsDefault),
			AvailableIpCount: getUint64Value(s.AvailableIpAddressCount),
			TotalIpCount:     getUint64Value(s.TotalIpAddressCount),
			RouteTableId:     getStringValue(s.RouteTableId),
			NetworkAclId:     getStringValue(s.NetworkAclId),
			CreatedTime:      getStringValue(s.CreatedTime),
		}
		result.Subnets = append(result.Subnets, info)
	}

	c.logger.WithField("subnet_count", len(result.Subnets)).Info("成功查询子网列表")
	return result, nil
}

func (c *Client) FormatSubnetsAsJSON(result *DescribeSubnetsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatSubnetsAsTable(result *DescribeSubnetsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("子网列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 150) + "\n")
	sb.WriteString(fmt.Sprintf("%-24s %-22s %-22s %-18s %-16s %-8s %-8s %-26s\n",
		"子网ID", "名称", "VPC ID", "CIDR", "可用区", "可用IP", "总IP", "创建时间"))
	sb.WriteString(strings.Repeat("-", 150) + "\n")

	for _, s := range result.Subnets {
		sb.WriteString(fmt.Sprintf("%-24s %-22s %-22s %-18s %-16s %-8d %-8d %-26s\n",
			s.SubnetId,
			truncateString(s.SubnetName, 20),
			s.VpcId,
			s.CidrBlock,
			s.Zone,
			s.AvailableIpCount,
			s.TotalIpCount,
			s.CreatedTime))
	}

	return sb.String()
}

// ===================== DescribeSecurityGroups =====================

// SecurityGroupInfo 安全组信息
type SecurityGroupInfo struct {
	SecurityGroupId   string `json:"security_group_id"`
	SecurityGroupName string `json:"security_group_name"`
	SecurityGroupDesc string `json:"security_group_desc"`
	ProjectId         string `json:"project_id"`
	IsDefault         bool   `json:"is_default"`
	CreatedTime       string `json:"created_time"`
	UpdateTime        string `json:"update_time"`
}

// DescribeSecurityGroupsResult 查询安全组结果
type DescribeSecurityGroupsResult struct {
	TotalCount     uint64              `json:"total_count"`
	SecurityGroups []SecurityGroupInfo `json:"security_groups"`
	Region         string              `json:"region"`
}

// DescribeSecurityGroups 查询安全组列表
func (c *Client) DescribeSecurityGroups(ctx context.Context, region string) (*DescribeSecurityGroupsResult, error) {
	c.logger.WithField("region", region).Debug("开始查询安全组列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeSecurityGroupsRequest()
	limit := "100"
	request.Limit = &limit

	response, err := client.DescribeSecurityGroups(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询安全组列表失败: %w", err)
	}

	result := &DescribeSecurityGroupsResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, sg := range response.Response.SecurityGroupSet {
		info := SecurityGroupInfo{
			SecurityGroupId:   getStringValue(sg.SecurityGroupId),
			SecurityGroupName: getStringValue(sg.SecurityGroupName),
			SecurityGroupDesc: getStringValue(sg.SecurityGroupDesc),
			ProjectId:         getStringValue(sg.ProjectId),
			IsDefault:         getBoolValue(sg.IsDefault),
			CreatedTime:       getStringValue(sg.CreatedTime),
			UpdateTime:        getStringValue(sg.UpdateTime),
		}
		result.SecurityGroups = append(result.SecurityGroups, info)
	}

	c.logger.WithField("sg_count", len(result.SecurityGroups)).Info("成功查询安全组列表")
	return result, nil
}

func (c *Client) FormatSecurityGroupsAsJSON(result *DescribeSecurityGroupsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatSecurityGroupsAsTable(result *DescribeSecurityGroupsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("安全组列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 130) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-25s %-30s %-10s %-26s\n",
		"安全组ID", "名称", "描述", "默认", "创建时间"))
	sb.WriteString(strings.Repeat("-", 130) + "\n")

	for _, sg := range result.SecurityGroups {
		isDefault := "否"
		if sg.IsDefault {
			isDefault = "是"
		}
		sb.WriteString(fmt.Sprintf("%-22s %-25s %-30s %-10s %-26s\n",
			sg.SecurityGroupId,
			truncateString(sg.SecurityGroupName, 23),
			truncateString(sg.SecurityGroupDesc, 28),
			isDefault,
			sg.CreatedTime))
	}

	return sb.String()
}

// ===================== DescribeNetworkInterfaces =====================

// NetworkInterfaceInfo 弹性网卡信息
type NetworkInterfaceInfo struct {
	NetworkInterfaceId   string   `json:"network_interface_id"`
	NetworkInterfaceName string   `json:"network_interface_name"`
	VpcId                string   `json:"vpc_id"`
	SubnetId             string   `json:"subnet_id"`
	MacAddress           string   `json:"mac_address"`
	State                string   `json:"state"`
	Primary              bool     `json:"primary"`
	PrivateIpAddresses   []string `json:"private_ip_addresses"`
	SecurityGroups       []string `json:"security_groups"`
	Zone                 string   `json:"zone"`
	CreatedTime          string   `json:"created_time"`
}

// DescribeNetworkInterfacesResult 查询弹性网卡结果
type DescribeNetworkInterfacesResult struct {
	TotalCount        uint64                 `json:"total_count"`
	NetworkInterfaces []NetworkInterfaceInfo `json:"network_interfaces"`
	Region            string                 `json:"region"`
}

// DescribeNetworkInterfaces 查询弹性网卡列表
func (c *Client) DescribeNetworkInterfaces(ctx context.Context, region string, vpcId string) (*DescribeNetworkInterfacesResult, error) {
	c.logger.WithFields(logrus.Fields{"region": region, "vpc_id": vpcId}).Debug("开始查询弹性网卡列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeNetworkInterfacesRequest()
	var limit uint64 = 100
	request.Limit = &limit

	if vpcId != "" {
		request.Filters = []*vpc.Filter{
			{
				Name:   common.StringPtr("vpc-id"),
				Values: common.StringPtrs([]string{vpcId}),
			},
		}
	}

	response, err := client.DescribeNetworkInterfaces(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询弹性网卡列表失败: %w", err)
	}

	result := &DescribeNetworkInterfacesResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, eni := range response.Response.NetworkInterfaceSet {
		var privateIPs []string
		for _, pip := range eni.PrivateIpAddressSet {
			if pip.PrivateIpAddress != nil {
				privateIPs = append(privateIPs, *pip.PrivateIpAddress)
			}
		}

		info := NetworkInterfaceInfo{
			NetworkInterfaceId:   getStringValue(eni.NetworkInterfaceId),
			NetworkInterfaceName: getStringValue(eni.NetworkInterfaceName),
			VpcId:                getStringValue(eni.VpcId),
			SubnetId:             getStringValue(eni.SubnetId),
			MacAddress:           getStringValue(eni.MacAddress),
			State:                getStringValue(eni.State),
			Primary:              getBoolValue(eni.Primary),
			PrivateIpAddresses:   privateIPs,
			SecurityGroups:       convertStringPtrSlice(eni.GroupSet),
			Zone:                 getStringValue(eni.Zone),
			CreatedTime:          getStringValue(eni.CreatedTime),
		}
		result.NetworkInterfaces = append(result.NetworkInterfaces, info)
	}

	c.logger.WithField("eni_count", len(result.NetworkInterfaces)).Info("成功查询弹性网卡列表")
	return result, nil
}

func (c *Client) FormatNetworkInterfacesAsJSON(result *DescribeNetworkInterfacesResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatNetworkInterfacesAsTable(result *DescribeNetworkInterfacesResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("弹性网卡列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 150) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-20s %-22s %-18s %-10s %-8s %-18s\n",
		"网卡ID", "名称", "VPC ID", "MAC地址", "状态", "主网卡", "内网IP"))
	sb.WriteString(strings.Repeat("-", 150) + "\n")

	for _, eni := range result.NetworkInterfaces {
		primaryStr := "否"
		if eni.Primary {
			primaryStr = "是"
		}
		privateIP := "-"
		if len(eni.PrivateIpAddresses) > 0 {
			privateIP = eni.PrivateIpAddresses[0]
		}
		sb.WriteString(fmt.Sprintf("%-22s %-20s %-22s %-18s %-10s %-8s %-18s\n",
			eni.NetworkInterfaceId,
			truncateString(eni.NetworkInterfaceName, 18),
			eni.VpcId,
			eni.MacAddress,
			eni.State,
			primaryStr,
			privateIP))
	}

	return sb.String()
}

// ===================== DescribeAddresses =====================

// AddressInfo 弹性公网IP信息
type AddressInfo struct {
	AddressId               string `json:"address_id"`
	AddressName             string `json:"address_name"`
	AddressIp               string `json:"address_ip"`
	AddressStatus           string `json:"address_status"`
	InstanceId              string `json:"instance_id,omitempty"`
	InstanceType            string `json:"instance_type,omitempty"`
	NetworkInterfaceId      string `json:"network_interface_id,omitempty"`
	PrivateAddressIp        string `json:"private_address_ip,omitempty"`
	Bandwidth               uint64 `json:"bandwidth"`
	InternetChargeType      string `json:"internet_charge_type"`
	InternetServiceProvider string `json:"internet_service_provider,omitempty"`
	CreatedTime             string `json:"created_time"`
	BandwidthPackageId      string `json:"bandwidth_package_id,omitempty"`
}

// DescribeAddressesResult 查询 EIP 结果
type DescribeAddressesResult struct {
	TotalCount int64         `json:"total_count"`
	Addresses  []AddressInfo `json:"addresses"`
	Region     string        `json:"region"`
}

// DescribeAddresses 查询弹性公网IP列表
func (c *Client) DescribeAddresses(ctx context.Context, region string) (*DescribeAddressesResult, error) {
	c.logger.WithField("region", region).Debug("开始查询弹性公网IP列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeAddressesRequest()
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeAddresses(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询弹性公网IP列表失败: %w", err)
	}

	result := &DescribeAddressesResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, addr := range response.Response.AddressSet {
		info := AddressInfo{
			AddressId:               getStringValue(addr.AddressId),
			AddressName:             getStringValue(addr.AddressName),
			AddressIp:               getStringValue(addr.AddressIp),
			AddressStatus:           getStringValue(addr.AddressStatus),
			InstanceId:              getStringValue(addr.InstanceId),
			InstanceType:            getStringValue(addr.InstanceType),
			NetworkInterfaceId:      getStringValue(addr.NetworkInterfaceId),
			PrivateAddressIp:        getStringValue(addr.PrivateAddressIp),
			Bandwidth:               getUint64Value(addr.Bandwidth),
			InternetChargeType:      getStringValue(addr.InternetChargeType),
			InternetServiceProvider: getStringValue(addr.InternetServiceProvider),
			CreatedTime:             getStringValue(addr.CreatedTime),
			BandwidthPackageId:      getStringValue(addr.BandwidthPackageId),
		}
		result.Addresses = append(result.Addresses, info)
	}

	c.logger.WithField("address_count", len(result.Addresses)).Info("成功查询弹性公网IP列表")
	return result, nil
}

func (c *Client) FormatAddressesAsJSON(result *DescribeAddressesResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatAddressesAsTable(result *DescribeAddressesResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("弹性公网IP列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 140) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-20s %-16s %-12s %-22s %-18s %-10s %-26s\n",
		"EIP ID", "名称", "公网IP", "状态", "绑定实例", "内网IP", "带宽(M)", "创建时间"))
	sb.WriteString(strings.Repeat("-", 140) + "\n")

	for _, addr := range result.Addresses {
		instanceId := addr.InstanceId
		if instanceId == "" {
			instanceId = "-"
		}
		privateIp := addr.PrivateAddressIp
		if privateIp == "" {
			privateIp = "-"
		}
		sb.WriteString(fmt.Sprintf("%-22s %-20s %-16s %-12s %-22s %-18s %-10d %-26s\n",
			addr.AddressId,
			truncateString(addr.AddressName, 18),
			addr.AddressIp,
			addr.AddressStatus,
			truncateString(instanceId, 20),
			privateIp,
			addr.Bandwidth,
			addr.CreatedTime))
	}

	return sb.String()
}

// ===================== DescribeBandwidthPackages =====================

// BandwidthPackageInfo 带宽包信息
type BandwidthPackageInfo struct {
	BandwidthPackageId   string `json:"bandwidth_package_id"`
	BandwidthPackageName string `json:"bandwidth_package_name"`
	NetworkType          string `json:"network_type"`
	ChargeType           string `json:"charge_type"`
	Bandwidth            int64  `json:"bandwidth"`
	Status               string `json:"status"`
	CreatedTime          string `json:"created_time"`
	Deadline             string `json:"deadline,omitempty"`
	ResourceCount        int    `json:"resource_count"`
}

// DescribeBandwidthPackagesResult 查询带宽包结果
type DescribeBandwidthPackagesResult struct {
	TotalCount        uint64                 `json:"total_count"`
	BandwidthPackages []BandwidthPackageInfo `json:"bandwidth_packages"`
	Region            string                 `json:"region"`
}

// DescribeBandwidthPackages 查询带宽包列表
func (c *Client) DescribeBandwidthPackages(ctx context.Context, region string) (*DescribeBandwidthPackagesResult, error) {
	c.logger.WithField("region", region).Debug("开始查询带宽包列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeBandwidthPackagesRequest()
	var limit uint64 = 100
	request.Limit = &limit

	response, err := client.DescribeBandwidthPackages(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询带宽包列表失败: %w", err)
	}

	result := &DescribeBandwidthPackagesResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, bp := range response.Response.BandwidthPackageSet {
		info := BandwidthPackageInfo{
			BandwidthPackageId:   getStringValue(bp.BandwidthPackageId),
			BandwidthPackageName: getStringValue(bp.BandwidthPackageName),
			NetworkType:          getStringValue(bp.NetworkType),
			ChargeType:           getStringValue(bp.ChargeType),
			Bandwidth:            getInt64Value(bp.Bandwidth),
			Status:               getStringValue(bp.Status),
			CreatedTime:          getStringValue(bp.CreatedTime),
			Deadline:             getStringValue(bp.Deadline),
			ResourceCount:        len(bp.ResourceSet),
		}
		result.BandwidthPackages = append(result.BandwidthPackages, info)
	}

	c.logger.WithField("bwp_count", len(result.BandwidthPackages)).Info("成功查询带宽包列表")
	return result, nil
}

func (c *Client) FormatBandwidthPackagesAsJSON(result *DescribeBandwidthPackagesResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatBandwidthPackagesAsTable(result *DescribeBandwidthPackagesResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("带宽包列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 140) + "\n")
	sb.WriteString(fmt.Sprintf("%-24s %-22s %-15s %-15s %-10s %-10s %-8s %-26s\n",
		"带宽包ID", "名称", "网络类型", "计费类型", "带宽(M)", "状态", "资源数", "创建时间"))
	sb.WriteString(strings.Repeat("-", 140) + "\n")

	for _, bp := range result.BandwidthPackages {
		sb.WriteString(fmt.Sprintf("%-24s %-22s %-15s %-15s %-10d %-10s %-8d %-26s\n",
			bp.BandwidthPackageId,
			truncateString(bp.BandwidthPackageName, 20),
			bp.NetworkType,
			bp.ChargeType,
			bp.Bandwidth,
			bp.Status,
			bp.ResourceCount,
			bp.CreatedTime))
	}

	return sb.String()
}

// ===================== DescribeVpcEndPoint =====================

// EndPointInfo 终端节点信息
type EndPointInfo struct {
	EndPointId        string `json:"endpoint_id"`
	EndPointName      string `json:"endpoint_name"`
	VpcId             string `json:"vpc_id"`
	SubnetId          string `json:"subnet_id"`
	EndPointVip       string `json:"endpoint_vip"`
	EndPointServiceId string `json:"endpoint_service_id"`
	ServiceVip        string `json:"service_vip,omitempty"`
	State             string `json:"state"`
	CreateTime        string `json:"create_time"`
}

// DescribeVpcEndPointResult 查询终端节点结果
type DescribeVpcEndPointResult struct {
	TotalCount uint64         `json:"total_count"`
	EndPoints  []EndPointInfo `json:"endpoints"`
	Region     string         `json:"region"`
}

// DescribeVpcEndPoint 查询终端节点列表
func (c *Client) DescribeVpcEndPoint(ctx context.Context, region string) (*DescribeVpcEndPointResult, error) {
	c.logger.WithField("region", region).Debug("开始查询终端节点列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeVpcEndPointRequest()
	var limit uint64 = 100
	request.Limit = &limit

	response, err := client.DescribeVpcEndPoint(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询终端节点列表失败: %w", err)
	}

	result := &DescribeVpcEndPointResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, ep := range response.Response.EndPointSet {
		info := EndPointInfo{
			EndPointId:        getStringValue(ep.EndPointId),
			EndPointName:      getStringValue(ep.EndPointName),
			VpcId:             getStringValue(ep.VpcId),
			SubnetId:          getStringValue(ep.SubnetId),
			EndPointVip:       getStringValue(ep.EndPointVip),
			EndPointServiceId: getStringValue(ep.EndPointServiceId),
			ServiceVip:        getStringValue(ep.ServiceVip),
			State:             getStringValue(ep.State),
			CreateTime:        getStringValue(ep.CreateTime),
		}
		result.EndPoints = append(result.EndPoints, info)
	}

	c.logger.WithField("endpoint_count", len(result.EndPoints)).Info("成功查询终端节点列表")
	return result, nil
}

func (c *Client) FormatVpcEndPointAsJSON(result *DescribeVpcEndPointResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatVpcEndPointAsTable(result *DescribeVpcEndPointResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("终端节点列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 140) + "\n")
	sb.WriteString(fmt.Sprintf("%-22s %-20s %-22s %-16s %-24s %-10s %-26s\n",
		"终端节点ID", "名称", "VPC ID", "VIP", "服务ID", "状态", "创建时间"))
	sb.WriteString(strings.Repeat("-", 140) + "\n")

	for _, ep := range result.EndPoints {
		sb.WriteString(fmt.Sprintf("%-22s %-20s %-22s %-16s %-24s %-10s %-26s\n",
			ep.EndPointId,
			truncateString(ep.EndPointName, 18),
			ep.VpcId,
			ep.EndPointVip,
			ep.EndPointServiceId,
			ep.State,
			ep.CreateTime))
	}

	return sb.String()
}

// ===================== DescribeVpcEndPointService =====================

// EndPointServiceInfo 终端节点服务信息
type EndPointServiceInfo struct {
	EndPointServiceId string `json:"endpoint_service_id"`
	ServiceName       string `json:"service_name"`
	VpcId             string `json:"vpc_id"`
	ServiceVip        string `json:"service_vip"`
	ServiceInstanceId string `json:"service_instance_id"`
	ServiceType       string `json:"service_type"`
	AutoAcceptFlag    bool   `json:"auto_accept_flag"`
	EndPointCount     uint64 `json:"endpoint_count"`
	CreateTime        string `json:"create_time"`
}

// DescribeVpcEndPointServiceResult 查询终端节点服务结果
type DescribeVpcEndPointServiceResult struct {
	TotalCount       uint64                `json:"total_count"`
	EndPointServices []EndPointServiceInfo `json:"endpoint_services"`
	Region           string                `json:"region"`
}

// DescribeVpcEndPointService 查询终端节点服务列表
func (c *Client) DescribeVpcEndPointService(ctx context.Context, region string) (*DescribeVpcEndPointServiceResult, error) {
	c.logger.WithField("region", region).Debug("开始查询终端节点服务列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeVpcEndPointServiceRequest()
	var limit uint64 = 100
	request.Limit = &limit

	response, err := client.DescribeVpcEndPointService(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询终端节点服务列表失败: %w", err)
	}

	result := &DescribeVpcEndPointServiceResult{
		TotalCount: getUint64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, eps := range response.Response.EndPointServiceSet {
		info := EndPointServiceInfo{
			EndPointServiceId: getStringValue(eps.EndPointServiceId),
			ServiceName:       getStringValue(eps.ServiceName),
			VpcId:             getStringValue(eps.VpcId),
			ServiceVip:        getStringValue(eps.ServiceVip),
			ServiceInstanceId: getStringValue(eps.ServiceInstanceId),
			ServiceType:       getStringValue(eps.ServiceType),
			AutoAcceptFlag:    getBoolValue(eps.AutoAcceptFlag),
			EndPointCount:     getUint64Value(eps.EndPointCount),
			CreateTime:        getStringValue(eps.CreateTime),
		}
		result.EndPointServices = append(result.EndPointServices, info)
	}

	c.logger.WithField("eps_count", len(result.EndPointServices)).Info("成功查询终端节点服务列表")
	return result, nil
}

func (c *Client) FormatVpcEndPointServiceAsJSON(result *DescribeVpcEndPointServiceResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatVpcEndPointServiceAsTable(result *DescribeVpcEndPointServiceResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("终端节点服务列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 150) + "\n")
	sb.WriteString(fmt.Sprintf("%-24s %-22s %-22s %-16s %-12s %-10s %-10s %-26s\n",
		"服务ID", "名称", "VPC ID", "VIP", "服务类型", "自动接受", "节点数", "创建时间"))
	sb.WriteString(strings.Repeat("-", 150) + "\n")

	for _, eps := range result.EndPointServices {
		autoAccept := "否"
		if eps.AutoAcceptFlag {
			autoAccept = "是"
		}
		sb.WriteString(fmt.Sprintf("%-24s %-22s %-22s %-16s %-12s %-10s %-10d %-26s\n",
			eps.EndPointServiceId,
			truncateString(eps.ServiceName, 20),
			eps.VpcId,
			eps.ServiceVip,
			eps.ServiceType,
			autoAccept,
			eps.EndPointCount,
			eps.CreateTime))
	}

	return sb.String()
}

// ===================== DescribeVpcPeeringConnections =====================

// PeerConnectionInfo 对等连接信息
type PeerConnectionInfo struct {
	PeeringConnectionId   string `json:"peering_connection_id"`
	PeeringConnectionName string `json:"peering_connection_name"`
	SourceVpcId           string `json:"source_vpc_id"`
	PeerVpcId             string `json:"peer_vpc_id"`
	DestinationVpcId      string `json:"destination_vpc_id,omitempty"`
	SourceRegion          string `json:"source_region"`
	DestinationRegion     string `json:"destination_region"`
	State                 string `json:"state"`
	Bandwidth             int64  `json:"bandwidth"`
	ChargeType            string `json:"charge_type,omitempty"`
	CreateTime            string `json:"create_time"`
}

// DescribeVpcPeeringConnectionsResult 查询对等连接结果
type DescribeVpcPeeringConnectionsResult struct {
	TotalCount      int64                `json:"total_count"`
	PeerConnections []PeerConnectionInfo `json:"peer_connections"`
	Region          string               `json:"region"`
}

// DescribeVpcPeeringConnections 查询对等连接列表
func (c *Client) DescribeVpcPeeringConnections(ctx context.Context, region string) (*DescribeVpcPeeringConnectionsResult, error) {
	c.logger.WithField("region", region).Debug("开始查询对等连接列表")

	client, err := c.newRegionClient(region)
	if err != nil {
		return nil, fmt.Errorf("创建 VPC 客户端失败: %w", err)
	}

	request := vpc.NewDescribeVpcPeeringConnectionsRequest()
	var limit int64 = 100
	request.Limit = &limit

	response, err := client.DescribeVpcPeeringConnections(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("VPC API 错误 [%s]: %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("查询对等连接列表失败: %w", err)
	}

	result := &DescribeVpcPeeringConnectionsResult{
		TotalCount: getInt64Value(response.Response.TotalCount),
		Region:     region,
	}

	for _, pc := range response.Response.PeerConnectionSet {
		info := PeerConnectionInfo{
			PeeringConnectionId:   getStringValue(pc.PeeringConnectionId),
			PeeringConnectionName: getStringValue(pc.PeeringConnectionName),
			SourceVpcId:           getStringValue(pc.SourceVpcId),
			PeerVpcId:             getStringValue(pc.PeerVpcId),
			DestinationVpcId:      getStringValue(pc.DestinationVpcId),
			SourceRegion:          getStringValue(pc.SourceRegion),
			DestinationRegion:     getStringValue(pc.DestinationRegion),
			State:                 getStringValue(pc.State),
			Bandwidth:             getInt64Value(pc.Bandwidth),
			ChargeType:            getStringValue(pc.ChargeType),
			CreateTime:            getStringValue(pc.CreateTime),
		}
		result.PeerConnections = append(result.PeerConnections, info)
	}

	c.logger.WithField("peering_count", len(result.PeerConnections)).Info("成功查询对等连接列表")
	return result, nil
}

func (c *Client) FormatVpcPeeringConnectionsAsJSON(result *DescribeVpcPeeringConnectionsResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return string(data), nil
}

func (c *Client) FormatVpcPeeringConnectionsAsTable(result *DescribeVpcPeeringConnectionsResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("对等连接列表 (地域: %s, 总数: %d)\n", result.Region, result.TotalCount))
	sb.WriteString(strings.Repeat("=", 160) + "\n")
	sb.WriteString(fmt.Sprintf("%-24s %-20s %-22s %-22s %-15s %-15s %-10s %-10s %-26s\n",
		"对等连接ID", "名称", "本端VPC", "对端VPC", "本端地域", "对端地域", "状态", "带宽(M)", "创建时间"))
	sb.WriteString(strings.Repeat("-", 160) + "\n")

	for _, pc := range result.PeerConnections {
		sb.WriteString(fmt.Sprintf("%-24s %-20s %-22s %-22s %-15s %-15s %-10s %-10d %-26s\n",
			pc.PeeringConnectionId,
			truncateString(pc.PeeringConnectionName, 18),
			pc.SourceVpcId,
			pc.PeerVpcId,
			pc.SourceRegion,
			pc.DestinationRegion,
			pc.State,
			pc.Bandwidth,
			pc.CreateTime))
	}

	return sb.String()
}
