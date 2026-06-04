package sampleschema

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
