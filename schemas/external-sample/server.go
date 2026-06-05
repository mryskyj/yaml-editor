package externalsample

type Server struct {
	Host string `yaml:"host" required:"true" default:"127.0.0.1" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080" desc:"listen port"`
}
