package externalsample

type Config struct {
	Server Server `yaml:"server" required:"true" desc:"server settings"`
	App    App    `yaml:"app"`
}
