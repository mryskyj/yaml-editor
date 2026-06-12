package rootschema

type Date struct {
	Day              string `yaml:"date"`
	IsHolidayTrading bool   `yaml:"holiday"`
}

type Common struct {
	SchemaVersion string           `yaml:"schema_version,omitempty"`
	Dates         map[string]*Date `yaml:"dates"`
	NumberOfDays  int64            `yaml:"number_of_days"`
	Schedules     map[string]int   `yaml:"schedules"`
}
