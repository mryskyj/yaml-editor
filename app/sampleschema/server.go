package sampleschema

type sampleServer struct {
	Host string `yaml:"host" required:"true" default:"localhost" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080" desc:"listen port"`
}
