package sampleschema

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
