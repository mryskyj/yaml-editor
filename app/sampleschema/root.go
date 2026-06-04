package sampleschema

type Config struct {
	Server         sampleServer         `yaml:"server" json:"server" required:"true" desc:"local editor server settings"`
	App            sampleApp            `yaml:"app" json:"app" required:"true" desc:"application runtime settings"`
	AWS            sampleAWS            `yaml:"aws" json:"aws" desc:"AWS account and region settings"`
	CloudFormation sampleCloudFormation `yaml:"cloudformation" json:"cloudformation" desc:"CloudFormation-style infrastructure definition"`
	ECS            sampleECS            `yaml:"ecs" json:"ecs" desc:"Amazon ECS task and service settings"`
	SSM            sampleSSM            `yaml:"ssm" json:"ssm" desc:"AWS Systems Manager Parameter Store settings"`
	Observability  sampleObservability  `yaml:"observability" json:"observability" desc:"logging, metrics, alarms, and tracing settings"`
	Deployment     sampleDeployment     `yaml:"deployment" json:"deployment" desc:"release and rollout settings"`
	Security       sampleSecurity       `yaml:"security" json:"security" desc:"IAM, encryption, and network policy settings"`
	JSONImport     sampleJSONImport     `json:"json_import"`
	XMLImport      sampleXMLImport      `xml:"xml_import"`
}
