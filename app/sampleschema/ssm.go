package sampleschema

type sampleSSM struct {
	Parameters []sampleSSMParameter `yaml:"parameters,omitempty" desc:"Parameter Store parameter declarations"`
	Policies   []sampleSSMPolicy    `yaml:"policies,omitempty" desc:"advanced parameter policies"`
	Paths      map[string]string    `yaml:"paths,omitempty" desc:"logical path aliases"`
}

type sampleSSMParameter struct {
	Name           string            `yaml:"name" required:"true" desc:"parameter name"`
	Type           string            `yaml:"type" required:"true" default:"String" enum:"String,StringList,SecureString" desc:"SSM parameter type"`
	Value          string            `yaml:"value,omitempty" desc:"parameter value"`
	DataType       string            `yaml:"data_type,omitempty" default:"text" enum:"text,aws:ec2:image,aws:ssm:integration" desc:"String parameter data type"`
	Tier           string            `yaml:"tier,omitempty" default:"Standard" enum:"Standard,Advanced,Intelligent-Tiering" desc:"parameter tier"`
	KeyID          string            `yaml:"key_id,omitempty" desc:"KMS key ID for SecureString"`
	AllowedPattern string            `yaml:"allowed_pattern,omitempty" desc:"regular expression validation"`
	Overwrite      bool              `yaml:"overwrite,omitempty" default:"false" desc:"overwrite existing parameter"`
	Tags           map[string]string `yaml:"tags,omitempty" desc:"parameter tags"`
}

type sampleSSMPolicy struct {
	Type       string            `yaml:"type" required:"true" enum:"Expiration,ExpirationNotification,NoChangeNotification" desc:"advanced parameter policy type"`
	Version    string            `yaml:"version,omitempty" default:"1.0" desc:"policy version"`
	Attributes map[string]string `yaml:"attributes,omitempty" desc:"policy-specific attributes"`
}
