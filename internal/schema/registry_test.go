package schema

import (
	"path/filepath"
	"testing"
)

type registryConfig struct {
	Server registryServer `yaml:"server"`
}

type registryServer struct {
	Host string `yaml:"host"`
}

func TestRegistryRegisterAndRoot(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register(registryConfig{}); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	root, err := registry.Root()
	if err != nil {
		t.Fatalf("Root() returned error: %v", err)
	}
	if root.Name != "registryConfig" {
		t.Fatalf("root.Name = %q, want %q", root.Name, "registryConfig")
	}
	if _, ok := root.FindChild("server"); !ok {
		t.Fatal("Root() schema does not include server field")
	}
}

func TestRegistryRootBeforeRegister(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	root, err := registry.Root()
	if err == nil {
		t.Fatal("Root() error = nil, want error")
	}
	if root != nil {
		t.Fatalf("Root() = %#v, want nil", root)
	}
}

func TestRegistryRegisterFromDirAndRoot(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.RegisterFromDir(filepath.Join("..", "..", "schemas", "external-sample"), "Config"); err != nil {
		t.Fatalf("RegisterFromDir() returned error: %v", err)
	}

	root, err := registry.Root()
	if err != nil {
		t.Fatalf("Root() returned error: %v", err)
	}
	if root.Name != "Config" {
		t.Fatalf("root.Name = %q, want Config", root.Name)
	}
	if _, ok := root.FindChild("server"); !ok {
		t.Fatal("Root() schema does not include server field")
	}
}

func TestRegistryRegisterInvalidSchema(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.Register("not struct"); err == nil {
		t.Fatal("Register() error = nil, want error")
	}

	root, err := registry.Root()
	if err == nil {
		t.Fatal("Root() error = nil, want error after failed register")
	}
	if root != nil {
		t.Fatalf("Root() = %#v, want nil", root)
	}
}

func TestRegistryNilReceiver(t *testing.T) {
	t.Parallel()

	var registry *Registry
	if err := registry.Register(registryConfig{}); err == nil {
		t.Fatal("Register() error = nil, want error")
	}

	root, err := registry.Root()
	if err == nil {
		t.Fatal("Root() error = nil, want error")
	}
	if root != nil {
		t.Fatalf("Root() = %#v, want nil", root)
	}
}
