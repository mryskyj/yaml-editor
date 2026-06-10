package gui

type accountKind string
type accountMetadataKey string

const defaultAccountKind accountKind = "standard"

type AddAccount struct {
	Name     string                        `yaml:"Name" required:"true" desc:"account display name"`
	Code     string                        `yaml:"Code" required:"true" desc:"account code"`
	Kind     accountKind                   `yaml:"Kind" default:"standard" enum:"standard,premium,temporary" desc:"account kind"`
	Tags     []string                      `yaml:"Tags" desc:"account tags"`
	Metadata map[accountMetadataKey]string `yaml:"Metadata" desc:"metadata values keyed by logical name"`
	Contacts []AddAccountContact           `yaml:"Contacts" desc:"contact list"`
	Ignored  accountRuntimeOnly            `json:"ignored"`
	runtime  accountRuntimeOnly
}

type AddAccounts []AddAccount

type AddAccountContact struct {
	Name  string `yaml:"Name" required:"true" desc:"contact name"`
	Email string `yaml:"Email" desc:"contact email address"`
}

type accountRuntimeOnly struct {
	SessionID string
}

func (request AddAccount) DefaultKind() accountKind {
	if request.Kind == "" {
		return defaultAccountKind
	}
	return request.Kind
}
