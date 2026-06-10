package ecs

type RunTask struct {
	Cluster        string            `yaml:"cluster" required:"true" desc:"ECS cluster name"`
	TaskDefinition string            `yaml:"task_definition" required:"true" desc:"task definition ARN or family"`
	LaunchType     string            `yaml:"launch_type" default:"FARGATE" enum:"EC2,FARGATE,EXTERNAL" desc:"ECS launch type"`
	Count          int               `yaml:"count" default:"1" desc:"number of tasks to run"`
	Overrides      map[string]string `yaml:"overrides" desc:"container override values"`
}
