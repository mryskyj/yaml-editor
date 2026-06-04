package sampleschema

type Config struct {
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
