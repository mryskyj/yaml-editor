package externalsample

type App struct {
	Mode    string            `yaml:"mode" enum:"dev,stg,prod" default:"dev" desc:"run mode"`
	Enabled bool              `yaml:"enabled" default:"true" desc:"feature enabled"`
	Weights []float64         `yaml:"weights" desc:"routing weights"`
	Labels  map[string]string `yaml:"labels" desc:"resource labels"`
}
