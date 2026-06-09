package validator

import (
	"strings"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func validatorSchema(t *testing.T) *schema.Field {
	t.Helper()

	type config struct {
		Server struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port" required:"true"`
		} `yaml:"server"`
		App struct {
			Debug bool             `yaml:"debug"`
			Mode  string           `yaml:"mode" enum:"dev,stg,prod"`
			Ratio float64          `yaml:"ratio"`
			Tags  []string         `yaml:"tags"`
			Ports [2]int           `yaml:"ports"`
			Meta  map[string]int   `yaml:"meta"`
			Nodes []map[string]int `yaml:"nodes"`
		} `yaml:"app"`
	}

	root, err := schema.Parse(config{})
	if err != nil {
		t.Fatalf("schema.Parse() returned error: %v", err)
	}
	return root
}

func TestValidateValidYAML(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(`
server:
  host: localhost
  port: 8080
app:
  debug: true
  mode: dev
  ratio: 1.5
  tags:
    - web
  ports:
    - 80
    - 443
  meta:
    retries: 3
  nodes:
    - weight: 10
`, validatorSchema(t))

	if len(diagnostics) != 0 {
		t.Fatalf("Validate() diagnostics = %#v, want none", diagnostics)
	}
}

func TestValidateReturnsYAMLSyntaxDiagnostic(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server:\n  host\n    bad: value\n", validatorSchema(t))
	if len(diagnostics) != 1 {
		t.Fatalf("Validate() diagnostics count = %d, want 1", len(diagnostics))
	}
	if !strings.Contains(diagnostics[0].Message, "YAML syntax error") {
		t.Fatalf("diagnostic.Message = %q", diagnostics[0].Message)
	}
}

func TestValidateUndefinedKey(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server:\n  unknown: value\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "undefined key \"unknown\"")
}

func TestValidateTypeMismatch(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server:\n  host: localhost\n  port: wrong\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "key \"port\" must be int")
}

func TestValidateNestedMismatch(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server: localhost\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "key \"server\" must be struct")
}

func TestValidateSequenceItemMismatch(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("app:\n  tags:\n    - web\n    - 10\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "key \"\" must be string")
}

func TestValidateMapValueMismatch(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("app:\n  meta:\n    retries: wrong\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "key \"\" must be int")
}

func TestValidateRequiredMissing(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server:\n  host: localhost\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "required key \"port\" is missing")
}

func TestValidateEnumMismatch(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("app:\n  mode: local\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "key \"mode\" must be one of: dev, stg, prod")
}

func TestValidateUnsupportedYAMLFeatureDiagnostic(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("defaults: &defaults\n  host: localhost\nserver: *defaults\n", validatorSchema(t))
	assertContainsDiagnostic(t, diagnostics, "YAML Alias is not supported")
}

func TestValidateNilSchema(t *testing.T) {
	t.Parallel()

	diagnostics := Validate("server:\n  host: localhost\n", nil)
	assertContainsDiagnostic(t, diagnostics, "root schema is not registered")
}

func assertContainsDiagnostic(t *testing.T, diagnostics []Diagnostic, message string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Message == message {
			if diagnostic.Severity != SeverityError {
				t.Fatalf("Severity = %q, want %q", diagnostic.Severity, SeverityError)
			}
			if diagnostic.Line <= 0 || diagnostic.Column <= 0 {
				t.Fatalf("diagnostic position = %d:%d, want non-zero", diagnostic.Line, diagnostic.Column)
			}
			return
		}
	}

	t.Fatalf("missing diagnostic %q in %#v", message, diagnostics)
}
