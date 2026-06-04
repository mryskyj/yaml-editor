package app

import "testing"

func TestNewReturnsApp(t *testing.T) {
	t.Parallel()

	got := New()
	if got == nil {
		t.Fatal("New() returned nil")
	}
}
