package sample

type Database struct {
	Driver string            `yaml:"driver" default:"postgres" enum:"postgres,mysql,sqlite" desc:"database driver"`
	Host   string            `yaml:"host" default:"127.0.0.1" desc:"database host"`
	Labels map[string]string `yaml:"labels" desc:"database labels"`
}
