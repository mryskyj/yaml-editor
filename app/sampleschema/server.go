package sampleschema

type sampleServer struct {
	Host string `yaml:"host" json:"host" required:"true" default:"localhost" desc:"listen host"`
	Port int    `yaml:"port" json:"port" required:"true" default:"8080" desc:"listen port"`
}
