package sampleschema

type sampleCloudFormation struct {
	TemplateFormatVersion string                     `yaml:"template_format_version" default:"2010-09-09" desc:"CloudFormation template format version"`
	Description           string                     `yaml:"description" desc:"template description"`
	Parameters            map[string]sampleParameter `yaml:"parameters" desc:"CloudFormation-style template parameters"`
	Resources             sampleResources            `yaml:"resources" desc:"resource declarations inspired by CloudFormation"`
	Outputs               map[string]sampleOutput    `yaml:"outputs" desc:"exported stack outputs"`
	StackPolicy           sampleStackPolicy          `yaml:"stack_policy" desc:"stack update protection policy"`
}

type sampleParameter struct {
	Type           string   `yaml:"type" required:"true" default:"String" enum:"String,Number,List<String>,CommaDelimitedList,AWS::EC2::KeyPair::KeyName,AWS::EC2::Subnet::Id,AWS::EC2::VPC::Id" desc:"CloudFormation parameter type"`
	Description    string   `yaml:"description" desc:"parameter description"`
	Default        string   `yaml:"default" desc:"default value"`
	AllowedValues  []string `yaml:"allowed_values" desc:"allowed values"`
	AllowedPattern string   `yaml:"allowed_pattern" desc:"regular expression constraint"`
	NoEcho         bool     `yaml:"no_echo" default:"false" desc:"mask value in stack events"`
}

type sampleResources struct {
	LaunchTemplate   sampleLaunchTemplate   `yaml:"launch_template" desc:"AWS::EC2::LaunchTemplate-like settings"`
	AutoScalingGroup sampleAutoScalingGroup `yaml:"auto_scaling_group" desc:"AWS::AutoScaling::AutoScalingGroup-like settings"`
	LoadBalancer     sampleLoadBalancer     `yaml:"load_balancer" desc:"Application Load Balancer settings"`
	TargetGroups     []sampleTargetGroup    `yaml:"target_groups" desc:"load balancer target groups"`
	SecurityGroups   []sampleSecurityGroup  `yaml:"security_groups" desc:"network security groups"`
}

type sampleLaunchTemplate struct {
	LaunchTemplateName string                        `yaml:"launch_template_name" desc:"launch template name"`
	VersionDescription string                        `yaml:"version_description" desc:"first version description"`
	LaunchTemplateData sampleLaunchTemplateData      `yaml:"launch_template_data" required:"true" desc:"instance launch settings"`
	TagSpecifications  []sampleTagSpecification      `yaml:"tag_specifications" desc:"tags for launched resources"`
	MetadataOptions    sampleInstanceMetadataOptions `yaml:"metadata_options" desc:"EC2 instance metadata options"`
}

type sampleLaunchTemplateData struct {
	ImageID               string                      `yaml:"image_id" required:"true" desc:"AMI ID"`
	InstanceType          string                      `yaml:"instance_type" required:"true" default:"t3.micro" enum:"t3.micro,t3.small,t3.medium,m6i.large,m6i.xlarge,c7g.large" desc:"EC2 instance type"`
	KeyName               string                      `yaml:"key_name" desc:"EC2 key pair name"`
	IAMInstanceProfile    sampleIAMInstanceProfile    `yaml:"iam_instance_profile" desc:"IAM instance profile"`
	SecurityGroupIDs      []string                    `yaml:"security_group_ids" desc:"security group IDs"`
	BlockDeviceMappings   []sampleBlockDeviceMapping  `yaml:"block_device_mappings" desc:"EBS block device mappings"`
	CreditSpecification   sampleCreditSpecification   `yaml:"credit_specification" desc:"burstable CPU credit behavior"`
	Monitoring            sampleDetailedMonitoring    `yaml:"monitoring" desc:"detailed CloudWatch monitoring"`
	NetworkInterfaces     []sampleNetworkInterface    `yaml:"network_interfaces" desc:"network interface settings"`
	UserData              string                      `yaml:"user_data" desc:"base64-encoded user data"`
	InstanceMarketOptions sampleInstanceMarketOptions `yaml:"instance_market_options" desc:"spot or on-demand market settings"`
	PrivateDNSNameOptions samplePrivateDNSNameOptions `yaml:"private_dns_name_options" desc:"private DNS name behavior"`
	DisableAPITermination bool                        `yaml:"disable_api_termination" default:"false" desc:"prevent accidental instance termination"`
	EBSOptimized          bool                        `yaml:"ebs_optimized" default:"true" desc:"enable EBS optimization"`
}

type sampleIAMInstanceProfile struct {
	Arn  string `yaml:"arn" desc:"instance profile ARN"`
	Name string `yaml:"name" desc:"instance profile name"`
}

type sampleBlockDeviceMapping struct {
	DeviceName string    `yaml:"device_name" required:"true" default:"/dev/xvda" desc:"device name"`
	EBS        sampleEBS `yaml:"ebs" required:"true" desc:"EBS volume settings"`
}

type sampleEBS struct {
	VolumeSize          int    `yaml:"volume_size" required:"true" default:"30" desc:"volume size in GiB"`
	VolumeType          string `yaml:"volume_type" required:"true" default:"gp3" enum:"gp2,gp3,io1,io2,st1,sc1" desc:"EBS volume type"`
	DeleteOnTermination bool   `yaml:"delete_on_termination" default:"true" desc:"delete volume when instance terminates"`
	Encrypted           bool   `yaml:"encrypted" default:"true" desc:"encrypt EBS volume"`
	KMSKeyID            string `yaml:"kms_key_id" desc:"KMS key ID for encryption"`
	IOPS                int    `yaml:"iops" default:"3000" desc:"provisioned IOPS"`
	Throughput          int    `yaml:"throughput" default:"125" desc:"gp3 throughput MiB/s"`
}

type sampleCreditSpecification struct {
	CPUCredits string `yaml:"cpu_credits" default:"standard" enum:"standard,unlimited" desc:"burstable CPU credit mode"`
}

type sampleDetailedMonitoring struct {
	Enabled bool `yaml:"enabled" default:"true" desc:"enable detailed monitoring"`
}

type sampleNetworkInterface struct {
	DeviceIndex              int      `yaml:"device_index" default:"0" desc:"network interface device index"`
	SubnetID                 string   `yaml:"subnet_id" desc:"subnet ID"`
	Groups                   []string `yaml:"groups" desc:"security groups"`
	AssociatePublicIPAddress bool     `yaml:"associate_public_ip_address" default:"false" desc:"assign public IP address"`
	DeleteOnTermination      bool     `yaml:"delete_on_termination" default:"true" desc:"delete interface on termination"`
}

type sampleInstanceMarketOptions struct {
	MarketType  string            `yaml:"market_type" default:"on-demand" enum:"on-demand,spot" desc:"instance market type"`
	SpotOptions sampleSpotOptions `yaml:"spot_options" desc:"Spot instance options"`
}

type sampleSpotOptions struct {
	MaxPrice                     string `yaml:"max_price" desc:"maximum spot price"`
	SpotInstanceType             string `yaml:"spot_instance_type" default:"one-time" enum:"one-time,persistent" desc:"spot request type"`
	InstanceInterruptionBehavior string `yaml:"instance_interruption_behavior" default:"terminate" enum:"hibernate,stop,terminate" desc:"spot interruption behavior"`
	BlockDurationMinutes         int    `yaml:"block_duration_minutes" desc:"spot block duration"`
}

type samplePrivateDNSNameOptions struct {
	HostnameType                    string `yaml:"hostname_type" default:"ip-name" enum:"ip-name,resource-name" desc:"private hostname type"`
	EnableResourceNameDNSARecord    bool   `yaml:"enable_resource_name_dns_a_record" default:"true" desc:"publish A record"`
	EnableResourceNameDNSAAAARecord bool   `yaml:"enable_resource_name_dns_aaaa_record" default:"false" desc:"publish AAAA record"`
}

type sampleInstanceMetadataOptions struct {
	HTTPEndpoint            string `yaml:"http_endpoint" default:"enabled" enum:"enabled,disabled" desc:"IMDS endpoint state"`
	HTTPTokens              string `yaml:"http_tokens" default:"required" enum:"optional,required" desc:"IMDSv2 token requirement"`
	HTTPPutResponseHopLimit int    `yaml:"http_put_response_hop_limit" default:"2" desc:"metadata response hop limit"`
	InstanceMetadataTags    string `yaml:"instance_metadata_tags" default:"enabled" enum:"enabled,disabled" desc:"allow instance tags in metadata"`
}

type sampleTagSpecification struct {
	ResourceType string            `yaml:"resource_type" required:"true" enum:"instance,volume,network-interface,spot-instances-request" desc:"resource type to tag"`
	Tags         map[string]string `yaml:"tags" desc:"resource tags"`
}

type sampleAutoScalingGroup struct {
	AutoScalingGroupName             string                        `yaml:"auto_scaling_group_name" desc:"Auto Scaling group name"`
	MinSize                          int                           `yaml:"min_size" required:"true" default:"2" desc:"minimum capacity"`
	MaxSize                          int                           `yaml:"max_size" required:"true" default:"6" desc:"maximum capacity"`
	DesiredCapacity                  int                           `yaml:"desired_capacity" default:"3" desc:"desired capacity"`
	DesiredCapacityType              string                        `yaml:"desired_capacity_type" default:"units" enum:"units,vcpu,memory-mib" desc:"capacity unit"`
	HealthCheckType                  string                        `yaml:"health_check_type" default:"ELB" enum:"EC2,ELB" desc:"health check source"`
	HealthCheckGracePeriod           int                           `yaml:"health_check_grace_period" default:"300" desc:"health check grace period seconds"`
	DefaultInstanceWarmup            int                           `yaml:"default_instance_warmup" default:"300" desc:"instance warmup seconds"`
	VPCZoneIdentifier                []string                      `yaml:"vpc_zone_identifier" required:"true" desc:"subnet IDs"`
	LaunchTemplate                   sampleLaunchTemplateReference `yaml:"launch_template" desc:"launch template reference"`
	MixedInstancesPolicy             sampleMixedInstancesPolicy    `yaml:"mixed_instances_policy" desc:"mixed instance policy"`
	TargetGroupARNs                  []string                      `yaml:"target_group_arns" desc:"load balancer target groups"`
	TerminationPolicies              []string                      `yaml:"termination_policies" desc:"instance termination policies"`
	NewInstancesProtectedFromScaleIn bool                          `yaml:"new_instances_protected_from_scale_in" default:"false" desc:"protect new instances from scale-in"`
	CapacityRebalance                bool                          `yaml:"capacity_rebalance" default:"true" desc:"replace spot instances before interruption"`
	MetricsCollection                []sampleMetricsCollection     `yaml:"metrics_collection" desc:"Auto Scaling metrics collection"`
}

type sampleLaunchTemplateReference struct {
	LaunchTemplateID   string `yaml:"launch_template_id" desc:"launch template ID"`
	LaunchTemplateName string `yaml:"launch_template_name" desc:"launch template name"`
	Version            string `yaml:"version" default:"$Latest" desc:"launch template version"`
}

type sampleMixedInstancesPolicy struct {
	InstancesDistribution sampleInstancesDistribution `yaml:"instances_distribution" desc:"on-demand and spot distribution"`
	Overrides             []sampleInstanceOverride    `yaml:"overrides" desc:"instance type overrides"`
}

type sampleInstancesDistribution struct {
	OnDemandAllocationStrategy          string `yaml:"on_demand_allocation_strategy" default:"prioritized" enum:"prioritized,lowest-price" desc:"on-demand allocation strategy"`
	OnDemandBaseCapacity                int    `yaml:"on_demand_base_capacity" default:"1" desc:"base on-demand capacity"`
	OnDemandPercentageAboveBaseCapacity int    `yaml:"on_demand_percentage_above_base_capacity" default:"50" desc:"on-demand percentage above base"`
	SpotAllocationStrategy              string `yaml:"spot_allocation_strategy" default:"capacity-optimized" enum:"lowest-price,capacity-optimized,price-capacity-optimized" desc:"spot allocation strategy"`
	SpotInstancePools                   int    `yaml:"spot_instance_pools" default:"2" desc:"number of spot pools"`
	SpotMaxPrice                        string `yaml:"spot_max_price" desc:"maximum spot price"`
}

type sampleInstanceOverride struct {
	InstanceType     string `yaml:"instance_type" required:"true" desc:"override instance type"`
	WeightedCapacity string `yaml:"weighted_capacity" default:"1" desc:"weighted capacity"`
}

type sampleMetricsCollection struct {
	Granularity string   `yaml:"granularity" default:"1Minute" enum:"1Minute" desc:"metrics granularity"`
	Metrics     []string `yaml:"metrics" desc:"Auto Scaling metrics to collect"`
}

type sampleLoadBalancer struct {
	Name           string            `yaml:"name" desc:"load balancer name"`
	Type           string            `yaml:"type" default:"application" enum:"application,network,gateway" desc:"load balancer type"`
	Scheme         string            `yaml:"scheme" default:"internet-facing" enum:"internet-facing,internal" desc:"load balancer scheme"`
	Subnets        []string          `yaml:"subnets" desc:"subnet IDs"`
	SecurityGroups []string          `yaml:"security_groups" desc:"security group IDs"`
	Attributes     map[string]string `yaml:"attributes" desc:"load balancer attributes"`
}

type sampleTargetGroup struct {
	Name               string `yaml:"name" required:"true" desc:"target group name"`
	Protocol           string `yaml:"protocol" default:"HTTP" enum:"HTTP,HTTPS,TCP,TLS,UDP" desc:"target group protocol"`
	Port               int    `yaml:"port" default:"80" desc:"target group port"`
	TargetType         string `yaml:"target_type" default:"ip" enum:"instance,ip,lambda,alb" desc:"target type"`
	HealthCheckPath    string `yaml:"health_check_path" default:"/health" desc:"health check path"`
	HealthCheckEnabled bool   `yaml:"health_check_enabled" default:"true" desc:"enable health checks"`
}

type sampleSecurityGroup struct {
	GroupName   string               `yaml:"group_name" required:"true" desc:"security group name"`
	Description string               `yaml:"description" desc:"security group description"`
	VpcID       string               `yaml:"vpc_id" desc:"VPC ID"`
	Ingress     []sampleSecurityRule `yaml:"ingress" desc:"inbound rules"`
	Egress      []sampleSecurityRule `yaml:"egress" desc:"outbound rules"`
	Tags        map[string]string    `yaml:"tags" desc:"security group tags"`
}

type sampleSecurityRule struct {
	Protocol    string `yaml:"protocol" default:"tcp" enum:"tcp,udp,icmp,-1" desc:"IP protocol"`
	FromPort    int    `yaml:"from_port" desc:"start port"`
	ToPort      int    `yaml:"to_port" desc:"end port"`
	CIDRIP      string `yaml:"cidr_ip" desc:"IPv4 CIDR"`
	CIDRIPv6    string `yaml:"cidr_ipv6" desc:"IPv6 CIDR"`
	Description string `yaml:"description" desc:"rule description"`
}

type sampleOutput struct {
	Description string `yaml:"description" desc:"output description"`
	Value       string `yaml:"value" required:"true" desc:"output value"`
	ExportName  string `yaml:"export_name" desc:"CloudFormation export name"`
}

type sampleStackPolicy struct {
	ProtectedLogicalIDs []string `yaml:"protected_logical_ids" desc:"resources protected from replacement"`
	DenyDelete          bool     `yaml:"deny_delete" default:"true" desc:"deny delete operations"`
	DenyReplace         bool     `yaml:"deny_replace" default:"true" desc:"deny replacement operations"`
}
