package rootschema

type TargetCode struct {
	Code string `yaml:"code"`
}

type Action struct {
	Tool     string            `yaml:"tool"`
	User     string            `yaml:"user,omitempty"`
	Password string            `yaml:"password,omitempty"`
	Path     string            `yaml:"path,omitempty"`
	Args     map[string]string `yaml:"args"`
}

type Doc struct {
	Name string `yaml:"name"`
}

type Step struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	DayRef      string `yaml:"day_ref"`
	Date        Date   `yaml:"-"`
	ScheduleRef string `yaml:"schedule_ref"`
	Schedule    int    `yaml:"-"`
	Action      Action `yaml:"action"`
}

type Scenario struct {
	ID          int    `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Docs        []Doc  `yaml:"docs"`
	Steps       []Step `yaml:"steps"`
}

type File struct {
	SchemaVersion string   `yaml:"schema_version"`
	Common        Common   `yaml:"common"`
	Scenario      Scenario `yaml:"scenario"`
}
