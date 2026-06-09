package sample

type Workspace struct {
	Project  Project  `yaml:"project" required:"true" desc:"project metadata"`
	Database Database `yaml:"database" desc:"database connection settings"`
}

type Project struct {
	Name string `yaml:"name" required:"true" desc:"project name"`
	Tier string `yaml:"tier" default:"dev" enum:"dev,stg,prod" desc:"deployment tier"`
}
