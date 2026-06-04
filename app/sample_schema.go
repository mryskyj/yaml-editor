package app

type sampleConfig struct {
	Server sampleServer `yaml:"server" required:"true" desc:"server settings"`
	App    sampleApp    `yaml:"app" required:"true" desc:"application settings"`
}

type sampleServer struct {
	Host string `yaml:"host" required:"true" default:"localhost" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080" desc:"listen port"`
}

type sampleApp struct {
	Mode string `yaml:"mode" required:"true" default:"dev" enum:"dev,stg,prod" desc:"runtime mode"`
}
