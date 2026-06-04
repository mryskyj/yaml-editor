package schemas

type Config struct {
	Server Server `yaml:"server" required:"true" desc:"server settings"`
	App    App    `yaml:"app" required:"true" desc:"application settings"`
}

type Server struct {
	Host string `yaml:"host" required:"true" default:"localhost" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080" desc:"listen port"`
}

type App struct {
	Mode string `yaml:"mode" required:"true" default:"dev" enum:"dev,stg,prod" desc:"runtime mode"`
}
