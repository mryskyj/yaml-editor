package sampleschema

type sampleECS struct {
	Cluster        sampleECSCluster        `yaml:"cluster,omitempty" desc:"ECS cluster settings"`
	TaskDefinition sampleECSTaskDefinition `yaml:"task_definition,omitempty" desc:"AWS::ECS::TaskDefinition-like settings"`
	Service        sampleECSService        `yaml:"service,omitempty" desc:"ECS service settings"`
	Capacity       sampleECSCapacity       `yaml:"capacity,omitempty" desc:"capacity provider settings"`
}

type sampleECSCluster struct {
	Name              string            `yaml:"name" required:"true" desc:"ECS cluster name"`
	ContainerInsights string            `yaml:"container_insights,omitempty" default:"enabled" enum:"enabled,disabled" desc:"CloudWatch Container Insights"`
	CapacityProviders []string          `yaml:"capacity_providers,omitempty" desc:"capacity providers"`
	Settings          map[string]string `yaml:"settings,omitempty" desc:"cluster settings"`
}

type sampleECSTaskDefinition struct {
	Family                  string                 `yaml:"family" required:"true" desc:"task definition family"`
	NetworkMode             string                 `yaml:"network_mode,omitempty" default:"awsvpc" enum:"awsvpc,bridge,host,none" desc:"task network mode"`
	RequiresCompatibilities []string               `yaml:"requires_compatibilities,omitempty" desc:"launch types such as FARGATE or EC2"`
	CPU                     string                 `yaml:"cpu,omitempty" default:"256" desc:"task CPU units"`
	Memory                  string                 `yaml:"memory,omitempty" default:"512" desc:"task memory MiB"`
	ExecutionRoleARN        string                 `yaml:"execution_role_arn,omitempty" desc:"task execution role ARN"`
	TaskRoleARN             string                 `yaml:"task_role_arn,omitempty" desc:"task role ARN"`
	RuntimePlatform         sampleRuntimePlatform  `yaml:"runtime_platform,omitempty" desc:"CPU architecture and operating system"`
	ContainerDefinitions    []sampleContainer      `yaml:"container_definitions" required:"true" desc:"container definitions"`
	Volumes                 []sampleECSVolume      `yaml:"volumes,omitempty" desc:"task volumes"`
	EphemeralStorage        sampleEphemeralStorage `yaml:"ephemeral_storage,omitempty" desc:"ephemeral storage for Fargate"`
	Tags                    map[string]string      `yaml:"tags,omitempty" desc:"task definition tags"`
}

type sampleRuntimePlatform struct {
	CPUArchitecture       string `yaml:"cpu_architecture,omitempty" default:"X86_64" enum:"X86_64,ARM64" desc:"task CPU architecture"`
	OperatingSystemFamily string `yaml:"operating_system_family,omitempty" default:"LINUX" enum:"LINUX,WINDOWS_SERVER_2019_CORE,WINDOWS_SERVER_2022_CORE" desc:"task OS family"`
}

type sampleContainer struct {
	Name              string                      `yaml:"name" required:"true" desc:"container name"`
	Image             string                      `yaml:"image" required:"true" desc:"container image"`
	Essential         bool                        `yaml:"essential,omitempty" default:"true" desc:"whether container is essential"`
	CPU               int                         `yaml:"cpu,omitempty" default:"0" desc:"container CPU units"`
	Memory            int                         `yaml:"memory,omitempty" desc:"hard memory limit MiB"`
	MemoryReservation int                         `yaml:"memory_reservation,omitempty" desc:"soft memory reservation MiB"`
	PortMappings      []samplePortMapping         `yaml:"port_mappings,omitempty" desc:"container port mappings"`
	Environment       []sampleNameValuePair       `yaml:"environment,omitempty" desc:"environment variables"`
	Secrets           []sampleSecret              `yaml:"secrets,omitempty" desc:"secret environment variables"`
	LogConfiguration  sampleLogConfiguration      `yaml:"log_configuration,omitempty" desc:"container log configuration"`
	HealthCheck       sampleContainerHealthCheck  `yaml:"health_check,omitempty" desc:"container health check"`
	MountPoints       []sampleMountPoint          `yaml:"mount_points,omitempty" desc:"volume mount points"`
	DependsOn         []sampleContainerDependency `yaml:"depends_on,omitempty" desc:"container dependencies"`
}

type samplePortMapping struct {
	ContainerPort int    `yaml:"container_port" required:"true" desc:"container port"`
	HostPort      int    `yaml:"host_port,omitempty" desc:"host port"`
	Protocol      string `yaml:"protocol,omitempty" default:"tcp" enum:"tcp,udp" desc:"port protocol"`
	AppProtocol   string `yaml:"app_protocol,omitempty" enum:"http,http2,grpc" desc:"application protocol"`
	Name          string `yaml:"name,omitempty" desc:"port mapping name"`
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
	LogDriver string            `yaml:"log_driver,omitempty" default:"awslogs" enum:"awslogs,firelens,json-file,none" desc:"container log driver"`
	Options   map[string]string `yaml:"options,omitempty" desc:"log driver options"`
}

type sampleContainerHealthCheck struct {
	Command     []string `yaml:"command,omitempty" desc:"health check command"`
	Interval    int      `yaml:"interval,omitempty" default:"30" desc:"interval seconds"`
	Timeout     int      `yaml:"timeout,omitempty" default:"5" desc:"timeout seconds"`
	Retries     int      `yaml:"retries,omitempty" default:"3" desc:"retry count"`
	StartPeriod int      `yaml:"start_period,omitempty" default:"0" desc:"start period seconds"`
}

type sampleMountPoint struct {
	SourceVolume  string `yaml:"source_volume" required:"true" desc:"volume name"`
	ContainerPath string `yaml:"container_path" required:"true" desc:"container path"`
	ReadOnly      bool   `yaml:"read_only,omitempty" default:"false" desc:"mount read-only"`
}

type sampleContainerDependency struct {
	ContainerName string `yaml:"container_name" required:"true" desc:"dependency container name"`
	Condition     string `yaml:"condition,omitempty" default:"START" enum:"START,COMPLETE,SUCCESS,HEALTHY" desc:"dependency condition"`
}

type sampleECSVolume struct {
	Name             string           `yaml:"name" required:"true" desc:"volume name"`
	EFSConfiguration sampleEFSConfig  `yaml:"efs_configuration,omitempty" desc:"EFS volume configuration"`
	Host             sampleHostVolume `yaml:"host,omitempty" desc:"host volume configuration"`
}

type sampleEFSConfig struct {
	FileSystemID      string `yaml:"file_system_id,omitempty" desc:"EFS file system ID"`
	RootDirectory     string `yaml:"root_directory,omitempty" default:"/" desc:"EFS root directory"`
	TransitEncryption string `yaml:"transit_encryption,omitempty" default:"ENABLED" enum:"ENABLED,DISABLED" desc:"encrypt traffic to EFS"`
	AuthorizationIAM  string `yaml:"authorization_iam,omitempty" default:"ENABLED" enum:"ENABLED,DISABLED" desc:"use IAM authorization"`
	AccessPointID     string `yaml:"access_point_id,omitempty" desc:"EFS access point ID"`
}

type sampleHostVolume struct {
	SourcePath string `yaml:"source_path,omitempty" desc:"host source path"`
}

type sampleEphemeralStorage struct {
	SizeInGiB int `yaml:"size_in_gib,omitempty" default:"21" desc:"Fargate ephemeral storage size"`
}

type sampleECSService struct {
	Name                    string                        `yaml:"name" required:"true" desc:"ECS service name"`
	LaunchType              string                        `yaml:"launch_type,omitempty" default:"FARGATE" enum:"EC2,FARGATE,EXTERNAL" desc:"ECS launch type"`
	DesiredCount            int                           `yaml:"desired_count,omitempty" default:"2" desc:"desired task count"`
	DeploymentController    string                        `yaml:"deployment_controller,omitempty" default:"ECS" enum:"ECS,CODE_DEPLOY,EXTERNAL" desc:"deployment controller"`
	PlatformVersion         string                        `yaml:"platform_version,omitempty" default:"LATEST" desc:"Fargate platform version"`
	AssignPublicIP          string                        `yaml:"assign_public_ip,omitempty" default:"DISABLED" enum:"ENABLED,DISABLED" desc:"assign public IP to tasks"`
	Subnets                 []string                      `yaml:"subnets,omitempty" desc:"service subnet IDs"`
	SecurityGroups          []string                      `yaml:"security_groups,omitempty" desc:"task security groups"`
	LoadBalancers           []sampleECSLoadBalancer       `yaml:"load_balancers,omitempty" desc:"service load balancers"`
	ServiceAutoScaling      sampleServiceAutoScaling      `yaml:"service_auto_scaling,omitempty" desc:"service auto scaling"`
	DeploymentConfiguration sampleDeploymentConfiguration `yaml:"deployment_configuration,omitempty" desc:"deployment percentages and circuit breaker"`
}

type sampleECSLoadBalancer struct {
	ContainerName  string `yaml:"container_name" required:"true" desc:"container name"`
	ContainerPort  int    `yaml:"container_port" required:"true" desc:"container port"`
	TargetGroupARN string `yaml:"target_group_arn,omitempty" desc:"target group ARN"`
}

type sampleServiceAutoScaling struct {
	Enabled             bool    `yaml:"enabled,omitempty" default:"true" desc:"enable service auto scaling"`
	MinCapacity         int     `yaml:"min_capacity,omitempty" default:"2" desc:"minimum tasks"`
	MaxCapacity         int     `yaml:"max_capacity,omitempty" default:"10" desc:"maximum tasks"`
	TargetCPUPercent    float64 `yaml:"target_cpu_percent,omitempty" default:"60" desc:"target CPU utilization"`
	TargetMemoryPercent float64 `yaml:"target_memory_percent,omitempty" default:"70" desc:"target memory utilization"`
}

type sampleDeploymentConfiguration struct {
	MinimumHealthyPercent int                            `yaml:"minimum_healthy_percent,omitempty" default:"100" desc:"minimum healthy task percentage"`
	MaximumPercent        int                            `yaml:"maximum_percent,omitempty" default:"200" desc:"maximum running task percentage"`
	CircuitBreaker        sampleDeploymentCircuitBreaker `yaml:"circuit_breaker,omitempty" desc:"deployment circuit breaker"`
}

type sampleDeploymentCircuitBreaker struct {
	Enable   bool `yaml:"enable,omitempty" default:"true" desc:"enable circuit breaker"`
	Rollback bool `yaml:"rollback,omitempty" default:"true" desc:"rollback failed deployments"`
}

type sampleECSCapacity struct {
	Providers []sampleCapacityProvider `yaml:"providers,omitempty" desc:"capacity provider strategies"`
}

type sampleCapacityProvider struct {
	Name   string `yaml:"name" required:"true" desc:"capacity provider name"`
	Weight int    `yaml:"weight,omitempty" default:"1" desc:"capacity provider weight"`
	Base   int    `yaml:"base,omitempty" default:"0" desc:"base tasks for provider"`
}
