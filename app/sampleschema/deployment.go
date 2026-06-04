package sampleschema

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
