package sampleschema

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
