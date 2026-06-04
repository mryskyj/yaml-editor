package sampleschema

type sampleApp struct {
	Mode string `yaml:"mode" required:"true" default:"dev" enum:"dev,stg,prod" desc:"runtime mode"`
}
