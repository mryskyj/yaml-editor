package sampleschema

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
