package rootschema

type Date struct {
	Day              string `yaml:"date"`
	IsHolidayTrading bool   `yaml:"holiday"`
}

type Common struct {
	SchemaVersion string           `yaml:"schema_version"`
	Dates         map[string]*Date `yaml:"dates"`
	Schedules     map[string]int   `yaml:"schedules"`
}
