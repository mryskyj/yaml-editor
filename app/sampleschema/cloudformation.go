package sampleschema

type sampleCloudFormation struct {
	TemplateFormatVersion string                     `yaml:"template_format_version,omitempty" default:"2010-09-09" desc:"CloudFormation template format version"`
	Description           string                     `yaml:"description,omitempty" desc:"template description"`
	Parameters            map[string]sampleParameter `yaml:"parameters,omitempty" desc:"CloudFormation-style template parameters"`
	Resources             sampleResources            `yaml:"resources,omitempty" desc:"resource declarations inspired by CloudFormation"`
	Outputs               map[string]sampleOutput    `yaml:"outputs,omitempty" desc:"exported stack outputs"`
	StackPolicy           sampleStackPolicy          `yaml:"stack_policy,omitempty" desc:"stack update protection policy"`
}

type sampleParameter struct {
	Type           string   `yaml:"type" required:"true" default:"String" enum:"String,Number,List<String>,CommaDelimitedList,AWS::EC2::KeyPair::KeyName,AWS::EC2::Subnet::Id,AWS::EC2::VPC::Id" desc:"CloudFormation parameter type"`
	Description    string   `yaml:"description,omitempty" desc:"parameter description"`
	Default        string   `yaml:"default,omitempty" desc:"default value"`
	AllowedValues  []string `yaml:"allowed_values,omitempty" desc:"allowed values"`
	AllowedPattern string   `yaml:"allowed_pattern,omitempty" desc:"regular expression constraint"`
	NoEcho         bool     `yaml:"no_echo,omitempty" default:"false" desc:"mask value in stack events"`
}

type sampleResources struct {
	LaunchTemplate   sampleLaunchTemplate   `yaml:"launch_template,omitempty" desc:"AWS::EC2::LaunchTemplate-like settings"`
	AutoScalingGroup sampleAutoScalingGroup `yaml:"auto_scaling_group,omitempty" desc:"AWS::AutoScaling::AutoScalingGroup-like settings"`
	LoadBalancer     sampleLoadBalancer     `yaml:"load_balancer,omitempty" desc:"Application Load Balancer settings"`
	TargetGroups     []sampleTargetGroup    `yaml:"target_groups,omitempty" desc:"load balancer target groups"`
	SecurityGroups   []sampleSecurityGroup  `yaml:"security_groups,omitempty" desc:"network security groups"`
}

type sampleLaunchTemplate struct {
	LaunchTemplateName string                        `yaml:"launch_template_name,omitempty" desc:"launch template name"`
	VersionDescription string                        `yaml:"version_description,omitempty" desc:"first version description"`
	LaunchTemplateData sampleLaunchTemplateData      `yaml:"launch_template_data" required:"true" desc:"instance launch settings"`
	TagSpecifications  []sampleTagSpecification      `yaml:"tag_specifications,omitempty" desc:"tags for launched resources"`
	MetadataOptions    sampleInstanceMetadataOptions `yaml:"metadata_options,omitempty" desc:"EC2 instance metadata options"`
}

type sampleLaunchTemplateData struct {
	ImageID               string                      `yaml:"image_id" required:"true" desc:"AMI ID"`
	InstanceType          string                      `yaml:"instance_type" required:"true" default:"t3.micro" enum:"t3.micro,t3.small,t3.medium,m6i.large,m6i.xlarge,c7g.large" desc:"EC2 instance type"`
	KeyName               string                      `yaml:"key_name,omitempty" desc:"EC2 key pair name"`
	IAMInstanceProfile    sampleIAMInstanceProfile    `yaml:"iam_instance_profile,omitempty" desc:"IAM instance profile"`
	SecurityGroupIDs      []string                    `yaml:"security_group_ids,omitempty" desc:"security group IDs"`
	BlockDeviceMappings   []sampleBlockDeviceMapping  `yaml:"block_device_mappings,omitempty" desc:"EBS block device mappings"`
	CreditSpecification   sampleCreditSpecification   `yaml:"credit_specification,omitempty" desc:"burstable CPU credit behavior"`
	Monitoring            sampleDetailedMonitoring    `yaml:"monitoring,omitempty" desc:"detailed CloudWatch monitoring"`
	NetworkInterfaces     []sampleNetworkInterface    `yaml:"network_interfaces,omitempty" desc:"network interface settings"`
	UserData              string                      `yaml:"user_data,omitempty" desc:"base64-encoded user data"`
	InstanceMarketOptions sampleInstanceMarketOptions `yaml:"instance_market_options,omitempty" desc:"spot or on-demand market settings"`
	PrivateDNSNameOptions samplePrivateDNSNameOptions `yaml:"private_dns_name_options,omitempty" desc:"private DNS name behavior"`
	DisableAPITermination bool                        `yaml:"disable_api_termination,omitempty" default:"false" desc:"prevent accidental instance termination"`
	EBSOptimized          bool                        `yaml:"ebs_optimized,omitempty" default:"true" desc:"enable EBS optimization"`
}

type sampleIAMInstanceProfile struct {
	Arn  string `yaml:"arn,omitempty" desc:"instance profile ARN"`
	Name string `yaml:"name,omitempty" desc:"instance profile name"`
}

type sampleBlockDeviceMapping struct {
	DeviceName string    `yaml:"device_name" required:"true" default:"/dev/xvda" desc:"device name"`
	EBS        sampleEBS `yaml:"ebs" required:"true" desc:"EBS volume settings"`
}

type sampleEBS struct {
	VolumeSize          int    `yaml:"volume_size" required:"true" default:"30" desc:"volume size in GiB"`
	VolumeType          string `yaml:"volume_type" required:"true" default:"gp3" enum:"gp2,gp3,io1,io2,st1,sc1" desc:"EBS volume type"`
	DeleteOnTermination bool   `yaml:"delete_on_termination,omitempty" default:"true" desc:"delete volume when instance terminates"`
	Encrypted           bool   `yaml:"encrypted,omitempty" default:"true" desc:"encrypt EBS volume"`
	KMSKeyID            string `yaml:"kms_key_id,omitempty" desc:"KMS key ID for encryption"`
	IOPS                int    `yaml:"iops,omitempty" default:"3000" desc:"provisioned IOPS"`
	Throughput          int    `yaml:"throughput,omitempty" default:"125" desc:"gp3 throughput MiB/s"`
}

type sampleCreditSpecification struct {
	CPUCredits string `yaml:"cpu_credits,omitempty" default:"standard" enum:"standard,unlimited" desc:"burstable CPU credit mode"`
}

type sampleDetailedMonitoring struct {
	Enabled bool `yaml:"enabled,omitempty" default:"true" desc:"enable detailed monitoring"`
}

type sampleNetworkInterface struct {
	DeviceIndex              int      `yaml:"device_index,omitempty" default:"0" desc:"network interface device index"`
	SubnetID                 string   `yaml:"subnet_id,omitempty" desc:"subnet ID"`
	Groups                   []string `yaml:"groups,omitempty" desc:"security groups"`
	AssociatePublicIPAddress bool     `yaml:"associate_public_ip_address,omitempty" default:"false" desc:"assign public IP address"`
	DeleteOnTermination      bool     `yaml:"delete_on_termination,omitempty" default:"true" desc:"delete interface on termination"`
}

type sampleInstanceMarketOptions struct {
	MarketType  string            `yaml:"market_type,omitempty" default:"on-demand" enum:"on-demand,spot" desc:"instance market type"`
	SpotOptions sampleSpotOptions `yaml:"spot_options,omitempty" desc:"Spot instance options"`
}

type sampleSpotOptions struct {
	MaxPrice                     string `yaml:"max_price,omitempty" desc:"maximum spot price"`
	SpotInstanceType             string `yaml:"spot_instance_type,omitempty" default:"one-time" enum:"one-time,persistent" desc:"spot request type"`
	InstanceInterruptionBehavior string `yaml:"instance_interruption_behavior,omitempty" default:"terminate" enum:"hibernate,stop,terminate" desc:"spot interruption behavior"`
	BlockDurationMinutes         int    `yaml:"block_duration_minutes,omitempty" desc:"spot block duration"`
}

type samplePrivateDNSNameOptions struct {
	HostnameType                    string `yaml:"hostname_type,omitempty" default:"ip-name" enum:"ip-name,resource-name" desc:"private hostname type"`
	EnableResourceNameDNSARecord    bool   `yaml:"enable_resource_name_dns_a_record,omitempty" default:"true" desc:"publish A record"`
	EnableResourceNameDNSAAAARecord bool   `yaml:"enable_resource_name_dns_aaaa_record,omitempty" default:"false" desc:"publish AAAA record"`
}

type sampleInstanceMetadataOptions struct {
	HTTPEndpoint            string `yaml:"http_endpoint,omitempty" default:"enabled" enum:"enabled,disabled" desc:"IMDS endpoint state"`
	HTTPTokens              string `yaml:"http_tokens,omitempty" default:"required" enum:"optional,required" desc:"IMDSv2 token requirement"`
	HTTPPutResponseHopLimit int    `yaml:"http_put_response_hop_limit,omitempty" default:"2" desc:"metadata response hop limit"`
	InstanceMetadataTags    string `yaml:"instance_metadata_tags,omitempty" default:"enabled" enum:"enabled,disabled" desc:"allow instance tags in metadata"`
}

type sampleTagSpecification struct {
	ResourceType string            `yaml:"resource_type" required:"true" enum:"instance,volume,network-interface,spot-instances-request" desc:"resource type to tag"`
	Tags         map[string]string `yaml:"tags,omitempty" desc:"resource tags"`
}

type sampleAutoScalingGroup struct {
	AutoScalingGroupName             string                        `yaml:"auto_scaling_group_name,omitempty" desc:"Auto Scaling group name"`
	MinSize                          int                           `yaml:"min_size" required:"true" default:"2" desc:"minimum capacity"`
	MaxSize                          int                           `yaml:"max_size" required:"true" default:"6" desc:"maximum capacity"`
	DesiredCapacity                  int                           `yaml:"desired_capacity,omitempty" default:"3" desc:"desired capacity"`
	DesiredCapacityType              string                        `yaml:"desired_capacity_type,omitempty" default:"units" enum:"units,vcpu,memory-mib" desc:"capacity unit"`
	HealthCheckType                  string                        `yaml:"health_check_type,omitempty" default:"ELB" enum:"EC2,ELB" desc:"health check source"`
	HealthCheckGracePeriod           int                           `yaml:"health_check_grace_period,omitempty" default:"300" desc:"health check grace period seconds"`
	DefaultInstanceWarmup            int                           `yaml:"default_instance_warmup,omitempty" default:"300" desc:"instance warmup seconds"`
	VPCZoneIdentifier                []string                      `yaml:"vpc_zone_identifier" required:"true" desc:"subnet IDs"`
	LaunchTemplate                   sampleLaunchTemplateReference `yaml:"launch_template,omitempty" desc:"launch template reference"`
	MixedInstancesPolicy             sampleMixedInstancesPolicy    `yaml:"mixed_instances_policy,omitempty" desc:"mixed instance policy"`
	TargetGroupARNs                  []string                      `yaml:"target_group_arns,omitempty" desc:"load balancer target groups"`
	TerminationPolicies              []string                      `yaml:"termination_policies,omitempty" desc:"instance termination policies"`
	NewInstancesProtectedFromScaleIn bool                          `yaml:"new_instances_protected_from_scale_in,omitempty" default:"false" desc:"protect new instances from scale-in"`
	CapacityRebalance                bool                          `yaml:"capacity_rebalance,omitempty" default:"true" desc:"replace spot instances before interruption"`
	MetricsCollection                []sampleMetricsCollection     `yaml:"metrics_collection,omitempty" desc:"Auto Scaling metrics collection"`
}

type sampleLaunchTemplateReference struct {
	LaunchTemplateID   string `yaml:"launch_template_id,omitempty" desc:"launch template ID"`
	LaunchTemplateName string `yaml:"launch_template_name,omitempty" desc:"launch template name"`
	Version            string `yaml:"version,omitempty" default:"$Latest" desc:"launch template version"`
}

type sampleMixedInstancesPolicy struct {
	InstancesDistribution sampleInstancesDistribution `yaml:"instances_distribution,omitempty" desc:"on-demand and spot distribution"`
	Overrides             []sampleInstanceOverride    `yaml:"overrides,omitempty" desc:"instance type overrides"`
}

type sampleInstancesDistribution struct {
	OnDemandAllocationStrategy          string `yaml:"on_demand_allocation_strategy,omitempty" default:"prioritized" enum:"prioritized,lowest-price" desc:"on-demand allocation strategy"`
	OnDemandBaseCapacity                int    `yaml:"on_demand_base_capacity,omitempty" default:"1" desc:"base on-demand capacity"`
	OnDemandPercentageAboveBaseCapacity int    `yaml:"on_demand_percentage_above_base_capacity,omitempty" default:"50" desc:"on-demand percentage above base"`
	SpotAllocationStrategy              string `yaml:"spot_allocation_strategy,omitempty" default:"capacity-optimized" enum:"lowest-price,capacity-optimized,price-capacity-optimized" desc:"spot allocation strategy"`
	SpotInstancePools                   int    `yaml:"spot_instance_pools,omitempty" default:"2" desc:"number of spot pools"`
	SpotMaxPrice                        string `yaml:"spot_max_price,omitempty" desc:"maximum spot price"`
}

type sampleInstanceOverride struct {
	InstanceType     string `yaml:"instance_type" required:"true" desc:"override instance type"`
	WeightedCapacity string `yaml:"weighted_capacity,omitempty" default:"1" desc:"weighted capacity"`
}

type sampleMetricsCollection struct {
	Granularity string   `yaml:"granularity,omitempty" default:"1Minute" enum:"1Minute" desc:"metrics granularity"`
	Metrics     []string `yaml:"metrics,omitempty" desc:"Auto Scaling metrics to collect"`
}

type sampleLoadBalancer struct {
	Name           string            `yaml:"name,omitempty" desc:"load balancer name"`
	Type           string            `yaml:"type,omitempty" default:"application" enum:"application,network,gateway" desc:"load balancer type"`
	Scheme         string            `yaml:"scheme,omitempty" default:"internet-facing" enum:"internet-facing,internal" desc:"load balancer scheme"`
	Subnets        []string          `yaml:"subnets,omitempty" desc:"subnet IDs"`
	SecurityGroups []string          `yaml:"security_groups,omitempty" desc:"security group IDs"`
	Attributes     map[string]string `yaml:"attributes,omitempty" desc:"load balancer attributes"`
}

type sampleTargetGroup struct {
	Name               string `yaml:"name" required:"true" desc:"target group name"`
	Protocol           string `yaml:"protocol,omitempty" default:"HTTP" enum:"HTTP,HTTPS,TCP,TLS,UDP" desc:"target group protocol"`
	Port               int    `yaml:"port,omitempty" default:"80" desc:"target group port"`
	TargetType         string `yaml:"target_type,omitempty" default:"ip" enum:"instance,ip,lambda,alb" desc:"target type"`
	HealthCheckPath    string `yaml:"health_check_path,omitempty" default:"/health" desc:"health check path"`
	HealthCheckEnabled bool   `yaml:"health_check_enabled,omitempty" default:"true" desc:"enable health checks"`
}

type sampleSecurityGroup struct {
	GroupName   string               `yaml:"group_name" required:"true" desc:"security group name"`
	Description string               `yaml:"description,omitempty" desc:"security group description"`
	VpcID       string               `yaml:"vpc_id,omitempty" desc:"VPC ID"`
	Ingress     []sampleSecurityRule `yaml:"ingress,omitempty" desc:"inbound rules"`
	Egress      []sampleSecurityRule `yaml:"egress,omitempty" desc:"outbound rules"`
	Tags        map[string]string    `yaml:"tags,omitempty" desc:"security group tags"`
}

type sampleSecurityRule struct {
	Protocol    string `yaml:"protocol,omitempty" default:"tcp" enum:"tcp,udp,icmp,-1" desc:"IP protocol"`
	FromPort    int    `yaml:"from_port,omitempty" desc:"start port"`
	ToPort      int    `yaml:"to_port,omitempty" desc:"end port"`
	CIDRIP      string `yaml:"cidr_ip,omitempty" desc:"IPv4 CIDR"`
	CIDRIPv6    string `yaml:"cidr_ipv6,omitempty" desc:"IPv6 CIDR"`
	Description string `yaml:"description,omitempty" desc:"rule description"`
}

type sampleOutput struct {
	Description string `yaml:"description,omitempty" desc:"output description"`
	Value       string `yaml:"value" required:"true" desc:"output value"`
	ExportName  string `yaml:"export_name,omitempty" desc:"CloudFormation export name"`
}

type sampleStackPolicy struct {
	ProtectedLogicalIDs []string `yaml:"protected_logical_ids,omitempty" desc:"resources protected from replacement"`
	DenyDelete          bool     `yaml:"deny_delete,omitempty" default:"true" desc:"deny delete operations"`
	DenyReplace         bool     `yaml:"deny_replace,omitempty" default:"true" desc:"deny replacement operations"`
}
