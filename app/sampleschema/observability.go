package sampleschema

type sampleObservability struct {
	LogGroups     []sampleLogGroup     `yaml:"log_groups,omitempty" desc:"CloudWatch log groups"`
	Dashboards    []sampleDashboard    `yaml:"dashboards,omitempty" desc:"CloudWatch dashboards"`
	Alarms        []sampleAlarm        `yaml:"alarms,omitempty" desc:"CloudWatch alarms"`
	Tracing       sampleTracing        `yaml:"tracing,omitempty" desc:"distributed tracing settings"`
	MetricFilters []sampleMetricFilter `yaml:"metric_filters,omitempty" desc:"log metric filters"`
}

type sampleLogGroup struct {
	Name                string            `yaml:"name" required:"true" desc:"log group name"`
	RetentionDays       int               `yaml:"retention_days,omitempty" default:"30" desc:"retention period in days"`
	KMSKeyID            string            `yaml:"kms_key_id,omitempty" desc:"KMS key for log encryption"`
	ContributorInsights bool              `yaml:"contributor_insights,omitempty" default:"false" desc:"enable Contributor Insights"`
	Tags                map[string]string `yaml:"tags,omitempty" desc:"log group tags"`
}

type sampleDashboard struct {
	Name    string   `yaml:"name" required:"true" desc:"dashboard name"`
	Widgets []string `yaml:"widgets,omitempty" desc:"dashboard widget JSON snippets"`
}

type sampleAlarm struct {
	Name               string   `yaml:"name" required:"true" desc:"alarm name"`
	Namespace          string   `yaml:"namespace,omitempty" default:"AWS/ECS" desc:"metric namespace"`
	MetricName         string   `yaml:"metric_name" required:"true" desc:"metric name"`
	Statistic          string   `yaml:"statistic,omitempty" default:"Average" enum:"Average,Minimum,Maximum,Sum,SampleCount" desc:"metric statistic"`
	Period             int      `yaml:"period,omitempty" default:"60" desc:"period seconds"`
	EvaluationPeriods  int      `yaml:"evaluation_periods,omitempty" default:"3" desc:"evaluation periods"`
	Threshold          float64  `yaml:"threshold" required:"true" desc:"alarm threshold"`
	ComparisonOperator string   `yaml:"comparison_operator,omitempty" default:"GreaterThanThreshold" enum:"GreaterThanThreshold,GreaterThanOrEqualToThreshold,LessThanThreshold,LessThanOrEqualToThreshold" desc:"comparison operator"`
	TreatMissingData   string   `yaml:"treat_missing_data,omitempty" default:"missing" enum:"breaching,notBreaching,ignore,missing" desc:"missing data behavior"`
	AlarmActions       []string `yaml:"alarm_actions,omitempty" desc:"SNS topics or actions"`
}

type sampleTracing struct {
	Enabled       bool    `yaml:"enabled,omitempty" default:"true" desc:"enable tracing"`
	Provider      string  `yaml:"provider,omitempty" default:"xray" enum:"xray,opentelemetry,none" desc:"tracing provider"`
	SamplingRate  float64 `yaml:"sampling_rate,omitempty" default:"0.05" desc:"trace sampling rate"`
	DaemonAddress string  `yaml:"daemon_address,omitempty" default:"127.0.0.1:2000" desc:"trace daemon address"`
}

type sampleMetricFilter struct {
	Name            string `yaml:"name" required:"true" desc:"metric filter name"`
	LogGroupName    string `yaml:"log_group_name" required:"true" desc:"source log group"`
	FilterPattern   string `yaml:"filter_pattern" required:"true" desc:"CloudWatch Logs filter pattern"`
	MetricName      string `yaml:"metric_name" required:"true" desc:"emitted metric name"`
	MetricNamespace string `yaml:"metric_namespace" required:"true" desc:"emitted metric namespace"`
	MetricValue     string `yaml:"metric_value,omitempty" default:"1" desc:"emitted metric value"`
}
