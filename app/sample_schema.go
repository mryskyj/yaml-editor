package app

type sampleConfig struct {
	Server         sampleServer         `yaml:"server" required:"true" desc:"local editor server settings"`
	App            sampleApp            `yaml:"app" required:"true" desc:"application runtime settings"`
	AWS            sampleAWS            `yaml:"aws" desc:"AWS account and region settings"`
	CloudFormation sampleCloudFormation `yaml:"cloudformation" desc:"CloudFormation-style infrastructure definition"`
	ECS            sampleECS            `yaml:"ecs" desc:"Amazon ECS task and service settings"`
	SSM            sampleSSM            `yaml:"ssm" desc:"AWS Systems Manager Parameter Store settings"`
	Observability  sampleObservability  `yaml:"observability" desc:"logging, metrics, alarms, and tracing settings"`
	Deployment     sampleDeployment     `yaml:"deployment" desc:"release and rollout settings"`
	Security       sampleSecurity       `yaml:"security" desc:"IAM, encryption, and network policy settings"`
}

type sampleServer struct {
	Host string `yaml:"host" required:"true" default:"localhost" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080" desc:"listen port"`
}

type sampleApp struct {
	Mode string `yaml:"mode" required:"true" default:"dev" enum:"dev,stg,prod" desc:"runtime mode"`
}

type sampleAWS struct {
	AccountID       string             `yaml:"account_id" desc:"AWS account ID"`
	Partition       string             `yaml:"partition" default:"aws" enum:"aws,aws-cn,aws-us-gov" desc:"AWS partition"`
	Region          string             `yaml:"region" required:"true" default:"ap-northeast-1" enum:"us-east-1,us-west-2,eu-west-1,ap-northeast-1,ap-southeast-1" desc:"primary AWS region"`
	BackupRegions   []string           `yaml:"backup_regions" desc:"secondary regions for disaster recovery"`
	DefaultTags     map[string]string  `yaml:"default_tags" desc:"tags applied to managed resources"`
	CostAllocation  sampleCostTags     `yaml:"cost_allocation" desc:"cost allocation tags and chargeback settings"`
	ServiceQuotas   sampleServiceQuota `yaml:"service_quotas" desc:"capacity and quota assumptions"`
	FeatureFlags    map[string]bool    `yaml:"feature_flags" desc:"account-level feature flags"`
	AllowedCIDRs    []string           `yaml:"allowed_cidrs" desc:"CIDR ranges allowed to access public endpoints"`
	AvailabilityAZs [3]string          `yaml:"availability_zones" desc:"preferred availability zones"`
}

type sampleCostTags struct {
	Enabled      bool     `yaml:"enabled" default:"true" desc:"enable cost allocation tags"`
	RequiredKeys []string `yaml:"required_keys" desc:"tag keys required on all resources"`
	OwnerKey     string   `yaml:"owner_key" default:"Owner" desc:"owner tag key"`
	ProjectKey   string   `yaml:"project_key" default:"Project" desc:"project tag key"`
}

type sampleServiceQuota struct {
	MaxEC2Instances     int     `yaml:"max_ec2_instances" default:"20" desc:"expected EC2 instance quota"`
	MaxFargateTasks     int     `yaml:"max_fargate_tasks" default:"100" desc:"expected Fargate task quota"`
	ReservedCPUCores    float64 `yaml:"reserved_cpu_cores" default:"8" desc:"reserved CPU cores for steady state"`
	ReservedMemoryGiB   float64 `yaml:"reserved_memory_gib" default:"32" desc:"reserved memory in GiB"`
	FailOnQuotaMismatch bool    `yaml:"fail_on_quota_mismatch" default:"false" desc:"fail deployment when quotas are below expectation"`
}

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

type sampleECS struct {
	Cluster        sampleECSCluster        `yaml:"cluster" desc:"ECS cluster settings"`
	TaskDefinition sampleECSTaskDefinition `yaml:"task_definition" desc:"AWS::ECS::TaskDefinition-like settings"`
	Service        sampleECSService        `yaml:"service" desc:"ECS service settings"`
	Capacity       sampleECSCapacity       `yaml:"capacity" desc:"capacity provider settings"`
}

type sampleECSCluster struct {
	Name              string            `yaml:"name" required:"true" desc:"ECS cluster name"`
	ContainerInsights string            `yaml:"container_insights" default:"enabled" enum:"enabled,disabled" desc:"CloudWatch Container Insights"`
	CapacityProviders []string          `yaml:"capacity_providers" desc:"capacity providers"`
	Settings          map[string]string `yaml:"settings" desc:"cluster settings"`
}

type sampleECSTaskDefinition struct {
	Family                  string                 `yaml:"family" required:"true" desc:"task definition family"`
	NetworkMode             string                 `yaml:"network_mode" default:"awsvpc" enum:"awsvpc,bridge,host,none" desc:"task network mode"`
	RequiresCompatibilities []string               `yaml:"requires_compatibilities" desc:"launch types such as FARGATE or EC2"`
	CPU                     string                 `yaml:"cpu" default:"256" desc:"task CPU units"`
	Memory                  string                 `yaml:"memory" default:"512" desc:"task memory MiB"`
	ExecutionRoleARN        string                 `yaml:"execution_role_arn" desc:"task execution role ARN"`
	TaskRoleARN             string                 `yaml:"task_role_arn" desc:"task role ARN"`
	RuntimePlatform         sampleRuntimePlatform  `yaml:"runtime_platform" desc:"CPU architecture and operating system"`
	ContainerDefinitions    []sampleContainer      `yaml:"container_definitions" required:"true" desc:"container definitions"`
	Volumes                 []sampleECSVolume      `yaml:"volumes" desc:"task volumes"`
	EphemeralStorage        sampleEphemeralStorage `yaml:"ephemeral_storage" desc:"ephemeral storage for Fargate"`
	Tags                    map[string]string      `yaml:"tags" desc:"task definition tags"`
}

type sampleRuntimePlatform struct {
	CPUArchitecture       string `yaml:"cpu_architecture" default:"X86_64" enum:"X86_64,ARM64" desc:"task CPU architecture"`
	OperatingSystemFamily string `yaml:"operating_system_family" default:"LINUX" enum:"LINUX,WINDOWS_SERVER_2019_CORE,WINDOWS_SERVER_2022_CORE" desc:"task OS family"`
}

type sampleContainer struct {
	Name              string                      `yaml:"name" required:"true" desc:"container name"`
	Image             string                      `yaml:"image" required:"true" desc:"container image"`
	Essential         bool                        `yaml:"essential" default:"true" desc:"whether container is essential"`
	CPU               int                         `yaml:"cpu" default:"0" desc:"container CPU units"`
	Memory            int                         `yaml:"memory" desc:"hard memory limit MiB"`
	MemoryReservation int                         `yaml:"memory_reservation" desc:"soft memory reservation MiB"`
	PortMappings      []samplePortMapping         `yaml:"port_mappings" desc:"container port mappings"`
	Environment       []sampleNameValuePair       `yaml:"environment" desc:"environment variables"`
	Secrets           []sampleSecret              `yaml:"secrets" desc:"secret environment variables"`
	LogConfiguration  sampleLogConfiguration      `yaml:"log_configuration" desc:"container log configuration"`
	HealthCheck       sampleContainerHealthCheck  `yaml:"health_check" desc:"container health check"`
	MountPoints       []sampleMountPoint          `yaml:"mount_points" desc:"volume mount points"`
	DependsOn         []sampleContainerDependency `yaml:"depends_on" desc:"container dependencies"`
}

type samplePortMapping struct {
	ContainerPort int    `yaml:"container_port" required:"true" desc:"container port"`
	HostPort      int    `yaml:"host_port" desc:"host port"`
	Protocol      string `yaml:"protocol" default:"tcp" enum:"tcp,udp" desc:"port protocol"`
	AppProtocol   string `yaml:"app_protocol" enum:"http,http2,grpc" desc:"application protocol"`
	Name          string `yaml:"name" desc:"port mapping name"`
}

type sampleNameValuePair struct {
	Name  string `yaml:"name" required:"true" desc:"environment variable name"`
	Value string `yaml:"value" required:"true" desc:"environment variable value"`
}

type sampleSecret struct {
	Name      string `yaml:"name" required:"true" desc:"secret variable name"`
	ValueFrom string `yaml:"value_from" required:"true" desc:"SSM parameter or Secrets Manager ARN"`
}

type sampleLogConfiguration struct {
	LogDriver string            `yaml:"log_driver" default:"awslogs" enum:"awslogs,firelens,json-file,none" desc:"container log driver"`
	Options   map[string]string `yaml:"options" desc:"log driver options"`
}

type sampleContainerHealthCheck struct {
	Command     []string `yaml:"command" desc:"health check command"`
	Interval    int      `yaml:"interval" default:"30" desc:"interval seconds"`
	Timeout     int      `yaml:"timeout" default:"5" desc:"timeout seconds"`
	Retries     int      `yaml:"retries" default:"3" desc:"retry count"`
	StartPeriod int      `yaml:"start_period" default:"0" desc:"start period seconds"`
}

type sampleMountPoint struct {
	SourceVolume  string `yaml:"source_volume" required:"true" desc:"volume name"`
	ContainerPath string `yaml:"container_path" required:"true" desc:"container path"`
	ReadOnly      bool   `yaml:"read_only" default:"false" desc:"mount read-only"`
}

type sampleContainerDependency struct {
	ContainerName string `yaml:"container_name" required:"true" desc:"dependency container name"`
	Condition     string `yaml:"condition" default:"START" enum:"START,COMPLETE,SUCCESS,HEALTHY" desc:"dependency condition"`
}

type sampleECSVolume struct {
	Name             string           `yaml:"name" required:"true" desc:"volume name"`
	EFSConfiguration sampleEFSConfig  `yaml:"efs_configuration" desc:"EFS volume configuration"`
	Host             sampleHostVolume `yaml:"host" desc:"host volume configuration"`
}

type sampleEFSConfig struct {
	FileSystemID      string `yaml:"file_system_id" desc:"EFS file system ID"`
	RootDirectory     string `yaml:"root_directory" default:"/" desc:"EFS root directory"`
	TransitEncryption string `yaml:"transit_encryption" default:"ENABLED" enum:"ENABLED,DISABLED" desc:"encrypt traffic to EFS"`
	AuthorizationIAM  string `yaml:"authorization_iam" default:"ENABLED" enum:"ENABLED,DISABLED" desc:"use IAM authorization"`
	AccessPointID     string `yaml:"access_point_id" desc:"EFS access point ID"`
}

type sampleHostVolume struct {
	SourcePath string `yaml:"source_path" desc:"host source path"`
}

type sampleEphemeralStorage struct {
	SizeInGiB int `yaml:"size_in_gib" default:"21" desc:"Fargate ephemeral storage size"`
}

type sampleECSService struct {
	Name                    string                        `yaml:"name" required:"true" desc:"ECS service name"`
	LaunchType              string                        `yaml:"launch_type" default:"FARGATE" enum:"EC2,FARGATE,EXTERNAL" desc:"ECS launch type"`
	DesiredCount            int                           `yaml:"desired_count" default:"2" desc:"desired task count"`
	DeploymentController    string                        `yaml:"deployment_controller" default:"ECS" enum:"ECS,CODE_DEPLOY,EXTERNAL" desc:"deployment controller"`
	PlatformVersion         string                        `yaml:"platform_version" default:"LATEST" desc:"Fargate platform version"`
	AssignPublicIP          string                        `yaml:"assign_public_ip" default:"DISABLED" enum:"ENABLED,DISABLED" desc:"assign public IP to tasks"`
	Subnets                 []string                      `yaml:"subnets" desc:"service subnet IDs"`
	SecurityGroups          []string                      `yaml:"security_groups" desc:"task security groups"`
	LoadBalancers           []sampleECSLoadBalancer       `yaml:"load_balancers" desc:"service load balancers"`
	ServiceAutoScaling      sampleServiceAutoScaling      `yaml:"service_auto_scaling" desc:"service auto scaling"`
	DeploymentConfiguration sampleDeploymentConfiguration `yaml:"deployment_configuration" desc:"deployment percentages and circuit breaker"`
}

type sampleECSLoadBalancer struct {
	ContainerName  string `yaml:"container_name" required:"true" desc:"container name"`
	ContainerPort  int    `yaml:"container_port" required:"true" desc:"container port"`
	TargetGroupARN string `yaml:"target_group_arn" desc:"target group ARN"`
}

type sampleServiceAutoScaling struct {
	Enabled             bool    `yaml:"enabled" default:"true" desc:"enable service auto scaling"`
	MinCapacity         int     `yaml:"min_capacity" default:"2" desc:"minimum tasks"`
	MaxCapacity         int     `yaml:"max_capacity" default:"10" desc:"maximum tasks"`
	TargetCPUPercent    float64 `yaml:"target_cpu_percent" default:"60" desc:"target CPU utilization"`
	TargetMemoryPercent float64 `yaml:"target_memory_percent" default:"70" desc:"target memory utilization"`
}

type sampleDeploymentConfiguration struct {
	MinimumHealthyPercent int                            `yaml:"minimum_healthy_percent" default:"100" desc:"minimum healthy task percentage"`
	MaximumPercent        int                            `yaml:"maximum_percent" default:"200" desc:"maximum running task percentage"`
	CircuitBreaker        sampleDeploymentCircuitBreaker `yaml:"circuit_breaker" desc:"deployment circuit breaker"`
}

type sampleDeploymentCircuitBreaker struct {
	Enable   bool `yaml:"enable" default:"true" desc:"enable circuit breaker"`
	Rollback bool `yaml:"rollback" default:"true" desc:"rollback failed deployments"`
}

type sampleECSCapacity struct {
	Providers []sampleCapacityProvider `yaml:"providers" desc:"capacity provider strategies"`
}

type sampleCapacityProvider struct {
	Name   string `yaml:"name" required:"true" desc:"capacity provider name"`
	Weight int    `yaml:"weight" default:"1" desc:"capacity provider weight"`
	Base   int    `yaml:"base" default:"0" desc:"base tasks for provider"`
}

type sampleSSM struct {
	Parameters []sampleSSMParameter `yaml:"parameters" desc:"Parameter Store parameter declarations"`
	Policies   []sampleSSMPolicy    `yaml:"policies" desc:"advanced parameter policies"`
	Paths      map[string]string    `yaml:"paths" desc:"logical path aliases"`
}

type sampleSSMParameter struct {
	Name           string            `yaml:"name" required:"true" desc:"parameter name"`
	Type           string            `yaml:"type" required:"true" default:"String" enum:"String,StringList,SecureString" desc:"SSM parameter type"`
	Value          string            `yaml:"value" desc:"parameter value"`
	DataType       string            `yaml:"data_type" default:"text" enum:"text,aws:ec2:image,aws:ssm:integration" desc:"String parameter data type"`
	Tier           string            `yaml:"tier" default:"Standard" enum:"Standard,Advanced,Intelligent-Tiering" desc:"parameter tier"`
	KeyID          string            `yaml:"key_id" desc:"KMS key ID for SecureString"`
	AllowedPattern string            `yaml:"allowed_pattern" desc:"regular expression validation"`
	Overwrite      bool              `yaml:"overwrite" default:"false" desc:"overwrite existing parameter"`
	Tags           map[string]string `yaml:"tags" desc:"parameter tags"`
}

type sampleSSMPolicy struct {
	Type       string            `yaml:"type" required:"true" enum:"Expiration,ExpirationNotification,NoChangeNotification" desc:"advanced parameter policy type"`
	Version    string            `yaml:"version" default:"1.0" desc:"policy version"`
	Attributes map[string]string `yaml:"attributes" desc:"policy-specific attributes"`
}

type sampleObservability struct {
	LogGroups     []sampleLogGroup     `yaml:"log_groups" desc:"CloudWatch log groups"`
	Dashboards    []sampleDashboard    `yaml:"dashboards" desc:"CloudWatch dashboards"`
	Alarms        []sampleAlarm        `yaml:"alarms" desc:"CloudWatch alarms"`
	Tracing       sampleTracing        `yaml:"tracing" desc:"distributed tracing settings"`
	MetricFilters []sampleMetricFilter `yaml:"metric_filters" desc:"log metric filters"`
}

type sampleLogGroup struct {
	Name                string            `yaml:"name" required:"true" desc:"log group name"`
	RetentionDays       int               `yaml:"retention_days" default:"30" desc:"retention period in days"`
	KMSKeyID            string            `yaml:"kms_key_id" desc:"KMS key for log encryption"`
	ContributorInsights bool              `yaml:"contributor_insights" default:"false" desc:"enable Contributor Insights"`
	Tags                map[string]string `yaml:"tags" desc:"log group tags"`
}

type sampleDashboard struct {
	Name    string   `yaml:"name" required:"true" desc:"dashboard name"`
	Widgets []string `yaml:"widgets" desc:"dashboard widget JSON snippets"`
}

type sampleAlarm struct {
	Name               string   `yaml:"name" required:"true" desc:"alarm name"`
	Namespace          string   `yaml:"namespace" default:"AWS/ECS" desc:"metric namespace"`
	MetricName         string   `yaml:"metric_name" required:"true" desc:"metric name"`
	Statistic          string   `yaml:"statistic" default:"Average" enum:"Average,Minimum,Maximum,Sum,SampleCount" desc:"metric statistic"`
	Period             int      `yaml:"period" default:"60" desc:"period seconds"`
	EvaluationPeriods  int      `yaml:"evaluation_periods" default:"3" desc:"evaluation periods"`
	Threshold          float64  `yaml:"threshold" required:"true" desc:"alarm threshold"`
	ComparisonOperator string   `yaml:"comparison_operator" default:"GreaterThanThreshold" enum:"GreaterThanThreshold,GreaterThanOrEqualToThreshold,LessThanThreshold,LessThanOrEqualToThreshold" desc:"comparison operator"`
	TreatMissingData   string   `yaml:"treat_missing_data" default:"missing" enum:"breaching,notBreaching,ignore,missing" desc:"missing data behavior"`
	AlarmActions       []string `yaml:"alarm_actions" desc:"SNS topics or actions"`
}

type sampleTracing struct {
	Enabled       bool    `yaml:"enabled" default:"true" desc:"enable tracing"`
	Provider      string  `yaml:"provider" default:"xray" enum:"xray,opentelemetry,none" desc:"tracing provider"`
	SamplingRate  float64 `yaml:"sampling_rate" default:"0.05" desc:"trace sampling rate"`
	DaemonAddress string  `yaml:"daemon_address" default:"127.0.0.1:2000" desc:"trace daemon address"`
}

type sampleMetricFilter struct {
	Name            string `yaml:"name" required:"true" desc:"metric filter name"`
	LogGroupName    string `yaml:"log_group_name" required:"true" desc:"source log group"`
	FilterPattern   string `yaml:"filter_pattern" required:"true" desc:"CloudWatch Logs filter pattern"`
	MetricName      string `yaml:"metric_name" required:"true" desc:"emitted metric name"`
	MetricNamespace string `yaml:"metric_namespace" required:"true" desc:"emitted metric namespace"`
	MetricValue     string `yaml:"metric_value" default:"1" desc:"emitted metric value"`
}

type sampleDeployment struct {
	Strategy         string                  `yaml:"strategy" default:"rolling" enum:"rolling,blue-green,canary,linear" desc:"deployment strategy"`
	Canary           sampleCanaryDeployment  `yaml:"canary" desc:"canary deployment settings"`
	BlueGreen        sampleBlueGreen         `yaml:"blue_green" desc:"blue-green deployment settings"`
	Approvals        []sampleApproval        `yaml:"approvals" desc:"manual approval gates"`
	RollbackTriggers []sampleRollbackTrigger `yaml:"rollback_triggers" desc:"rollback trigger alarms"`
	ChangeWindows    []sampleChangeWindow    `yaml:"change_windows" desc:"allowed deployment windows"`
}

type sampleCanaryDeployment struct {
	Enabled         bool    `yaml:"enabled" default:"false" desc:"enable canary releases"`
	InitialPercent  float64 `yaml:"initial_percent" default:"10" desc:"initial traffic percentage"`
	IntervalMinutes int     `yaml:"interval_minutes" default:"10" desc:"interval between increments"`
	Steps           int     `yaml:"steps" default:"5" desc:"number of rollout steps"`
}

type sampleBlueGreen struct {
	Enabled                bool   `yaml:"enabled" default:"false" desc:"enable blue-green deployments"`
	TerminationWaitMinutes int    `yaml:"termination_wait_minutes" default:"30" desc:"wait before terminating old environment"`
	TrafficRouting         string `yaml:"traffic_routing" default:"all-at-once" enum:"all-at-once,canary,linear" desc:"traffic routing mode"`
}

type sampleApproval struct {
	Name      string   `yaml:"name" required:"true" desc:"approval name"`
	Required  bool     `yaml:"required" default:"true" desc:"approval required"`
	Approvers []string `yaml:"approvers" desc:"approver identifiers"`
}

type sampleRollbackTrigger struct {
	AlarmName string `yaml:"alarm_name" required:"true" desc:"CloudWatch alarm name"`
	Type      string `yaml:"type" default:"cloudwatch-alarm" enum:"cloudwatch-alarm,manual,synthetic-check" desc:"trigger type"`
}

type sampleChangeWindow struct {
	DayOfWeek string `yaml:"day_of_week" required:"true" enum:"Mon,Tue,Wed,Thu,Fri,Sat,Sun" desc:"allowed day"`
	StartUTC  string `yaml:"start_utc" required:"true" desc:"start time UTC"`
	EndUTC    string `yaml:"end_utc" required:"true" desc:"end time UTC"`
}

type sampleSecurity struct {
	Encryption    sampleEncryption    `yaml:"encryption" desc:"encryption defaults"`
	IAM           sampleIAM           `yaml:"iam" desc:"IAM policy settings"`
	NetworkPolicy sampleNetworkPolicy `yaml:"network_policy" desc:"network access rules"`
	Compliance    sampleCompliance    `yaml:"compliance" desc:"compliance controls"`
}

type sampleEncryption struct {
	EnabledByDefault bool              `yaml:"enabled_by_default" default:"true" desc:"enable encryption by default"`
	DefaultKMSKeyID  string            `yaml:"default_kms_key_id" desc:"default KMS key ID"`
	KeyAliases       map[string]string `yaml:"key_aliases" desc:"logical key aliases"`
	RotationEnabled  bool              `yaml:"rotation_enabled" default:"true" desc:"enable key rotation"`
}

type sampleIAM struct {
	PermissionsBoundaryARN string               `yaml:"permissions_boundary_arn" desc:"permissions boundary ARN"`
	ManagedPolicyARNs      []string             `yaml:"managed_policy_arns" desc:"managed policy ARNs"`
	InlinePolicies         []sampleInlinePolicy `yaml:"inline_policies" desc:"inline IAM policies"`
	AllowedPrincipalARNs   []string             `yaml:"allowed_principal_arns" desc:"trusted principal ARNs"`
}

type sampleInlinePolicy struct {
	Name      string   `yaml:"name" required:"true" desc:"policy name"`
	Effect    string   `yaml:"effect" default:"Allow" enum:"Allow,Deny" desc:"policy effect"`
	Actions   []string `yaml:"actions" required:"true" desc:"IAM actions"`
	Resources []string `yaml:"resources" required:"true" desc:"resource ARNs"`
}

type sampleNetworkPolicy struct {
	PublicIngressAllowed bool                 `yaml:"public_ingress_allowed" default:"false" desc:"allow public ingress"`
	AllowedPorts         []int                `yaml:"allowed_ports" desc:"allowed inbound ports"`
	AllowedCIDRs         []string             `yaml:"allowed_cidrs" desc:"allowed CIDRs"`
	PrivateSubnetsOnly   bool                 `yaml:"private_subnets_only" default:"true" desc:"restrict workloads to private subnets"`
	EgressRules          []sampleSecurityRule `yaml:"egress_rules" desc:"egress rules"`
}

type sampleCompliance struct {
	Frameworks       []string        `yaml:"frameworks" desc:"compliance frameworks"`
	DataClasses      []string        `yaml:"data_classes" desc:"data classification labels"`
	RequiredControls map[string]bool `yaml:"required_controls" desc:"required controls by name"`
	AuditLogS3Bucket string          `yaml:"audit_log_s3_bucket" desc:"audit log bucket"`
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
